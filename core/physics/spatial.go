package physics

import "math"

// SpatialHash implements a grid-based spatial hash for broad-phase collision culling.
// It partitions 2D space into cells and maps entity IDs to cells based on their AABB.
// This reduces the number of collision pairs from O(n²) to amortised O(n).
type SpatialHash struct {
	cellSize float64
	cells    map[int]map[int][]int // cell (gx,gy) -> entity IDs
	bodies   map[int]*RigidBody    // entity ID -> body
}

// NewSpatialHash creates a new spatial hash with the given cell size.
// cellSize should be roughly 2–3× the average body size for best performance.
func NewSpatialHash(cellSize float64) *SpatialHash {
	return &SpatialHash{
		cellSize: cellSize,
		cells:    make(map[int]map[int][]int),
		bodies:   make(map[int]*RigidBody),
	}
}

// Clear removes all bodies from the hash. Call each frame before re-inserting.
func (sh *SpatialHash) Clear() {
	sh.cells = make(map[int]map[int][]int)
	sh.bodies = make(map[int]*RigidBody)
}

// cellCoords returns the grid cell coordinates for a given world position.
func (sh *SpatialHash) cellCoords(x, y float32) (int, int) {
	gx := int(math.Floor(float64(x) / sh.cellSize))
	gy := int(math.Floor(float64(y) / sh.cellSize))
	return gx, gy
}

// Insert adds a body to all cells its AABB overlaps.
func (sh *SpatialHash) Insert(id int, body *RigidBody) {
	sh.bodies[id] = body

	minX, minY, maxX, maxY := body.AABB()
	gxMin, gyMin := sh.cellCoords(minX, minY)
	gxMax, gyMax := sh.cellCoords(maxX, maxY)

	for gx := gxMin; gx <= gxMax; gx++ {
		if sh.cells[gx] == nil {
			sh.cells[gx] = make(map[int][]int)
		}
		for gy := gyMin; gy <= gyMax; gy++ {
			sh.cells[gx][gy] = append(sh.cells[gx][gy], id)
		}
	}
}

// Remove removes a body from all cells its AABB overlaps.
func (sh *SpatialHash) Remove(id int, body *RigidBody) {
	delete(sh.bodies, id)

	minX, minY, maxX, maxY := body.AABB()
	gxMin, gyMin := sh.cellCoords(minX, minY)
	gxMax, gyMax := sh.cellCoords(maxX, maxY)

	for gx := gxMin; gx <= gxMax; gx++ {
		row, ok := sh.cells[gx]
		if !ok {
			continue
		}
		for gy := gyMin; gy <= gyMax; gy++ {
			cell := row[gy]
			for i, eid := range cell {
				if eid == id {
					// Swap-with-last removal (order does not matter)
					cell[i] = cell[len(cell)-1]
					row[gy] = cell[:len(cell)-1]
					break
				}
			}
		}
	}
}

// GetCandidates returns all body IDs that could potentially collide with
// the given body (same cell or overlapping cells).  The result excludes
// the query body itself and contains no duplicates.
func (sh *SpatialHash) GetCandidates(id int, body *RigidBody) []int {
	minX, minY, maxX, maxY := body.AABB()
	gxMin, gyMin := sh.cellCoords(minX, minY)
	gxMax, gyMax := sh.cellCoords(maxX, maxY)

	seen := make(map[int]bool)
	seen[id] = true // exclude self

	var result []int

	for gx := gxMin; gx <= gxMax; gx++ {
		row, ok := sh.cells[gx]
		if !ok {
			continue
		}
		for gy := gyMin; gy <= gyMax; gy++ {
			cell, ok := row[gy]
			if !ok {
				continue
			}
			for _, eid := range cell {
				if !seen[eid] {
					seen[eid] = true
					result = append(result, eid)
				}
			}
		}
	}

	return result
}

// Update re-inserts a body that has moved (remove + insert).
func (sh *SpatialHash) Update(id int, body *RigidBody) {
	sh.Remove(id, body)
	sh.Insert(id, body)
}
