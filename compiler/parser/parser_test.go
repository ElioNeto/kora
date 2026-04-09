package parser_test

import (
	"testing"

	"github.com/ElioNeto/kora/compiler/lexer"
	"github.com/ElioNeto/kora/compiler/parser"
)

func mustTokenise(t *testing.T, src string) []lexer.Token {
	t.Helper()
	toks, err := lexer.New(src).Tokenise()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	return toks
}

func TestParseEmptyObject(t *testing.T) {
	src := `object Empty {}`
	prog, err := parser.New(mustTokenise(t, src)).Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(prog.Objects) != 1 {
		t.Fatalf("expected 1 object, got %d", len(prog.Objects))
	}
	if prog.Objects[0].Name != "Empty" {
		t.Errorf("expected name Empty, got %s", prog.Objects[0].Name)
	}
}

func TestParseObjectWithFields(t *testing.T) {
	src := `
object Player {
  var speed: float = 180
  var hp: int = 5
}`
	prog, err := parser.New(mustTokenise(t, src)).Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	obj := prog.Objects[0]
	if len(obj.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(obj.Fields))
	}
	if obj.Fields[0].Name != "speed" || obj.Fields[0].Type != "float" {
		t.Errorf("unexpected field[0]: %+v", obj.Fields[0])
	}
}

func TestParseMethod(t *testing.T) {
	src := `
object Enemy {
  update(dt: float) {
    this.x += 1
  }
}`
	prog, err := parser.New(mustTokenise(t, src)).Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	obj := prog.Objects[0]
	if len(obj.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(obj.Methods))
	}
	m := obj.Methods[0]
	if m.Name != "update" || m.Async {
		t.Errorf("unexpected method: %+v", m)
	}
}

func TestParseAsyncMethod(t *testing.T) {
	src := `
object Intro {
  async create() {
    await wait(0.5)
    await tween(this, 0.3)
    emit "ready"
  }
}`
	prog, err := parser.New(mustTokenise(t, src)).Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	m := prog.Objects[0].Methods[0]
	if !m.Async {
		t.Error("expected method to be async")
	}
	if len(m.Body) != 3 {
		t.Errorf("expected 3 stmts, got %d", len(m.Body))
	}
}

func TestParseIfElse(t *testing.T) {
	src := `
object T {
  update(dt: float) {
    if (this.hp <= 0) {
      this.destroy()
    } else {
      this.hp -= 1
    }
  }
}`
	_, err := parser.New(mustTokenise(t, src)).Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
}

func TestParseImport(t *testing.T) {
	src := `
import { Input, Audio } from "kora"
object T {}
`
	prog, err := parser.New(mustTokenise(t, src)).Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(prog.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(prog.Imports))
	}
	if prog.Imports[0].Module != "kora" {
		t.Errorf("unexpected module: %s", prog.Imports[0].Module)
	}
}
