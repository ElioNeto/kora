// Package runner wires every Kora core module into a runnable Ebitengine game.
//
// Responsibilities:
//   - Implement ebiten.Game (Update / Draw / Layout).
//   - Drive scene.Update and scene.Draw each frame.
//   - Call input.Update every tick.
//   - Expose a SceneLoader so user code can switch scenes.
//   - Provide an optional debug overlay (FPS, entity count, task count).
package runner

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/ElioNeto/kora/core/input"
	"github.com/ElioNeto/kora/core/render"
	"github.com/ElioNeto/kora/core/scene"
)

// ----------------------------------------------------------------------------
// Config
// ----------------------------------------------------------------------------

// Config holds startup parameters for the Kora game runner.
type Config struct {
	Title        string
	Width        int     // logical screen width in pixels
	Height       int     // logical screen height in pixels
	TargetFPS    int     // 0 = use Ebitengine default (60)
	ClearColor   color.Color
	DebugOverlay bool
}

func (c *Config) apply() {
	if c.Width == 0 {
		c.Width = 360
	}
	if c.Height == 0 {
		c.Height = 640
	}
	if c.Title == "" {
		c.Title = "Kora Game"
	}
	if c.ClearColor == nil {
		c.ClearColor = color.Black
	}
}

// ----------------------------------------------------------------------------
// SceneFactory
// ----------------------------------------------------------------------------

// SceneFactory is a function that populates a fresh scene.Scene.
// Return from a SceneFactory by calling scene.Spawn for each entity.
type SceneFactory func(s *scene.Scene)

// ----------------------------------------------------------------------------
// Game — implements ebiten.Game
// ----------------------------------------------------------------------------

// Game is the central game object. Create one with New, then call Run.
type Game struct {
	cfg     Config
	scene   *scene.Scene
	renderer *render.Renderer
	pending SceneFactory // non-nil when a transition is requested
	fade    scene.FadeState
	ticks   uint64
}

// New creates a Game with the given config and initial scene factory.
func New(cfg Config, initial SceneFactory) *Game {
	cfg.apply()
	g := &Game{cfg: cfg}
	g.scene = scene.New()
	initial(g.scene)
	return g
}

// Run starts the Ebitengine event loop. Blocks until the window is closed.
func Run(g *Game) error {
	ebiten.SetWindowTitle(g.cfg.Title)
	ebiten.SetWindowSize(g.cfg.Width, g.cfg.Height)
	if g.cfg.TargetFPS > 0 {
		ebiten.SetTPS(g.cfg.TargetFPS)
	}
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	return ebiten.RunGame(g)
}

// ----------------------------------------------------------------------------
// ebiten.Game interface
// ----------------------------------------------------------------------------

// Update is called by Ebitengine every tick (default 60 TPS).
func (g *Game) Update() error {
	dt := 1.0 / float64(ebiten.TPS())
	g.ticks++

	// Input must be sampled first.
	input.Update()

	// Fade transition tick.
	if g.fade.Active() {
		g.fade.Tick(dt)
		return nil // freeze world during transition
	}

	// Scene switch committed after fade-out.
	if g.pending != nil {
		g.scene.DestroyAll()
		g.scene = scene.New()
		g.pending(g.scene)
		g.pending = nil
		g.fade.FadeIn(0.3, nil)
	}

	// Normal world update.
	g.scene.Update(dt)
	return nil
}

// Draw is called by Ebitengine every frame (vsync).
func (g *Game) Draw(screen *ebiten.Image) {
	g.renderer = render.NewRenderer(screen)
	g.renderer.Clear(g.cfg.ClearColor)

	// Draw all scene entities.
	g.scene.Draw(g.renderer)

	// Fade overlay.
	if g.fade.Active() || g.fade.Alpha > 0 {
		drawFade(screen, g.fade.Alpha, g.cfg.Width, g.cfg.Height)
	}

	// Debug overlay.
	if g.cfg.DebugOverlay {
		g.drawDebug()
	}
}

// Layout returns the logical screen size (Ebitengine scales to the window).
func (g *Game) Layout(_, _ int) (int, int) {
	return g.cfg.Width, g.cfg.Height
}

// ----------------------------------------------------------------------------
// Scene transitions
// ----------------------------------------------------------------------------

// GotoScene requests a scene change with a fade transition.
// The current scene is kept alive until the fade-out completes.
func (g *Game) GotoScene(factory SceneFactory) {
	g.pending = factory
	g.fade.FadeOut(0.3, nil)
}

// Scene returns the currently active scene (read-only use recommended).
func (g *Game) Scene() *scene.Scene { return g.scene }

// Renderer returns the most recent Renderer (valid during Draw only).
func (g *Game) Renderer() *render.Renderer { return g.renderer }

// ----------------------------------------------------------------------------
// Fade overlay helper
// ----------------------------------------------------------------------------

func drawFade(screen *ebiten.Image, alpha float32, w, h int) {
	if alpha <= 0 {
		return
	}
	overlayImg := ebiten.NewImage(w, h)
	a := uint8(alpha * 255)
	overlayImg.Fill(color.RGBA{0, 0, 0, a})
	screen.DrawImage(overlayImg, nil)
}

// ----------------------------------------------------------------------------
// Debug overlay
// ----------------------------------------------------------------------------

func (g *Game) drawDebug() {
	if g.renderer == nil {
		return
	}
	msg := fmt.Sprintf(
		"FPS: %0.1f  TPS: %0.1f\nEntities: %d  Tasks: %d  Tick: %d",
		ebiten.ActualFPS(),
		ebiten.ActualTPS(),
		g.scene.Count(),
		g.scene.Scheduler().Len(),
		g.ticks,
	)
	g.renderer.DrawDebugText(4, 4, msg)
}
