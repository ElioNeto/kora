package render

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// ----------------------------------------------------------------------------
// Batch
// ----------------------------------------------------------------------------

// Batch accumulates draw calls and flushes them grouped by texture.
// This reduces state changes and improves GPU performance.
type Batch struct {
	draws map[*ebiten.Image][]batchDraw
	count int
}

// batchDraw represents a single draw call within a batch group.
// The source rectangle (srcX, srcY, srcW, srcH) defines the sub-region of the
// texture image to draw.
type batchDraw struct {
	opts            ebiten.DrawImageOptions
	srcX, srcY, srcW, srcH int
}

// NewBatch creates an empty Batch.
func NewBatch() *Batch {
	return &Batch{
		draws: make(map[*ebiten.Image][]batchDraw),
	}
}

// Add adds a draw call to the batch using the full image as the source.
// If img is nil, nothing is added.
// If the same texture is used multiple times, they are grouped together.
func (b *Batch) Add(img *ebiten.Image, opts *ebiten.DrawImageOptions) {
	if img == nil {
		return
	}
	bounds := img.Bounds()
	b.add(img, bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy(), opts)
}

// AddSprite adds a sprite draw to the batch with the given options.
// The sprite's source rectangle is used to determine which sub-region of the
// texture image to draw. If sp is nil or sp.image is nil, nothing is added.
func (b *Batch) AddSprite(sp *Sprite, opts *ebiten.DrawImageOptions) {
	if sp == nil || sp.image == nil {
		return
	}
	b.add(sp.image, sp.Bounds.Min.X, sp.Bounds.Min.Y, sp.Bounds.Dx(), sp.Bounds.Dy(), opts)
}

// add is an internal helper that appends a draw call to the batch.
func (b *Batch) add(img *ebiten.Image, srcX, srcY, srcW, srcH int, opts *ebiten.DrawImageOptions) {
	d := batchDraw{
		srcX: srcX,
		srcY: srcY,
		srcW: srcW,
		srcH: srcH,
	}
	if opts != nil {
		d.opts = *opts
	}
	b.draws[img] = append(b.draws[img], d)
	b.count++
}

// Flush draws all accumulated draws to the screen, grouped by texture.
// After flushing, the batch is cleared.
func (b *Batch) Flush(screen *ebiten.Image) {
	if screen == nil || b.count == 0 {
		b.Clear()
		return
	}
	for img, draws := range b.draws {
		for _, d := range draws {
			rect := image.Rect(d.srcX, d.srcY, d.srcX+d.srcW, d.srcY+d.srcH)
			sub := img.SubImage(rect).(*ebiten.Image)
			screen.DrawImage(sub, &d.opts)
		}
	}
	b.Clear()
}

// Clear removes all pending draws without drawing.
func (b *Batch) Clear() {
	b.draws = make(map[*ebiten.Image][]batchDraw)
	b.count = 0
}

// Len returns the number of pending draws.
func (b *Batch) Len() int {
	return b.count
}

// ----------------------------------------------------------------------------
// TilemapBatch
// ----------------------------------------------------------------------------

// TilemapBatch renders a Tilemap using batched draw calls.
// It groups tiles by tileset texture for efficient rendering.
type TilemapBatch struct {
	tilemap *Tilemap
	batch   *Batch
}

// NewTilemapBatch creates a TilemapBatch that renders the given Tilemap.
func NewTilemapBatch(tm *Tilemap) *TilemapBatch {
	return &TilemapBatch{
		tilemap: tm,
		batch:   NewBatch(),
	}
}

// Draw renders all visible tiles using the batch. After drawing,
// flushes the batch to the screen.
func (tb *TilemapBatch) Draw(screen *ebiten.Image) {
	if tb.tilemap == nil || tb.tilemap.Tileset == nil || tb.tilemap.Tileset.atlas == nil {
		return
	}
	tm := tb.tilemap
	tw := float64(tm.Tileset.TileW)
	th := float64(tm.Tileset.TileH)
	atlas := tm.Tileset.atlas
	for row := 0; row < tm.Rows; row++ {
		for col := 0; col < tm.Cols; col++ {
			tid := tm.Data[row][col]
			if tid < 0 {
				continue
			}
			rect := tm.Tileset.TileRect(tid)
			wx := tm.OffsetX + float64(col)*tw
			wy := tm.OffsetY + float64(row)*th
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(wx, wy)
			tb.batch.add(atlas, rect.Min.X, rect.Min.Y, rect.Dx(), rect.Dy(), op)
		}
	}
	tb.batch.Flush(screen)
}
