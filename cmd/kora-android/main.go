// Command kora-android is the Android entry point for Kora games.
//
// gomobile build targets this package to produce the native .apk / .aab.
// The game logic (scenes, entities) is linked in via the generated
// code in the `gen/` package produced by the KScript compiler.
//
// To build:
//
//	./android/build.sh debug
package main

import (
	"image/color"
	"log"

	"github.com/ElioNeto/kora/core/input"
	"github.com/ElioNeto/kora/core/render"
	"github.com/ElioNeto/kora/core/scene"
	"github.com/ElioNeto/kora/runner"
)

// androidPlayer is a placeholder entity used until the KScript compiler
// generates the real game objects into the gen/ package.
type androidPlayer struct {
	scene.BaseEntity
	alive bool
	X, Y  float64
	Speed float64
}

func (p *androidPlayer) IsAlive() bool { return p.alive }
func (p *androidPlayer) Destroy()      { p.alive = false }
func (p *androidPlayer) Update(dt float64) {
	p.X += input.AxisX() * p.Speed * dt
	p.Y += input.AxisY() * p.Speed * dt
	// Touch movement.
	if input.AnyTouch() {
		tx, ty := input.TouchPos()
		dx, dy := tx-p.X, ty-p.Y
		spd := p.Speed * dt
		if dx*dx+dy*dy > spd*spd {
			norm := spd / (dx*dx + dy*dy)
			p.X += dx * norm * spd
			p.Y += dy * norm * spd
		}
	}
}
func (p *androidPlayer) Draw(r interface{}) {
	renderer := r.(*render.Renderer)
	renderer.DrawRect(p.X-16, p.Y-16, 32, 32, color.RGBA{0x00, 0xff, 0x88, 0xff})
}

func main() {
	// Register virtual D-pad zones for the bottom of the screen.
	// Left half = move left/right via joystick; right side = jump button.
	input.RegisterZone(0, 480, 80, 80, input.ActionLeft)
	input.RegisterZone(100, 480, 80, 80, input.ActionRight)
	input.RegisterZone(280, 480, 80, 80, input.ActionJump)

	g := runner.New(runner.Config{
		Title:        "Kora Game",
		Width:        360,
		Height:       640,
		ClearColor:   color.RGBA{0x0d, 0x0d, 0x1a, 0xff},
		DebugOverlay: false,
	}, func(s *scene.Scene) {
		p := &androidPlayer{alive: true, Speed: 200}
		p.X, p.Y = 180, 320
		s.Spawn(p)
	})

	if err := runner.Run(g); err != nil {
		log.Fatal(err)
	}
}
