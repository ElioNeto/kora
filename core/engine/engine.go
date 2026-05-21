// Package engine provides the Kora runtime entry point.
//
// Deprecated: Use github.com/ElioNeto/kora/runner instead. This package now
// delegates all functionality to the runner package and is kept for backward
// compatibility only.
package engine

import (
	"github.com/ElioNeto/kora/runner"
	"github.com/hajimehoshi/ebiten/v2"
)

// Config holds the initial engine configuration.
//
// Deprecated: Use runner.Config instead.
type Config struct {
	Title  string
	Width  int
	Height int
	FPS    int
}

// Engine is the main Kora runtime.
//
// Deprecated: Use runner.Game instead. This type delegates to runner internally.
type Engine struct {
	game *runner.Game
}

// New creates and initialises a new Engine.
//
// Deprecated: Use runner.New instead.
func New(cfg Config) (*Engine, error) {
	g := runner.New(runner.Config{
		Title:     cfg.Title,
		Width:     cfg.Width,
		Height:    cfg.Height,
		TargetFPS: cfg.FPS,
	})
	return &Engine{game: g}, nil
}

// Run starts the game loop using Ebitengine.
//
// Deprecated: Use runner.Run instead.
func (e *Engine) Run() error {
	return runner.Run(e.game)
}

// Update is called every tick by Ebitengine.
func (e *Engine) Update() error {
	return e.game.Update()
}

// Draw is called every frame by Ebitengine.
func (e *Engine) Draw(screen *ebiten.Image) {
	e.game.Draw(screen)
}

// Layout returns the logical screen size.
func (e *Engine) Layout(outsideWidth, outsideHeight int) (int, int) {
	return e.game.Layout(outsideWidth, outsideHeight)
}
