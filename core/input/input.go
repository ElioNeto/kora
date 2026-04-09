// Package input provides a unified input layer for Kora.
//
// Supports:
//   - Multi-touch (Android primary input)
//   - Virtual gamepad (on-screen D-pad + buttons)
//   - Physical keyboard (desktop / emulator)
//
// Usage (call once per frame from the game loop):
//
//	input.Update()
//	if input.JustPressed(input.ActionJump) { ... }
//	ax := input.AxisX()   // -1..+1
package input

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ----------------------------------------------------------------------------
// Actions
// ----------------------------------------------------------------------------

// Action is a logical input event independent of the physical source.
type Action int

const (
	ActionLeft Action = iota
	ActionRight
	ActionUp
	ActionDown
	ActionJump
	ActionAttack
	ActionPause
	actionCount
)

// ActionName returns the string label for an Action.
func ActionName(a Action) string {
	switch a {
	case ActionLeft:   return "Left"
	case ActionRight:  return "Right"
	case ActionUp:     return "Up"
	case ActionDown:   return "Down"
	case ActionJump:   return "Jump"
	case ActionAttack: return "Attack"
	case ActionPause:  return "Pause"
	}
	return "Unknown"
}

// ----------------------------------------------------------------------------
// Input state
// ----------------------------------------------------------------------------

var (
	state    [actionCount]bool // current frame
	previous [actionCount]bool // previous frame

	touches    []ebiten.TouchID
	touchPos   = make(map[ebiten.TouchID][2]float64)

	// Virtual gamepad zones (set by the UI layer).
	vpad virtualPad
)

// defaultKeyBindings maps keyboard keys to actions.
var defaultKeyBindings = map[ebiten.Key]Action{
	ebiten.KeyArrowLeft:  ActionLeft,
	ebiten.KeyA:          ActionLeft,
	ebiten.KeyArrowRight: ActionRight,
	ebiten.KeyD:          ActionRight,
	ebiten.KeyArrowUp:    ActionUp,
	ebiten.KeyW:          ActionUp,
	ebiten.KeyArrowDown:  ActionDown,
	ebiten.KeyS:          ActionDown,
	ebiten.KeySpace:      ActionJump,
	ebiten.KeyZ:          ActionAttack,
	ebiten.KeyEscape:     ActionPause,
}

// ----------------------------------------------------------------------------
// Update — call once per frame
// ----------------------------------------------------------------------------

// Update samples all input sources and updates the action state.
func Update() {
	copy(previous[:], state[:])
	for i := range state {
		state[i] = false
	}

	// Keyboard.
	for key, action := range defaultKeyBindings {
		if ebiten.IsKeyPressed(key) {
			state[action] = true
		}
	}

	// Touch.
	touches = inpututil.AppendJustPressedTouchIDs(touches[:0])
	for _, id := range ebiten.AppendTouchIDs(nil) {
		x, y := ebiten.TouchPosition(id)
		touchPos[id] = [2]float64{float64(x), float64(y)}
	}
	// Remove released touches.
	for id := range touchPos {
		if inpututil.IsTouchJustReleased(id) {
			delete(touchPos, id)
		}
	}

	// Virtual pad.
	vpad.sample()
}

// ----------------------------------------------------------------------------
// Action queries
// ----------------------------------------------------------------------------

// Pressed reports whether action is held this frame.
func Pressed(a Action) bool { return state[a] }

// JustPressed reports whether action was pressed this frame (rising edge).
func JustPressed(a Action) bool { return state[a] && !previous[a] }

// JustReleased reports whether action was released this frame (falling edge).
func JustReleased(a Action) bool { return !state[a] && previous[a] }

// AxisX returns a horizontal axis value in [-1, +1].
func AxisX() float64 {
	switch {
	case state[ActionLeft] && !state[ActionRight]:
		return -1
	case state[ActionRight] && !state[ActionLeft]:
		return 1
	}
	return 0
}

// AxisY returns a vertical axis value in [-1, +1] (up = -1).
func AxisY() float64 {
	switch {
	case state[ActionUp] && !state[ActionDown]:
		return -1
	case state[ActionDown] && !state[ActionUp]:
		return 1
	}
	return 0
}

// ----------------------------------------------------------------------------
// Touch queries
// ----------------------------------------------------------------------------

// TouchCount returns the number of active touches.
func TouchCount() int { return len(touchPos) }

// TouchPos returns the screen position of the first active touch, or (0,0).
func TouchPos() (float64, float64) {
	for _, pos := range touchPos {
		return pos[0], pos[1]
	}
	return 0, 0
}

// AnyTouch reports whether at least one finger is currently touching.
func AnyTouch() bool { return len(touchPos) > 0 }

// JustTouched reports whether a new touch started this frame.
func JustTouched() bool { return len(touches) > 0 }
