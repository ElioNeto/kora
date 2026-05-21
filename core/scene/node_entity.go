package scene

import (
	"github.com/ElioNeto/kora/core/node"
	"github.com/ElioNeto/kora/core/render"
)

// NodeEntity wraps a *node.Node2D tree as a scene.Entity, bridging the
// two previously incompatible object systems.
//
// When spawned into a Scene via Spawn(), the NodeEntity makes the entire
// Node2D tree participate in the Entity-based update/draw loop. The tree's
// Update() and Draw() methods propagate to all children automatically.
type NodeEntity struct {
	BaseEntity
	root *node.Node2D
}

// NewNodeEntity creates a NodeEntity wrapping the given Node2D root.
// The root may be nil (useful as a placeholder).
func NewNodeEntity(root *node.Node2D) *NodeEntity {
	return &NodeEntity{
		root: root,
	}
}

// Root returns the underlying Node2D root, or nil.
func (ne *NodeEntity) Root() *node.Node2D {
	return ne.root
}

// SetRoot replaces the underlying Node2D root.
func (ne *NodeEntity) SetRoot(root *node.Node2D) {
	ne.root = root
}

// IsAlive returns true as long as the entity and its root are alive.
func (ne *NodeEntity) IsAlive() bool {
	return ne.root != nil && ne.root.IsAlive()
}

// Destroy marks the entity and its root as destroyed.
func (ne *NodeEntity) Destroy() {
	if ne.root != nil {
		ne.root.Destroy()
	}
}

// Update delegates to Node2D.Update, which propagates to all children.
func (ne *NodeEntity) Update(dt float64) {
	if ne.root != nil {
		ne.root.Update(dt)
	}
}

// Draw delegates to Node2D.Draw, which propagates to all children.
func (ne *NodeEntity) Draw(r interface{}) {
	if ne.root == nil {
		return
	}
	// The Node2D.Draw expects *ebiten.Image.
	// The render.Renderer is passed as interface{} from Scene.Draw.
	// Extract the screen from the renderer.
	if renderer, ok := r.(*render.Renderer); ok {
		screen := renderer.Screen()
		if screen != nil {
			ne.root.Draw(screen)
		}
	}
}

// Compile-time interface checks
var (
	_ Entity   = (*NodeEntity)(nil)
	_ Updater  = (*NodeEntity)(nil)
	_ Drawer   = (*NodeEntity)(nil)
)
