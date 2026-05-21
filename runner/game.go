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

	"github.com/ElioNeto/kora/core/autoload"
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
	Width        int               // logical screen width in pixels
	Height       int               // logical screen height in pixels
	TargetFPS    int               // 0 = use Ebitengine default (60)
	ClearColor   color.Color
	DebugOverlay bool
	OnCreate     func(s *scene.Scene) // optional callback; called with the initial scene
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
	cfg          Config
	tree         *scene.SceneTree // SceneTree orchestrates the game loop
	sceneManager *scene.SceneManager
	renderer     *render.Renderer
	fade         scene.FadeState
	ticks        uint64
}

var gameInstance *Game

// gameTree is the global SceneTree, accessible by KScript-generated code via GameTree().
var gameTree *scene.SceneTree

// GameTree returns the global SceneTree for use by KScript built-ins.
// It is set once when New() is called and never changes for the lifetime of the process.
func GameTree() *scene.SceneTree {
	return gameTree
}

// GameSceneManager returns the global SceneManager for KScript built-ins.
func GameSceneManager() *scene.SceneManager {
	if gameInstance == nil {
		return nil
	}
	return gameInstance.sceneManager
}

// New creates a Game with the given config and optional initial scene factory.
//
// The initial scene can be populated in one of three ways (first match wins):
//  1. Pass a SceneFactory as a variadic argument.
//  2. Set Config.OnCreate.
//  3. Leave both unset to start with an empty scene.
func New(cfg Config, initial ...SceneFactory) *Game {
	cfg.apply()
	g := &Game{cfg: cfg}
	gameInstance = g
	g.tree = scene.NewSceneTree()
	gameTree = g.tree // expose globally for KScript built-ins
	g.sceneManager = scene.NewSceneManager("scenes")
	// Register KScript API
	scene.SetKScriptAPIManager(g.sceneManager)
	scene.SetKScriptAPITree(g.tree)
	g.tree.SetCurrentScene(scene.New())

	// Determine which factory to use.
	var fn SceneFactory
	if len(initial) > 0 {
		fn = initial[0]
	} else if cfg.OnCreate != nil {
		fn = cfg.OnCreate
	}
	if fn != nil {
		fn(g.tree.GetCurrentScene())
	}
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

	// Update persistent AutoLoad singletons before scene logic.
	autoload.UpdateAll(dt)

	// Delegate to SceneManager
	g.sceneManager.Update(dt)
	return nil
}

// syncCamera finds the active Camera2D node (set via SceneManager.SetActiveCamera)
// and applies its transform to the renderer's camera. If no Camera2D is active,
// the renderer camera is reset to identity (centered at origin, zoom=1).
func (g *Game) syncCamera() {
	camNode := g.sceneManager.ActiveCamera()
	if camNode == nil {
		// No active Camera2D -> reset to defaults
		g.renderer.Camera = render.NewCamera(float64(g.cfg.Width), float64(g.cfg.Height))
		return
	}

	// Set viewport on the camera node (ensures WorldToScreen works)
	camNode.SetViewport(float64(g.cfg.Width), float64(g.cfg.Height))

	// Sync node Camera2D -> render.Camera
	pos := camNode.GetWorldPosition()
	g.renderer.Camera.X = float64(pos.X) + float64(camNode.GetShakeOffset().X)
	g.renderer.Camera.Y = float64(pos.Y) + float64(camNode.GetShakeOffset().Y)
	g.renderer.Camera.Zoom = camNode.Zoom
	g.renderer.Camera.W = float64(g.cfg.Width)
	g.renderer.Camera.H = float64(g.cfg.Height)
}

// Draw is called by Ebitengine every frame (vsync).
func (g *Game) Draw(screen *ebiten.Image) {
	if g.renderer == nil {
		g.renderer = render.NewRenderer()
	}
	g.renderer.SetScreen(screen)

	// Sync active camera before rendering.
	g.syncCamera()

	g.renderer.Clear(g.cfg.ClearColor)

	// Delegate to SceneManager
	g.sceneManager.Draw(g.renderer)

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
	s := scene.New()
	factory(s)
	g.tree.RegisterScene("next", s)
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
	s := g.tree.GetCurrentScene()
	msg := fmt.Sprintf(
		"FPS: %0.1f  TPS: %0.1f\nEntities: %d  Tasks: %d  Tick: %d",
		ebiten.ActualFPS(),
		ebiten.ActualTPS(),
		s.Count(),
		s.Scheduler().Len(),
		g.ticks,
	)
	g.renderer.DrawDebugText(4, 4, msg)
}
