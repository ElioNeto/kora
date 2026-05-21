package render

import (
	"testing"
)

func TestTileCollisionMapIsSolid(t *testing.T) {
	cm := TileCollisionMap{
		0: TileEmpty,
		1: TileSolid,
		2: TilePlatform,
	}

	if cm.IsSolid(0) {
		t.Error("tile 0 (empty) should not be solid")
	}
	if !cm.IsSolid(1) {
		t.Error("tile 1 (solid) should be solid")
	}
	if !cm.IsSolid(99) {
		t.Error("unknown tile should default to solid if >= 0")
	}
}

func TestTileCollisionMapIsPlatform(t *testing.T) {
	cm := TileCollisionMap{
		1: TileSolid,
		2: TilePlatform,
	}

	if cm.IsPlatform(1) {
		t.Error("solid tile should not be platform")
	}
	if !cm.IsPlatform(2) {
		t.Error("platform tile should be platform")
	}
	if cm.IsPlatform(99) {
		t.Error("unknown tile should not be platform")
	}
}

func TestNewTileLayer(t *testing.T) {
	layer := NewTileLayer("overlay", 10, 5)
	if layer.Name != "overlay" {
		t.Errorf("expected name 'overlay', got '%s'", layer.Name)
	}
	if layer.Cols != 10 || layer.Rows != 5 {
		t.Errorf("expected (10, 5), got (%d, %d)", layer.Cols, layer.Rows)
	}
	if !layer.Visible {
		t.Error("expected layer visible by default")
	}

	// Should be filled with -1
	for row := 0; row < 5; row++ {
		for col := 0; col < 10; col++ {
			if layer.Get(col, row) != -1 {
				t.Errorf("expected -1 at (%d, %d), got %d", col, row, layer.Get(col, row))
			}
		}
	}
}

func TestTileLayerSetGet(t *testing.T) {
	layer := NewTileLayer("test", 10, 10)
	layer.Set(3, 4, 42)
	if layer.Get(3, 4) != 42 {
		t.Errorf("expected 42 at (3,4), got %d", layer.Get(3, 4))
	}
	if layer.Get(-1, 0) != -1 {
		t.Error("expected -1 for out-of-bounds col")
	}
	if layer.Get(0, 99) != -1 {
		t.Error("expected -1 for out-of-bounds row")
	}
}

func TestTileLayerSetCollision(t *testing.T) {
	layer := NewTileLayer("test", 10, 10)
	layer.SetCollision(1, TileSolid)
	if !layer.Collision.IsSolid(1) {
		t.Error("expected tile 1 to be solid after setting collision")
	}
}

func TestComputeAutoTileMask(t *testing.T) {
	// Create a 3x3 grid with only the centre tile filled
	grid := map[[2]int]int{
		{1, 1}: 0, // centre is solid
	}

	tilesFn := func(c, r int) int {
		if tid, ok := grid[[2]int{c, r}]; ok {
			return tid
		}
		return -1
	}

	mask := ComputeAutoTileMask(1, 1, tilesFn)
	if mask != 0 {
		t.Errorf("expected mask 0 for isolated tile, got %d", mask)
	}

	// Now with all 8 neighbours solid
	for dr := -1; dr <= 1; dr++ {
		for dc := -1; dc <= 1; dc++ {
			if dr == 0 && dc == 0 {
				continue
			}
			grid[[2]int{1 + dc, 1 + dr}] = 0
		}
	}

	mask = ComputeAutoTileMask(1, 1, tilesFn)
	expectedMask := 0xFF // all 8 bits set
	if mask != expectedMask {
		t.Errorf("expected mask 0xFF for surrounded tile, got 0x%02X", mask)
	}
}

func TestTo4BitMask(t *testing.T) {
	tests := []struct {
		mask8 int
		mask4 int
	}{
		{0x00, 0x00},       // no neighbours
		{0xFF, 0x0F},       // all 8 neighbours -> all 4 cardinals
		{0x02, 0x02},       // only N -> N
		{0x08, 0x01},       // only W -> W
		{0x40, 0x08},       // only S -> S
		{0x10, 0x04},       // only E -> E
	}

	for _, tt := range tests {
		got := To4BitMask(tt.mask8)
		if got != tt.mask4 {
			t.Errorf("To4BitMask(0x%02X) = 0x%02X, want 0x%02X", tt.mask8, got, tt.mask4)
		}
	}
}

func TestAutoTileRuleFromGrid(t *testing.T) {
	// Create a mock tileset
	ts := &Tileset{Cols: 16} // 16 cols

	rule, count := AutoTileRuleFromGrid(ts, 0, 0, 4)
	if count != 16 {
		t.Errorf("expected 16 tiles for 4-bit, got %d", count)
	}

	// First tile should be at column 0
	if rule[0] != 0 {
		t.Errorf("expected rule[0]=0, got %d", rule[0])
	}

	// 16th tile should be at column 15
	if rule[15] != 15 {
		t.Errorf("expected rule[15]=15, got %d", rule[15])
	}
}

func TestNeighbourOffsets(t *testing.T) {
	if len(neighbourOffsets) != 8 {
		t.Errorf("expected 8 neighbour offsets, got %d", len(neighbourOffsets))
	}

	// Verify ordering
	expected := [8][2]int{
		{-1, -1}, {0, -1}, {1, -1}, // top row
		{-1, 0}, {1, 0}, // middle row
		{-1, 1}, {0, 1}, {1, 1}, // bottom row
	}
	for i, off := range expected {
		if neighbourOffsets[i] != off {
			t.Errorf("offset %d: expected (%d,%d), got (%d,%d)",
				i, off[0], off[1], neighbourOffsets[i][0], neighbourOffsets[i][1])
		}
	}
}

func TestNewLayerTilemap(t *testing.T) {
	ts := &Tileset{Cols: 16}
	ltm := NewLayerTilemap(ts, 20, 15)

	if ltm.Tileset != ts {
		t.Error("expected tileset to match")
	}
	if ltm.Base.Cols != 20 || ltm.Base.Rows != 15 {
		t.Errorf("expected base (20, 15), got (%d, %d)", ltm.Base.Cols, ltm.Base.Rows)
	}
	if len(ltm.Layers) != 0 {
		t.Errorf("expected 0 layers, got %d", len(ltm.Layers))
	}
}

func TestAddLayer(t *testing.T) {
	ts := &Tileset{Cols: 16}
	ltm := NewLayerTilemap(ts, 10, 10)
	layer := NewTileLayer("decor", 10, 10)
	ltm.AddLayer(layer)

	if len(ltm.Layers) != 1 {
		t.Errorf("expected 1 layer, got %d", len(ltm.Layers))
	}
	if ltm.Layers[0] != layer {
		t.Error("expected added layer to match")
	}
}

func TestIsSolidMultiLayer(t *testing.T) {
	ts := &Tileset{Cols: 16}
	ltm := NewLayerTilemap(ts, 10, 10)

	// Base has no tiles -> not solid
	if ltm.IsSolid(0, 0) {
		t.Error("expected empty base to not be solid")
	}

	// Add a tile to base
	ltm.Base.Set(0, 0, 1)
	if !ltm.IsSolid(0, 0) {
		t.Error("expected base tile to be solid")
	}

	// Add overlay with collision
	layer := NewTileLayer("decor", 10, 10)
	layer.SetCollision(5, TileEmpty)
	layer.Set(1, 1, 5)
	ltm.AddLayer(layer)

	if ltm.IsSolid(1, 1) {
		t.Error("expected TileEmpty overlay to not be solid")
	}
}

func TestTileLayerSetCollisionIntegration(t *testing.T) {
	layer := NewTileLayer("test", 10, 10)
	layer.SetCollision(1, TileSolid)
	layer.SetCollision(2, TilePlatform)
	layer.SetCollision(3, TileEmpty)
	layer.SetCollision(4, TileSlope45)

	// Check each collision type
	if !layer.Collision.IsSolid(1) {
		t.Error("TileSolid should be IsSolid")
	}
	if !layer.Collision.IsSolid(2) {
		t.Error("TilePlatform should be IsSolid")
	}
	if layer.Collision.IsSolid(3) {
		t.Error("TileEmpty should not be IsSolid")
	}
	if !layer.Collision.IsPlatform(2) {
		t.Error("TilePlatform should be IsPlatform")
	}
	if layer.Collision.IsPlatform(1) {
		t.Error("TileSolid should not be IsPlatform")
	}
}
