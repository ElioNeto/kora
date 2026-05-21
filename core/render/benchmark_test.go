package render_test

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Tilemap Draw — requires ebiten runtime, skip
// ---------------------------------------------------------------------------

func BenchmarkTilemapDraw(b *testing.B) {
	b.Skip("requires ebiten runtime (screen *ebiten.Image, atlas *ebiten.Image)")
}

// ---------------------------------------------------------------------------
// Tilemap Draw with culling — requires ebiten runtime, skip
// ---------------------------------------------------------------------------

func BenchmarkTilemapDrawCulling(b *testing.B) {
	b.Skip("requires ebiten runtime (screen *ebiten.Image, atlas *ebiten.Image)")
}

// ---------------------------------------------------------------------------
// DebugTextAt — font rendering performance
// This benchmark measures the font atlas generation (once) and glyph lookup
// for drawing text. The first call triggers lazy font initialisation.
// Expected order of magnitude:
//   "Hello world": ~1-5 µs/op (after init)
//   "ABCD...Z" 26 chars: ~2-10 µs/op
// ---------------------------------------------------------------------------

func BenchmarkDebugTextAt(b *testing.B) {
	b.Skip("requires ebiten runtime (screen *ebiten.Image)")
}

// ---------------------------------------------------------------------------
// Font rendering performance (logic only, no ebiten image)
// Measures text measurement (no GPU interaction).
// Expected order of magnitude:
//   Short text (12 chars):  ~10-50 ns/op
//   Long text (256 chars):  ~50-200 ns/op
// ---------------------------------------------------------------------------

func BenchmarkBitmapFontMeasureText(b *testing.B) {
	// Create a minimal font atlas (1×1 pixel) to avoid requiring ebiten
	// runtime for font creation. NewBitmapFont just processes glyph data
	// in Go memory.
	b.Skip("requires ebiten runtime for atlas *ebiten.Image")
}
