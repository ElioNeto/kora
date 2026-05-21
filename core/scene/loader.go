package scene

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/ElioNeto/kora/core/node"
)

type sceneMeta struct {
	Name     string `json:"name"`
	Version  int    `json:"version"`
	LogicalW int    `json:"logicalW"`
	LogicalH int    `json:"logicalH"`
}

type sceneEntity struct {
	ID       int            `json:"id"`
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	X        float64        `json:"x"`
	Y        float64        `json:"y"`
	W        float64        `json:"w"`
	H        float64        `json:"h"`
	AssetID  string         `json:"assetId"`
	Script   string         `json:"script"`
	ParentID int            `json:"parentId,omitempty"`
	Children []*sceneEntity `json:"children,omitempty"`
}

type sceneJSON struct {
	Meta        sceneMeta      `json:"meta"`
	ParentScene string         `json:"parentScene,omitempty"`
	Entities    []*sceneEntity `json:"entities"`
}

func loadSceneJSON(path string) (*sceneJSON, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s sceneJSON
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func entityToNode(e *sceneEntity) *node.Node2D {
	n := node.NewNode2D(e.Name, uint64(e.ID))
	n.SetPosition(float32(e.X), float32(e.Y))
	for _, child := range e.Children {
		n.AddChild(entityToNode(child))
	}
	return n
}

func LoadScene(path string) (*node.Node2D, error) {
	s, err := loadSceneJSON(path)
	if err != nil {
		return nil, err
	}

	// Load parent scene first if specified
	var root *node.Node2D
	if s.ParentScene != "" {
		parentPath := filepath.Join(filepath.Dir(path), s.ParentScene)
		parentRoot, err := LoadScene(parentPath)
		if err != nil {
			return nil, err
		}
		root = parentRoot
	} else {
		root = node.NewNode2D(s.Meta.Name, 0)
	}

	// Add current scene entities
	for _, e := range s.Entities {
		root.AddChild(entityToNode(e))
	}
	return root, nil
}

// LoadSceneEntity loads a scene file and returns a NodeEntity wrapping the
// Node2D tree. This can be Spawn()'d directly into a Scene, bridging the
// Node2D tree into the Entity-based update/draw loop.
func LoadSceneEntity(path string) (*NodeEntity, error) {
	root, err := LoadScene(path)
	if err != nil {
		return nil, err
	}
	return NewNodeEntity(root), nil
}
