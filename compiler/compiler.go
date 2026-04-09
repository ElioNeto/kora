// Package compiler is the public API for the KScript compilation pipeline.
//
// Usage:
//
//	result, err := compiler.CompileFile("path/to/player.ks")
//	if err != nil { ... }
//	fmt.Println(result.GoSource)
package compiler

import (
	"fmt"
	"os"

	"github.com/ElioNeto/kora/compiler/checker"
	"github.com/ElioNeto/kora/compiler/emitter"
	"github.com/ElioNeto/kora/compiler/lexer"
	"github.com/ElioNeto/kora/compiler/parser"
	"github.com/ElioNeto/kora/compiler/transform"
)

// Result holds the output of a successful compilation.
type Result struct {
	// GoSource is the emitted Go source code (ready to write to a .go file).
	GoSource string
	// Warnings are non-fatal diagnostic messages.
	Warnings []string
}

// CompileSource compiles KScript source code and returns the generated Go code.
func CompileSource(src string, opts *emitter.Options) (*Result, error) {
	// 1. Lex.
	toks, err := lexer.New(src).Tokenise()
	if err != nil {
		return nil, fmt.Errorf("lexer: %w", err)
	}

	// 2. Parse.
	p := parser.New(toks)
	prog, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("parser: %w", err)
	}

	// 3. Type check.
	ch := checker.New(prog)
	errors := ch.Check()
	if len(errors) > 0 {
		msg := "type errors:\n"
		for _, e := range errors {
			msg += "  " + e.Message + "\n"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	// 4. Async transform.
	asyncMap := transform.TransformProgram(prog)

	// 5. Emit Go.
	goSrc := emitter.New(prog, asyncMap, opts).Emit()

	return &Result{GoSource: goSrc}, nil
}

// CompileFile reads a .ks file and compiles it.
func CompileFile(path string, opts *emitter.Options) (*Result, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return CompileSource(string(src), opts)
}
