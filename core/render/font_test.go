package render_test

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/ElioNeto/kora/core/render"
)

// safeNewImage creates an *ebiten.Image or skips the test if ebiten is not
// initialised (e.g. in CI without a GPU).
func safeNewImage(t *testing.T, w, h int) *ebiten.Image {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Skip("ebiten not initialised; skipping GPU-dependent test")
		}
	}()
	return ebiten.NewImage(w, h)
}

// ---------------------------------------------------------------------------
// DefaultFont
// ---------------------------------------------------------------------------

func TestDefaultFontInitialized(t *testing.T) {
	// initDefaultFont is called implicitly by DebugTextAt; we trigger it here
	// by calling DebugTextAt with a dummy image.
	img := safeNewImage(t, 10, 10)
	render.DebugTextAt(img, "", 0, 0)

	if render.DefaultFont == nil {
		t.Fatal("DefaultFont should be initialised after DebugTextAt call")
	}
}

func TestDefaultFontLineHeight(t *testing.T) {
	img := safeNewImage(t, 10, 10)
	render.DebugTextAt(img, "", 0, 0)

	if lh := render.DefaultFont.LineHeight(); lh != 8 {
		t.Errorf("expected line height 8, got %d", lh)
	}
}

// ---------------------------------------------------------------------------
// NewBitmapFont / MeasureText
// ---------------------------------------------------------------------------

func TestNewBitmapFont(t *testing.T) {
	atlas := safeNewImage(t, 80, 8) // 10 columns × 8 px each
	font := render.NewBitmapFont(atlas, ' ', 10, 1, 8, 8)
	if font == nil {
		t.Fatal("expected non-nil font")
	}
	if lh := font.LineHeight(); lh != 8 {
		t.Errorf("expected line height 8, got %d", lh)
	}
}

func TestMeasureText(t *testing.T) {
	atlas := safeNewImage(t, 95*8, 8)
	font := render.NewBitmapFont(atlas, 32, 95, 1, 8, 8)

	tests := []struct {
		text   string
		width  float64
		height float64
	}{
		{"", 0, 8},
		{" ", 8, 8},
		{"!", 8, 8},
		{"A", 8, 8},
		{"AB", 16, 8},
		{"ABC", 24, 8},
		{"Hello", 40, 8},
	}

	for _, tt := range tests {
		w, h := font.MeasureText(tt.text)
		if w != tt.width {
			t.Errorf("MeasureText(%q) width = %v, want %v", tt.text, w, tt.width)
		}
		if h != tt.height {
			t.Errorf("MeasureText(%q) height = %v, want %v", tt.text, h, tt.height)
		}
	}
}

func TestMeasureTextUnknownChar(t *testing.T) {
	// Characters outside the font's range should not cause errors;
	// they are mapped to '?' internally.
	atlas := safeNewImage(t, 95*8, 8)
	font := render.NewBitmapFont(atlas, 32, 95, 1, 8, 8)

	// Unicode chars above 126
	w, h := font.MeasureText("\u00E9\u00E0")
	// Each unknown char should be measured as 8px wide (the '?' fallback)
	if w != 16 {
		t.Errorf("expected width 16 for two unknown chars, got %v", w)
	}
	if h != 8 {
		t.Errorf("expected height 8, got %v", h)
	}
}

// ---------------------------------------------------------------------------
// DrawText
// ---------------------------------------------------------------------------

func TestDrawTextNoPanic(t *testing.T) {
	atlas := safeNewImage(t, 95*8, 8)
	font := render.NewBitmapFont(atlas, 32, 95, 1, 8, 8)

	dst := safeNewImage(t, 200, 50)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("DrawText panicked: %v", r)
		}
	}()

	font.DrawText(dst, "Hello, World!", 0, 0, 1.0, nil)

	// Test with scale > 1
	font.DrawText(dst, "Scaled", 10, 10, 2.0, nil)

	// Test with nil screen (should not panic)
	font.DrawText(nil, "nil screen", 0, 0, 1.0, nil)
}

func TestDrawTextColorScale(t *testing.T) {
	atlas := safeNewImage(t, 95*8, 8)
	font := render.NewBitmapFont(atlas, 32, 95, 1, 8, 8)
	dst := safeNewImage(t, 100, 50)

	// Custom color scale (red tint)
	cs := &ebiten.ColorScale{}
	cs.Scale(1, 0, 0, 1)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("DrawText with ColorScale panicked: %v", r)
		}
	}()

	font.DrawText(dst, "Red", 0, 0, 1.0, cs)
}

// ---------------------------------------------------------------------------
// DebugTextAt
// ---------------------------------------------------------------------------

func TestDebugTextAtNoPanic(t *testing.T) {
	img := safeNewImage(t, 200, 50)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("DebugTextAt panicked: %v", r)
		}
	}()

	// Should initialise the default font and draw text
	render.DebugTextAt(img, "FPS: 60.0  Entities: 42", 4, 4)

	// nil screen should not panic
	render.DebugTextAt(nil, "test", 0, 0)
}

// ---------------------------------------------------------------------------
// DrawRect cached pixel (regression test)
// ---------------------------------------------------------------------------

func TestDrawRectDoesNotPanic(t *testing.T) {
	img := safeNewImage(t, 100, 100)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("DrawRect panicked: %v", r)
		}
	}()

	render.DrawRect(img, 10, 10, 50, 30, nil)

	// nil screen should not panic
	render.DrawRect(nil, 0, 0, 10, 10, nil)
}
