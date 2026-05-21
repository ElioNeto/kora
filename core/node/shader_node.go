package node

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/ElioNeto/kora/core/render"
)

// ShaderNode is a node that renders with a custom shader applied.
// If a shader is set, the node renders a full-screen quad with the shader
// effect. Children are drawn underneath the shader overlay.
type ShaderNode struct {
	*Node2D

	// ShaderName is the lookup key used when fetching from a Manager.
	ShaderName string

	// Shader is the compiled Kage shader to render.
	Shader *ebiten.Shader

	// Uniforms is a map of uniform name → value for the shader.
	Uniforms map[string]interface{}

	// Manager optionally holds a *render.ShaderManager for named lookups.
	Manager *render.ShaderManager

	// Time accumulates elapsed seconds. Updated each frame in Update.
	Time float64

	whitePixel *ebiten.Image // cached 1×1 white texture for simple effects
}

// NewShaderNode creates a new ShaderNode with the given name.
func NewShaderNode(name string) *ShaderNode {
	return &ShaderNode{
		Node2D:   NewNode2D(name, 0),
		Uniforms: make(map[string]interface{}),
	}
}

// SetShaderByName looks up a compiled shader from the attached Manager
// by name and sets it on this node. Returns false if the Manager is nil
// or if the shader name is not found.
func (sn *ShaderNode) SetShaderByName(name string) bool {
	if sn.Manager == nil {
		return false
	}
	shader := sn.Manager.GetShader(name)
	if shader == nil {
		return false
	}
	sn.ShaderName = name
	sn.Shader = shader
	return true
}

// SetShaderSource compiles shader source and sets it directly on the node.
func (sn *ShaderNode) SetShaderSource(name string, source []byte) error {
	shader, err := render.CompileShader(source)
	if err != nil {
		return err
	}
	sn.ShaderName = name
	sn.Shader = shader
	return nil
}

// SetUniform sets a shader uniform value. Creates the Uniforms map
// if it is nil.
func (sn *ShaderNode) SetUniform(key string, value interface{}) {
	if sn.Uniforms == nil {
		sn.Uniforms = make(map[string]interface{})
	}
	sn.Uniforms[key] = value
}

// RemoveUniform removes a shader uniform value. Safe to call for
// non-existent keys.
func (sn *ShaderNode) RemoveUniform(key string) {
	delete(sn.Uniforms, key)
}

// Update increments the internal time counter and propagates the
// update to children.
func (sn *ShaderNode) Update(dt float64) {
	sn.Time += dt
	sn.Node2D.Update(dt)
}

// Draw renders the shader node. If a shader is set, it renders a
// full-screen quad with the shader effect on top of the children.
// Otherwise the node draws nothing (transparent).
func (sn *ShaderNode) Draw(screen *ebiten.Image) {
	if !sn.visible || !sn.alive {
		return
	}

	// Draw children first (they appear underneath the shader overlay).
	for _, child := range sn.children {
		child.Draw(screen)
	}

	// Without a shader this node is transparent.
	if sn.Shader == nil {
		return
	}

	// Lazily create a 1×1 white texture used when the shader only needs
	// a solid-colour source (Images[0] is required by DrawRectShader).
	if sn.whitePixel == nil {
		sn.whitePixel = ebiten.NewImage(1, 1)
		sn.whitePixel.Fill(color.White)
	}

	opts := &ebiten.DrawRectShaderOptions{
		Uniforms: sn.Uniforms,
	}
	opts.Images[0] = sn.whitePixel

	bounds := screen.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	screen.DrawRectShader(w, h, sn.Shader, opts)
}

// Compile-time check that *ShaderNode satisfies the Node interface.
var _ Node = (*ShaderNode)(nil)
