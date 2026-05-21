// Package node implements the core node system for Kora Engine
// Each node represents a game element with 2D transformation properties
package node

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/ElioNeto/kora/core/math"
	"github.com/ElioNeto/kora/core/render"
)

// Node2D is the base class for all 2D nodes in the engine
// All game elements should inherit from Node2D
type Node2D struct {
	// Basic properties
	name       string
	id         uint64
	parent     *Node2D
	children   []*Node2D
	script     ScriptHandler

	// Position in parent space
	pos math.Vector2

	// Rotation in degrees
	rotation float32

	// Scale factors
	scaleX, scaleY float32

	// Visibility
	visible bool

	// Entity lifecycle
	alive bool

	// Input handling hooks
	inputHandler  InputHandler
	collisionHook CollisionHook
}

// ScriptHandler is the type for KScript handlers associated with nodes
type ScriptHandler interface {
	Update(dt float64)
	Input(event InputEvent)
}

// InputHandler is called when the node receives input
type InputHandler func(event InputEvent)

// CollisionHook is called when collision events occur
type CollisionHook func(other *Node2D, eventType CollisionType)

// InputEvent represents input events that nodes can handle
type InputEvent struct {
	Type     string  // "pressed", "released", "down"
	Key      string  // Keyboard key name
	Axis     float32 // Axis value (-1.0 to 1.0)
	X, Y     float32 // Mouse/touch position
	Button   int     // Mouse button index
	DeltaX   float32 // Mouse deltaX
	DeltaY   float32 // Mouse deltaY
	TouchPos Point   // Touch position
}

// CollisionType represents types of collision events
type CollisionType int

const (
	CollisionTypeEnter CollisionType = iota
	CollisionTypeExit
	CollisionTypeOverlap
	CollisionTypeCollide
)

// Point represents a 2D point
type Point struct {
	X, Y float32
}

// NewNode2D creates a new Node2D instance
func NewNode2D(name string, id uint64) *Node2D {
	return &Node2D{
		name:    name,
		id:      id,
		pos:     math.NewVector2(0, 0),
		scaleX:  1.0,
		scaleY:  1.0,
		visible: true,
		alive:   true,
	}
}

// IsAlive implements scene.Entity
func (n *Node2D) IsAlive() bool {
	return n.alive
}

// Destroy implements scene.Entity
func (n *Node2D) Destroy() {
	n.alive = false
}

// Render renders the node and its children using the internal renderer
func (n *Node2D) Render(r *render.Renderer) {
	if !n.visible || !n.alive {
		return
	}
	for _, child := range n.children {
		child.Render(r)
	}
}

// Draw satisfies the Node interface (renders to ebiten.Image)
func (n *Node2D) Draw(screen *ebiten.Image) {
	if !n.visible || !n.alive {
		return
	}
	for _, child := range n.children {
		child.Draw(screen)
	}
}

// GetNode2D returns the underlying *Node2D pointer.
// This is used by AddChild to support types that embed *Node2D.
func (n *Node2D) GetNode2D() *Node2D { return n }

// FindCamera searches this node and its children recursively for a Camera2D.
// Returns the first Camera2D found, or nil if none exists.
// This is used by the runner to determine the active camera each frame.
//
// Since children are stored as *Node2D pointers, and Camera2D embeds *Node2D,
// this method can only find a Camera2D when called directly on it (as a root),
// or when the child implements a type-assertable interface.
// For reliable camera detection, use the runner's ActiveCamera field.
func (n *Node2D) FindCamera() *Camera2D {
	if cam, ok := interface{}(n).(*Camera2D); ok {
		return cam
	}
	for _, child := range n.children {
		if cam := child.FindCamera(); cam != nil {
			return cam
		}
	}
	return nil
}

// Compile-time interface check
var _ Node = (*Node2D)(nil)

// GetName returns the node name
func (n *Node2D) GetName() string {
	return n.name
}

// Name returns the node name (satisfies Node interface)
func (n *Node2D) Name() string {
	return n.name
}

// GetID returns the node id
func (n *Node2D) GetID() uint64 {
	return n.id
}

// SetName sets the node name
func (n *Node2D) SetName(name string) {
	n.name = name
}

// GetParent returns the parent node or nil
func (n *Node2D) GetParent() *Node2D {
	return n.parent
}

// Parent returns the parent node as Node interface (satisfies Node interface)
func (n *Node2D) Parent() Node {
	if n.parent == nil {
		return nil
	}
	return n.parent
}

// Children returns children as []Node (satisfies Node interface)
func (n *Node2D) Children() []Node {
	result := make([]Node, len(n.children))
	for i, child := range n.children {
		result[i] = child
	}
	return result
}

// AddChild adds a child node to this node (satisfies Node interface)
func (n *Node2D) AddChild(child Node) {
	if child == nil {
		return
	}

	// Extract the underlying *Node2D from any type that embeds it.
	node2d := extractNode2D(child)
	if node2d == nil {
		panic("Node2D.AddChild: child must embed *Node2D, got " + getTypeName(child))
	}

	// Check if already child
	for _, c := range n.children {
		if c == node2d {
			return
		}
	}

	// Set parent
	node2d.parent = n
	n.children = append(n.children, node2d)
}

// extractNode2D attempts to extract the underlying *Node2D from a Node value.
// It supports direct *Node2D values and any type that embeds *Node2D.
func extractNode2D(child Node) *Node2D {
	// Direct *Node2D
	if n, ok := child.(*Node2D); ok {
		return n
	}

	// Check for common embedded types by trying interface conversion.
	// This handles PhysicsBody2D, Camera2D, Area2D, AudioPlayer2D, etc.
	type node2DProvider interface {
		GetNode2D() *Node2D
	}
	if p, ok := child.(node2DProvider); ok {
		return p.GetNode2D()
	}

	return nil
}

// getTypeName returns the type name of an interface for panic messages.
func getTypeName(i interface{}) string {
	return fmt.Sprintf("%T", i)
}

// RemoveChild removes a child node by name (satisfies Node interface)
func (n *Node2D) RemoveChild(name string) {
	for i, c := range n.children {
		if c.GetName() == name {
			// Remove from slice
			n.children = append(n.children[:i], n.children[i+1:]...)
			c.parent = nil
			return
		}
	}
}

// RemoveAllChildren removes all children from this node
func (n *Node2D) RemoveAllChildren() {
	for _, child := range n.children {
		child.parent = nil
	}
	n.children = nil
}

// GetChild returns the child by name
func (n *Node2D) GetChild(name string) *Node2D {
	for _, child := range n.children {
		if child.GetName() == name {
			return child
		}
	}
	return nil
}

// GetChildren returns all children
func (n *Node2D) GetChildren() []*Node2D {
	return n.children
}

// GetNode returns a node by path (e.g., "Player/Sprite")
// Returns nil if the path does not exist
func (n *Node2D) GetNode(path string) Node {
	if path == "" {
		return nil
	}

	parts := splitPath(path)
	if len(parts) == 0 {
		return nil
	}

	// If the first part matches the current node's name, skip it
	// (i.e., "A/B/C" starting from node A means start from A)
	startIdx := 0
	if parts[0] == n.GetName() {
		startIdx = 1
	}

	current := n
	for i := startIdx; i < len(parts); i++ {
		part := parts[i]
		found := false
		for _, child := range current.children {
			if child.GetName() == part {
				current = child
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	return current
}

// splitPath splits a path by "/" separator
func splitPath(path string) []string {
	parts := strings.Split(path, "/")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// GetChildCount returns the number of children
func (n *Node2D) GetChildCount() int {
	return len(n.children)
}

// SetPosition sets the node position
func (n *Node2D) SetPosition(x, y float32) {
	n.pos.X = x
	n.pos.Y = y
}

// GetPosition returns the node position
func (n *Node2D) GetPosition() math.Vector2 {
	return n.pos
}

// SetX sets the X position
func (n *Node2D) SetX(x float32) {
	n.pos.X = x
}

// SetY sets the Y position
func (n *Node2D) SetY(y float32) {
	n.pos.Y = y
}

// SetRotation sets the rotation in degrees
func (n *Node2D) SetRotation(degrees float32) {
	n.rotation = degrees
}

// GetRotation returns the rotation in degrees
func (n *Node2D) GetRotation() float32 {
	return n.rotation
}

// SetScaleX sets the X scale factor
func (n *Node2D) SetScaleX(x float32) {
	n.scaleX = x
}

// SetScaleY sets the Y scale factor
func (n *Node2D) SetScaleY(y float32) {
	n.scaleY = y
}

// GetX returns the X position
func (n *Node2D) GetX() float32 {
	return n.pos.X
}

// GetY returns the Y position
func (n *Node2D) GetY() float32 {
	return n.pos.Y
}

// GetScaleX returns the X scale factor
func (n *Node2D) GetScaleX() float32 {
	return n.scaleX
}

// GetScaleY returns the Y scale factor
func (n *Node2D) GetScaleY() float32 {
	return n.scaleY
}

// SetScale sets both scale factors
func (n *Node2D) SetScale(x float32, y float32) {
	n.scaleX = x
	n.scaleY = y
}

// SetVisible sets the visibility
func (n *Node2D) SetVisible(visible bool) {
	n.visible = visible
}

// IsVisible returns whether the node is visible
func (n *Node2D) IsVisible() bool {
	return n.visible
}

// SetWorldPosition sets position (same as SetPosition for root nodes)
func (n *Node2D) SetWorldPosition(x, y float32) {
	n.pos.X = x
	n.pos.Y = y
}

// GetWorldPosition returns the world position
func (n *Node2D) GetWorldPosition() math.Vector2 {
	if n.parent == nil {
		return n.pos
	}
	return n.parent.TransformPoint(n.pos)
}

// GetWorldRotation returns the absolute rotation in world space
func (n *Node2D) GetWorldRotation() float32 {
	if n.parent == nil {
		return n.rotation
	}
	return n.parent.rotation + n.rotation
}

// TransformPoint transforms a point from node space to world space
func (n *Node2D) TransformPoint(point math.Vector2) math.Vector2 {
	if n.parent == nil {
		return n.pos.Add(point)
	}
	return n.parent.TransformPoint(n.pos.Add(point))
}

// SetScript sets the KScript handler for this node
func (n *Node2D) SetScript(script ScriptHandler) {
	n.script = script
}

// GetScript returns the KScript handler
func (n *Node2D) GetScript() ScriptHandler {
	return n.script
}

// Update processes KScript update for the node
func (n *Node2D) Update(dt float64) {
	if n.script != nil {
		n.script.Update(dt)
	}

	// Propagate to children
	for _, child := range n.children {
		if child != nil {
			child.Update(dt)
		}
	}
}

// ProcessInput sends input event to the node and its children
func (n *Node2D) ProcessInput(event InputEvent) {
	if n.inputHandler != nil {
		n.inputHandler(event)
	}

	// Propagate to children (in reverse order so children get processed first)
	for i := len(n.children) - 1; i >= 0; i-- {
		child := n.children[i]
		if child != nil && child.inputHandler != nil {
			child.inputHandler(event)
		}
	}
}

// SetInputHandler sets the input handler for this node
func (n *Node2D) SetInputHandler(handler InputHandler) {
	n.inputHandler = handler
}

// SetCollisionHook sets the collision callback for this node
func (n *Node2D) SetCollisionHook(hook CollisionHook) {
	n.collisionHook = hook
}

// TriggerCollision triggers the collision hook
func (n *Node2D) TriggerCollision(other *Node2D, eventType CollisionType) {
	if n.collisionHook != nil {
		n.collisionHook(other, eventType)
	}
}
