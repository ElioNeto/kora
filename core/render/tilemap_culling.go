package render

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// ---------------------------------------------------------------------------
// VisibleTileRange
// ---------------------------------------------------------------------------

// VisibleTileRange computes the range of visible tiles for a given camera view.
// camX, camY is the camera centre in world coordinates.
// viewW, viewH is the visible world dimensions (screen size / zoom).
// tileW, tileH is the tile size in world units.
// Returns (startCol, startRow, endCol, endRow) with 1-tile padding.
func VisibleTileRange(camX, camY float64, viewW, viewH float64, tileW, tileH float64) (int, int, int, int) {
	if tileW <= 0 || tileH <= 0 {
		return 0, 0, 0, 0
	}

	left := camX - viewW/2
	top := camY - viewH/2
	right := camX + viewW/2
	bottom := camY + viewH/2

	// 1-tile padding to prevent edge artifacts
	startCol := int(math.Floor(left/tileW)) - 1
	startRow := int(math.Floor(top/tileH)) - 1
	endCol := int(math.Ceil(right/tileW)) + 1
	endRow := int(math.Ceil(bottom/tileH)) + 1

	return startCol, startRow, endCol, endRow
}

// ---------------------------------------------------------------------------
// CullingTilemap
// ---------------------------------------------------------------------------

// CullingTilemap wraps a Tilemap and adds viewport-based culling.
type CullingTilemap struct {
	*Tilemap
}

// NewCullingTilemap creates a CullingTilemap from an existing Tilemap.
func NewCullingTilemap(tm *Tilemap) *CullingTilemap {
	return &CullingTilemap{Tilemap: tm}
}

// DrawVisible renders only the tiles visible in the given viewport.
// camX, camY is the camera world position; viewW, viewH is the viewport size
// in world units (already divided by zoom, if any).
func (ct *CullingTilemap) DrawVisible(screen *ebiten.Image, camX, camY, viewW, viewH float64) {
	tm := ct.Tilemap
	if tm == nil || tm.Tileset == nil || tm.Tileset.atlas == nil || screen == nil {
		return
	}

	tw := float64(tm.Tileset.TileW)
	th := float64(tm.Tileset.TileH)

	absStartCol, absStartRow, absEndCol, absEndRow := VisibleTileRange(
		camX, camY, viewW, viewH, tw, th,
	)

	// Convert absolute tile coords → local tilemap indices.
	startCol := int(math.Floor(float64(absStartCol) - tm.OffsetX/tw))
	startRow := int(math.Floor(float64(absStartRow) - tm.OffsetY/th))
	endCol := int(math.Ceil(float64(absEndCol) - tm.OffsetX/tw))
	endRow := int(math.Ceil(float64(absEndRow) - tm.OffsetY/th))

	// Clamp to map bounds.
	if startCol < 0 {
		startCol = 0
	}
	if startRow < 0 {
		startRow = 0
	}
	if endCol > tm.Cols {
		endCol = tm.Cols
	}
	if endRow > tm.Rows {
		endRow = tm.Rows
	}

	// Draw visible tiles.
	for row := startRow; row < endRow; row++ {
		for col := startCol; col < endCol; col++ {
			tid := tm.Data[row][col]
			if tid < 0 {
				continue
			}
			rect := tm.Tileset.TileRect(tid)
			wx := tm.OffsetX + float64(col)*tw
			wy := tm.OffsetY + float64(row)*th

			// World → screen (no zoom — caller is expected to pass viewW/viewH
			// already scaled by zoom if needed).
			sx := wx - camX + viewW/2
			sy := wy - camY + viewH/2

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(sx, sy)
			screen.DrawImage(
				tm.Tileset.atlas.SubImage(rect).(*ebiten.Image),
				op,
			)
		}
	}
}

// ---------------------------------------------------------------------------
// Tilemap.DrawWithCamera
// ---------------------------------------------------------------------------

// DrawWithCamera renders only visible tiles using the given camera.
// It computes the visible world rect from the camera's position, zoom
// and screen dimensions, then only iterates over tiles that intersect it.
func (tm *Tilemap) DrawWithCamera(screen *ebiten.Image, camera *Camera) {
	if tm == nil || tm.Tileset == nil || tm.Tileset.atlas == nil || screen == nil || camera == nil {
		return
	}

	tw := float64(tm.Tileset.TileW)
	th := float64(tm.Tileset.TileH)

	// Convert screen pixel dimensions to visible world units via zoom.
	zoom := camera.Zoom
	if zoom == 0 {
		zoom = 1
	}
	viewW := camera.W / zoom
	viewH := camera.H / zoom

	absStartCol, absStartRow, absEndCol, absEndRow := VisibleTileRange(
		camera.X, camera.Y, viewW, viewH, tw, th,
	)

	// Convert absolute tile coords → local tilemap indices.
	startCol := int(math.Floor(float64(absStartCol) - tm.OffsetX/tw))
	startRow := int(math.Floor(float64(absStartRow) - tm.OffsetY/th))
	endCol := int(math.Ceil(float64(absEndCol) - tm.OffsetX/tw))
	endRow := int(math.Ceil(float64(absEndRow) - tm.OffsetY/th))

	// Clamp to map bounds.
	if startCol < 0 {
		startCol = 0
	}
	if startRow < 0 {
		startRow = 0
	}
	if endCol > tm.Cols {
		endCol = tm.Cols
	}
	if endRow > tm.Rows {
		endRow = tm.Rows
	}

	// Draw visible tiles.
	for row := startRow; row < endRow; row++ {
		for col := startCol; col < endCol; col++ {
			tid := tm.Data[row][col]
			if tid < 0 {
				continue
			}
			rect := tm.Tileset.TileRect(tid)
			wx := tm.OffsetX + float64(col)*tw
			wy := tm.OffsetY + float64(row)*th
			sx, sy := camera.WorldToScreen(wx, wy)

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(zoom, zoom)
			op.GeoM.Translate(sx, sy)
			screen.DrawImage(
				tm.Tileset.atlas.SubImage(rect).(*ebiten.Image),
				op,
			)
		}
	}
}
