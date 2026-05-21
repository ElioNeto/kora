// Package navigation provides A* pathfinding and AI navigation for the Kora engine.
package navigation

// Point represents a grid coordinate for pathfinding.
type Point struct {
	X, Y int
}

// Node represents a node in the A* search graph.
type Node struct {
	Point
	G, H, F float64
	Parent  *Node
}

// Pathfinder implements the A* pathfinding algorithm on a 2D grid.
type Pathfinder struct {
	width, height int
	grid          [][]bool // true = walkable
}

// NewPathfinder creates a new pathfinder with the given grid dimensions.
// All cells are walkable by default.
func NewPathfinder(width, height int) *Pathfinder {
	grid := make([][]bool, height)
	for y := range grid {
		grid[y] = make([]bool, width)
		for x := range grid[y] {
			grid[y][x] = true
		}
	}
	return &Pathfinder{
		width:  width,
		height: height,
		grid:   grid,
	}
}

// SetWalkable sets whether a grid cell is walkable.
func (pf *Pathfinder) SetWalkable(x, y int, walkable bool) {
	if x < 0 || x >= pf.width || y < 0 || y >= pf.height {
		return
	}
	pf.grid[y][x] = walkable
}

// IsWalkable returns whether a grid cell is walkable.
// Out-of-bounds cells are reported as not walkable.
func (pf *Pathfinder) IsWalkable(x, y int) bool {
	if x < 0 || x >= pf.width || y < 0 || y >= pf.height {
		return false
	}
	return pf.grid[y][x]
}

// FindPath finds a path from start to end using A* with Manhattan distance.
// Returns the path as a slice of Points (excluding start, including end).
// If no path exists, returns nil.
// If start == end, returns an empty slice.
func (pf *Pathfinder) FindPath(startX, startY, endX, endY int) []Point {
	// Bounds check
	if startX < 0 || startX >= pf.width || startY < 0 || startY >= pf.height {
		return nil
	}
	if endX < 0 || endX >= pf.width || endY < 0 || endY >= pf.height {
		return nil
	}

	// Check if start or end is blocked
	if !pf.grid[startY][startX] || !pf.grid[endY][endX] {
		return nil
	}

	// If start == end, return empty path
	if startX == endX && startY == endY {
		return []Point{}
	}

	// Closed set: 2D bool array
	closed := make([][]bool, pf.height)
	for y := range closed {
		closed[y] = make([]bool, pf.width)
	}

	// Open set: simple slice
	open := make([]*Node, 0, pf.width*pf.height)

	// Start node
	startNode := &Node{
		Point: Point{X: startX, Y: startY},
		G:     0,
		H:     manhattan(startX, startY, endX, endY),
	}
	startNode.F = startNode.G + startNode.H
	open = append(open, startNode)

	// 4-directional movement: up, down, left, right
	dirs := [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

	for len(open) > 0 {
		// Find node with lowest F score
		currentIdx := 0
		for i := 1; i < len(open); i++ {
			if open[i].F < open[currentIdx].F {
				currentIdx = i
			}
		}
		current := open[currentIdx]

		// Move from open to closed
		open = append(open[:currentIdx], open[currentIdx+1:]...)
		closed[current.Y][current.X] = true

		// Reached the goal
		if current.X == endX && current.Y == endY {
			return reconstructPath(current)
		}

		// Explore neighbors
		for _, dir := range dirs {
			nx, ny := current.X+dir[0], current.Y+dir[1]

			// Bounds check
			if nx < 0 || nx >= pf.width || ny < 0 || ny >= pf.height {
				continue
			}

			// Skip if not walkable or already evaluated
			if !pf.grid[ny][nx] || closed[ny][nx] {
				continue
			}

			// G cost: uniform cost of 1 per step
			g := current.G + 1

			// Check if already in open set
			found := false
			for _, n := range open {
				if n.X == nx && n.Y == ny {
					found = true
					if g < n.G {
						// Better path found
						n.G = g
						n.F = n.G + n.H
						n.Parent = current
					}
					break
				}
			}

			if !found {
				h := manhattan(nx, ny, endX, endY)
				neighbor := &Node{
					Point:  Point{X: nx, Y: ny},
					G:      g,
					H:      h,
					F:      g + h,
					Parent: current,
				}
				open = append(open, neighbor)
			}
		}
	}

	// No path found
	return nil
}

// FindPathSmooth returns a smoothed path using simplified string pulling.
// Redundant waypoints (collinear points) are removed.
func (pf *Pathfinder) FindPathSmooth(startX, startY, endX, endY int) []Point {
	path := pf.FindPath(startX, startY, endX, endY)
	if path == nil {
		return nil
	}
	return smoothPath(path)
}

// Width returns the grid width.
func (pf *Pathfinder) Width() int {
	return pf.width
}

// Height returns the grid height.
func (pf *Pathfinder) Height() int {
	return pf.height
}

// manhattan calculates the Manhattan distance between two points.
func manhattan(x1, y1, x2, y2 int) float64 {
	dx := x1 - x2
	dy := y1 - y2
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return float64(dx + dy)
}

// reconstructPath rebuilds the path from the end node back to the start.
func reconstructPath(end *Node) []Point {
	var path []Point
	current := end
	for current.Parent != nil {
		path = append([]Point{current.Point}, path...)
		current = current.Parent
	}
	return path
}

// smoothPath removes redundant waypoints that lie on a straight line
// between their neighbours (simplified string pulling).
func smoothPath(path []Point) []Point {
	if len(path) <= 2 {
		return path
	}

	result := make([]Point, 0, len(path))
	// Always include the first point (adjacent to start)
	result = append(result, path[0])

	for i := 1; i < len(path)-1; i++ {
		prev := result[len(result)-1]
		curr := path[i]
		next := path[i+1]

		// Check if direction changes using cross product.
		// If cross product is 0, points are collinear (same direction).
		dx1 := curr.X - prev.X
		dy1 := curr.Y - prev.Y
		dx2 := next.X - curr.X
		dy2 := next.Y - curr.Y

		cross := dx1*dy2 - dy1*dx2
		if cross != 0 {
			// Direction changed, keep this waypoint
			result = append(result, curr)
		}
		// Else: collinear, skip redundant waypoint
	}

	// Always include the end point
	result = append(result, path[len(path)-1])
	return result
}
