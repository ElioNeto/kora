package render

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// ----------------------------------------------------------------------------
// Tileset
// ----------------------------------------------------------------------------

// Tileset maps tile IDs (0-based) to sub-regions of an atlas.
type Tileset struct {
	atlas  *ebiten.Image
	TileW  int
	TileH  int
	Cols   int
}

// NewTileset creates a Tileset from an atlas image with uniform tile size.
func NewTileset(atlas *ebiten.Image, tileW, tileH int) *Tileset {
	return &Tileset{
		atlas: atlas,
		TileW: tileW,
		TileH: tileH,
		Cols:  atlas.Bounds().Dx() / tileW,
	}
}

// TileRect returns the source rectangle for tileID.
func (ts *Tileset) TileRect(tileID int) image.Rectangle {
	col := tileID % ts.Cols
	row := tileID / ts.Cols
	return image.Rect(
		col*ts.TileW, row*ts.TileH,
		(col+1)*ts.TileW, (row+1)*ts.TileH,
	)
}

// ----------------------------------------------------------------------------
// Tilemap
// ----------------------------------------------------------------------------

// Tilemap is a 2D grid of tile IDs rendered in a single batched pass.
// Tile ID -1 means empty (not drawn).
type Tilemap struct {
	Tileset *Tileset
	Data    [][]int // [row][col] tile IDs
	Cols    int
	Rows    int
	OffsetX float64 // world-space top-left X
	OffsetY float64 // world-space top-left Y
}

// NewTilemap creates a Tilemap filled with -1 (empty).
func NewTilemap(ts *Tileset, cols, rows int) *Tilemap {
	data := make([][]int, rows)
	for i := range data {
		row := make([]int, cols)
		for j := range row {
			row[j] = -1
		}
		data[i] = row
	}
	return &Tilemap{Tileset: ts, Data: data, Cols: cols, Rows: rows}
}

// Set writes tileID at (col, row).
func (tm *Tilemap) Set(col, row, tileID int) {
	if row < 0 || row >= tm.Rows || col < 0 || col >= tm.Cols {
		return
	}
	tm.Data[row][col] = tileID
}

// Get returns the tileID at (col, row), or -1 if out of bounds.
func (tm *Tilemap) Get(col, row int) int {
	if row < 0 || row >= tm.Rows || col < 0 || col >= tm.Cols {
		return -1
	}
	return tm.Data[row][col]
}

// TileAtWorld returns the tile at a world-space position.
func (tm *Tilemap) TileAtWorld(wx, wy float64) int {
	col := int((wx - tm.OffsetX) / float64(tm.Tileset.TileW))
	row := int((wy - tm.OffsetY) / float64(tm.Tileset.TileH))
	return tm.Get(col, row)
}

// Draw renders all non-empty tiles onto r.
func (tm *Tilemap) Draw(r *Renderer) {
	tw := float64(tm.Tileset.TileW)
	th := float64(tm.Tileset.TileH)
	for row := 0; row < tm.Rows; row++ {
		for col := 0; col < tm.Cols; col++ {
			tid := tm.Data[row][col]
			if tid < 0 {
				continue
			}
			rect := tm.Tileset.TileRect(tid)
			wx := tm.OffsetX + float64(col)*tw
			wy := tm.OffsetY + float64(row)*th
			sx, sy := r.Camera.WorldToScreen(wx, wy)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(r.Camera.Zoom, r.Camera.Zoom)
			op.GeoM.Translate(sx, sy)
			r.screen.DrawImage(
				tm.Tileset.atlas.SubImage(rect).(*ebiten.Image),
				op,
			)
		}
	}
}
