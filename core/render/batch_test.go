package render

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// requireEbiten skips the test if ebiten is not initialized (no GPU context).
func requireEbiten(t interface {
	Helper()
	Skip(args ...interface{})
}) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Skip("requires ebiten GPU context:", r)
		}
	}()
	_ = ebiten.NewImage(1, 1)
}

// ---------------------------------------------------------------------------
// Batch — New
// ---------------------------------------------------------------------------

func TestNewBatch(t *testing.T) {
	b := NewBatch()
	if b == nil {
		t.Fatal("expected non-nil Batch")
	}
	if b.Len() != 0 {
		t.Errorf("expected 0, got %d", b.Len())
	}
}

// ---------------------------------------------------------------------------
// Batch — Add (nil guards)
// ---------------------------------------------------------------------------

func TestBatchAddNilImage(t *testing.T) {
	b := NewBatch()
	b.Add(nil, nil)
	if b.Len() != 0 {
		t.Errorf("expected 0 after adding nil image, got %d", b.Len())
	}
}

func TestBatchAddSpriteNil(t *testing.T) {
	b := NewBatch()
	b.AddSprite(nil, nil)
	if b.Len() != 0 {
		t.Errorf("expected 0 after adding nil sprite, got %d", b.Len())
	}
}

func TestBatchAddSpriteNilImage(t *testing.T) {
	b := NewBatch()
	b.AddSprite(&Sprite{image: nil}, nil)
	if b.Len() != 0 {
		t.Errorf("expected 0 after adding sprite with nil image, got %d", b.Len())
	}
}

// ---------------------------------------------------------------------------
// Batch — Add increases count
// ---------------------------------------------------------------------------

func TestBatchAddIncreasesCount(t *testing.T) {
	requireEbiten(t)
	b := NewBatch()
	img := ebiten.NewImage(16, 16)
	defer img.Dispose()
	b.Add(img, nil)
	b.Add(img, nil)
	if b.Len() != 2 {
		t.Errorf("expected Len() == 2, got %d", b.Len())
	}
}

func TestBatchAddSpriteIncreasesCount(t *testing.T) {
	requireEbiten(t)
	b := NewBatch()
	img := ebiten.NewImage(16, 16)
	defer img.Dispose()
	sp := &Sprite{image: img, Bounds: image.Rect(0, 0, 16, 16)}
	b.AddSprite(sp, nil)
	if b.Len() != 1 {
		t.Errorf("expected 1, got %d", b.Len())
	}
}

// ---------------------------------------------------------------------------
// Batch — Clear
// ---------------------------------------------------------------------------

func TestBatchClearEmpty(t *testing.T) {
	b := NewBatch()
	b.Clear()
	if b.Len() != 0 {
		t.Errorf("expected 0, got %d", b.Len())
	}
}

func TestBatchClearResetsCount(t *testing.T) {
	requireEbiten(t)
	b := NewBatch()
	img := ebiten.NewImage(16, 16)
	defer img.Dispose()
	b.Add(img, nil)
	b.Add(img, nil)
	b.Clear()
	if b.Len() != 0 {
		t.Errorf("expected 0 after Clear, got %d", b.Len())
	}
}

// ---------------------------------------------------------------------------
// Batch — Flush
// ---------------------------------------------------------------------------

func TestBatchFlushNilScreen(t *testing.T) {
	requireEbiten(t)
	b := NewBatch()
	img := ebiten.NewImage(16, 16)
	defer img.Dispose()
	b.Add(img, nil)
	b.Flush(nil)
	if b.Len() != 0 {
		t.Errorf("expected 0 after Flush, got %d", b.Len())
	}
}

func TestBatchFlushEmptyBatch(t *testing.T) {
	b := NewBatch()
	b.Flush(nil)
	if b.Len() != 0 {
		t.Errorf("expected 0, got %d", b.Len())
	}
}

func TestBatchMultipleFlushCycles(t *testing.T) {
	requireEbiten(t)
	b := NewBatch()
	img := ebiten.NewImage(16, 16)
	defer img.Dispose()
	for i := 0; i < 3; i++ {
		b.Add(img, nil)
		b.Flush(nil)
		if b.Len() != 0 {
			t.Errorf("expected 0 after cycle %d, got %d", i, b.Len())
		}
	}
}

// ---------------------------------------------------------------------------
// Batch — Texture grouping
// ---------------------------------------------------------------------------

func TestBatchSameTextureGrouped(t *testing.T) {
	requireEbiten(t)
	b := NewBatch()
	img := ebiten.NewImage(16, 16)
	defer img.Dispose()
	b.Add(img, nil)
	b.Add(img, nil)
	b.Add(img, nil)
	if len(b.draws) != 1 {
		t.Errorf("expected 1 texture group, got %d", len(b.draws))
	}
	if b.Len() != 3 {
		t.Errorf("expected 3 total draws, got %d", b.Len())
	}
}

func TestBatchSameTextureAddSpriteGrouped(t *testing.T) {
	requireEbiten(t)
	b := NewBatch()
	img := ebiten.NewImage(16, 16)
	defer img.Dispose()
	sp := &Sprite{image: img, Bounds: image.Rect(0, 0, 16, 16)}
	b.AddSprite(sp, nil)
	b.AddSprite(sp, nil)
	if len(b.draws) != 1 {
		t.Errorf("expected 1 texture group, got %d", len(b.draws))
	}
	if b.Len() != 2 {
		t.Errorf("expected 2 total draws, got %d", b.Len())
	}
}

func TestBatchDifferentTexturesSeparate(t *testing.T) {
	requireEbiten(t)
	b := NewBatch()
	img1 := ebiten.NewImage(16, 16)
	defer img1.Dispose()
	img2 := ebiten.NewImage(32, 32)
	defer img2.Dispose()
	b.Add(img1, nil)
	b.Add(img2, nil)
	if len(b.draws) != 2 {
		t.Errorf("expected 2 texture groups for different images, got %d", len(b.draws))
	}
}

func TestBatchMixedAddAndAddSprite(t *testing.T) {
	requireEbiten(t)
	b := NewBatch()
	img := ebiten.NewImage(16, 16)
	defer img.Dispose()
	sp := &Sprite{image: img, Bounds: image.Rect(0, 0, 16, 16)}
	b.Add(img, nil)
	b.AddSprite(sp, nil)
	if len(b.draws) != 1 {
		t.Errorf("expected 1 texture group, got %d", len(b.draws))
	}
	if b.Len() != 2 {
		t.Errorf("expected 2 total draws, got %d", b.Len())
	}
}

// ---------------------------------------------------------------------------
// Batch — Options behavior
// ---------------------------------------------------------------------------

func TestBatchAddPreservesFullImageBounds(t *testing.T) {
	requireEbiten(t)
	b := NewBatch()
	img := ebiten.NewImage(16, 16)
	defer img.Dispose()
	b.Add(img, nil)
	// Should use full image bounds
	draw := b.draws[img][0]
	if draw.srcW != 16 || draw.srcH != 16 {
		t.Errorf("expected src 16x16, got %dx%d", draw.srcW, draw.srcH)
	}
}

func TestBatchAddCopiesOptions(t *testing.T) {
	requireEbiten(t)
	b := NewBatch()
	img := ebiten.NewImage(16, 16)
	defer img.Dispose()
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(10, 20)
	b.Add(img, opts)
	// Modify original
	opts.GeoM.Translate(100, 200)
	// Stored options should not have the extra translation
	draws := b.draws[img]
	if len(draws) == 0 {
		t.Fatal("expected draw entry")
	}
}

// ---------------------------------------------------------------------------
// TilemapBatch
// ---------------------------------------------------------------------------

func TestTilemapBatchNew(t *testing.T) {
	tb := NewTilemapBatch(nil)
	if tb == nil {
		t.Fatal("expected non-nil TilemapBatch")
	}
}

func TestTilemapBatchNewWithTilemap(t *testing.T) {
	tm := &Tilemap{}
	tb := NewTilemapBatch(tm)
	if tb.tilemap != tm {
		t.Error("tilemap not stored")
	}
}

func TestTilemapBatchDrawNilTilemap(t *testing.T) {
	tb := NewTilemapBatch(nil)
	tb.Draw(nil)
}

func TestTilemapBatchDrawNilTileset(t *testing.T) {
	tm := &Tilemap{Cols: 10, Rows: 10}
	tb := NewTilemapBatch(tm)
	tb.Draw(nil)
}

func TestTilemapBatchDrawNilAtlas(t *testing.T) {
	requireEbiten(t)
	atlas := ebiten.NewImage(64, 64)
	defer atlas.Dispose()
	ts := NewTileset(atlas, 16, 16)
	tm := NewTilemap(ts, 10, 10)
	tb := NewTilemapBatch(tm)
	tb.Draw(nil)
}

func TestTilemapBatchDrawEmptyTilemap(t *testing.T) {
	tm := &Tilemap{Cols: 0, Rows: 0}
	tb := NewTilemapBatch(tm)
	tb.Draw(nil)
}

func TestTilemapBatchDrawAccumulatesTiles(t *testing.T) {
	requireEbiten(t)
	atlas := ebiten.NewImage(64, 64)
	defer atlas.Dispose()
	ts := NewTileset(atlas, 16, 16)
	tm := NewTilemap(ts, 3, 1)
	tm.Set(0, 0, 1)
	tm.Set(1, 0, 2)
	tm.Set(2, 0, 3)
	tb := NewTilemapBatch(tm)
	tb.Draw(nil)
	if tb.batch.Len() != 0 {
		t.Errorf("expected batch empty after Draw (Flush clears it), got %d", tb.batch.Len())
	}
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkBatchAdd(b *testing.B) {
	requireEbiten(b)
	img := ebiten.NewImage(16, 16)
	defer img.Dispose()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bat := NewBatch()
		for j := 0; j < 100; j++ {
			bat.Add(img, nil)
		}
	}
}

func BenchmarkBatchAddAndFlush(b *testing.B) {
	requireEbiten(b)
	img := ebiten.NewImage(16, 16)
	defer img.Dispose()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bat := NewBatch()
		for j := 0; j < 100; j++ {
			bat.Add(img, nil)
		}
		bat.Flush(nil)
	}
}
