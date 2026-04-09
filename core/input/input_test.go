package input_test

import (
	"testing"

	"github.com/ElioNeto/kora/core/input"
)

// input.Update() calls Ebitengine which requires a display — we test
// only the pure-logic helpers that don’t need a GPU context.

func TestActionNames(t *testing.T) {
	cases := []struct {
		action input.Action
		want   string
	}{
		{input.ActionLeft, "Left"},
		{input.ActionRight, "Right"},
		{input.ActionUp, "Up"},
		{input.ActionDown, "Down"},
		{input.ActionJump, "Jump"},
		{input.ActionAttack, "Attack"},
		{input.ActionPause, "Pause"},
	}
	for _, c := range cases {
		if got := input.ActionName(c.action); got != c.want {
			t.Errorf("ActionName(%d) = %q, want %q", c.action, got, c.want)
		}
	}
}

func TestClearZones(t *testing.T) {
	input.RegisterZone(0, 0, 100, 100, input.ActionJump)
	input.ClearZones()
	// No panic = pass.
}
