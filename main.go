package main

import (
	"log"

	"github.com/ElioNeto/kora/core/engine"
)

func main() {
	e, err := engine.New(engine.Config{
		Title:  "Kora Engine",
		Width:  360,
		Height: 640,
		FPS:    60,
	})
	if err != nil {
		log.Fatalf("failed to create engine: %v", err)
	}

	if err := e.Run(); err != nil {
		log.Fatalf("engine error: %v", err)
	}
}
