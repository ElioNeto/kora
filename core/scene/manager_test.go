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
	if prefab.GetName() != "Prefab" {
		t.Errorf("expected prefab name Prefab, got %s", prefab.GetName())
	}
}
