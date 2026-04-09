package emitter_test

import (
	"strings"
	"testing"

	"github.com/ElioNeto/kora/compiler/checker"
	"github.com/ElioNeto/kora/compiler/emitter"
	"github.com/ElioNeto/kora/compiler/lexer"
	"github.com/ElioNeto/kora/compiler/parser"
	"github.com/ElioNeto/kora/compiler/transform"
)

func compile(t *testing.T, src string) string {
	t.Helper()
	toks, err := lexer.New(src).Tokenise()
	if err != nil {
		t.Fatalf("lexer: %v", err)
	}
	prog, err := parser.New(toks).Parse()
	if err != nil {
		t.Fatalf("parser: %v", err)
	}
	errs := checker.New(prog).Check()
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("checker: %s", e.Message)
		}
		t.Fatalf("%d checker error(s)", len(errs))
	}
	asyncMap := transform.TransformProgram(prog)
	return emitter.New(prog, asyncMap, nil).Emit()
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("expected output to contain:\n  %q\ngot:\n%s", want, got)
	}
}

func TestEmitSimpleObject(t *testing.T) {
	src := `
object Player {
  var speed: float = 180
  var hp: int = 5
}`
	out := compile(t, src)
	assertContains(t, out, "type Player struct")
	assertContains(t, out, "Speed float64")
	assertContains(t, out, "Hp int")
	assertContains(t, out, "func NewPlayer() *Player")
	assertContains(t, out, "o.Speed = 180")
	assertContains(t, out, "o.Hp = 5")
}

func TestEmitSyncMethod(t *testing.T) {
	src := `
object Enemy {
  var speed: float = 60

  update(dt: float) {
    this.x += this.speed * dt
  }
}`
	out := compile(t, src)
	assertContains(t, out, "func (o *Enemy) Update(")
	assertContains(t, out, "o.X += (o.Speed * dt)")
}

func TestEmitAsyncMethod(t *testing.T) {
	src := `
object Intro {
  async create() {
    await wait(0.5)
    await wait(0.3)
  }
}`
	out := compile(t, src)
	assertContains(t, out, "type Intro_Create_Task struct")
	assertContains(t, out, "func (t *Intro_Create_Task) Tick(dt float64) async.Status")
	assertContains(t, out, "async.Wait(0.5)")
	assertContains(t, out, "async.Wait(0.3)")
	assertContains(t, out, "func (o *Intro) Create() *Intro_Create_Task")
}

func TestEmitIfElse(t *testing.T) {
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
	out := compile(t, src)
	assertContains(t, out, "if (o.Hp <= 0)")
	assertContains(t, out, "} else {")
}

func TestEmitPackageHeader(t *testing.T) {
	src := `object T {}`
	out := compile(t, src)
	assertContains(t, out, "package gen")
	assertContains(t, out, "DO NOT EDIT")
}
