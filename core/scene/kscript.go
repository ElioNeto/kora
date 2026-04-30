package scene

import (
	"github.com/ElioNeto/kora/core/node"
)

// KScript API bridge functions.
// These functions are intended to be called from KScript runtime.
// They rely on the global SceneManager and SceneTree set by the runner.
// The runner must call SetKScriptAPIManager and SetKScriptAPITree to provide access.

var (
	kscriptManager *SceneManager
	kscriptTree    *SceneTree
)

// SetKScriptAPIManager sets the SceneManager for KScript API.
// Called by runner during initialization.
func SetKScriptAPIManager(sm *SceneManager) {
	kscriptManager = sm
}

// SetKScriptAPITree sets the SceneTree for KScript API.
// Called by runner during initialization.
func SetKScriptAPITree(st *SceneTree) {
	kscriptTree = st
}

// SceneLoad loads a scene by path (KScript: Scene.load).
func SceneLoad(name string) {
	if kscriptManager == nil {
		return
	}
	_, _ = kscriptManager.Load(name)
}

// SceneInstantiate instantiates a scene as a prefab (KScript: Scene.instantiate).
// Returns a node.Node2D that can be added to a scene tree.
func SceneInstantiate(name string) *node.Node2D {
	if kscriptManager == nil {
		return nil
	}
	n, _ := kscriptManager.Instantiate(name)
	return n
}

// SceneTreeChangeScene changes the current scene via SceneTree (KScript: SceneTree.changeScene).
func SceneTreeChangeScene(name string) {
	if kscriptTree == nil {
		return
	}
	kscriptTree.ChangeScene(name)
}
