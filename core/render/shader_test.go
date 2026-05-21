package render_test

import (
	"testing"

	"github.com/ElioNeto/kora/core/render"
)

// --- ShaderManager Creation ---

func TestNewShaderManager(t *testing.T) {
	sm := render.NewShaderManager()
	if sm == nil {
		t.Fatal("expected non-nil ShaderManager")
	}
}

// --- Cache Queries ---

func TestGetShaderReturnsNilForUnknown(t *testing.T) {
	sm := render.NewShaderManager()
	shader := sm.GetShader("nonexistent")
	if shader != nil {
		t.Error("expected nil for unknown shader name")
	}
}

// --- Compilation Errors ---

func TestLoadShaderInvalidSource(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Skip("ebiten not initialised; skipping shader compilation test")
		}
	}()

	sm := render.NewShaderManager()
	_, err := sm.LoadShader("invalid", []byte("invalid shader source"))
	if err == nil {
		t.Error("expected error for invalid shader source")
	}
}

func TestCompileShaderInvalidSource(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Skip("ebiten not initialised; skipping shader compilation test")
		}
	}()

	_, err := render.CompileShader([]byte("bad shader"))
	if err == nil {
		t.Error("expected error for invalid shader source")
	}
}

// --- Clear / Unload ---

func TestClearAndUnloadShader(t *testing.T) {
	sm := render.NewShaderManager()

	// Unload on an empty cache must not panic.
	sm.UnloadShader("nonexistent")

	// Clear on an empty cache must not panic.
	sm.Clear()

	// After Clear, GetShader must still return nil.
	if shader := sm.GetShader("any"); shader != nil {
		t.Error("expected nil after Clear")
	}
}

// --- DefaultUniforms ---

func TestDefaultUniforms(t *testing.T) {
	uniforms := render.DefaultUniforms(1.5, 800, 600, 100.0, 200.0)
	if uniforms == nil {
		t.Fatal("expected non-nil uniforms map")
	}

	// Time
	v, ok := uniforms["Time"]
	if !ok {
		t.Fatal("expected 'Time' uniform")
	}
	timeVal, ok := v.(float32)
	if !ok {
		t.Fatalf("expected float32 for Time, got %T", v)
	}
	if timeVal != 1.5 {
		t.Errorf("expected Time=1.5, got %f", timeVal)
	}

	// Resolution
	v, ok = uniforms["Resolution"]
	if !ok {
		t.Fatal("expected 'Resolution' uniform")
	}
	res, ok := v.([]float32)
	if !ok {
		t.Fatalf("expected []float32 for Resolution, got %T", v)
	}
	if len(res) != 2 {
		t.Fatalf("expected len 2 for Resolution, got %d", len(res))
	}
	if res[0] != 800 || res[1] != 600 {
		t.Errorf("unexpected Resolution values: [%f, %f]", res[0], res[1])
	}

	// Mouse
	v, ok = uniforms["Mouse"]
	if !ok {
		t.Fatal("expected 'Mouse' uniform")
	}
	mouse, ok := v.([]float32)
	if !ok {
		t.Fatalf("expected []float32 for Mouse, got %T", v)
	}
	if len(mouse) != 2 {
		t.Fatalf("expected len 2 for Mouse, got %d", len(mouse))
	}
	if mouse[0] != 100 || mouse[1] != 200 {
		t.Errorf("unexpected Mouse values: [%f, %f]", mouse[0], mouse[1])
	}
}
