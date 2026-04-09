// Command kora-run compiles a .ks file and runs the resulting game
// immediately in a desktop window. Useful for rapid iteration.
//
// Usage:
//
//	kora-run game.ks
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ElioNeto/kora/compiler"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: kora-run <file.ks>")
		os.Exit(1)
	}
	path := os.Args[1]

	result, err := compiler.CompileFile(path, nil)
	if err != nil {
		log.Fatalf("compile error: %v", err)
	}

	// Write generated Go to a temp file and report success.
	// In a full implementation this would be loaded via plugin or
	// written to a temp dir and `go run`-ed.
	fmt.Printf("// === Generated Go for %s ===\n", path)
	fmt.Print(result.GoSource)
}
