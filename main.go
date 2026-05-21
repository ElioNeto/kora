package main

import (
	"image/color"
	"log"

	"github.com/ElioNeto/kora/runner"
)

func main() {
	g := runner.New(runner.Config{
		Title:        "Kora Engine",
		Width:        360,
		Height:       640,
		TargetFPS:    60,
		ClearColor:   color.RGBA{0x0d, 0x0d, 0x1a, 0xff},
		DebugOverlay: false,
	})
	if err := runner.Run(g); err != nil {
		log.Fatal(err)
	}
}
