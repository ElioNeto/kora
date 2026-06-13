package editor

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ElioNeto/kora/core/node"
	"github.com/ElioNeto/kora/core/scene"
)

// ─── SceneEntity → Node2D ────────────────────────────────────────────────────

// SceneToNode converts a SceneFile into a Node2D tree ready for the runtime.
// The root node returned is a container that holds all top-level entities.
//
// Usage:
//
//	root := editor.SceneToNode(&sceneFile)
//	nodeEntity := scene.NewNodeEntity(root)
//	myScene.SetNodeRoot(nodeEntity)
func SceneToNode(sf *SceneFile) *node.Node2D {
	if sf == nil {
		return node.NewNode2D("root", 0)
	}

	root := node.NewNode2D(sf.Meta.Name, 0)

	// Build lookup by ID
	byID := make(map[int]*SceneEntity, len(sf.Entities))
	for _, ent := range sf.Entities {
		byID[ent.ID] = ent
	}

	// Convert each entity, maintaining parent-child hierarchy
	converted := make(map[int]*node.Node2D, len(sf.Entities))

	for _, ent := range sf.Entities {
		if ent.ParentID != 0 {
			continue // children handled after parents
		}
		n := entityToNode(ent, byID, converted)
		root.AddChild(n)
	}

	return root
}

// entityToNode converts a single SceneEntity (and its children) to Node2D.
func entityToNode(ent *SceneEntity, byID map[int]*SceneEntity, converted map[int]*node.Node2D) node.Node {
	var n node.Node

	switch ent.Type {
	case "sprite":
		sprite := node.NewSprite2D(ent.Name)
		sprite.Node2D.SetPosition(float32(ent.X), float32(ent.Y))
		sprite.Node2D.SetRotation(float32(ent.Rotation))
		sprite.Node2D.SetVisible(ent.Visible)
		if ent.W > 0 {
			sprite.SetSize(float32(ent.W), float32(ent.H))
		}
		if ent.Color != "" {
			sprite.SetColorString(ent.Color)
		}
		n = sprite

	case "camera":
		cam := node.NewCamera2D(ent.Name)
		cam.Node2D.SetPosition(float32(ent.X), float32(ent.Y))
		cam.Node2D.SetRotation(float32(ent.Rotation))
		cam.Node2D.SetVisible(ent.Visible)
		n = cam

	case "tilemap":
		// Tilemaps use base Node2D (actual tile data is in render system)
		tileNode := node.NewNode2D(ent.Name, uint64(ent.ID))
		tileNode.SetPosition(float32(ent.X), float32(ent.Y))
		tileNode.SetVisible(ent.Visible)
		n = tileNode

	case "audio":
		audio := node.NewAudioPlayer2D(ent.Name)
		audio.Node2D.SetPosition(float32(ent.X), float32(ent.Y))
		audio.Node2D.SetVisible(ent.Visible)
		n = audio

	default: // "custom" or unknown
		base := node.NewNode2D(ent.Name, uint64(ent.ID))
		base.SetPosition(float32(ent.X), float32(ent.Y))
		base.SetRotation(float32(ent.Rotation))
		base.SetVisible(ent.Visible)
		n = base
	}

	converted[ent.ID] = extractNode2D(n)

	// Convert children from the ent.Children slice
	for _, child := range ent.Children {
		childNode := entityToNode(child, byID, converted)
		n.AddChild(childNode)
	}

	// Also find children by ParentID (flat list support)
	for _, other := range byID {
		if other.ParentID == ent.ID && other.ID != ent.ID {
			if _, done := converted[other.ID]; !done {
				childNode := entityToNode(other, byID, converted)
				n.AddChild(childNode)
			}
		}
	}

	return n
}

// extractNode2D gets the underlying *node.Node2D from any node type.
func extractNode2D(n node.Node) *node.Node2D {
	if n == nil {
		return nil
	}
	if n2d, ok := n.(*node.Node2D); ok {
		return n2d
	}
	type node2DProvider interface{ GetNode2D() *node.Node2D }
	if p, ok := n.(node2DProvider); ok {
		return p.GetNode2D()
	}
	return nil
}

// ─── Node2D → SceneEntity ────────────────────────────────────────────────────

// NodeToScene converts a Node2D tree back into a SceneFile.
// This is the reverse of SceneToNode, used when saving the scene
// after runtime edits.
func NodeToScene(root *node.Node2D, meta SceneMeta) *SceneFile {
	sf := &SceneFile{
		Meta:     meta,
		Entities: make([]*SceneEntity, 0),
	}

	if root == nil {
		return sf
	}

	collectNodes(root, nil, sf, make(map[*node.Node2D]int))
	return sf
}

func collectNodes(n *node.Node2D, parent *SceneEntity, sf *SceneFile, visited map[*node.Node2D]int) {
	if n == nil {
		return
	}
	if _, seen := visited[n]; seen {
		return
	}
	visited[n] = len(sf.Entities) + 1

	ent := nodeToEntity(n, parent)
	sf.Entities = append(sf.Entities, ent)

	for _, child := range n.GetChildren() {
		collectNodes(child, ent, sf, visited)
	}
}

func nodeToEntity(n *node.Node2D, parent *SceneEntity) *SceneEntity {
	ent := &SceneEntity{
		ID:       int(n.GetID()),
		Name:     n.GetName(),
		X:        float64(n.GetPosition().X),
		Y:        float64(n.GetPosition().Y),
		W:        32,
		H:        32,
		Rotation: float64(n.GetRotation()),
		Visible:  n.IsVisible(),
		Children: make([]*SceneEntity, 0),
	}

	if parent != nil {
		ent.ParentID = parent.ID
	}

	ent.Type = detectNodeType(n)
	return ent
}

func detectNodeType(n *node.Node2D) string {
	typeName := fmt.Sprintf("%T", n)
	switch {
	case strings.Contains(typeName, "Sprite2D"):
		return "sprite"
	case strings.Contains(typeName, "Camera2D"):
		return "camera"
	case strings.Contains(typeName, "AudioPlayer2D"):
		return "audio"
	default:
		return "custom"
	}
}

// ─── Scene Instantiation ─────────────────────────────────────────────────────

// Instantiate creates a ready-to-use Scene filled with the entities
// from the SceneFile. Returns the scene and a cleanup function.
func Instantiate(sf *SceneFile) (*scene.Scene, func()) {
	s := scene.New()
	root := SceneToNode(sf)
	nodeEntity := scene.NewNodeEntity(root)
	s.SetNodeRoot(nodeEntity)

	return s, func() {
		nodeEntity.Destroy()
	}
}

// ─── Color Parsing ───────────────────────────────────────────────────────────

// ParseHexColor parses "#RRGGBB" or "#RRGGBBAA" into RGBA values.
func ParseHexColor(hex string) (r, g, b, a uint8) {
	a = 255
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) < 6 {
		return 0, 0, 0, 255
	}
	r64, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g64, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b64, _ := strconv.ParseUint(hex[4:6], 16, 8)
	if len(hex) >= 8 {
		a64, _ := strconv.ParseUint(hex[6:8], 16, 8)
		a = uint8(a64)
	}
	return uint8(r64), uint8(g64), uint8(b64), a
}

// ─── Entity Lookup ───────────────────────────────────────────────────────────

// FindEntity returns the entity with the given ID, or nil.
func FindEntity(sf *SceneFile, id int) *SceneEntity {
	for _, ent := range sf.Entities {
		if ent.ID == id {
			return ent
		}
		if ent.Children != nil {
			if found := findInChildren(ent.Children, id); found != nil {
				return found
			}
		}
	}
	return nil
}

func findInChildren(children []*SceneEntity, id int) *SceneEntity {
	for _, child := range children {
		if child.ID == id {
			return child
		}
		if child.Children != nil {
			if found := findInChildren(child.Children, id); found != nil {
				return found
			}
		}
	}
	return nil
}

// NextID returns the next available entity ID.
func NextID(entities []*SceneEntity) int {
	maxID := 0
	for _, ent := range entities {
		if ent.ID > maxID {
			maxID = ent.ID
		}
	}
	return maxID + 1
}
