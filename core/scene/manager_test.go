package scene

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSceneValidJSON(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "test.kora.json")
	jsonContent := `{
		"meta": { "name": "TestScene", "version": 1, "logicalW": 360, "logicalH": 640 },
		"entities": [
			{
				"id": 1, "name": "Player", "type": "sprite",
				"x": 180, "y": 320, "w": 48, "h": 48,
				"assetId": "asset_player", "script": "test.js"
			}
		]
	}`
	if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("failed to write test json: %v", err)
	}

	scene, err := LoadScene(jsonPath)
	if err != nil {
		t.Fatalf("LoadScene failed: %v", err)
	}
	if scene.GetName() != "TestScene" {
		t.Errorf("expected scene name TestScene, got %s", scene.GetName())
	}
	if len(scene.GetChildren()) != 1 {
		t.Errorf("expected 1 child, got %d", len(scene.GetChildren()))
	}
}

func TestSceneManagerChangeScene(t *testing.T) {
	dir := t.TempDir()
	sm := NewSceneManager(dir)

	json1 := `{
		"meta": { "name": "Scene1", "version": 1, "logicalW": 360, "logicalH": 640 },
		"entities": []
	}`
	json2 := `{
		"meta": { "name": "Scene2", "version": 1, "logicalW": 360, "logicalH": 640 },
		"entities": []
	}`
	if err := os.WriteFile(filepath.Join(dir, "scene1.kora.json"), []byte(json1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "scene2.kora.json"), []byte(json2), 0644); err != nil {
		t.Fatal(err)
	}

	if err := sm.ChangeScene("scene1.kora.json"); err != nil {
		t.Fatalf("ChangeScene failed: %v", err)
	}
	sm.Update(0)
	if sm.CurrentScene().GetName() != "Scene1" {
		t.Errorf("expected current scene Scene1, got %s", sm.CurrentScene().GetName())
	}

	if err := sm.ChangeScene("scene2.kora.json"); err != nil {
		t.Fatalf("ChangeScene failed: %v", err)
	}
	sm.Update(0)
	if sm.CurrentScene().GetName() != "Scene2" {
		t.Errorf("expected current scene Scene2, got %s", sm.CurrentScene().GetName())
	}
}

func TestInstantiatePrefab(t *testing.T) {
	dir := t.TempDir()
	sm := NewSceneManager(dir)

	jsonContent := `{
		"meta": { "name": "Prefab", "version": 1, "logicalW": 360, "logicalH": 640 },
		"entities": [
			{ "id": 1, "name": "Bullet", "type": "sprite", "x": 0, "y": 0, "w": 10, "h": 10, "assetId": "bullet", "script": "" }
		]
	}`
	if err := os.WriteFile(filepath.Join(dir, "bullet.kora.json"), []byte(jsonContent), 0644); err != nil {
		t.Fatal(err)
	}

	prefab, err := sm.Instantiate("bullet.kora.json")
	if err != nil {
		t.Fatalf("Instantiate failed: %v", err)
	}
	if prefab.Name() != "Prefab" {
		t.Errorf("expected prefab name Prefab, got %s", prefab.Name())
	}
}

func TestLoadAdditive(t *testing.T) {
	dir := t.TempDir()
	sm := NewSceneManager(dir)

	base := `{
		"meta": { "name": "Base", "version": 1, "logicalW": 360, "logicalH": 640 },
		"entities": [ { "id": 1, "name": "Bg", "type": "sprite", "x": 0, "y": 0, "w": 360, "h": 640, "assetId": "bg", "script": "" } ]
	}`
	overlay := `{
		"meta": { "name": "Overlay", "version": 1, "logicalW": 360, "logicalH": 640 },
		"entities": [ { "id": 2, "name": "UI", "type": "sprite", "x": 0, "y": 0, "w": 100, "h": 50, "assetId": "ui", "script": "" } ]
	}`
	if err := os.WriteFile(filepath.Join(dir, "base.kora.json"), []byte(base), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "overlay.kora.json"), []byte(overlay), 0644); err != nil {
		t.Fatal(err)
	}

	if err := sm.LoadAdditive("base.kora.json"); err != nil {
		t.Fatalf("LoadAdditive base failed: %v", err)
	}
	if err := sm.LoadAdditive("overlay.kora.json"); err != nil {
		t.Fatalf("LoadAdditive overlay failed: %v", err)
	}
	// Trigger update to load pending additive scenes? Actually LoadAdditive already adds.
	// But we need to verify they are in additiveScenes.
	sm.Update(0)
	if len(sm.additiveScenes) != 2 {
		t.Errorf("expected 2 additive scenes, got %d", len(sm.additiveScenes))
	}
}

func TestChangeSceneFrameSafe(t *testing.T) {
	dir := t.TempDir()
	sm := NewSceneManager(dir)

	json1 := `{
		"meta": { "name": "Scene1", "version": 1, "logicalW": 360, "logicalH": 640 },
		"entities": []
	}`
	json2 := `{
		"meta": { "name": "Scene2", "version": 1, "logicalW": 360, "logicalH": 640 },
		"entities": []
	}`
	if err := os.WriteFile(filepath.Join(dir, "s1.kora.json"), []byte(json1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "s2.kora.json"), []byte(json2), 0644); err != nil {
		t.Fatal(err)
	}

	// Initially no current scene
	if sm.CurrentScene() != nil {
		t.Error("expected no current scene")
	}

	// ChangeScene should not change current scene immediately
	sm.ChangeScene("s1.kora.json")
	if sm.CurrentScene() != nil {
		t.Error("expected current scene still nil before Update")
	}

	// After Update, scene should be loaded and set
	sm.Update(0)
	if sm.CurrentScene() == nil {
		t.Fatal("expected current scene after Update")
	}
	if sm.CurrentScene().GetName() != "Scene1" {
		t.Errorf("expected Scene1, got %s", sm.CurrentScene().GetName())
	}

	// Now change to scene2, should not change until next Update
	sm.ChangeScene("s2.kora.json")
	if sm.CurrentScene().GetName() != "Scene1" {
		t.Errorf("expected still Scene1 before Update, got %s", sm.CurrentScene().GetName())
	}
	sm.Update(0)
	if sm.CurrentScene().GetName() != "Scene2" {
		t.Errorf("expected Scene2 after Update, got %s", sm.CurrentScene().GetName())
	}
}

func TestParentSceneInheritance(t *testing.T) {
	dir := t.TempDir()
	sm := NewSceneManager(dir)

	parent := `{
		"meta": { "name": "Parent", "version": 1, "logicalW": 360, "logicalH": 640 },
		"entities": [ { "id": 1, "name": "ParentNode", "type": "sprite", "x": 0, "y": 0, "w": 10, "h": 10, "assetId": "p", "script": "" } ]
	}`
	child := `{
		"meta": { "name": "Child", "version": 1, "logicalW": 360, "logicalH": 640 },
		"parentScene": "parent.kora.json",
		"entities": [ { "id": 2, "name": "ChildNode", "type": "sprite", "x": 10, "y": 10, "w": 10, "h": 10, "assetId": "c", "script": "" } ]
	}`
	if err := os.WriteFile(filepath.Join(dir, "parent.kora.json"), []byte(parent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "child.kora.json"), []byte(child), 0644); err != nil {
		t.Fatal(err)
	}

	scene, err := sm.Load("child.kora.json")
	if err != nil {
		t.Fatalf("Load child scene failed: %v", err)
	}
	// The root is the parent scene's root (named "Parent").
	// Child entities are added as children of this root.
	// So we expect root name "Parent" and at least two children (ParentNode and ChildNode).
	if scene.GetName() != "Parent" {
		t.Errorf("expected root name Parent, got %s", scene.GetName())
	}
	children := scene.GetChildren()
	// Collect child names
	childNames := make([]string, 0)
	for _, c := range children {
		childNames = append(childNames, c.Name())
	}
	hasParentNode := false
	hasChildNode := false
	for _, name := range childNames {
		if name == "ParentNode" {
			hasParentNode = true
		}
		if name == "ChildNode" {
			hasChildNode = true
		}
	}
	if !hasParentNode {
		t.Error("expected ParentNode from parent scene")
	}
	if !hasChildNode {
		t.Error("expected ChildNode from child scene")
	}
}
