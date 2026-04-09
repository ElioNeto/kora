package runner_test

import (
	"testing"

	"github.com/ElioNeto/kora/core/scene"
	"github.com/ElioNeto/kora/runner"
)

// We can’t call ebiten.RunGame in a headless test, but we can verify
// that New + Scene wiring works without a GPU context.

func TestNewGame(t *testing.T) {
	spawned := false
	g := runner.New(runner.Config{
		Title:  "Test",
		Width:  360,
		Height: 640,
	}, func(s *scene.Scene) {
		spawned = true
	})
	if !spawned {
		t.Error("initial SceneFactory should be called during New")
	}
	if g.Scene() == nil {
		t.Error("game.Scene() should not be nil after New")
	}
}

func TestDefaultConfig(t *testing.T) {
	g := runner.New(runner.Config{}, func(_ *scene.Scene) {})
	// Config.apply() fills in defaults — scene should still be valid.
	if g.Scene() == nil {
		t.Error("scene should be initialised with default config")
	}
}
