// Kora Engine — main entry point
//
// Run without arguments to start the engine with an empty scene (debug overlay).
// Run with a scene path to load a specific scene file:
//
//	go run main.go scenes/mygame.kora.json
//
// The simplest way to see the engine running is:
//
//	go run ./examples/hello

package main

import (
	"flag"
	"image/color"
	"log"
	"os"

	"github.com/ElioNeto/kora/core/scene"
	"github.com/ElioNeto/kora/runner"
)

func main() {
	flag.Parse()

	// Default config
	cfg := runner.Config{
		Title:        "Kora Engine",
		Width:        360,
		Height:       640,
		TargetFPS:    60,
		ClearColor:   color.RGBA{0x0d, 0x0d, 0x1a, 0xff},
		DebugOverlay: true,
	}

	// If a scene file was provided as argument, load it
	if args := flag.Args(); len(args) > 0 {
		scenePath := args[0]
		g := runner.New(cfg, func(s *scene.Scene) {
			entity, err := scene.LoadSceneEntity(scenePath)
			if err != nil {
				log.Printf("warning: could not load scene %q: %v", scenePath, err)
				return
			}
			s.SetNodeRoot(entity)
		})
		if err := runner.Run(g); err != nil {
			log.Fatal(err)
		}
		return
	}

	// Otherwise start with an empty interactive scene
	showCLIHelp()
	g := runner.New(cfg)
	if err := runner.Run(g); err != nil {
		log.Fatal(err)
	}
}

func showCLIHelp() {
	log.Println("Kora Engine — use go run ./examples/hello for a demo")
	log.Println("  go run main.go scenes/file.kora.json  — load a scene")
	log.Println("  go run ./examples/hello               — player demo")
	log.Println("  Press F3 in-game for debug overlay")
	log.Println("")
	// Print to stderr so it doesn't interfere with the game
	os.Stderr.WriteString("Kora Engine starting...\n")
}
