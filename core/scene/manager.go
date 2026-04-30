package scene

import (
	"fmt"
	"sync"

	"github.com/ElioNeto/kora/core/async"
	"github.com/ElioNeto/kora/core/node"
)

// SceneManager handles scene loading, transitions, and additive loading
// It ensures scene changes happen between frames to avoid corruption
type SceneManager struct {
	activeScene   *node.Node2D
	pendingScene  *node.Node2D
	additiveScenes map[string]*node.Node2D
	loader        *Loader
	scheduler     *async.Scheduler
	
	// Pending change tracking
	pendingLoad    string
	pendingAdditive string
	changeMutex    sync.Mutex
}

// NewSceneManager creates a new SceneManager
func NewSceneManager(basePath string, scheduler *async.Scheduler) *SceneManager {
	return &SceneManager{
		loader:         NewLoader(basePath),
		scheduler:      scheduler,
		additiveScenes: make(map[string]*node.Node2D),
	}
}

// Load loads a scene from a .kora.json file and returns the root Node2D
func (sm *SceneManager) Load(path string) (*node.Node2D, error) {
	sceneRoot, err := sm.loader.LoadScene(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load scene %s: %w", path, err)
	}
	return sceneRoot, nil
}

// ChangeScene enqueues a scene change to happen after the current frame
// This ensures we never change scenes mid-frame
func (sm *SceneManager) ChangeScene(path string) error {
	sm.changeMutex.Lock()
	defer sm.changeMutex.Unlock()
	
	// Load the new scene immediately
	newScene, err := sm.Load(path)
	if err != nil {
		return fmt.Errorf("failed to prepare scene change to %s: %w", path, err)
	}
	
	// Schedule the actual scene swap for the next frame
	sm.scheduler.Enqueue(func() {
		sm.changeMutex.Lock()
		oldScene := sm.activeScene
		sm.activeScene = newScene
		sm.changeMutex.Unlock()
		
		// Clean up old scene if needed
		if oldScene != nil {
			sm.destroyScene(oldScene)
		}
	})
	
	return nil
}

// LoadAdditive adds a scene on top of the current scene without destroying it
func (sm *SceneManager) LoadAdditive(path string) error {
	sceneRoot, err := sm.Load(path)
	if err != nil {
		return fmt.Errorf("failed to load additive scene %s: %w", path, err)
	}
	
	sm.changeMutex.Lock()
	sm.additiveScenes[path] = sceneRoot
	sm.changeMutex.Unlock()
	
	return nil
}

// Instantiate creates a scene as a prefab at any point in the tree
func (sm *SceneManager) Instantiate(path string, parent *node.Node2D) (*node.Node2D, error) {
	sceneRoot, err := sm.Load(path)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate scene %s: %w", path, err)
	}
	
	// If parent is provided, add the instantiated scene as a child
	if parent != nil {
		parent.AddChild(sceneRoot)
	}
	
	return sceneRoot, nil
}

// GetActiveScene returns the current active scene root
func (sm *SceneManager) GetActiveScene() *node.Node2D {
	sm.changeMutex.Lock()
	defer sm.changeMutex.Unlock()
	return sm.activeScene
}

// GetAdditiveScenes returns all additive scenes
func (sm *SceneManager) GetAdditiveScenes() map[string]*node.Node2D {
	sm.changeMutex.Lock()
	defer sm.changeMutex.Unlock()
	
	// Return a copy to avoid race conditions
	result := make(map[string]*node.Node2D)
	for k, v := range sm.additiveScenes {
		result[k] = v
	}
	return result
}

// Update processes updates for the active scene and all additive scenes
func (sm *SceneManager) Update(dt float64) {
	sm.changeMutex.Lock()
	active := sm.activeScene
	additive := make([]*node.Node2D, 0, len(sm.additiveScenes))
	for _, scene := range sm.additiveScenes {
		additive = append(additive, scene)
	}
	sm.changeMutex.Unlock()
	
	// Update active scene
	if active != nil {
		active.Update(dt)
	}
	
	// Update additive scenes
	for _, scene := range additive {
		scene.Update(dt)
	}
}

// Draw renders the active scene and all additive scenes
func (sm *SceneManager) Draw() {
	sm.changeMutex.Lock()
	active := sm.activeScene
	additive := make([]*node.Node2D, 0, len(sm.additiveScenes))
	for _, scene := range sm.additiveScenes {
		additive = append(additive, scene)
	}
	sm.changeMutex.Unlock()
	
	// Draw active scene
	if active != nil {
		// Note: Actual drawing would be handled by a render system
		// This is a placeholder for the draw call
	}
	
	// Draw additive scenes
	for _, scene := range additive {
		// Draw each additive scene
	}
}

// destroyScene cleans up a scene and all its children
func (sm *SceneManager) destroyScene(scene *node.Node2D) {
	if scene == nil {
		return
	}
	
	// Remove all children
	scene.RemoveAllChildren()
	
	// Additional cleanup could be done here
	// For now, just clear the reference
}

// Reload reloads the current active scene
func (sm *SceneManager) Reload() error {
	sm.changeMutex.Lock()
	active := sm.activeScene
	sm.changeMutex.Unlock()
	
	if active == nil {
		return fmt.Errorf("no active scene to reload")
	}
	
	// In a real implementation, we would need to track the path of the active scene
	// For now, this is a placeholder
	return fmt.Errorf("reload not fully implemented - need to track scene paths")
}