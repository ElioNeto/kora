package render

// ---------------------------------------------------------------------------
// Collision data
// ---------------------------------------------------------------------------

// TileCollision defines the collision behaviour of a tile ID.
type TileCollision int

const (
	TileEmpty     TileCollision = iota // no tile / passable
	TileSolid                           // fully solid (wall / floor)
	TilePlatform                        // one-way platform (solid from above)
	TileSlope45                         // 45-degree slope placeholder
)

// TileCollisionMap stores collision data per tile ID.
// Indexed by tile ID; returns TileEmpty for unknown tiles.
type TileCollisionMap map[int]TileCollision

// IsSolid returns true if the given tile ID is solid (not passable).
func (cm TileCollisionMap) IsSolid(tileID int) bool {
	c, ok := cm[tileID]
	if !ok {
		return tileID >= 0 // non-empty tiles default to solid
	}
	return c == TileSolid || c == TilePlatform
}

// IsPlatform returns true if the tile is a one-way platform.
func (cm TileCollisionMap) IsPlatform(tileID int) bool {
	c, ok := cm[tileID]
	return ok && c == TilePlatform
}

// ---------------------------------------------------------------------------
// TileLayer — independent overlay with its own data and collision
// ---------------------------------------------------------------------------

// TileLayer is a named overlay on top of the base Tilemap.
// Each layer can have its own Data grid, collision map, and Z-order.
type TileLayer struct {
	Name   string
	Data   [][]int // [row][col], -1 = empty
	Cols   int
	Rows   int
	Collision TileCollisionMap
	Visible   bool
	ZOrder    int // higher = drawn on top
}

// NewTileLayer creates an empty TileLayer filled with -1.
func NewTileLayer(name string, cols, rows int) *TileLayer {
	data := make([][]int, rows)
	for i := range data {
		row := make([]int, cols)
		for j := range row {
			row[j] = -1
		}
		data[i] = row
	}
	return &TileLayer{
		Name:   name,
		Data:   data,
		Cols:   cols,
		Rows:   rows,
		Collision: make(TileCollisionMap),
		Visible:   true,
	}
}

// Set writes a tile ID at (col, row).
func (tl *TileLayer) Set(col, row, tileID int) {
	if row < 0 || row >= tl.Rows || col < 0 || col >= tl.Cols {
		return
	}
	tl.Data[row][col] = tileID
}

// Get returns the tile ID at (col, row), or -1 if out of bounds.
func (tl *TileLayer) Get(col, row int) int {
	if row < 0 || row >= tl.Rows || col < 0 || col >= tl.Cols {
		return -1
	}
	return tl.Data[row][col]
}

// SetCollision sets the collision type for a tile ID in this layer.
func (tl *TileLayer) SetCollision(tileID int, c TileCollision) {
	tl.Collision[tileID] = c
}

// ---------------------------------------------------------------------------
// Auto-tile (bitmask method)
// ---------------------------------------------------------------------------

// AutoTileRule maps a neighbour bitmask to a tile index in the tileset.
// The bitmask encodes which of the 8 neighbours are solid (bits 0-7):
//
//	bit 0: top-left      (NW)   bit 4: bottom-left  (SW)
//	bit 1: top-center    (N)    bit 5: bottom-center (S)
//	bit 2: top-right     (NE)   bit 6: bottom-right  (SE)
//	bit 3: middle-left   (W)    bit 7: middle-right  (E)
type AutoTileRule [256]int

// AutoTileRuleFromGrid generates an AutoTileRule from a rule tileset.
// The tileset must be laid out in the standard auto-tile pattern:
// 16 columns x N rows of auto-tile variants (bitmask 0-15 for 4-bit mode)
// or 47 tiles for 8-bit auto-tile.
//
// For 4-bit auto-tile (default), the layout is:
//
//	Index  0: mask 0 (isolated)
//	Index  1: mask 1 (W)
//	Index  2: mask 2 (N)
//	Index  3: mask 3 (W | N)
//	...
//	Index 15: mask 15 (W | N | E | S)
//
// Returns the generated rule and the number of tiles used.
func AutoTileRuleFromGrid(tileset *Tileset, startCol, startRow int, bitCount int) (AutoTileRule, int) {
	var rule AutoTileRule

	if bitCount <= 0 {
		bitCount = 4 // default 4-bit auto-tile
	}

	// 4-bit uses 16 tiles, 8-bit uses 47 tiles
	tileCount := 16
	if bitCount >= 8 {
		tileCount = 47
	}

	for i := 0; i < tileCount && i < 256; i++ {
		col := startCol + i
		row := startRow
		for col >= tileset.Cols {
			col -= tileset.Cols
			row++
		}
		rule[i] = row*tileset.Cols + col
	}

	return rule, tileCount
}

// neighbourOffsets maps bit positions to (dc, dr) neighbour offsets.
// Order: NW, N, NE, W, E, SW, S, SE
var neighbourOffsets = [8][2]int{
	{-1, -1}, {0, -1}, {1, -1}, // top row
	{-1, 0}, {1, 0}, // middle row
	{-1, 1}, {0, 1}, {1, 1}, // bottom row
}

// ComputeAutoTileMask computes the 8-bit auto-tile mask at (col, row).
// A neighbour is considered "solid" if its tile ID >= 0.
// tilesFunc returns the tile ID at a given (col, row), or -1 if out of bounds.
func ComputeAutoTileMask(col, row int, tilesFunc func(c, r int) int) int {
	mask := 0
	for bit := 0; bit < 8; bit++ {
		nc := col + neighbourOffsets[bit][0]
		nr := row + neighbourOffsets[bit][1]
		if tid := tilesFunc(nc, nr); tid >= 0 {
			mask |= 1 << bit
		}
	}
	return mask
}

// To4BitMask converts an 8-bit auto-tile mask to a 4-bit mask using
// the standard method: the 4 corners propagate inward.
func To4BitMask(mask8 int) int {
	// bits: 0=NW, 1=N, 2=NE, 3=W, 4=E, 5=SW, 6=S, 7=SE
	n := (mask8 >> 1) & 1
	w := (mask8 >> 3) & 1
	e := (mask8 >> 4) & 1
	s := (mask8 >> 6) & 1

	// 4-bit mask: bit 0 = W, bit 1 = N, bit 2 = E, bit 3 = S
	mask4 := 0
	if w != 0 {
		mask4 |= 1 << 0
	}
	if n != 0 {
		mask4 |= 1 << 1
	}
	if e != 0 {
		mask4 |= 1 << 2
	}
	if s != 0 {
		mask4 |= 1 << 3
	}
	return mask4
}

// ApplyAutoTile fills the auto-tile region of the tilemap using the given rule.
// It computes the mask for each cell in [startCol..endCol) x [startRow..endRow)
// and replaces tile IDs using the AutoTileRule mapping.
//
// baseTileID is the tile ID that represents "solid" for neighbour detection.
// If baseTileID < 0, any non-negative tile ID is considered solid.
//
// tilesFunc returns the tile ID at a given (col, row); this can be from any
// data source (base tilemap or a tile layer).
//
// setFunc sets the tile ID at a given (col, row) to the auto-tiled result.
func ApplyAutoTile(
	startCol, startRow, endCol, endRow int,
	baseTileID int,
	tilesFunc func(c, r int) int,
	setFunc func(c, r, tileID int),
	rule AutoTileRule,
) {
	// Helper that checks if a cell is solid according to the base tile.
	isSolid := func(c, r int) bool {
		tid := tilesFunc(c, r)
		if baseTileID >= 0 {
			return tid == baseTileID
		}
		return tid >= 0
	}

	for row := startRow; row < endRow; row++ {
		for col := startCol; col < endCol; col++ {
			if !isSolid(col, row) {
				continue
			}

			mask8 := 0
			for bit := 0; bit < 8; bit++ {
				nc := col + neighbourOffsets[bit][0]
				nr := row + neighbourOffsets[bit][1]
				if isSolid(nc, nr) {
					mask8 |= 1 << bit
				}
			}

			mask4 := To4BitMask(mask8)
			autoTileID := rule[mask4]
			setFunc(col, row, autoTileID)
		}
	}
}

// ---------------------------------------------------------------------------
// LayerTilemap — multi-layer tilemap with auto-tile support
// ---------------------------------------------------------------------------

// LayerTilemap extends the base Tilemap with multiple layers and auto-tile.
type LayerTilemap struct {
	Tileset *Tileset
	Base    *Tilemap      // base (ground) layer
	Layers  []*TileLayer  // additional overlays
	Rule    AutoTileRule  // auto-tile rule for the base layer
}

// NewLayerTilemap creates a multi-layer tilemap.
func NewLayerTilemap(ts *Tileset, cols, rows int) *LayerTilemap {
	return &LayerTilemap{
		Tileset: ts,
		Base:    NewTilemap(ts, cols, rows),
		Layers:  make([]*TileLayer, 0),
	}
}

// AddLayer adds an overlay layer.
func (ltm *LayerTilemap) AddLayer(layer *TileLayer) {
	ltm.Layers = append(ltm.Layers, layer)
}

// AutoTileBase applies auto-tile to the base layer.
func (ltm *LayerTilemap) AutoTileBase(baseTileID int) {
	ApplyAutoTile(
		0, 0, ltm.Base.Cols, ltm.Base.Rows,
		baseTileID,
		func(c, r int) int { return ltm.Base.Get(c, r) },
		func(c, r, tid int) { ltm.Base.Set(c, r, tid) },
		ltm.Rule,
	)
}

// AutoTileLayer applies auto-tile to a named layer.
func (ltm *LayerTilemap) AutoTileLayer(layerName string, baseTileID int) {
	for _, layer := range ltm.Layers {
		if layer.Name == layerName {
			ApplyAutoTile(
				0, 0, layer.Cols, layer.Rows,
				baseTileID,
				func(c, r int) int { return layer.Get(c, r) },
				func(c, r, tid int) { layer.Set(c, r, tid) },
				ltm.Rule,
			)
			return
		}
	}
}

// IsSolid returns true if any layer has a solid tile at (col, row).
func (ltm *LayerTilemap) IsSolid(col, row int) bool {
	// Check base layer
	if ltm.Base.Get(col, row) >= 0 {
		return true
	}
	// Check overlay layers
	for _, layer := range ltm.Layers {
		if !layer.Visible {
			continue
		}
		tid := layer.Get(col, row)
		if tid >= 0 {
			if layer.Collision.IsSolid(tid) {
				return true
			}
		}
	}
	return false
}

// TileAtWorld returns the tile ID at a world-space position, checking
// all layers (topmost visible layer first).
func (ltm *LayerTilemap) TileAtWorld(wx, wy float64) int {
	col := int((wx - ltm.Base.OffsetX) / float64(ltm.Tileset.TileW))
	row := int((wy - ltm.Base.OffsetY) / float64(ltm.Tileset.TileH))

	// Check layers in reverse order (topmost first)
	for i := len(ltm.Layers) - 1; i >= 0; i-- {
		layer := ltm.Layers[i]
		if !layer.Visible {
			continue
		}
		if tid := layer.Get(col, row); tid >= 0 {
			return tid
		}
	}

	return ltm.Base.Get(col, row)
}
