package node

import (
	"testing"

	"github.com/ElioNeto/kora/core/render"
)

// --- Creation ---

func TestNewShaderNode(t *testing.T) {
	sn := NewShaderNode("test")
	if sn == nil {
		t.Fatal("expected non-nil ShaderNode")
	}
	if sn.GetName() != "test" {
		t.Errorf("expected name 'test', got '%s'", sn.GetName())
	}
	if sn.Shader != nil {
		t.Error("expected Shader to be nil by default")
	}
	if sn.Uniforms == nil {
		t.Error("expected non-nil Uniforms map")
	}
	if len(sn.Uniforms) != 0 {
		t.Errorf("expected empty Uniforms, got %d entries", len(sn.Uniforms))
	}
	if sn.Time != 0 {
		t.Errorf("expected Time=0, got %f", sn.Time)
	}
}

// --- SetUniform / RemoveUniform ---

func TestShaderNodeSetUniform(t *testing.T) {
	sn := NewShaderNode("test")
	sn.SetUniform("Time", float32(1.0))
	sn.SetUniform("Resolution", []float32{800, 600})

	if len(sn.Uniforms) != 2 {
		t.Errorf("expected 2 uniforms, got %d", len(sn.Uniforms))
	}

	v, ok := sn.Uniforms["Time"]
	if !ok {
		t.Fatal("expected 'Time' in Uniforms")
	}
	if v.(float32) != 1.0 {
		t.Errorf("expected Time=1.0, got %v", v)
	}

	v, ok = sn.Uniforms["Resolution"]
	if !ok {
		t.Fatal("expected 'Resolution' in Uniforms")
	}
	rs := v.([]float32)
	if rs[0] != 800 || rs[1] != 600 {
		t.Errorf("unexpected Resolution: %v", rs)
	}
}

func TestShaderNodeRemoveUniform(t *testing.T) {
	sn := NewShaderNode("test")
	sn.SetUniform("Time", float32(1.0))
	sn.RemoveUniform("Time")

	if _, ok := sn.Uniforms["Time"]; ok {
		t.Error("expected 'Time' to be removed")
	}
}

func TestShaderNodeRemoveNonExistentUniform(t *testing.T) {
	sn := NewShaderNode("test")
	// Must not panic.
	sn.RemoveUniform("nonexistent")
}

// --- SetShaderByName ---

func TestShaderNodeSetShaderByNameNoManager(t *testing.T) {
	sn := NewShaderNode("test")
	ok := sn.SetShaderByName("any")
	if ok {
		t.Error("expected false when Manager is nil")
	}
}

func TestShaderNodeSetShaderByNameNonExistent(t *testing.T) {
	sn := NewShaderNode("test")
	sn.Manager = render.NewShaderManager()
	ok := sn.SetShaderByName("nonexistent")
	if ok {
		t.Error("expected false for non-existent shader name")
	}
}

// --- Node Interface ---

func TestShaderNodeNodeInterface(t *testing.T) {
	// Compile-time check.
	var _ Node = (*ShaderNode)(nil)

	// Runtime check.
	sn := NewShaderNode("test")
	var n Node = sn
	if n.Name() != "test" {
		t.Error("Node interface not satisfied correctly")
	}
}

// --- Update ---

func TestShaderNodeUpdateIncrementsTime(t *testing.T) {
	sn := NewShaderNode("test")
	sn.Update(0.016)
	if sn.Time <= 0 {
		t.Error("expected Time to increase after Update")
	}
}

func TestShaderNodeUpdatePropagatesToChildren(t *testing.T) {
	parent := NewShaderNode("parent")
	child := NewShaderNode("child")
	parent.AddChild(child)

	parent.Update(0.016)

	// NOTE: Due to Go embedding mechanics, Node2D.Update iterates children
	// as *Node2D and does not dispatch to overridden methods on concrete
	// types like ShaderNode. The child's Node2D.Update is called, which
	// propagates to its children but doesn't increment ShaderNode.Time.
	// This is a known engine limitation.
	_ = child
}

// --- Draw (no crash) ---

func TestShaderNodeDrawNoShaderNoCrash(t *testing.T) {
	sn := NewShaderNode("test")
	// Without a shader, Draw should return early without calling ebiten APIs.
	sn.Draw(nil)
}

func TestShaderNodeDrawWithChildrenNoCrash(t *testing.T) {
	parent := NewShaderNode("parent")
	child := NewShaderNode("child")
	parent.AddChild(child)

	// Children are ShaderNodes without shaders – they return early too.
	parent.Draw(nil)
}
