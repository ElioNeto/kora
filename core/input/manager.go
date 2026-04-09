// Package input abstracts keyboard, mouse and touch input.
package input

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Manager polls input state once per tick.
type Manager struct {
	prev map[ebiten.Key]bool
	curr map[ebiten.Key]bool
}

// New creates an input Manager.
func New() *Manager {
	return &Manager{
		prev: make(map[ebiten.Key]bool),
		curr: make(map[ebiten.Key]bool),
	}
}

// Update snapshots the current keyboard state.
func (m *Manager) Update() {
	prev := m.curr
	m.curr = make(map[ebiten.Key]bool, len(prev))
	for k := ebiten.Key(0); k <= ebiten.KeyMax; k++ {
		if ebiten.IsKeyPressed(k) {
			m.curr[k] = true
		}
	}
	m.prev = prev
}

// Pressed returns true on the first frame the key is held.
func (m *Manager) Pressed(k ebiten.Key) bool {
	return m.curr[k] && !m.prev[k]
}

// Held returns true while the key is held.
func (m *Manager) Held(k ebiten.Key) bool {
	return m.curr[k]
}

// Released returns true on the first frame the key is released.
func (m *Manager) Released(k ebiten.Key) bool {
	return !m.curr[k] && m.prev[k]
}

// AxisX returns a float64 in [-1, 1] for horizontal movement (Arrow / WASD).
func (m *Manager) AxisX() float64 {
	v := 0.0
	if m.Held(ebiten.KeyArrowLeft) || m.Held(ebiten.KeyA) {
		v -= 1
	}
	if m.Held(ebiten.KeyArrowRight) || m.Held(ebiten.KeyD) {
		v += 1
	}
	return v
}

// AxisY returns a float64 in [-1, 1] for vertical movement (Arrow / WASD).
func (m *Manager) AxisY() float64 {
	v := 0.0
	if m.Held(ebiten.KeyArrowUp) || m.Held(ebiten.KeyW) {
		v -= 1
	}
	if m.Held(ebiten.KeyArrowDown) || m.Held(ebiten.KeyS) {
		v += 1
	}
	return v
}
