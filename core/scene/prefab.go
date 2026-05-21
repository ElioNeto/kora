package scene

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/ElioNeto/kora/core/node"
)

// Prefab represents a reusable node template.
type Prefab struct {
	Name     string
	Root     *node.Node2D
	Tags     []string
	Category string
}

// PrefabManager manages named prefabs for the scene system.
// It provides thread-safe registration, retrieval, and instantiation of
// prefabs as deep-copied Node2D trees or NodeEntity wrappers.
type PrefabManager struct {
	prefabs  map[string]*Prefab
	sceneDir string
	mu       sync.RWMutex
}

// NewPrefabManager creates a new PrefabManager.
// sceneDir is reserved for editor integration (may be empty).
func NewPrefabManager(sceneDir string) *PrefabManager {
	return &PrefabManager{
		prefabs:  make(map[string]*Prefab),
		sceneDir: sceneDir,
	}
}

// Register adds a prefab by name. If a prefab with the same name exists,
// it is replaced.
func (pm *PrefabManager) Register(name string, root *node.Node2D) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.prefabs[name] = &Prefab{
		Name: name,
		Root: root,
	}
}

// ---------------------------------------------------------------------------
// JSON file format for .kora.prefab files
// ---------------------------------------------------------------------------

type prefabJSON struct {
	Name     string           `json:"name"`
	Category string           `json:"category"`
	Tags     []string         `json:"tags"`
	Root     *prefabNodeJSON  `json:"root"`
}

type prefabNodeJSON struct {
	Name     string            `json:"name"`
	Type     string            `json:"type,omitempty"`
	X        float64           `json:"x"`
	Y        float64           `json:"y"`
	Width    float64           `json:"w,omitempty"`
	Height   float64           `json:"h,omitempty"`
	Color    string            `json:"color,omitempty"`
	Children []*prefabNodeJSON `json:"children,omitempty"`
}

// Load loads a prefab from a .kora.prefab JSON file and registers it.
// The file's "name" field is used as the registration key.
func (pm *PrefabManager) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var pj prefabJSON
	if err := json.Unmarshal(data, &pj); err != nil {
		return err
	}

	root := prefabNodeToNode(pj.Root)

	pm.mu.Lock()
	pm.prefabs[pj.Name] = &Prefab{
		Name:     pj.Name,
		Root:     root,
		Tags:     pj.Tags,
		Category: pj.Category,
	}
	pm.mu.Unlock()

	return nil
}

// prefabNodeToNode converts a prefabNodeJSON into a *node.Node2D tree.
func prefabNodeToNode(pn *prefabNodeJSON) *node.Node2D {
	if pn == nil {
		return nil
	}
	n := node.NewNode2D(pn.Name, 0)
	n.SetPosition(float32(pn.X), float32(pn.Y))
	for _, child := range pn.Children {
		n.AddChild(prefabNodeToNode(child))
	}
	return n
}

// Get returns a deep copy of the prefab root (a new Node2D tree).
// The caller owns the returned tree and may mutate it freely.
// Returns nil if no prefab with the given name exists.
func (pm *PrefabManager) Get(name string) *node.Node2D {
	pm.mu.RLock()
	prefab, ok := pm.prefabs[name]
	pm.mu.RUnlock()
	if !ok {
		return nil
	}
	return cloneNode(prefab.Root)
}

// Instantiate creates a NodeEntity from the prefab, ready to Spawn().
// The returned NodeEntity wraps a deep copy of the prefab's root tree.
// Returns nil if no prefab with the given name exists.
func (pm *PrefabManager) Instantiate(name string) *NodeEntity {
	root := pm.Get(name)
	if root == nil {
		return nil
	}
	return NewNodeEntity(root)
}

// Has returns whether a prefab with the given name exists.
func (pm *PrefabManager) Has(name string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	_, ok := pm.prefabs[name]
	return ok
}

// Remove removes a prefab from the manager. No-op if the name does not exist.
func (pm *PrefabManager) Remove(name string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.prefabs, name)
}

// Names returns all registered prefab names in no particular order.
func (pm *PrefabManager) Names() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	names := make([]string, 0, len(pm.prefabs))
	for name := range pm.prefabs {
		names = append(names, name)
	}
	return names
}

// Count returns the number of registered prefabs.
func (pm *PrefabManager) Count() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.prefabs)
}

// ---------------------------------------------------------------------------
// cloneNode recursively deep-copies a Node2D tree using exported methods.
// Runtime state (ScriptHandler, InputHandler, CollisionHook) is NOT copied
// because those are embeded runtime behaviours that should be set fresh on
// each instance.
// ---------------------------------------------------------------------------

func cloneNode(n *node.Node2D) *node.Node2D {
	if n == nil {
		return nil
	}

	clone := node.NewNode2D(n.Name(), n.GetID())

	// Copy position
	pos := n.GetPosition()
	clone.SetPosition(pos.X, pos.Y)

	// Copy rotation
	clone.SetRotation(n.GetRotation())

	// Copy scale
	clone.SetScale(n.GetScaleX(), n.GetScaleY())

	// Copy visibility
	clone.SetVisible(n.IsVisible())

	// Copy alive state (typically true, but preserve if explicitly destroyed)
	if !n.IsAlive() {
		clone.Destroy()
	}

	// Recursively clone children
	for _, child := range n.GetChildren() {
		clone.AddChild(cloneNode(child))
	}

	return clone
}
