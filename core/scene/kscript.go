package scene

import (
	"github.com/ElioNeto/kora/core/autoload"
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

// SceneLoad loads a scene by path and sets it as current (KScript: Scene.load).
func SceneLoad(name string) {
	if kscriptManager == nil {
		return
	}
	kscriptManager.ChangeScene(name)
}

// SceneInstantiate instantiates a scene as a prefab (KScript: Scene.instantiate).
// Returns a node.Node that can be added to a scene tree.
func SceneInstantiate(name string) node.Node {
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

// ----------------------------------------------------------------------------
// AutoLoad API (KScript bridge)
// ----------------------------------------------------------------------------

// AutoLoadSet registers an AutoLoad singleton globally (KScript: AutoLoad.set).
// name is the global identifier; obj is the singleton instance.
func AutoLoadSet(name string, obj interface{}) {
	// The KScript compiler emits structs that implement autoload.AutoLoad;
	// this bridge provides a registration entry point from scripts.
	if a, ok := obj.(autoload.AutoLoad); ok {
		autoload.Set(a)
	}
}

// AutoLoadGet returns a registered AutoLoad by name (KScript: AutoLoad.get).
// Returns nil if not found.
func AutoLoadGet(name string) interface{} {
	return autoload.Get(name)
}

// init ensures the autoload package is imported.
var _ = autoload.Len
