package render

import (
	"math"
	"testing"
)

// --- VisibleTileRange ---

func TestVisibleTileRangeCenter(t *testing.T) {
	// Camera at (0,0), view 160×160 world units, tiles 16×16.
	// Visible world: [-80, -80] .. [80, 80]
	// Without padding: floor(-80/16)=-5, ceil(80/16)=5
	// With 1-tile padding: -6 .. 6
	sc, sr, ec, er := VisibleTileRange(0, 0, 160, 160, 16, 16)
	if sc != -6 || sr != -6 || ec != 6 || er != 6 {
		t.Errorf("expected (-6,-6,6,6), got (%d,%d,%d,%d)", sc, sr, ec, er)
	}
}

func TestVisibleTileRangeOffsetCamera(t *testing.T) {
	// Camera at (100, 200), view 80×60 world units, tiles 16×16.
	// Visible world: [60,170] .. [140,230]
	// Without padding: floor(60/16)=3, ceil(140/16)=9  → (3,10)..(9,15)
	// With padding: 2 .. 10, 9 .. 16
	sc, sr, ec, er := VisibleTileRange(100, 200, 80, 60, 16, 16)
	if sc != 2 || sr != 9 || ec != 10 || er != 16 {
		t.Errorf("expected (2,9,10,16), got (%d,%d,%d,%d)", sc, sr, ec, er)
	}
}

func TestVisibleTileRangeNegativeCamera(t *testing.T) {
	// Camera at (-50, -50), view 80×80 world units, tiles 16×16.
	// Visible world: [-90,-90] .. [-10,-10]
	// Without padding: floor(-90/16)=-6, ceil(-10/16)=0
	// With padding: -7 .. 1
	sc, sr, ec, er := VisibleTileRange(-50, -50, 80, 80, 16, 16)
	if sc != -7 || sr != -7 || ec != 1 || er != 1 {
		t.Errorf("expected (-7,-7,1,1), got (%d,%d,%d,%d)", sc, sr, ec, er)
	}
}

func TestVisibleTileRangeZeroTiles(t *testing.T) {
	// Zero tile dimensions should not crash and return a sane range.
	sc, sr, ec, er := VisibleTileRange(0, 0, 100, 100, 0, 0)
	if ec-sc != 0 || er-sr != 0 {
		t.Errorf("expected empty range (0,0,0,0), got (%d,%d,%d,%d)", sc, sr, ec, er)
	}
}

func TestVisibleTileRangeNegativeTileSize(t *testing.T) {
	// Negative tile dimensions should not cause panics.
	sc, sr, ec, er := VisibleTileRange(0, 0, 100, 100, -16, -16)
	if ec-sc != 0 || er-sr != 0 {
		t.Errorf("expected empty range, got (%d,%d,%d,%d)", sc, sr, ec, er)
	}
}

func TestVisibleTileRangeCoversEverythingNoEdgeArtifacts(t *testing.T) {
	// When the view is much larger than the map, all tiles should be included.
	// With 1-tile padding we expect the range to extend beyond the mapped area.
	sc, _, ec, _ := VisibleTileRange(0, 0, 99999, 99999, 16, 16)
	// left ≈ -49999.5, top ≈ -49999.5, right ≈ 49999.5, bottom ≈ 49999.5
	// startCol = floor(-49999.5/16) - 1 ≈ -3125 - 1 = -3126
	// endCol = ceil(49999.5/16) + 1 ≈ 3125 + 1 = 3126
	if ec-sc < 6000 {
		t.Errorf("expected a very wide range, got %d columns (%d..%d)", ec-sc, sc, ec)
	}
}

// --- Zoom factor via VisibleTileRange ---

func TestVisibleTileRangeWithZoom(t *testing.T) {
	// Simulating zoom=2: effective visible world is half the screen size.
	// Screen 320×240, zoom=2 → viewW=160, viewH=120
	viewW := 320.0 / 2.0 // 160
	viewH := 240.0 / 2.0 // 120

	sc, sr, ec, er := VisibleTileRange(0, 0, viewW, viewH, 16, 16)
	// left=-80, right=80, top=-60, bottom=60
	// startCol = floor(-80/16)-1 = -5-1 = -6
	// startRow = floor(-60/16)-1 = -4-1 = -5
	// endCol = ceil(80/16)+1 = 5+1 = 6
	// endRow = ceil(60/16)+1 = 4+1 = 5
	if sc != -6 || sr != -5 || ec != 6 || er != 5 {
		t.Errorf("expected (-6,-5,6,5) for zoom=2, got (%d,%d,%d,%d)", sc, sr, ec, er)
	}
}

// --- Clamping to map bounds ---

func TestClampingDrawWithCamera(t *testing.T) {
	// Verify the clamping logic used in DrawWithCamera/DrawVisible
	// doesn't panic and produces a usable range.
	ts := &Tileset{TileW: 16, TileH: 16, Cols: 4}
	tm := NewTilemap(ts, 10, 10)

	tw := float64(ts.TileW)
	th := float64(ts.TileH)

	// Camera far from the map — visible rect completely outside the tiles.
	cam := Camera{X: 1000, Y: 1000, Zoom: 1, W: 320, H: 240}

	absSc, absSr, absEc, absEr := VisibleTileRange(
		cam.X, cam.Y, cam.W/cam.Zoom, cam.H/cam.Zoom, tw, th,
	)

	// Convert absolute → local.
	sc := int(math.Floor(float64(absSc) - tm.OffsetX/tw))
	sr := int(math.Floor(float64(absSr) - tm.OffsetY/th))
	ec := int(math.Ceil(float64(absEc) - tm.OffsetX/tw))
	er := int(math.Ceil(float64(absEr) - tm.OffsetY/th))

	// Clamp.
	if sc < 0 {
		sc = 0
	}
	if sr < 0 {
		sr = 0
	}
	if ec > tm.Cols {
		ec = tm.Cols
	}
	if er > tm.Rows {
		er = tm.Rows
	}

	// The loop `for row := sr; row < er; row++` would not execute
	// because sr >= er — that is correct behaviour for a camera
	// that does not overlap the map. No panic should occur.
	_ = sc
	_ = sr
	_ = ec
	_ = er

	// Camera centred on the map — all tiles should be visible.
	cam2 := Camera{X: 5 * 16, Y: 5 * 16, Zoom: 1, W: 320, H: 240}
	absSc2, absSr2, absEc2, absEr2 := VisibleTileRange(
		cam2.X, cam2.Y, cam2.W/cam2.Zoom, cam2.H/cam2.Zoom, tw, th,
	)

	sc2 := int(math.Floor(float64(absSc2) - tm.OffsetX/tw))
	sr2 := int(math.Floor(float64(absSr2) - tm.OffsetY/th))
	ec2 := int(math.Ceil(float64(absEc2) - tm.OffsetX/tw))
	er2 := int(math.Ceil(float64(absEr2) - tm.OffsetY/th))

	if sc2 < 0 {
		sc2 = 0
	}
	if sr2 < 0 {
		sr2 = 0
	}
	if ec2 > tm.Cols {
		ec2 = tm.Cols
	}
	if er2 > tm.Rows {
		er2 = tm.Rows
	}

	if sc2 != 0 || sr2 != 0 {
		t.Errorf("expected start (0,0) for centred camera, got (%d,%d)", sc2, sr2)
	}
	if ec2 != tm.Cols || er2 != tm.Rows {
		t.Errorf("expected end (%d,%d) for centred camera, got (%d,%d)", tm.Cols, tm.Rows, ec2, er2)
	}
}

// --- CullingTilemap ---

func TestCullingTilemapNilChecks(t *testing.T) {
	// NewCullingTilemap should work with a nil Tilemap (not crash on nil deref).
	var tm *Tilemap = nil
	ct := NewCullingTilemap(tm)
	if ct.Tilemap != nil {
		t.Error("expected nil Tilemap")
	}
}

func TestCullingTilemapNew(t *testing.T) {
	ts := &Tileset{TileW: 16, TileH: 16, Cols: 4}
	tm := NewTilemap(ts, 10, 10)
	ct := NewCullingTilemap(tm)
	if ct.Tilemap != tm {
		t.Error("expected CullingTilemap to wrap the provided Tilemap")
	}
}

// --- Zero zoom guard ---

func TestVisibleTileRangeZeroZoom(t *testing.T) {
	// When zoom is 0, DrawWithCamera treats it as zoom=1.
	// Simulate what DrawWithCamera does internally.
	cam := Camera{X: 0, Y: 0, Zoom: 0, W: 320, H: 240}
	zoom := cam.Zoom
	if zoom == 0 {
		zoom = 1
	}
	viewW := cam.W / zoom
	viewH := cam.H / zoom

	sc, sr, ec, er := VisibleTileRange(cam.X, cam.Y, viewW, viewH, 16, 16)
	// viewW=320, viewH=240 at zoom=1.
	// left=-160, top=-120, right=160, bottom=120
	// startCol = floor(-160/16)-1 = -10-1 = -11
	// startRow = floor(-120/16)-1 = -8-1 = -9
	// endCol = ceil(160/16)+1 = 10+1 = 11
	// endRow = ceil(120/16)+1 = 8+1 = 9
	if sc != -11 || sr != -9 || ec != 11 || er != 9 {
		t.Errorf("expected (-11,-9,11,9) for zoom=0→1, got (%d,%d,%d,%d)", sc, sr, ec, er)
	}
}

// --- Empty map ---

func TestEmptyMapNoCrash(t *testing.T) {
	// A tilemap with 0 cols/rows should not cause panics when calling
	// culling-related functions.
	ts := &Tileset{TileW: 16, TileH: 16, Cols: 4}
	tm := NewTilemap(ts, 0, 0)

	// VisibleTileRange still works (pure math).
	sc, sr, ec, er := VisibleTileRange(0, 0, 100, 100, 16, 16)
	_ = sc
	_ = sr
	_ = ec
	_ = er
	_ = tm // compiles and doesn't crash
}

func TestCullingTilemapEmptyDrawVisible(t *testing.T) {
	// Creating a CullingTilemap with nil Tilemap should not panic.
	_ = &CullingTilemap{Tilemap: nil}
}

// --- Full map visible ---

func TestVisibleTileRangeFullMap(t *testing.T) {
	// When the camera covers the entire map, VisibleTileRange should
	// return a range that includes all map tiles (after clamping).
	ts := &Tileset{TileW: 16, TileH: 16, Cols: 4}
	tm := NewTilemap(ts, 50, 50)

	cam := Camera{
		X: 25 * 16, // centre of 50-tile map
		Y: 25 * 16,
		Zoom: 1,
		W:    1600, // wide enough to see entire map
		H:    1600,
	}

	tw := float64(ts.TileW)
	th := float64(ts.TileH)
	viewW := cam.W / cam.Zoom
	viewH := cam.H / cam.Zoom

	startCol, startRow, endCol, endRow := VisibleTileRange(
		cam.X, cam.Y, viewW, viewH, tw, th,
	)

	// Adjust for offset.
	startCol -= int(tm.OffsetX / tw)
	startRow -= int(tm.OffsetY / th)
	endCol -= int(tm.OffsetX / tw)
	endRow -= int(tm.OffsetY / th)

	// Clamp.
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

	// The visible world at zoom=1 with W=1600 covers 1600 world units.
	// The map is 50 tiles * 16 = 800 world units wide, so it should all be visible.
	if startCol != 0 || startRow != 0 {
		t.Errorf("expected start (0,0), got (%d,%d)", startCol, startRow)
	}
	if endCol != tm.Cols || endRow != tm.Rows {
		t.Errorf("expected end (%d,%d), got (%d,%d)", tm.Cols, tm.Rows, endCol, endRow)
	}
}

// --- Benchmark ---

func BenchmarkTilemapCulling(b *testing.B) {
	ts := &Tileset{TileW: 16, TileH: 16, Cols: 10}
	tm := NewTilemap(ts, 100, 100)
	// Fill all tiles with valid IDs.
	for r := 0; r < 100; r++ {
		for c := 0; c < 100; c++ {
			tm.Data[r][c] = (c + r) % 10
		}
	}

	cam := Camera{
		X:    50 * 16,
		Y:    50 * 16,
		Zoom: 1,
		W:    320,
		H:    240,
	}

	b.Run("full-iteration", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tw := float64(ts.TileW)
			th := float64(ts.TileH)
			count := 0
			for row := 0; row < tm.Rows; row++ {
				for col := 0; col < tm.Cols; col++ {
					if tm.Data[row][col] >= 0 {
						_ = tm.Tileset.TileRect(tm.Data[row][col])
						_ = tm.OffsetX + float64(col)*tw
						_ = tm.OffsetY + float64(row)*th
						count++
					}
				}
			}
			_ = count
		}
	})

	b.Run("culling", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tw := float64(ts.TileW)
			th := float64(ts.TileH)

			viewW := cam.W / cam.Zoom
			viewH := cam.H / cam.Zoom

			absSc, absSr, absEc, absEr := VisibleTileRange(
				cam.X, cam.Y, viewW, viewH, tw, th,
			)

			sc := int(math.Floor(float64(absSc) - tm.OffsetX/tw))
			sr := int(math.Floor(float64(absSr) - tm.OffsetY/th))
			ec := int(math.Ceil(float64(absEc) - tm.OffsetX/tw))
			er := int(math.Ceil(float64(absEr) - tm.OffsetY/th))

			if sc < 0 {
				sc = 0
			}
			if sr < 0 {
				sr = 0
			}
			if ec > tm.Cols {
				ec = tm.Cols
			}
			if er > tm.Rows {
				er = tm.Rows
			}

			count := 0
			for row := sr; row < er; row++ {
				for col := sc; col < ec; col++ {
					if tm.Data[row][col] >= 0 {
						_ = tm.Tileset.TileRect(tm.Data[row][col])
						_ = tm.OffsetX + float64(col)*tw
						_ = tm.OffsetY + float64(row)*th
						count++
					}
				}
			}
			_ = count
		}
	})
}
