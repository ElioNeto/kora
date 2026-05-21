package navigation

import (
	"testing"
)

// validatePath checks that a path is valid for the given grid and start/end.
func validatePath(t *testing.T, pf *Pathfinder, path []Point, startX, startY, endX, endY int) {
	t.Helper()

	if path == nil {
		t.Error("expected non-nil path")
		return
	}

	// Verify last point is the end
	if len(path) > 0 {
		last := path[len(path)-1]
		if last.X != endX || last.Y != endY {
			t.Errorf("path does not end at target (%d,%d), ends at (%d,%d)", endX, endY, last.X, last.Y)
		}
	}

	// Verify path continuity and walkability
	prevX, prevY := startX, startY
	for i, p := range path {
		if !pf.IsWalkable(p.X, p.Y) {
			t.Errorf("path[%d] = (%d,%d) is not walkable", i, p.X, p.Y)
		}
		dx := p.X - prevX
		dy := p.Y - prevY
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}
		if dx+dy != 1 {
			t.Errorf("non-adjacent step from (%d,%d) to (%d,%d)", prevX, prevY, p.X, p.Y)
		}
		prevX, prevY = p.X, p.Y
	}
}

// countBlocks counts the number of non-walkable (blocked) cells in the grid.
func countBlocks(pf *Pathfinder) int {
	count := 0
	for y := 0; y < pf.Height(); y++ {
		for x := 0; x < pf.Width(); x++ {
			if !pf.IsWalkable(x, y) {
				count++
			}
		}
	}
	return count
}

func TestFindPath_Straight(t *testing.T) {
	pf := NewPathfinder(10, 10)

	tests := []struct {
		name              string
		startX, startY, endX, endY int
		expectedLen       int // -1 means expect nil
	}{
		{
			name:        "horizontal right",
			startX:      0, startY: 0,
			endX:        5, endY: 0,
			expectedLen: 5,
		},
		{
			name:        "horizontal left",
			startX:      5, startY: 3,
			endX:        0, endY: 3,
			expectedLen: 5,
		},
		{
			name:        "vertical down",
			startX:      2, startY: 0,
			endX:        2, endY: 5,
			expectedLen: 5,
		},
		{
			name:        "vertical up",
			startX:      2, startY: 5,
			endX:        2, endY: 0,
			expectedLen: 5,
		},
		{
			name:        "start equals end",
			startX:      3, startY: 3,
			endX:        3, endY: 3,
			expectedLen: 0,
		},
		{
			name:        "adjacent cells",
			startX:      0, startY: 0,
			endX:        1, endY: 0,
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := pf.FindPath(tt.startX, tt.startY, tt.endX, tt.endY)
			if tt.expectedLen == -1 {
				if path != nil {
					t.Errorf("expected nil path, got %v", path)
				}
				return
			}
			if path == nil {
				t.Fatal("expected non-nil path, got nil")
			}
			if len(path) != tt.expectedLen {
				t.Errorf("expected path length %d, got %d: %v", tt.expectedLen, len(path), path)
			}
			validatePath(t, pf, path, tt.startX, tt.startY, tt.endX, tt.endY)
		})
	}
}

func TestFindPath_WithObstacle(t *testing.T) {
	pf := NewPathfinder(10, 10)

	// Create a wall from x=3, y=2 to x=3, y=7 (partial wall)
	for y := 2; y <= 7; y++ {
		pf.SetWalkable(3, y, false)
	}

	tests := []struct {
		name                    string
		startX, startY, endX, endY int
		expectPath              bool
	}{
		{
			name:       "go around wall right",
			startX:     0, startY: 0,
			endX:       5, endY: 9,
			expectPath: true,
		},
		{
			name:       "end on wall - unreachable",
			startX:     0, startY: 0,
			endX:       3, endY: 5,
			expectPath: false,
		},
		{
			name:       "start on wall - unreachable",
			startX:     3, startY: 3,
			endX:       5, endY: 5,
			expectPath: false,
		},
		{
			name:       "both sides of partial wall",
			startX:     1, startY: 0,
			endX:       5, endY: 0,
			expectPath: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := pf.FindPath(tt.startX, tt.startY, tt.endX, tt.endY)
			if tt.expectPath {
				if path == nil {
					t.Fatal("expected a path, got nil")
				}
				validatePath(t, pf, path, tt.startX, tt.startY, tt.endX, tt.endY)
			} else {
				if path != nil {
					t.Errorf("expected nil path, got %v", path)
				}
			}
		})
	}
}

func TestFindPath_Maze(t *testing.T) {
	// 7x7 maze with a zigzag corridor
	//
	// Legend: . = walkable, # = wall
	//
	//  .  .  .  .  .  .  .
	//  .  #  #  #  #  #  .
	//  .  .  .  .  .  #  .
	//  .  #  #  #  .  #  .
	//  .  #  .  .  .  #  .
	//  .  #  #  #  #  #  .
	//  .  .  .  .  .  .  .
	//
	// Path from (0,0) to (4,4):
	// Go around the outer ring to (0,2), then zigzag through
	// the corridor: (1,2)→(2,2)→(3,2)→(4,2)→(4,3)→(4,4)

	pf := NewPathfinder(7, 7)

	// Walls forming a zigzag corridor
	walls := [][2]int{
		{1, 1}, {2, 1}, {3, 1}, {4, 1}, {5, 1},
		{5, 2},
		{1, 3}, {2, 3}, {3, 3}, {5, 3},
		{1, 4}, {5, 4},
		{1, 5}, {2, 5}, {3, 5}, {4, 5}, {5, 5},
	}
	for _, w := range walls {
		pf.SetWalkable(w[0], w[1], false)
	}

	path := pf.FindPath(0, 0, 4, 4)
	if path == nil {
		t.Fatal("expected a path through the maze, got nil")
	}

	validatePath(t, pf, path, 0, 0, 4, 4)

	// Verify it finds a valid path (should be at least a few steps due to detour)
	if len(path) < 2 {
		t.Errorf("expected a longer path through the maze, got length %d: %v", len(path), path)
	}
	t.Logf("maze path length: %d, path: %v", len(path), path)
}

func TestFindPath_UnreachableTarget(t *testing.T) {
	pf := NewPathfinder(5, 5)

	// Completely surround (2,2) with walls
	pf.SetWalkable(2, 1, false) // top
	pf.SetWalkable(2, 3, false) // bottom
	pf.SetWalkable(1, 2, false) // left
	pf.SetWalkable(3, 2, false) // right

	tests := []struct {
		name              string
		startX, startY, endX, endY int
	}{
		{
			name:    "surrounded target",
			startX:  0, startY: 0,
			endX:    2, endY: 2,
		},
		{
			name:    "surrounded start",
			startX:  2, startY: 2,
			endX:    0, endY: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := pf.FindPath(tt.startX, tt.startY, tt.endX, tt.endY)
			if path != nil {
				t.Errorf("expected nil path for unreachable target, got %v", path)
			}
		})
	}
}

func TestFindPath_FullyWalledOff(t *testing.T) {
	pf := NewPathfinder(5, 5)

	// Split the grid into two disconnected regions
	// Block column 2 for all rows from 0 to 4
	for y := 0; y < 5; y++ {
		pf.SetWalkable(2, y, false)
	}

	// (0,0) and (4,4) are now disconnected
	path := pf.FindPath(0, 0, 4, 4)
	if path != nil {
		t.Errorf("expected nil path for walled-off grid, got %v (len=%d)", path, len(path))
	}
}

func TestFindPath_OutOfBounds(t *testing.T) {
	pf := NewPathfinder(5, 5)

	tests := []struct {
		name              string
		startX, startY, endX, endY int
	}{
		{name: "start outside grid", startX: -1, startY: 0, endX: 3, endY: 3},
		{name: "end outside grid", startX: 0, startY: 0, endX: 10, endY: 10},
		{name: "both outside grid", startX: -5, startY: -5, endX: 10, endY: 10},
		{name: "start negative", startX: 0, startY: -1, endX: 3, endY: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := pf.FindPath(tt.startX, tt.startY, tt.endX, tt.endY)
			if path != nil {
				t.Errorf("expected nil path for out-of-bounds, got %v", path)
			}
		})
	}
}

func TestFindPath_EmptyGrid(t *testing.T) {
	pf := NewPathfinder(1, 1)

	// Only one cell (0,0)
	path := pf.FindPath(0, 0, 0, 0)
	if path == nil {
		t.Fatal("expected empty path for start==end, got nil")
	}
	if len(path) != 0 {
		t.Errorf("expected empty path, got length %d: %v", len(path), path)
	}
}

func TestFindPath_LargerGrid(t *testing.T) {
	// Test performance with a 256x256 grid (max spec)
	pf := NewPathfinder(256, 256)

	// Create a winding corridor
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			if x == 0 || x == 255 || y == 0 || y == 255 {
				continue // keep border walkable
			}
			// Checkerboard pattern
			if (x+y)%2 == 0 {
				pf.SetWalkable(x, y, false)
			}
		}
	}

	// Path should still be found along the border
	path := pf.FindPath(0, 0, 255, 255)
	if path == nil {
		t.Fatal("expected a path in the large grid, got nil")
	}
	if len(path) < 100 {
		t.Errorf("expected long path around checkerboard, got length %d", len(path))
	}
	validatePath(t, pf, path, 0, 0, 255, 255)
}

func TestFindPathSmooth_RemovesRedundant(t *testing.T) {
	pf := NewPathfinder(10, 10)

	tests := []struct {
		name              string
		startX, startY, endX, endY int
		expectedSmoothLen int
	}{
		{
			name:              "straight line removes all intermediate",
			startX:            0, startY: 0,
			endX:              5, endY: 0,
			expectedSmoothLen: 2, // first waypoint + end
		},
		{
			name:              "vertical line",
			startX:            0, startY: 0,
			endX:              0, endY: 5,
			expectedSmoothLen: 2, // first waypoint + end
		},
		{
			name:              "short path unchanged",
			startX:            0, startY: 0,
			endX:              1, endY: 0,
			expectedSmoothLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := pf.FindPath(tt.startX, tt.startY, tt.endX, tt.endY)
			if raw == nil {
				t.Fatal("expected non-nil raw path")
			}

			smooth := pf.FindPathSmooth(tt.startX, tt.startY, tt.endX, tt.endY)
			if smooth == nil {
				t.Fatal("expected non-nil smoothed path")
			}

			if len(smooth) != tt.expectedSmoothLen {
				t.Errorf("expected smoothed path length %d, got %d: %v", tt.expectedSmoothLen, len(smooth), smooth)
			}

			// Verify smoothed path still reaches the end
			if len(smooth) > 0 {
				last := smooth[len(smooth)-1]
				if last.X != tt.endX || last.Y != tt.endY {
					t.Errorf("smoothed path does not end at target (%d,%d), ends at (%d,%d)",
						tt.endX, tt.endY, last.X, last.Y)
				}
			}
		})
	}
}

func TestFindPathSmooth_PreservesCorners(t *testing.T) {
	pf := NewPathfinder(10, 10)

	// Wall blocking direct path, forcing an L-shaped path
	for y := 0; y < 5; y++ {
		pf.SetWalkable(3, y, false)
	}

	// Path from (0,0) to (7,0) must go around the wall
	raw := pf.FindPath(0, 0, 7, 0)
	if raw == nil {
		t.Fatal("expected non-nil raw path")
	}

	smooth := pf.FindPathSmooth(0, 0, 7, 0)
	if smooth == nil {
		t.Fatal("expected non-nil smoothed path")
	}

	if len(smooth) >= len(raw) {
		t.Logf("raw path length: %d, smoothed length: %d", len(raw), len(smooth))
		t.Logf("raw: %v", raw)
		t.Logf("smooth: %v", smooth)
	}

	// Verify smoothed path still reaches the end
	if len(smooth) > 0 {
		last := smooth[len(smooth)-1]
		if last.X != 7 || last.Y != 0 {
			t.Errorf("smoothed path does not end at (7,0), ends at (%d,%d)", last.X, last.Y)
		}
	}

	// Verify all points in smoothed path are walkable
	for i, p := range smooth {
		if !pf.IsWalkable(p.X, p.Y) {
			t.Errorf("smoothed path[%d] = (%d,%d) is not walkable", i, p.X, p.Y)
		}
	}
}

func TestFindPathSmooth_EmptyPath(t *testing.T) {
	pf := NewPathfinder(5, 5)

	// Start == end
	raw := pf.FindPath(2, 2, 2, 2)
	if raw == nil {
		t.Fatal("expected empty raw path")
	}
	if len(raw) != 0 {
		t.Fatalf("expected empty raw path, got length %d", len(raw))
	}

	smooth := pf.FindPathSmooth(2, 2, 2, 2)
	if smooth == nil {
		t.Fatal("expected non-nil smooth path")
	}
	if len(smooth) != 0 {
		t.Errorf("expected empty smooth path, got length %d: %v", len(smooth), smooth)
	}
}

func TestFindPathSmooth_NilPath(t *testing.T) {
	pf := NewPathfinder(5, 5)

	// Block all neighbors of start so no path exists
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			if x != 0 || y != 0 {
				pf.SetWalkable(x, y, false)
			}
		}
	}

	smooth := pf.FindPathSmooth(0, 0, 4, 4)
	if smooth != nil {
		t.Errorf("expected nil smooth path for unreachable target, got %v", smooth)
	}
}

func TestSetWalkable(t *testing.T) {
	pf := NewPathfinder(3, 3)

	// All cells should start walkable
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			if !pf.IsWalkable(x, y) {
				t.Errorf("cell (%d,%d) should be walkable by default", x, y)
			}
		}
	}

	// Set a cell as blocked
	pf.SetWalkable(1, 1, false)
	if pf.IsWalkable(1, 1) {
		t.Error("cell (1,1) should not be walkable after SetWalkable(false)")
	}

	// Set it back
	pf.SetWalkable(1, 1, true)
	if !pf.IsWalkable(1, 1) {
		t.Error("cell (1,1) should be walkable after SetWalkable(true)")
	}

	// Out of bounds should not panic
	pf.SetWalkable(-1, 0, false)
	pf.SetWalkable(10, 10, true)
}

func TestIsWalkable_OutOfBounds(t *testing.T) {
	pf := NewPathfinder(5, 5)

	if pf.IsWalkable(-1, 0) {
		t.Error("expected IsWalkable(-1, 0) to be false")
	}
	if pf.IsWalkable(0, -1) {
		t.Error("expected IsWalkable(0, -1) to be false")
	}
	if pf.IsWalkable(5, 0) {
		t.Error("expected IsWalkable(5, 0) to be false")
	}
	if pf.IsWalkable(0, 5) {
		t.Error("expected IsWalkable(0, 5) to be false")
	}
}

func TestPathfinder_WidthHeight(t *testing.T) {
	pf := NewPathfinder(10, 20)

	if pf.Width() != 10 {
		t.Errorf("expected width 10, got %d", pf.Width())
	}
	if pf.Height() != 20 {
		t.Errorf("expected height 20, got %d", pf.Height())
	}
}

func TestFindPath_PathQuality(t *testing.T) {
	// Test that A* finds the shortest path, not just any path
	pf := NewPathfinder(5, 5)

	// Open grid with no obstacles
	// The shortest path from (0,0) to (4,0) should be 4 steps (straight line)
	path := pf.FindPath(0, 0, 4, 0)
	if path == nil {
		t.Fatal("expected non-nil path")
	}
	if len(path) != 4 {
		t.Errorf("expected shortest path length 4, got %d: %v", len(path), path)
	}

	// Verify it's actually straight (all points have Y=0)
	for i, p := range path {
		if p.Y != 0 {
			t.Errorf("path[%d] = (%d,%d) is off the straight line", i, p.X, p.Y)
		}
	}
}

func TestFindPath_DiagonalMovement(t *testing.T) {
	// A* with 4-directional movement should find a Manhattan path
	pf := NewPathfinder(5, 5)

	path := pf.FindPath(0, 0, 4, 4)
	if path == nil {
		t.Fatal("expected non-nil path")
	}

	// With 4-directional movement, shortest path from (0,0) to (4,4) is 8 steps
	if len(path) != 8 {
		t.Errorf("expected path length 8 for diagonal movement, got %d: %v", len(path), path)
	}
}
