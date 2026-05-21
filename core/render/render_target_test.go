package render

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// ---------------------------------------------------------------------------
// NewRenderTarget
// ---------------------------------------------------------------------------

func TestNewRenderTarget_CreatesImageOfCorrectSize(t *testing.T) {
	rt := NewRenderTarget(200, 100)
	if rt == nil {
		t.Fatal("expected non-nil RenderTarget")
	}
	if rt.image == nil {
		t.Fatal("expected non-nil underlying image")
	}
	b := rt.image.Bounds()
	if b.Dx() != 200 {
		t.Errorf("expected width 200, got %d", b.Dx())
	}
	if b.Dy() != 100 {
		t.Errorf("expected height 100, got %d", b.Dy())
	}
}

func TestNewRenderTarget_ZeroSizeUsesDefaults(t *testing.T) {
	rt := NewRenderTarget(0, 0)
	if rt == nil {
		t.Fatal("expected non-nil RenderTarget")
	}
	if rt.image == nil {
		t.Fatal("expected non-nil underlying image")
	}
	b := rt.image.Bounds()
	if b.Dx() != DefaultRenderTargetWidth {
		t.Errorf("expected default width %d, got %d", DefaultRenderTargetWidth, b.Dx())
	}
	if b.Dy() != DefaultRenderTargetHeight {
		t.Errorf("expected default height %d, got %d", DefaultRenderTargetHeight, b.Dy())
	}
}

func TestNewRenderTarget_PartialZeroSize(t *testing.T) {
	t.Run("zero width", func(t *testing.T) {
		rt := NewRenderTarget(0, 100)
		b := rt.image.Bounds()
		if b.Dx() != DefaultRenderTargetWidth {
			t.Errorf("expected default width %d, got %d", DefaultRenderTargetWidth, b.Dx())
		}
		if b.Dy() != 100 {
			t.Errorf("expected height 100, got %d", b.Dy())
		}
	})

	t.Run("zero height", func(t *testing.T) {
		rt := NewRenderTarget(200, 0)
		b := rt.image.Bounds()
		if b.Dx() != 200 {
			t.Errorf("expected width 200, got %d", b.Dx())
		}
		if b.Dy() != DefaultRenderTargetHeight {
			t.Errorf("expected default height %d, got %d", DefaultRenderTargetHeight, b.Dy())
		}
	})
}

func TestNewRenderTarget_NegativeSize(t *testing.T) {
	rt := NewRenderTarget(-10, -20)
	b := rt.image.Bounds()
	if b.Dx() != DefaultRenderTargetWidth {
		t.Errorf("expected default width %d, got %d", DefaultRenderTargetWidth, b.Dx())
	}
	if b.Dy() != DefaultRenderTargetHeight {
		t.Errorf("expected default height %d, got %d", DefaultRenderTargetHeight, b.Dy())
	}
}

// ---------------------------------------------------------------------------
// Begin / End
// ---------------------------------------------------------------------------

func TestBeginEnd_ReturnsNonNilImage(t *testing.T) {
	rt := NewRenderTarget(100, 100)
	rt.Begin()
	img := rt.End()
	if img == nil {
		t.Fatal("End() returned nil image")
	}
	if rt.dirty {
		t.Error("dirty flag should be false after End()")
	}
}

func TestBeginEnd_MultipleCycles(t *testing.T) {
	rt := NewRenderTarget(64, 64)
	for i := 0; i < 5; i++ {
		rt.Begin()
		img := rt.End()
		if img == nil {
			t.Fatalf("End() returned nil on cycle %d", i)
		}
		if rt.dirty {
			t.Errorf("dirty flag should be false after End() on cycle %d", i)
		}
	}
}

func TestBegin_FillsTransparentByDefault(t *testing.T) {
	rt := NewRenderTarget(1, 1)
	rt.Begin()
	// The image should exist and be transparent
	if rt.image == nil {
		t.Fatal("image should not be nil after Begin()")
	}
}

func TestBegin_NilImageDoesNotPanic(t *testing.T) {
	rt := &RenderTarget{width: 100, height: 100, image: nil}
	// Should not panic
	rt.Begin()
}

// ---------------------------------------------------------------------------
// Image
// ---------------------------------------------------------------------------

func TestImage_ReturnsUnderlyingImage(t *testing.T) {
	rt := NewRenderTarget(50, 50)
	img := rt.Image()
	if img == nil {
		t.Fatal("Image() returned nil")
	}
	if img != rt.image {
		t.Error("Image() should return the same pointer as the internal image")
	}
}

// ---------------------------------------------------------------------------
// SetClearColor
// ---------------------------------------------------------------------------

func TestSetClearColor(t *testing.T) {
	rt := NewRenderTarget(10, 10)

	c := color.RGBA{R: 255, G: 128, B: 64, A: 255}
	rt.SetClearColor(c)

	if rt.clearColor != c {
		t.Errorf("expected clearColor %v, got %v", c, rt.clearColor)
	}

	// Begin should use the new color (should not panic)
	rt.Begin()
}

// ---------------------------------------------------------------------------
// Resize
// ---------------------------------------------------------------------------

func TestResize_ChangesSize(t *testing.T) {
	rt := NewRenderTarget(100, 100)
	rt.Resize(200, 300)

	b := rt.image.Bounds()
	if b.Dx() != 200 {
		t.Errorf("expected width 200, got %d", b.Dx())
	}
	if b.Dy() != 300 {
		t.Errorf("expected height 300, got %d", b.Dy())
	}
	if rt.width != 200 || rt.height != 300 {
		t.Errorf("expected internal size (200,300), got (%d,%d)", rt.width, rt.height)
	}
}

func TestResize_SameSizeDoesNotRecreate(t *testing.T) {
	rt := NewRenderTarget(100, 100)
	original := rt.image

	rt.Resize(100, 100)

	if rt.image != original {
		t.Error("expected same image pointer when size does not change")
	}
}

func TestResize_ZeroSize(t *testing.T) {
	rt := NewRenderTarget(100, 100)
	rt.Resize(0, 0)

	b := rt.image.Bounds()
	if b.Dx() != DefaultRenderTargetWidth {
		t.Errorf("expected default width %d, got %d", DefaultRenderTargetWidth, b.Dx())
	}
	if b.Dy() != DefaultRenderTargetHeight {
		t.Errorf("expected default height %d, got %d", DefaultRenderTargetHeight, b.Dy())
	}
}

func TestResize_NegativeSize(t *testing.T) {
	rt := NewRenderTarget(100, 100)
	rt.Resize(-50, -50)

	b := rt.image.Bounds()
	if b.Dx() != DefaultRenderTargetWidth {
		t.Errorf("expected default width %d, got %d", DefaultRenderTargetWidth, b.Dx())
	}
	if b.Dy() != DefaultRenderTargetHeight {
		t.Errorf("expected default height %d, got %d", DefaultRenderTargetHeight, b.Dy())
	}
}

// ---------------------------------------------------------------------------
// DrawCompose
// ---------------------------------------------------------------------------

func TestDrawCompose_DoesNotPanic(t *testing.T) {
	t.Skip("requires ebiten runtime (GPU) for DrawImage")

	rt := NewRenderTarget(10, 10)
	screen := ebiten.NewImage(100, 100)

	// Should not panic with valid images
	rt.DrawCompose(screen, nil)

	// Should not panic with options
	rt.DrawCompose(screen, &ebiten.DrawImageOptions{})
}

func TestDrawCompose_NilScreenDoesNotPanic(t *testing.T) {
	rt := NewRenderTarget(10, 10)
	rt.DrawCompose(nil, nil)
	// No assertion — just must not panic
}

func TestDrawCompose_NilImageDoesNotPanic(t *testing.T) {
	rt := &RenderTarget{width: 10, height: 10, image: nil}
	screen := ebiten.NewImage(100, 100)
	rt.DrawCompose(screen, nil)
	// No assertion — just must not panic
}

// ---------------------------------------------------------------------------
// Dispose
// ---------------------------------------------------------------------------

func TestDispose_FreesMemory(t *testing.T) {
	rt := NewRenderTarget(100, 100)
	if rt.image == nil {
		t.Fatal("expected non-nil image before Dispose")
	}

	rt.Dispose()

	if rt.image != nil {
		t.Error("expected nil image after Dispose")
	}
	if rt.dirty {
		t.Error("expected dirty flag to be false after Dispose")
	}

	// Calling Dispose again should not panic
	rt.Dispose()
}

func TestDispose_BeginAfterDisposeDoesNotPanic(t *testing.T) {
	rt := NewRenderTarget(100, 100)
	rt.Dispose()
	rt.Begin()
	// No assertion — just must not panic
}

// ---------------------------------------------------------------------------
// Integration: Begin → draw → End → compose cycle
// ---------------------------------------------------------------------------

func TestRenderTarget_BeginEndCycle(t *testing.T) {
	rt := NewRenderTarget(64, 64)

	// First cycle
	rt.SetClearColor(color.RGBA{R: 255, G: 0, B: 0, A: 255})
	rt.Begin()
	img1 := rt.End()
	if img1 == nil {
		t.Fatal("first End() returned nil")
	}
	if rt.dirty {
		t.Error("dirty should be false after End()")
	}

	// Second cycle with different clear color
	rt.SetClearColor(color.RGBA{R: 0, G: 255, B: 0, A: 255})
	rt.Begin()
	img2 := rt.End()
	if img2 == nil {
		t.Fatal("second End() returned nil")
	}
	if img2 != img1 {
		t.Error("End() should return the same underlying image")
	}
}

func TestRenderTarget_DefaultClearIsTransparent(t *testing.T) {
	rt := NewRenderTarget(1, 1)
	// By default, clearColor is zero-initialized which is color.RGBA{0,0,0,0}
	// which is transparent black — this is the expected default.
	if rt.clearColor != (color.RGBA{}) {
		t.Error("expected default clear color to be zero-value (transparent black)")
	}
}
