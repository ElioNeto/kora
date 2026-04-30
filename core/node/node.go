// Package node implements the core node system for Kora Engine
// Each node represents a game element with 2D transformation properties
package node

import "github.com/hajimehoshi/ebiten/v2"

// Node is the interface that all nodes must satisfy
type Node interface {
	Update(dt float64)
	Draw(screen *ebiten.Image)
	Name() string
	Parent() Node
	Children() []Node
	AddChild(n Node)
	RemoveChild(name string)
	GetNode(path string) Node
}
