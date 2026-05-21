package autoload

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

type counter struct {
	Base
	name  string
	count int
}

func (c *counter) Name() string { return c.name }
func (c *counter) Update(dt float64) {
	c.count++
}

type noopAutoLoad struct {
	Base
	name string
}

func (n *noopAutoLoad) Name() string { return n.name }

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r.Len() != 0 {
		t.Fatalf("expected empty registry, got %d", r.Len())
	}
}

func TestSetAndGet(t *testing.T) {
	r := NewRegistry()
	a := &counter{name: "test", count: 0}
	r.Set(a)

	got := r.Get("test")
	if got == nil {
		t.Fatal("expected to find instance 'test'")
	}
	if got != a {
		t.Fatal("expected same instance")
	}
}

func TestGetMissing(t *testing.T) {
	r := NewRegistry()
	got := r.Get("nonexistent")
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestRemove(t *testing.T) {
	r := NewRegistry()
	r.Set(&counter{name: "a"})
	r.Set(&counter{name: "b"})

	if r.Len() != 2 {
		t.Fatalf("expected 2, got %d", r.Len())
	}

	r.Remove("a")
	if r.Len() != 1 {
		t.Fatalf("expected 1 after remove, got %d", r.Len())
	}
	if r.Get("a") != nil {
		t.Fatal("expected nil after remove")
	}
	if r.Get("b") == nil {
		t.Fatal("expected 'b' to still exist")
	}
}

func TestUpdate(t *testing.T) {
	r := NewRegistry()
	a := &counter{name: "a"}
	b := &counter{name: "b"}
	r.Set(a)
	r.Set(b)

	r.Update(1.0)
	if a.count != 1 {
		t.Fatalf("expected a.count=1, got %d", a.count)
	}
	if b.count != 1 {
		t.Fatalf("expected b.count=1, got %d", b.count)
	}

	r.Update(1.0)
	if a.count != 2 {
		t.Fatalf("expected a.count=2, got %d", a.count)
	}
}

func TestUpdateOrder(t *testing.T) {
	r := NewRegistry()
	var order []string
	a := &counter{name: "a"}
	b := &counter{name: "b"}
	r.Set(a)
	r.Set(b)

	// Override Update to record order
	r.instances["a"] = &fnAutoLoad{nameFn: func() string { return "a" }, updateFn: func(dt float64) { order = append(order, "a") }}
	r.instances["b"] = &fnAutoLoad{nameFn: func() string { return "b" }, updateFn: func(dt float64) { order = append(order, "b") }}

	r.Update(1.0)
	if len(order) != 2 || order[0] != "a" || order[1] != "b" {
		t.Fatalf("expected [a b], got %v", order)
	}
}

type fnAutoLoad struct {
	Base
	nameFn   func() string
	updateFn func(dt float64)
}

func (f *fnAutoLoad) Name() string {
	if f.nameFn != nil {
		return f.nameFn()
	}
	return "fn"
}
func (f *fnAutoLoad) Update(dt float64) {
	if f.updateFn != nil {
		f.updateFn(dt)
	}
}

func TestClear(t *testing.T) {
	r := NewRegistry()
	r.Set(&counter{name: "a"})
	r.Set(&counter{name: "b"})
	r.Clear()
	if r.Len() != 0 {
		t.Fatalf("expected 0 after clear, got %d", r.Len())
	}
}

func TestNames(t *testing.T) {
	r := NewRegistry()
	r.Set(&counter{name: "z"})
	r.Set(&counter{name: "a"})
	r.Set(&counter{name: "m"})
	names := r.Names()
	if len(names) != 3 {
		t.Fatalf("expected 3 names, got %d", len(names))
	}
	// Must be insertion order
	if names[0] != "z" || names[1] != "a" || names[2] != "m" {
		t.Fatalf("expected [z a m], got %v", names)
	}
}

func TestMustGet(t *testing.T) {
	r := NewRegistry()
	r.Set(&counter{name: "exists"})
	if MustGetFrom("exists", r) == nil {
		t.Fatal("expected instance")
	}
}

func TestMustGetPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for missing singleton")
		}
	}()
	r := NewRegistry()
	_ = MustGetFrom("missing", r)
}

func TestGlobalRegistry(t *testing.T) {
	// Save state and restore after test
	old := Global
	Global = NewRegistry()
	defer func() { Global = old }()

	a := &counter{name: "global"}
	Set(a)
	if Get("global") != a {
		t.Fatal("expected to get instance from global registry")
	}
	if Len() != 1 {
		t.Fatalf("expected Len=1, got %d", Len())
	}
}

func TestReplace(t *testing.T) {
	r := NewRegistry()
	a := &counter{name: "x", count: 0}
	b := &counter{name: "x", count: 100}
	r.Set(a)
	r.Set(b)

	if r.Len() != 1 {
		t.Fatalf("expected 1 after replace, got %d", r.Len())
	}

	got := r.Get("x")
	if got != b {
		t.Fatal("expected second instance to replace first")
	}
}

// MustGetFrom is like MustGet but operates on a given registry.
func MustGetFrom(name string, r *Registry) AutoLoad {
	a := r.Get(name)
	if a == nil {
		panic("autoload: singleton not found")
	}
	return a
}

var _ AutoLoad = (*counter)(nil)
var _ AutoLoad = (*noopAutoLoad)(nil)
