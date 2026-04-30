package scene

import (
	"errors"
	"path/filepath"

	"github.com/ElioNeto/kora/core/node"
	"github.com/ElioNeto/kora/core/render"
)

var ErrSceneManagerNil = errors.New("scene manager is nil")

type SceneManager struct {
	currentScene   *node.Node2D
	pendingScene   *node.Node2D
	pendingPath    string // path for deferred loading
	additiveScenes []*node.Node2D
	sceneDir       string
}

func NewSceneManager(sceneDir string) *SceneManager {
	return &SceneManager{
		sceneDir:       sceneDir,
		additiveScenes: make([]*node.Node2D, 0),
	}
}

func (sm *SceneManager) Load(path string) (*node.Node2D, error) {
	fullPath := filepath.Join(sm.sceneDir, path)
	return LoadScene(fullPath)
}

func (sm *SceneManager) ChangeScene(path string) error {
	// Queue the path for loading in the next Update, ensuring we never
	// load mid-frame. The actual scene switch happens in Update.
	sm.pendingPath = path
	return nil
}

func (sm *SceneManager) LoadAdditive(path string) error {
	fullPath := filepath.Join(sm.sceneDir, path)
	scene, err := LoadScene(fullPath)
	if err != nil {
		return err
	}
	sm.additiveScenes = append(sm.additiveScenes, scene)
	return nil
}

func (sm *SceneManager) Instantiate(path string) (*node.Node2D, error) {
	return sm.Load(path)
}

func (sm *SceneManager) Update(dt float64) {
	// Handle deferred scene loading (ChangeScene queued a path)
	if sm.pendingPath != "" {
		fullPath := filepath.Join(sm.sceneDir, sm.pendingPath)
		scene, err := LoadScene(fullPath)
		if err == nil {
			sm.pendingScene = scene
		}
		// If error, we just skip; maybe log later
		sm.pendingPath = ""
	}
	// Apply pending scene change
	if sm.pendingScene != nil {
		sm.currentScene = sm.pendingScene
		sm.pendingScene = nil
	}
	if sm.currentScene != nil {
		sm.currentScene.Update(dt)
	}
	for _, s := range sm.additiveScenes {
		s.Update(dt)
	}
}

func (sm *SceneManager) CurrentScene() *node.Node2D {
	return sm.currentScene
}

func (sm *SceneManager) Draw(r *render.Renderer) {
	if sm.currentScene != nil {
		sm.currentScene.Render(r)
	}
	for _, s := range sm.additiveScenes {
		s.Render(r)
	}
}
