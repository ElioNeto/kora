package transform_test

import (
	"testing"

	"github.com/ElioNeto/kora/compiler/ast"
	"github.com/ElioNeto/kora/compiler/lexer"
	"github.com/ElioNeto/kora/compiler/parser"
	"github.com/ElioNeto/kora/compiler/transform"
)

func parse(t *testing.T, src string) *ast.Program {
	t.Helper()
	toks, err := lexer.New(src).Tokenise()
	if err != nil {
		t.Fatalf("lexer: %v", err)
	}
	prog, err := parser.New(toks).Parse()
	if err != nil {
		t.Fatalf("parser: %v", err)
	}
	return prog
}

func TestAsyncSingleAwait(t *testing.T) {
	src := `
object T {
  async create() {
    await wait(0.5)
    this.hp = 5
  }
}`
	prog := parse(t, src)
	res := transform.TransformProgram(prog)
	if len(res) != 1 {
		t.Fatalf("expected 1 async method, got %d", len(res))
	}
	for _, am := range res {
		if len(am.States) != 2 {
			t.Errorf("expected 2 states, got %d", len(am.States))
		}
		if am.States[0].Await == nil {
			t.Error("state 0 should have an await expression")
		}
		if am.States[1].Await != nil {
			t.Error("state 1 (terminal) should have no await")
		}
	}
}

func TestAsyncMultipleAwaits(t *testing.T) {
	src := `
object Intro {
  async create() {
    await wait(0.3)
    await wait(0.2)
    await wait(0.1)
  }
}`
	prog := parse(t, src)
	res := transform.TransformProgram(prog)
	for _, am := range res {
		if len(am.States) != 4 {
			t.Errorf("expected 4 states (3 awaits + terminal), got %d", len(am.States))
		}
	}
}

func TestSyncMethodNotTransformed(t *testing.T) {
	src := `
object T {
  update(dt: float) {
    this.x += 1
  }
}`
	prog := parse(t, src)
	res := transform.TransformProgram(prog)
	if len(res) != 0 {
		t.Errorf("sync methods should not be in the result map")
	}
}

func TestLiveVarPromotion(t *testing.T) {
	src := `
object T {
  async create() {
    var speed: float = 180
    await wait(1.0)
    this.x += 1
  }
}`
	prog := parse(t, src)
	res := transform.TransformProgram(prog)
	for _, am := range res {
		if len(am.LiveVars) == 0 {
			t.Error("expected speed to be promoted as a live var")
		}
	}
}
