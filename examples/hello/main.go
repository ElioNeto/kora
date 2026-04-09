// Example: minimal Kora game that spawns a player and moves it with input.
// Run with: go run ./examples/hello
package main

import (
	"image/color"
	"log"

	"github.com/ElioNeto/kora/core/input"
	"github.com/ElioNeto/kora/core/render"
	"github.com/ElioNeto/kora/core/scene"
	"github.com/ElioNeto/kora/runner"
)

// ----------------------------------------------------------------------------
// Player entity (hand-written; normally emitted by the compiler)
// ----------------------------------------------------------------------------

type Player struct {
	scene.BaseEntity
	alive  bool
	X, Y   float64
	Speed  float64
	sprite *render.Sprite
}

func NewPlayer() *Player {
	return &Player{alive: true, Speed: 180}
}

func (p *Player) IsAlive() bool { return p.alive }
func (p *Player) Destroy()      { p.alive = false }

func (p *Player) Update(dt float64) {
	p.X += input.AxisX() * p.Speed * dt
	p.Y += input.AxisY() * p.Speed * dt
}

func (p *Player) Draw(r interface{}) {
	renderer := r.(*render.Renderer)
	renderer.DrawRect(p.X-12, p.Y-12, 24, 24, color.RGBA{0x00, 0xc8, 0xff, 0xff})
}

// ----------------------------------------------------------------------------
// main
// ----------------------------------------------------------------------------

func main() {
	g := runner.New(runner.Config{
		Title:        "Kora Hello",
		Width:        360,
		Height:       640,
		ClearColor:   color.RGBA{0x1a, 0x1a, 0x2e, 0xff},
		DebugOverlay: true,
	}, func(s *scene.Scene) {
		p := NewPlayer()
		p.X, p.Y = 180, 320
		s.Spawn(p)
	})

	if err := runner.Run(g); err != nil {
		log.Fatal(err)
	}
}
