// Package engine is the entry point of the Kora runtime.
// It wires together the renderer, input, audio, scene and async scheduler.
package engine

import (
	"github.com/ElioNeto/kora/core/async"
	"github.com/ElioNeto/kora/core/input"
	"github.com/ElioNeto/kora/core/render"
	"github.com/ElioNeto/kora/core/scene"
	"github.com/hajimehoshi/ebiten/v2"
)

// Config holds the initial engine configuration.
type Config struct {
	Title  string
	Width  int
	Height int
	FPS    int
}

// Engine is the main Kora runtime.
type Engine struct {
	cfg       Config
	renderer  *render.Renderer
	input     *input.Manager
	scheduler *async.Scheduler
	scene     *scene.Scene
}

// New creates and initialises a new Engine.
func New(cfg Config) (*Engine, error) {
	e := &Engine{
		cfg:       cfg,
		renderer:  render.NewRenderer(),
		input:     input.New(),
		scheduler: async.NewScheduler(),
		scene:     scene.New(),
	}
	return e, nil
}

// Run starts the game loop using Ebitengine.
func (e *Engine) Run() error {
	ebiten.SetWindowSize(e.cfg.Width, e.cfg.Height)
	ebiten.SetWindowTitle(e.cfg.Title)
	ebiten.SetTPS(e.cfg.FPS)
	return ebiten.RunGame(e)
}

// Update is called every tick by Ebitengine.
func (e *Engine) Update() error {
	dt := 1.0 / float64(e.cfg.FPS)
	e.input.Update()
	e.scene.Update(dt)
	e.scheduler.Tick(dt)
	return nil
}

// Draw is called every frame by Ebitengine.
func (e *Engine) Draw(screen *ebiten.Image) {
	e.renderer.SetScreen(screen)
	e.scene.Draw(e.renderer)
}

// Layout returns the logical screen size.
func (e *Engine) Layout(outsideWidth, outsideHeight int) (int, int) {
	return e.cfg.Width, e.cfg.Height
}
