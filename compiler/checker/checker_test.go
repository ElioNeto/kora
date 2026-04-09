package checker_test

import (
	"testing"

	"github.com/ElioNeto/kora/compiler/checker"
	"github.com/ElioNeto/kora/compiler/lexer"
	"github.com/ElioNeto/kora/compiler/parser"
)

func check(t *testing.T, src string) []*checker.Error {
	t.Helper()
	toks, err := lexer.New(src).Tokenise()
	if err != nil {
		t.Fatalf("lexer: %v", err)
	}
	prog, err := parser.New(toks).Parse()
	if err != nil {
		t.Fatalf("parser: %v", err)
	}
	return checker.New(prog).Check()
}

func expectNoErrors(t *testing.T, src string) {
	t.Helper()
	errs := check(t, src)
	for _, e := range errs {
		t.Errorf("unexpected error: %s", e.Message)
	}
}

func expectError(t *testing.T, src, contains string) {
	t.Helper()
	errs := check(t, src)
	for _, e := range errs {
		if len(e.Message) > 0 && containsStr(e.Message, contains) {
			return
		}
	}
	t.Errorf("expected error containing %q, but got: %v", contains, errs)
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		})())
}

// --- Tests ---

func TestValidObject(t *testing.T) {
	expectNoErrors(t, `
object Player {
  var speed: float = 180
  var hp: int = 5

  update(dt: float) {
    this.x += dt
  }
}`)
}

func TestAwaitOutsideAsync(t *testing.T) {
	expectError(t, `
object T {
  update(dt: float) {
    await wait(1.0)
  }
}`, "await")
}

func TestAwaitInsideAsync(t *testing.T) {
	expectNoErrors(t, `
object T {
  async create() {
    await wait(0.5)
  }
}`)
}

func TestUnknownType(t *testing.T) {
	expectError(t, `
object T {
  var x: GhostType = 0
}`, "unknown type")
}

func TestUndeclaredVariable(t *testing.T) {
	expectError(t, `
object T {
  update(dt: float) {
    x += 1
  }
}`, "undeclared")
}

func TestDuplicateObject(t *testing.T) {
	expectError(t, `
object Player {}
object Player {}
`, "duplicate")
}

func TestEngineAPICall(t *testing.T) {
	expectNoErrors(t, `
object T {
  async create() {
    await wait(1.0)
    await waitFrames(3)
  }
}`)
}

func TestWrongArgCount(t *testing.T) {
	expectError(t, `
object T {
  async create() {
    await wait(1.0, 2.0)
  }
}`, "expects 1")
}

func TestConstReassign(t *testing.T) {
	expectError(t, `
object T {
  update(dt: float) {
    const limit: int = 10
    limit = 5
  }
}`, "const")
}

func TestValidImport(t *testing.T) {
	expectNoErrors(t, `
import { Input } from "kora"
object T {
  update(dt: float) {
    const ax: float = 1.0
  }
}`)
}

func TestInvalidImport(t *testing.T) {
	expectError(t, `
import { X } from "npm"
object T {}
`, "unknown module")
}
