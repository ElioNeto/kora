package scene_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/ElioNeto/kora/core/node"
	"github.com/ElioNeto/kora/core/scene"
)

// ---------------------------------------------------------------------------
// PrefabManager creation
// ---------------------------------------------------------------------------

func TestPrefabManagerCreation(t *testing.T) {
	pm := scene.NewPrefabManager("")
	if pm == nil {
		t.Fatal("NewPrefabManager returned nil")
	}
	if pm.Count() != 0 {
		t.Errorf("expected 0 prefabs initially, got %d", pm.Count())
	}

	names := pm.Names()
	if len(names) != 0 {
		t.Errorf("expected empty Names(), got %v", names)
	}
}

// ---------------------------------------------------------------------------
// Register / Get — deep copy verification
// ---------------------------------------------------------------------------

func TestRegisterAndGetDeepCopy(t *testing.T) {
	pm := scene.NewPrefabManager("")

	// Build a simple tree: root -> child
	root := node.NewNode2D("root", 1)
	child := node.NewNode2D("child", 2)
	root.AddChild(child)
	root.SetPosition(10, 20)
	root.SetRotation(45)
	root.SetScale(2, 3)
	root.SetVisible(false)

	pm.Register("test", root)

	// Get returns a deep copy
	clone := pm.Get("test")
	if clone == nil {
		t.Fatal("Get returned nil for registered prefab")
	}

	// Verify clone has same properties
	if clone.Name() != "root" {
		t.Errorf("expected name 'root', got '%s'", clone.Name())
	}
	pos := clone.GetPosition()
	if pos.X != 10 || pos.Y != 20 {
		t.Errorf("expected position (10,20), got (%f,%f)", pos.X, pos.Y)
	}
	if clone.GetRotation() != 45 {
		t.Errorf("expected rotation 45, got %f", clone.GetRotation())
	}
	sx, sy := clone.GetScaleX(), clone.GetScaleY()
	if sx != 2 || sy != 3 {
		t.Errorf("expected scale (2,3), got (%f,%f)", sx, sy)
	}
	if clone.IsVisible() {
		t.Error("expected visible=false")
	}
	if clone.GetChildCount() != 1 {
		t.Errorf("expected 1 child, got %d", clone.GetChildCount())
	}
	gotChild := clone.GetChild("child")
	if gotChild == nil {
		t.Fatal("expected child 'child' in clone")
	}

	// --- Mutating the clone must NOT affect the original ---
	clone.SetPosition(99, 99)
	clone.RemoveAllChildren()

	origPos := root.GetPosition()
	if origPos.X != 10 || origPos.Y != 20 {
		t.Errorf("original position changed after clone mutation: (%f,%f)", origPos.X, origPos.Y)
	}
	if root.GetChildCount() != 1 {
		t.Errorf("original children count changed after clone mutation: %d", root.GetChildCount())
	}

	// --- Successive Get calls return independent copies ---
	clone2 := pm.Get("test")
	if clone2 == clone {
		t.Error("successive Get calls should return different copies")
	}
	clone2.SetPosition(1, 2)
	pos2 := clone.GetPosition()
	if pos2.X == 1 {
		t.Error("mutating clone2 should not affect clone1")
	}
}

// ---------------------------------------------------------------------------
// Has
// ---------------------------------------------------------------------------

func TestHas(t *testing.T) {
	pm := scene.NewPrefabManager("")

	if pm.Has("nonexistent") {
		t.Error("Has should return false for unregistered prefab")
	}

	root := node.NewNode2D("test", 1)
	pm.Register("my_prefab", root)
	if !pm.Has("my_prefab") {
		t.Error("Has should return true after Register")
	}

	pm.Remove("my_prefab")
	if pm.Has("my_prefab") {
		t.Error("Has should return false after Remove")
	}
}

// ---------------------------------------------------------------------------
// Remove
// ---------------------------------------------------------------------------

func TestRemove(t *testing.T) {
	pm := scene.NewPrefabManager("")

	pm.Register("a", node.NewNode2D("a_root", 1))
	if pm.Count() != 1 {
		t.Errorf("expected count 1, got %d", pm.Count())
	}

	pm.Remove("a")
	if pm.Count() != 0 {
		t.Errorf("expected count 0 after Remove, got %d", pm.Count())
	}
	if pm.Has("a") {
		t.Error("prefab should not exist after Remove")
	}

	// Remove nonexistent should not panic
	pm.Remove("does_not_exist")
}

// ---------------------------------------------------------------------------
// Names / Count
// ---------------------------------------------------------------------------

func TestNamesCount(t *testing.T) {
	pm := scene.NewPrefabManager("")

	// Empty manager
	if pm.Count() != 0 {
		t.Errorf("expected count 0, got %d", pm.Count())
	}
	if len(pm.Names()) != 0 {
		t.Errorf("expected empty Names, got %v", pm.Names())
	}

	// Register three prefabs
	pm.Register("a", node.NewNode2D("a_root", 1))
	pm.Register("b", node.NewNode2D("b_root", 2))
	pm.Register("c", node.NewNode2D("c_root", 3))

	if pm.Count() != 3 {
		t.Errorf("expected count 3, got %d", pm.Count())
	}

	names := pm.Names()
	if len(names) != 3 {
		t.Errorf("expected 3 names, got %d: %v", len(names), names)
	}

	found := make(map[string]bool)
	for _, n := range names {
		found[n] = true
	}
	for _, want := range []string{"a", "b", "c"} {
		if !found[want] {
			t.Errorf("Names() missing expected name %q", want)
		}
	}
}

// ---------------------------------------------------------------------------
// Two prefabs are independent
// ---------------------------------------------------------------------------

func TestPrefabsIndependent(t *testing.T) {
	pm := scene.NewPrefabManager("")

	root1 := node.NewNode2D("prefab1", 1)
	root2 := node.NewNode2D("prefab2", 2)

	pm.Register("p1", root1)
	pm.Register("p2", root2)

	clone1 := pm.Get("p1")
	clone2 := pm.Get("p2")

	clone1.SetName("modified")
	if clone2.Name() != "prefab2" {
		t.Error("modifying one prefab clone should not affect another")
	}

	// Removing one should not affect the other
	pm.Remove("p1")
	if !pm.Has("p2") {
		t.Error("p2 should still exist after p1 is removed")
	}
	if pm.Has("p1") {
		t.Error("p1 should not exist after removal")
	}
	if pm.Count() != 1 {
		t.Errorf("expected count 1, got %d", pm.Count())
	}
}

// ---------------------------------------------------------------------------
// Instantiate
// ---------------------------------------------------------------------------

func TestInstantiate(t *testing.T) {
	pm := scene.NewPrefabManager("")

	root := node.NewNode2D("test_root", 1)
	child := node.NewNode2D("child", 2)
	root.AddChild(child)

	pm.Register("test", root)

	// Instantiate an existing prefab
	entity := pm.Instantiate("test")
	if entity == nil {
		t.Fatal("Instantiate returned nil for existing prefab")
	}

	// Ensure it wraps a deep copy
	entityRoot := entity.Root()
	if entityRoot == nil {
		t.Fatal("entity root is nil")
	}
	if entityRoot.Name() != "test_root" {
		t.Errorf("expected root name 'test_root', got '%s'", entityRoot.Name())
	}
	if entityRoot == root {
		t.Error("Instantiate should return a clone, not the original")
	}

	// Ensure the entity satisfies the Entity interface
	var _ scene.Entity = entity

	// Instantiate nonexistent should return nil
	if nilEntity := pm.Instantiate("nonexistent"); nilEntity != nil {
		t.Error("Instantiate for nonexistent prefab should return nil")
	}
}

// ---------------------------------------------------------------------------
// Register replaces existing prefab
// ---------------------------------------------------------------------------

func TestRegisterReplaces(t *testing.T) {
	pm := scene.NewPrefabManager("")

	root1 := node.NewNode2D("original", 1)
	pm.Register("dup", root1)

	root2 := node.NewNode2D("replacement", 2)
	pm.Register("dup", root2)

	if pm.Count() != 1 {
		t.Errorf("expected count 1 after replace, got %d", pm.Count())
	}

	clone := pm.Get("dup")
	if clone.Name() != "replacement" {
		t.Errorf("expected name 'replacement' after replace, got '%s'", clone.Name())
	}
}

// ---------------------------------------------------------------------------
// Load from .kora.prefab JSON
// ---------------------------------------------------------------------------

func TestLoadPrefabJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "enemy_slime.kora.prefab")

	content := `{
		"name": "enemy_slime",
		"category": "enemies",
		"tags": ["slime", "enemy", "green"],
		"root": {
			"name": "Slime",
			"type": "sprite",
			"x": 0, "y": 0, "w": 32, "h": 32,
			"color": "#00ff88",
			"children": [
				{ "name": "Collider", "type": "custom", "x": 0, "y": 16, "w": 28, "h": 12 }
			]
		}
	}`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write prefab file: %v", err)
	}

	pm := scene.NewPrefabManager(dir)
	if err := pm.Load(path); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !pm.Has("enemy_slime") {
		t.Fatal("prefab 'enemy_slime' should exist after Load")
	}
	if pm.Count() != 1 {
		t.Errorf("expected count 1, got %d", pm.Count())
	}

	// Instantiate and verify structure
	entity := pm.Instantiate("enemy_slime")
	if entity == nil {
		t.Fatal("Instantiate after Load returned nil")
	}

	rootNode := entity.Root()
	if rootNode.Name() != "Slime" {
		t.Errorf("expected root name 'Slime', got '%s'", rootNode.Name())
	}
	pos := rootNode.GetPosition()
	if pos.X != 0 || pos.Y != 0 {
		t.Errorf("expected root position (0,0), got (%f,%f)", pos.X, pos.Y)
	}
	if rootNode.GetChildCount() != 1 {
		t.Errorf("expected 1 child, got %d", rootNode.GetChildCount())
	}
	child := rootNode.GetChild("Collider")
	if child == nil {
		t.Fatal("expected child 'Collider'")
	}
	cpos := child.GetPosition()
	if cpos.X != 0 || cpos.Y != 16 {
		t.Errorf("expected Collider position (0,16), got (%f,%f)", cpos.X, cpos.Y)
	}
}

func TestLoadPrefabFileNotFound(t *testing.T) {
	pm := scene.NewPrefabManager("")
	err := pm.Load("/nonexistent/path.prefab")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

// ---------------------------------------------------------------------------
// Thread safety (best-effort verification; run with -race for full check)
// ---------------------------------------------------------------------------

func TestConcurrentAccess(t *testing.T) {
	pm := scene.NewPrefabManager("")

	// Pre-register a prefab so reads have something to work with.
	root := node.NewNode2D("conc", 1)
	pm.Register("conc", root)

	var wg sync.WaitGroup

	// Launch concurrent readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				pm.Has("conc")
				pm.Get("conc")
				pm.Names()
				pm.Count()
			}
		}()
	}

	// Launch concurrent writers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				r := node.NewNode2D("r", uint64(id*1000+j))
				pm.Register("conc", r)
				pm.Register("tmp", r)
				pm.Remove("tmp")
			}
		}(i)
	}

	wg.Wait()

	// After all goroutines finish, the manager should still be in a
	// consistent state — no data races or panics.
	if pm.Count() == 0 {
		t.Error("expected at least one prefab remaining after concurrent access")
	}
	if !pm.Has("conc") {
		t.Error("expected 'conc' prefab to exist after concurrent access")
	}
}
