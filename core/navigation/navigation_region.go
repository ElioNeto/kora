package navigation

import (
	"github.com/ElioNeto/kora/core/math"
	"github.com/ElioNeto/kora/core/node"
)

// NavigationRegion2D defines a navigable area that can be used for pathfinding.
// It embeds *Node2D so it can be added to the scene tree.
type NavigationRegion2D struct {
	*node.Node2D
	pathfinder *Pathfinder
	cellSize   float64 // world units per grid cell (e.g., 32 pixels)
	offsetX    float64 // world offset X for the grid origin
	offsetY    float64 // world offset Y for the grid origin
}

// NewNavigationRegion2D creates a navigation region with the given grid dimensions and cell size.
func NewNavigationRegion2D(name string, width, height int, cellSize float64) *NavigationRegion2D {
	n := node.NewNode2D(name, 0)
	return &NavigationRegion2D{
		Node2D:     n,
		pathfinder: NewPathfinder(width, height),
		cellSize:   cellSize,
	}
}

// BakeFromTileMap generates the walkable grid from a tile collision map.
// The tilemap parameter must implement TileCollisionAt(tileX, tileY int) bool,
// returning true when a tile at grid coordinates is blocked.
func (nr *NavigationRegion2D) BakeFromTileMap(tilemap interface {
	TileCollisionAt(tileX, tileY int) bool
}) {
	for y := 0; y < nr.pathfinder.height; y++ {
		for x := 0; x < nr.pathfinder.width; x++ {
			nr.pathfinder.SetWalkable(x, y, !tilemap.TileCollisionAt(x, y))
		}
	}
}

// GetPath finds a path from one world position to another.
// Returns world-space points as Vector2 (excluding start, including end).
// Returns nil if no path exists.
func (nr *NavigationRegion2D) GetPath(fromX, fromY, toX, toY float64) []math.Vector2 {
	sx, sy := nr.WorldToGrid(fromX, fromY)
	ex, ey := nr.WorldToGrid(toX, toY)

	points := nr.pathfinder.FindPath(sx, sy, ex, ey)
	if points == nil {
		return nil
	}

	path := make([]math.Vector2, len(points))
	for i, p := range points {
		wx, wy := nr.GridToWorld(p.X, p.Y)
		path[i] = math.NewVector2(float32(wx), float32(wy))
	}
	return path
}

// GetPathSmooth finds a smoothed path from one world position to another.
func (nr *NavigationRegion2D) GetPathSmooth(fromX, fromY, toX, toY float64) []math.Vector2 {
	sx, sy := nr.WorldToGrid(fromX, fromY)
	ex, ey := nr.WorldToGrid(toX, toY)

	points := nr.pathfinder.FindPathSmooth(sx, sy, ex, ey)
	if points == nil {
		return nil
	}

	path := make([]math.Vector2, len(points))
	for i, p := range points {
		wx, wy := nr.GridToWorld(p.X, p.Y)
		path[i] = math.NewVector2(float32(wx), float32(wy))
	}
	return path
}

// WorldToGrid converts world coordinates to grid coordinates.
func (nr *NavigationRegion2D) WorldToGrid(wx, wy float64) (int, int) {
	gx := int((wx - nr.offsetX) / nr.cellSize)
	gy := int((wy - nr.offsetY) / nr.cellSize)
	return gx, gy
}

// GridToWorld converts grid coordinates to the center of the cell in world coordinates.
func (nr *NavigationRegion2D) GridToWorld(gx, gy int) (float64, float64) {
	wx := nr.offsetX + float64(gx)*nr.cellSize + nr.cellSize/2
	wy := nr.offsetY + float64(gy)*nr.cellSize + nr.cellSize/2
	return wx, wy
}

// SetObstacle marks a world position as blocked (dynamic obstacle) or unblocked.
func (nr *NavigationRegion2D) SetObstacle(wx, wy float64, blocked bool) {
	gx, gy := nr.WorldToGrid(wx, wy)
	nr.pathfinder.SetWalkable(gx, gy, !blocked)
}

// SetOffset sets the world offset for the grid origin.
func (nr *NavigationRegion2D) SetOffset(offsetX, offsetY float64) {
	nr.offsetX = offsetX
	nr.offsetY = offsetY
}

// GetCellSize returns the cell size in world units.
func (nr *NavigationRegion2D) GetCellSize() float64 {
	return nr.cellSize
}

// GetPathfinder returns the internal Pathfinder for direct grid manipulation.
func (nr *NavigationRegion2D) GetPathfinder() *Pathfinder {
	return nr.pathfinder
}

// Update satisfies the Node interface by delegating to Node2D.
func (nr *NavigationRegion2D) Update(dt float64) {
	nr.Node2D.Update(dt)
}

// Compile-time interface check: NavigationRegion2D must satisfy node.Node.
var _ node.Node = (*NavigationRegion2D)(nil)
