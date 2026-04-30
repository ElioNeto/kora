package scene

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ElioNeto/kora/core/node"
)

// KoraScene represents the structure of a .kora.json file
type KoraScene struct {
	Kora     string             `json:"kora"`
	Name     string             `json:"name"`
	Version  int                `json:"version"`
	LogicalW int                `json:"logicalW"`
	LogicalH int                `json:"logicalH"`
	Entities []KoraEntity      `json:"entities"`
	Meta     map[string]interface{} `json:"meta,omitempty"`
}

// KoraEntity represents an entity in the scene JSON
type KoraEntity struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	X        float32                `json:"x"`
	Y        float32                `json:"y"`
	W        float32                `json:"w"`
	H        float32                `json:"h"`
	Rotation float32                `json:"rotation"`
	Visible  bool                   `json:"visible"`
	Locked   bool                   `json:"locked"`
	Color    string                 `json:"color"`
	Script   string                 `json:"script"`
	Parent   string                 `json:"parent,omitempty"`
	Children []string               `json:"children,omitempty"`
	Fields   map[string]interface{} `json:"fields,omitempty"`
}

// Loader handles deserialization of .kora.json files into Node2D trees
type Loader struct {
	basePath string
}

// NewLoader creates a new scene loader
func NewLoader(basePath string) *Loader {
	return &Loader{
		basePath: basePath,
	}
}

// LoadScene loads a .kora.json file and returns the root Node2D
func (l *Loader) LoadScene(path string) (*node.Node2D, error) {
	fullPath := filepath.Join(l.basePath, path)
	
	// Read the JSON file
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read scene file %s: %w", path, err)
	}

	// Parse the JSON
	var scene KoraScene
	if err := json.Unmarshal(data, &scene); err != nil {
		return nil, fmt.Errorf("failed to parse scene JSON %s: %w", path, err)
	}

	// Validate scene format
	if scene.Kora != "1.0" {
		return nil, fmt.Errorf("unsupported scene format: %s", scene.Kora)
	}

	// Convert entities to Node2D tree
	root, err := l.buildNodeTree(scene.Entities)
	if err != nil {
		return nil, fmt.Errorf("failed to build node tree: %w", err)
	}

	return root, nil
}

// buildNodeTree converts KoraEntity slice to Node2D tree structure
func (l *Loader) buildNodeTree(entities []KoraEntity) (*node.Node2D, error) {
	entityMap := make(map[string]*node.Node2D)
	var root *node.Node2D

	// First pass: create all nodes
	for _, entity := range entities {
		node := node.NewNode2D(entity.Name, l.generateID())
		node.SetPosition(entity.X, entity.Y)
		node.SetRotation(entity.Rotation)
		node.SetVisible(entity.Visible)
		
		// Store in map
		entityMap[entity.ID] = node
		
		// Set as potential root (first node if no parent specified)
		if root == nil && entity.Parent == "" {
			root = node
		}
	}

	// Second pass: build hierarchy
	for _, entity := range entities {
		currentNode := entityMap[entity.ID]
		
		if entity.Parent != "" {
			parentNode, exists := entityMap[entity.Parent]
			if exists {
				parentNode.AddChild(currentNode)
			} else {
				return nil, fmt.Errorf("parent node %s not found for entity %s", entity.Parent, entity.ID)
			}
		}
		
		// Add children nodes
		for _, childID := range entity.Children {
			childNode, exists := entityMap[childID]
			if exists {
				currentNode.AddChild(childNode)
			}
		}
	}

	if root == nil {
		return nil, fmt.Errorf("no root node found in scene")
	}

	return root, nil
}

// generateID generates a simple ID counter (in real implementation, use proper ID generation)
func (l *Loader) generateID() uint64 {
	return 1 // Placeholder - should use proper ID generation
}

// SerializeScene converts a Node2D tree to KoraScene format
func (l *Loader) SerializeScene(root *node.Node2D) (*KoraScene, error) {
	scene := &KoraScene{
		Kora:     "1.0",
		Name:     "Generated Scene",
		Version:  1,
		LogicalW: 360,
		LogicalH: 640,
		Entities: []KoraEntity{},
	}

	// Traverse the tree and serialize nodes
	l.serializeNode(root, scene, make(map[string]bool))
	
	return scene, nil
}

// serializeNode recursively serializes a node and its children
func (l *Loader) serializeNode(node *node.Node2D, scene *KoraScene, visited map[string]bool) {
	nodeID := fmt.Sprintf("node_%d", node.GetID())
	if visited[nodeID] {
		return
	}
	visited[nodeID] = true

	entity := KoraEntity{
		ID:       nodeID,
		Name:     node.GetName(),
		Type:     "node", // Default type, could be determined by node type
		X:        node.GetX(),
		Y:        node.GetY(),
		W:        48, // Default width
		H:        48, // Default height
		Rotation: node.GetRotation(),
		Visible:  node.IsVisible(),
		Color:    "#00e5a0", // Default color
	}

	scene.Entities = append(scene.Entities, entity)

	// Serialize children
	for _, child := range node.GetChildren() {
		l.serializeNode(child, scene, visited)
	}
}