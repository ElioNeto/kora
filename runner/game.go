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
	cfg      Config
	tree     *scene.SceneTree       // SceneTree orchestrates the game loop
	renderer *render.Renderer
	fade     scene.FadeState
	ticks    uint64
}

// gameTree returns the global SceneTree for use by KScript built-ins.
var gameTree *scene.SceneTree

// New creates a Game with the given config and initial scene factory.
func New(cfg Config, initial SceneFactory) *Game {
	cfg.apply()
	g := &Game{cfg: cfg}
	g.tree = scene.NewSceneTree()
	gameTree = g.tree // Expose globally for KScript built-ins
	g.tree.SetCurrentScene(scene.New())
	initial(g.tree.GetCurrentScene())
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

	// SceneTree Tick drives the entire game loop (physics + logic).
	g.tree.Tick(dt)
	return nil
}

// Draw is called by Ebitengine every frame (vsync).
func (g *Game) Draw(screen *ebiten.Image) {
	if g.renderer == nil {
		g.renderer = render.NewRenderer()
	}
	g.renderer.SetScreen(screen)
	g.renderer.Clear(g.cfg.ClearColor)

	// SceneTree Draw renders everything (runs even when paused).
	g.tree.Draw(g.renderer)

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

// GotoScene requests a scene change.
func (g *Game) GotoScene(factory SceneFactory) {
	scene := scene.New()
	factory(scene)
	g.tree.RegisterScene("next", scene)
	g.tree.ChangeScene("next")
}

// Scene returns the currently active scene (read-only use recommended).
func (g *Game) Scene() *scene.Scene {
	if g.tree == nil {
		return nil
	}
	return g.tree.GetCurrentScene()
}

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
	scene := g.tree.GetCurrentScene()
	msg := fmt.Sprintf(
		"FPS: %0.1f  TPS: %0.1f\nEntities: %d  Tasks: %d  Tick: %d",
		ebiten.ActualFPS(),
		ebiten.ActualTPS(),
		scene.Count(),
		scene.Scheduler().Len(),
		g.ticks,
	)
	g.renderer.DrawDebugText(4, 4, msg)
}
