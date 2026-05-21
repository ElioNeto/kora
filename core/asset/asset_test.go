package asset_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/ElioNeto/kora/core/asset"
)

// ---------------------------------------------------------------------------
// Test NewAssetManager
// ---------------------------------------------------------------------------

func TestNewAssetManager(t *testing.T) {
	am := asset.NewAssetManager(".")
	if am == nil {
		t.Fatal("expected non-nil AssetManager")
	}
	if count := am.Count(); count != 0 {
		t.Errorf("expected 0 assets, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// Test RegisterLoader and Load
// ---------------------------------------------------------------------------

func TestRegisterLoaderAndLoad(t *testing.T) {
	am := asset.NewAssetManager(".")
	am.RegisterLoader(asset.AssetAudio, func(path string) (interface{}, error) {
		return "mock-audio-data", nil
	})

	ref, err := am.Load(asset.AssetAudio, "test.ogg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref == nil {
		t.Fatal("expected non-nil AssetRef")
	}
	if ref.Type != asset.AssetAudio {
		t.Errorf("expected type AssetAudio, got %v", ref.Type)
	}
	if ref.Name != "test.ogg" {
		t.Errorf("expected name 'test.ogg', got %q", ref.Name)
	}
	if ref.Data != "mock-audio-data" {
		t.Errorf("expected data 'mock-audio-data', got %v", ref.Data)
	}
	if ref.RefCount != 1 {
		t.Errorf("expected refcount 1, got %d", ref.RefCount)
	}
}

// ---------------------------------------------------------------------------
// Test Load returns cached asset (same *AssetRef for same path)
// ---------------------------------------------------------------------------

func TestLoadReturnsCachedAsset(t *testing.T) {
	am := asset.NewAssetManager(".")
	callCount := 0
	am.RegisterLoader(asset.AssetTexture, func(path string) (interface{}, error) {
		callCount++
		return "texture-data", nil
	})

	ref1, err := am.Load(asset.AssetTexture, "sprite.png")
	if err != nil {
		t.Fatalf("first load: %v", err)
	}
	ref2, err := am.Load(asset.AssetTexture, "sprite.png")
	if err != nil {
		t.Fatalf("second load: %v", err)
	}

	if ref1 != ref2 {
		t.Error("expected the same *AssetRef pointer for the same path")
	}
	if callCount != 1 {
		t.Errorf("loader should be called exactly once, called %d times", callCount)
	}
	// RefCount should be 2 after two Load calls.
	if ref1.RefCount != 2 {
		t.Errorf("expected refcount 2, got %d", ref1.RefCount)
	}
}

// ---------------------------------------------------------------------------
// Test Unload decrements refcount and frees at zero
// ---------------------------------------------------------------------------

func TestUnloadDecrementsRefCount(t *testing.T) {
	am := asset.NewAssetManager(".")
	am.RegisterLoader(asset.AssetTexture, func(path string) (interface{}, error) {
		return "texture", nil
	})

	ref, _ := am.Load(asset.AssetTexture, "tex.png")
	if ref.RefCount != 1 {
		t.Fatalf("expected refcount 1 after load, got %d", ref.RefCount)
	}

	// Unload once: refcount goes to 0, asset should be freed.
	am.Unload(ref)
	if ref.RefCount != 0 {
		t.Errorf("expected refcount 0 after unload, got %d", ref.RefCount)
	}

	// Asset should no longer be in cache.
	if got := am.Get("tex.png"); got != nil {
		t.Error("expected nil after unload, asset was not freed")
	}
}

func TestUnloadWithMultipleRefs(t *testing.T) {
	am := asset.NewAssetManager(".")
	am.RegisterLoader(asset.AssetTexture, func(path string) (interface{}, error) {
		return "multi-ref", nil
	})

	ref, _ := am.Load(asset.AssetTexture, "multi.png")
	ref2, _ := am.Load(asset.AssetTexture, "multi.png")
	if ref != ref2 {
		t.Fatal("expected same ref")
	}
	if ref.RefCount != 2 {
		t.Fatalf("expected refcount 2, got %d", ref.RefCount)
	}

	// First unload: refcount goes to 1, asset stays.
	am.Unload(ref)
	if ref.RefCount != 1 {
		t.Errorf("expected refcount 1 after first unload, got %d", ref.RefCount)
	}
	if am.Get("multi.png") == nil {
		t.Error("asset should still be cached after one unload")
	}

	// Second unload: refcount goes to 0, asset freed.
	am.Unload(ref2)
	if ref.RefCount != 0 {
		t.Errorf("expected refcount 0 after second unload, got %d", ref.RefCount)
	}
	if am.Get("multi.png") != nil {
		t.Error("asset should be freed after refcount reaches 0")
	}
}

// ---------------------------------------------------------------------------
// Test Get returns nil for unknown path
// ---------------------------------------------------------------------------

func TestGetReturnsNilForUnknown(t *testing.T) {
	am := asset.NewAssetManager(".")
	if got := am.Get("nonexistent.asset"); got != nil {
		t.Error("expected nil for unknown asset")
	}
}

// ---------------------------------------------------------------------------
// Test Preload loads multiple assets
// ---------------------------------------------------------------------------

func TestPreload(t *testing.T) {
	am := asset.NewAssetManager(".")
	loaded := make(map[string]bool)
	am.RegisterLoader(asset.AssetTexture, func(path string) (interface{}, error) {
		loaded[path] = true
		return "tex", nil
	})
	am.RegisterLoader(asset.AssetAudio, func(path string) (interface{}, error) {
		loaded[path] = true
		return "audio", nil
	})

	errs := am.Preload(map[asset.AssetType][]string{
		asset.AssetTexture: {"a.png", "b.png"},
		asset.AssetAudio:   {"c.ogg"},
	})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if !loaded["a.png"] {
		t.Error("a.png was not loaded")
	}
	if !loaded["b.png"] {
		t.Error("b.png was not loaded")
	}
	if !loaded["c.ogg"] {
		t.Error("c.ogg was not loaded")
	}
	if am.Count() != 3 {
		t.Errorf("expected 3 loaded assets, got %d", am.Count())
	}
}

func TestPreloadCollectsErrors(t *testing.T) {
	am := asset.NewAssetManager(".")
	am.RegisterLoader(asset.AssetTexture, func(path string) (interface{}, error) {
		if path == "broken.png" {
			return nil, os.ErrNotExist
		}
		return "ok", nil
	})

	errs := am.Preload(map[asset.AssetType][]string{
		asset.AssetTexture: {"good.png", "broken.png"},
	})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	// The good asset should still be loaded.
	if am.Get("good.png") == nil {
		t.Error("good.png should be loaded despite errors")
	}
}

// ---------------------------------------------------------------------------
// Test ReleaseUnused
// ---------------------------------------------------------------------------

func TestReleaseUnused(t *testing.T) {
	am := asset.NewAssetManager(".")
	am.RegisterLoader(asset.AssetTexture, func(path string) (interface{}, error) {
		return "tex", nil
	})

	// Load three assets.
	refA, _ := am.Load(asset.AssetTexture, "a.png")
	refB, _ := am.Load(asset.AssetTexture, "b.png")
	refC, _ := am.Load(asset.AssetTexture, "c.png")

	// Unload a (refcount becomes 0, freed immediately by Unload).
	// Unload b after adding an extra ref so it stays alive.
	am.Unload(refA)
	am.Unload(refB) // refcount 1→0, freed

	// ReleaseUnused should find nothing (both already freed by Unload).
	freed := am.ReleaseUnused()
	if freed != 0 {
		t.Errorf("expected 0 freed assets (already freed by Unload), got %d", freed)
	}

	// a and b should be gone.
	if am.Get("a.png") != nil {
		t.Error("a.png should be freed by Unload")
	}
	if am.Get("b.png") != nil {
		t.Error("b.png should be freed by Unload")
	}

	// c should still be cached.
	if am.Get("c.png") == nil {
		t.Error("c.png should still be cached")
	}

	// Unload c (refcount becomes 0).
	am.Unload(refC)
	if am.Get("c.png") != nil {
		t.Error("c.png should be freed after Unload")
	}
}

// ---------------------------------------------------------------------------
// Test Stats
// ---------------------------------------------------------------------------

func TestStats(t *testing.T) {
	am := asset.NewAssetManager(".")
	am.RegisterLoader(asset.AssetTexture, func(path string) (interface{}, error) {
		return "tex", nil
	})

	// Nothing loaded yet.
	stats := am.Stats()
	if stats.Loaded != 0 {
		t.Errorf("expected Loaded=0, got %d", stats.Loaded)
	}
	if stats.Pending != 0 {
		t.Errorf("expected Pending=0, got %d", stats.Pending)
	}
	if stats.Errors != 0 {
		t.Errorf("expected Errors=0, got %d", stats.Errors)
	}

	// Load one asset.
	_, _ = am.Load(asset.AssetTexture, "tex.png")
	stats = am.Stats()
	if stats.Loaded != 1 {
		t.Errorf("expected Loaded=1, got %d", stats.Loaded)
	}

	// Trigger an error.
	am.RegisterLoader(asset.AssetAudio, func(path string) (interface{}, error) {
		return nil, os.ErrInvalid
	})
	_, _ = am.Load(asset.AssetAudio, "bad.ogg")
	stats = am.Stats()
	if stats.Errors != 1 {
		t.Errorf("expected Errors=1, got %d", stats.Errors)
	}
	if stats.Loaded != 1 {
		t.Errorf("expected still Loaded=1, got %d", stats.Loaded)
	}
}

// ---------------------------------------------------------------------------
// Test concurrent Load / Unload
// ---------------------------------------------------------------------------

func TestConcurrentLoadUnload(t *testing.T) {
	am := asset.NewAssetManager(".")
	am.RegisterLoader(asset.AssetTexture, func(path string) (interface{}, error) {
		return "concurrent", nil
	})

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			ref, err := am.Load(asset.AssetTexture, "shared.png")
			if err != nil {
				t.Errorf("concurrent load error: %v", err)
				return
			}
			// Simulate some work with the ref.
			am.Unload(ref)
		}()
	}

	wg.Wait()

	// After all goroutines finish:
	// refcount should be 0 (each Load was paired with an Unload)
	stats := am.Stats()
	// The asset may have been freed by the last Unload (if refcount reached 0)
	// or may still be in cache with refcount 0 (before ReleaseUnused).
	// The important thing is no panics, races, or corruption.
	ref := am.Get("shared.png")
	if ref != nil && ref.RefCount != 0 {
		t.Errorf("expected refcount 0 or nil, got refcount=%d", ref.RefCount)
	}
	t.Logf("Stats after concurrent: Loaded=%d, Errors=%d", stats.Loaded, stats.Errors)
}

func TestConcurrentDifferentPaths(t *testing.T) {
	am := asset.NewAssetManager(".")
	am.RegisterLoader(asset.AssetTexture, func(path string) (interface{}, error) {
		return "concurrent-different", nil
	})

	const n = 100
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		path := filepath.Join("assets", "sprite_"+string(rune('a'+i%26))+".png")
		go func(p string) {
			defer wg.Done()
			ref, err := am.Load(asset.AssetTexture, p)
			if err != nil {
				t.Errorf("load %s: %v", p, err)
				return
			}
			am.Unload(ref)
		}(path)
	}

	wg.Wait()

	// All refs should be at 0 (all Load/Unload pairs matched).
	freed := am.ReleaseUnused()
	t.Logf("Freed %d unused assets after concurrent different-path test", freed)
}

// ---------------------------------------------------------------------------
// Test LoadAsync
// ---------------------------------------------------------------------------

func TestLoadAsync(t *testing.T) {
	am := asset.NewAssetManager(".")
	am.RegisterLoader(asset.AssetTexture, func(path string) (interface{}, error) {
		return "async-tex", nil
	})

	done := make(chan struct{})
	am.LoadAsync(asset.AssetTexture, "async.png", func(ref *asset.AssetRef, err error) {
		defer close(done)
		if err != nil {
			t.Errorf("async load error: %v", err)
			return
		}
		if ref == nil {
			t.Error("expected non-nil ref from async load")
			return
		}
		if ref.Data != "async-tex" {
			t.Errorf("expected 'async-tex', got %v", ref.Data)
		}
		// Clean up.
		am.Unload(ref)
	})

	<-done
}

// ---------------------------------------------------------------------------
// Test SceneAssets with temp file
// ---------------------------------------------------------------------------

func TestSceneAssets(t *testing.T) {
	dir := t.TempDir()

	// Create a dummy scene JSON referencing assetIds.
	sceneContent := `{
		"meta": { "name": "test", "version": 1, "logicalW": 800, "logicalH": 600 },
		"entities": [
			{ "id": 1, "name": "sprite1", "type": "Sprite2D", "x": 0, "y": 0, "assetId": "player.png" },
			{ "id": 2, "name": "sprite2", "type": "Sprite2D", "x": 100, "y": 100, "assetId": "enemy.png" },
			{ "id": 3, "name": "bg", "type": "Sprite2D", "x": 0, "y": 0, "assetId": "player.png" }
		]
	}`
	scenePath := filepath.Join(dir, "scene.json")
	if err := os.WriteFile(scenePath, []byte(sceneContent), 0644); err != nil {
		t.Fatal(err)
	}

	am := asset.NewAssetManager(dir)
	am.RegisterLoader(asset.AssetTexture, func(path string) (interface{}, error) {
		return "mock-texture", nil
	})

	if err := am.SceneAssets(scenePath); err != nil {
		t.Fatalf("SceneAssets: %v", err)
	}

	// player.png should be loaded (appears twice but loaded once).
	ref1 := am.Get("player.png")
	if ref1 == nil {
		t.Error("player.png should be loaded")
	}
	// enemy.png should be loaded.
	ref2 := am.Get("enemy.png")
	if ref2 == nil {
		t.Error("enemy.png should be loaded")
	}

	if am.Count() != 2 {
		t.Errorf("expected 2 unique assets, got %d", am.Count())
	}
	_ = ref1
	_ = ref2
}

// ---------------------------------------------------------------------------
// Test edge cases
// ---------------------------------------------------------------------------

func TestUnloadNil(t *testing.T) {
	am := asset.NewAssetManager(".")
	// Should not panic.
	am.Unload(nil)
}

func TestLoadWithNoLoader(t *testing.T) {
	am := asset.NewAssetManager(".")
	// AssetScene has no default loader.
	_, err := am.Load(asset.AssetScene, "scene.json")
	if err == nil {
		t.Error("expected error for type with no loader")
	}
}

func TestGetAfterReleaseUnused(t *testing.T) {
	am := asset.NewAssetManager(".")
	am.RegisterLoader(asset.AssetTexture, func(path string) (interface{}, error) {
		return "tex", nil
	})

	ref, _ := am.Load(asset.AssetTexture, "tmp.png")
	am.Unload(ref)

	// The asset should be freed (refcount 0), but might still be in the map
	// until ReleaseUnused is called.
	_ = am.ReleaseUnused()
	if got := am.Get("tmp.png"); got != nil {
		t.Error("asset should not be retrievable after release")
	}
}

func TestMultipleTypes(t *testing.T) {
	am := asset.NewAssetManager(".")
	am.RegisterLoader(asset.AssetTexture, func(p string) (interface{}, error) { return "tex", nil })
	am.RegisterLoader(asset.AssetAudio, func(p string) (interface{}, error) { return "audio", nil })
	am.RegisterLoader(asset.AssetShader, func(p string) (interface{}, error) { return "shader", nil })

	refs := make([]*asset.AssetRef, 0, 3)

	r1, _ := am.Load(asset.AssetTexture, "a.png")
	refs = append(refs, r1)
	r2, _ := am.Load(asset.AssetAudio, "b.ogg")
	refs = append(refs, r2)
	r3, _ := am.Load(asset.AssetShader, "c.kage")
	refs = append(refs, r3)

	if am.Count() != 3 {
		t.Errorf("expected 3 assets, got %d", am.Count())
	}

	for _, r := range refs {
		am.Unload(r)
	}
	if am.Count() != 0 {
		t.Errorf("expected 0 after unloading all, got %d", am.Count())
	}
}
