package editor

import (
	"fmt"
	"testing"

	"github.com/ElioNeto/kora/core/node"
)

func TestSceneToNode_EmptyScene(t *testing.T) {
	sf := &SceneFile{
		Meta:     SceneMeta{Name: "test", Version: 1, LogicalW: 360, LogicalH: 640},
		Entities: []*SceneEntity{},
	}

	root := SceneToNode(sf)
	if root == nil {
		t.Fatal("expected non-nil root")
	}
	if root.GetName() != "test" {
		t.Errorf("expected name 'test', got %q", root.GetName())
	}
}

func TestSceneToNode_SingleSprite(t *testing.T) {
	sf := &SceneFile{
		Meta: SceneMeta{Name: "test", Version: 1, LogicalW: 360, LogicalH: 640},
		Entities: []*SceneEntity{
			{ID: 1, Name: "Player", Type: "sprite", X: 100, Y: 200, W: 32, H: 32, Color: "#00e5a0", Visible: true},
		},
	}

	root := SceneToNode(sf)
	if root == nil {
		t.Fatal("expected non-nil root")
	}

	children := root.GetChildren()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}

	sprite := children[0]
	if sprite.GetName() != "Player" {
		t.Errorf("expected name 'Player', got %q", sprite.GetName())
	}
	if sprite.GetPosition().X != 100 || sprite.GetPosition().Y != 200 {
		t.Errorf("expected position (100,200), got (%.0f,%.0f)",
			sprite.GetPosition().X, sprite.GetPosition().Y)
	}
}

func TestSceneToNode_Hierarchy(t *testing.T) {
	sf := &SceneFile{
		Meta: SceneMeta{Name: "test", Version: 1},
		Entities: []*SceneEntity{
			{ID: 1, Name: "Parent", Type: "custom", X: 0, Y: 0, Children: []*SceneEntity{
				{ID: 2, Name: "Child", Type: "sprite", X: 10, Y: 20},
			}},
			{ID: 3, Name: "Other", Type: "custom", X: 50, Y: 50},
		},
	}

	root := SceneToNode(sf)
	children := root.GetChildren()
	if len(children) != 2 {
		t.Fatalf("expected 2 top-level children, got %d", len(children))
	}

	// First top-level should have 1 child (nested)
	parent := children[0]
	parentChildren := parent.GetChildren()
	if len(parentChildren) != 1 {
		t.Fatalf("expected parent to have 1 child, got %d", len(parentChildren))
	}
	if parentChildren[0].GetName() != "Child" {
		t.Errorf("expected child name 'Child', got %q", parentChildren[0].GetName())
	}
}

func TestSceneToNode_Camera(t *testing.T) {
	sf := &SceneFile{
		Meta: SceneMeta{Name: "test", Version: 1},
		Entities: []*SceneEntity{
			{ID: 1, Name: "MainCam", Type: "camera", X: 160, Y: 120},
		},
	}

	root := SceneToNode(sf)
	children := root.GetChildren()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}

	cam := children[0]
	if cam.GetName() != "MainCam" {
		t.Errorf("expected 'MainCam', got %q", cam.GetName())
	}
}

func TestSceneToNode_Audio(t *testing.T) {
	sf := &SceneFile{
		Meta: SceneMeta{Name: "test", Version: 1},
		Entities: []*SceneEntity{
			{ID: 1, Name: "BGM", Type: "audio", X: 0, Y: 0},
		},
	}

	root := SceneToNode(sf)
	children := root.GetChildren()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}

	audio := children[0]
	if audio.GetName() != "BGM" {
		t.Errorf("expected 'BGM', got %q", audio.GetName())
	}
}

func TestNodeToScene_RoundTrip(t *testing.T) {
	// Create original scene
	sf := &SceneFile{
		Meta: SceneMeta{Name: "roundtrip", Version: 1, LogicalW: 360, LogicalH: 640},
		Entities: []*SceneEntity{
			{ID: 1, Name: "Player", Type: "sprite", X: 100, Y: 200, W: 32, H: 32, Color: "#ff0000", Visible: true},
			{ID: 2, Name: "Cam", Type: "camera", X: 180, Y: 320, Visible: true},
			{ID: 3, Name: "Enemy", Type: "custom", X: 50, Y: 50, Visible: true},
		},
	}

	// Convert to Node2D
	root := SceneToNode(sf)
	if root == nil {
		t.Fatal("SceneToNode returned nil")
	}

	// Convert back to SceneFile
	result := NodeToScene(root, sf.Meta)
	if result == nil {
		t.Fatal("NodeToScene returned nil")
	}

	// Verify the round trip preserved entities (names at minimum)
	names := make(map[string]bool)
	for _, ent := range result.Entities {
		names[ent.Name] = true
	}

	for _, original := range sf.Entities {
		if !names[original.Name] {
			t.Errorf("round trip lost entity %q", original.Name)
		}
	}
}

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		hex    string
		r, g, b, a uint8
	}{
		{"#ff0000", 255, 0, 0, 255},
		{"#00ff00", 0, 255, 0, 255},
		{"#0000ff", 0, 0, 255, 255},
		{"#4a9eff", 74, 158, 255, 255},
		{"", 0, 0, 0, 255},
		{"invalid", 0, 0, 0, 255},
	}

	for _, tt := range tests {
		r, g, b, a := ParseHexColor(tt.hex)
		if r != tt.r || g != tt.g || b != tt.b || a != tt.a {
			t.Errorf("ParseHexColor(%q) = (%d,%d,%d,%d), want (%d,%d,%d,%d)",
				tt.hex, r, g, b, a, tt.r, tt.g, tt.b, tt.a)
		}
	}
}

func TestFindEntity(t *testing.T) {
	sf := &SceneFile{
		Entities: []*SceneEntity{
			{ID: 1, Name: "A"},
			{ID: 2, Name: "B"},
			{ID: 5, Name: "E"},
		},
	}

	if ent := FindEntity(sf, 1); ent == nil || ent.Name != "A" {
		t.Error("failed to find entity 1")
	}
	if ent := FindEntity(sf, 5); ent == nil || ent.Name != "E" {
		t.Error("failed to find entity 5")
	}
	if ent := FindEntity(sf, 99); ent != nil {
		t.Error("found non-existent entity")
	}
}

func TestNextID(t *testing.T) {
	entities := []*SceneEntity{
		{ID: 1}, {ID: 3}, {ID: 7},
	}

	next := NextID(entities)
	if next != 8 {
		t.Errorf("expected 8, got %d", next)
	}
}

func TestTimelineState_BasicPlayback(t *testing.T) {
	clip := NewDefaultClip("test_clip", 2.0, 1)
	ts := NewTimelineState(clip)

	if ts.IsPlaying {
		t.Error("should not be playing initially")
	}

	ts.Play()
	if !ts.IsPlaying {
		t.Error("should be playing after Play()")
	}

	// Tick a few times
	ts.Tick(0.5)
	if ts.CurrentTime < 0.4 || ts.CurrentTime > 0.6 {
		t.Errorf("expected time ~0.5, got %f", ts.CurrentTime)
	}

	ts.Pause()
	if ts.IsPlaying {
		t.Error("should not be playing after Pause()")
	}

	ts.Stop()
	if ts.CurrentTime != 0 {
		t.Errorf("expected time 0 after Stop(), got %f", ts.CurrentTime)
	}
}

func TestTimelineState_Seek(t *testing.T) {
	clip := NewDefaultClip("seek_test", 5.0, 1)
	ts := NewTimelineState(clip)

	ts.Seek(2.5)
	if ts.CurrentTime != 2.5 {
		t.Errorf("expected 2.5, got %f", ts.CurrentTime)
	}

	// Seek beyond end should clamp
	ts.Seek(10)
	if ts.CurrentTime != 5.0 {
		t.Errorf("expected 5.0 (clamped), got %f", ts.CurrentTime)
	}
}

func TestHotReloadState_EnableDisable(t *testing.T) {
	hr := NewHotReloadState(".")

	if hr.IsEnabled() {
		t.Error("should not be enabled initially")
	}

	hr.Enable()
	if !hr.IsEnabled() {
		t.Error("should be enabled after Enable()")
	}

	hr.Disable()
	if hr.IsEnabled() {
		t.Error("should not be enabled after Disable()")
	}
}

func TestNewDefaultClip(t *testing.T) {
	clip := NewDefaultClip("walk", 1.0, 5)

	if clip.Name != "walk" {
		t.Errorf("expected 'walk', got %q", clip.Name)
	}
	if clip.Duration != 1.0 {
		t.Errorf("expected 1.0, got %f", clip.Duration)
	}
	if clip.EntityID != 5 {
		t.Errorf("expected 5, got %d", clip.EntityID)
	}
	if len(clip.Tracks) != 4 {
		t.Errorf("expected 4 tracks, got %d", len(clip.Tracks))
	}
}

func TestInstantiate(t *testing.T) {
	sf := &SceneFile{
		Meta: SceneMeta{Name: "test_instantiate", Version: 1},
		Entities: []*SceneEntity{
			{ID: 1, Name: "Root", Type: "custom", X: 0, Y: 0},
			{ID: 2, Name: "Sprite", Type: "sprite", X: 100, Y: 100, W: 32, H: 32, Color: "#ff0000", Visible: true},
		},
	}

	s, cleanup := Instantiate(sf)
	defer cleanup()

	if s == nil {
		t.Fatal("Instantiate returned nil scene")
	}

	// Scene should be valid
	if s.IsPaused() {
		t.Error("scene should not be paused initially")
	}
}

func BenchmarkSceneToNode(b *testing.B) {
	// Create a scene with 100 entities
	sf := &SceneFile{Meta: SceneMeta{Name: "bench", Version: 1}}
	for i := 0; i < 100; i++ {
		sf.Entities = append(sf.Entities, &SceneEntity{
			ID: i, Name: "Entity" + fmt.Sprintf("%d", i), Type: "sprite",
			X: float64(i) * 10, Y: float64(i) * 10, Visible: true,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		root := SceneToNode(sf)
		if root == nil {
			b.Fatal("SceneToNode returned nil")
		}
	}
}

// Helper to ensure the package compiles with node/scene deps
func TestImports(t *testing.T) {
	_ = node.NewNode2D("test", 0)
}
