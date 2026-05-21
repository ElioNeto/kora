package navigation

import (
	"fmt"
	"math/rand"
	"testing"
)

// makeOpenGrid creates an empty grid with all cells walkable.
func makeOpenGrid(w, h int) *Pathfinder {
	return NewPathfinder(w, h)
}

// makeMazeGrid creates a grid with a zigzag corridor maze pattern.
func makeMazeGrid(w, h int) *Pathfinder {
	pf := NewPathfinder(w, h)
	// Create a checkerboard-like wall pattern that forces pathfinding detours.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if x == 0 || x == w-1 || y == 0 || y == h-1 {
				continue // keep border walkable
			}
			// Block every other cell in a checkerboard with some randomness.
			if (x+y)%3 == 0 {
				pf.SetWalkable(x, y, false)
			}
		}
	}
	return pf
}

// makeRandomObstacles adds random blocked cells to a pathfinder.
func makeRandomObstacles(pf *Pathfinder, pct float64) {
	w, h := pf.Width(), pf.Height()
	total := w * h
	blocked := int(float64(total) * pct)
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < blocked; i++ {
		x := rng.Intn(w)
		y := rng.Intn(h)
		pf.SetWalkable(x, y, false)
	}
}

// ---------------------------------------------------------------------------
// A* on open grids of various sizes
// Expected order of magnitude:
//   10x10:     ~0.5-5 µs/op
//   100x100:   ~50-500 µs/op
//   1000x1000: ~5-50 ms/op
// ---------------------------------------------------------------------------

func BenchmarkAStar_10x10(b *testing.B) {
	pf := makeOpenGrid(10, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := pf.FindPath(0, 0, 9, 9)
		if path == nil {
			b.Fatal("expected path on open grid")
		}
	}
}

func BenchmarkAStar_100x100(b *testing.B) {
	pf := makeOpenGrid(100, 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := pf.FindPath(0, 0, 99, 99)
		if path == nil {
			b.Fatal("expected path on open grid")
		}
	}
}

func BenchmarkAStar_1000x1000(b *testing.B) {
	pf := makeOpenGrid(1000, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := pf.FindPath(0, 0, 999, 999)
		if path == nil {
			b.Fatal("expected path on open grid")
		}
	}
}

// BenchmarkAStar_Maze runs A* on a grid with maze-like obstacles.
// Expected order of magnitude: similar to open grid but slower due to more
// nodes explored before finding the goal.
func BenchmarkAStar_Maze(b *testing.B) {
	sizes := []struct {
		name string
		w, h int
	}{
		{"10x10", 10, 10},
		{"100x100", 100, 100},
	}
	for _, sz := range sizes {
		b.Run(sz.name, func(b *testing.B) {
			pf := makeMazeGrid(sz.w, sz.h)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				path := pf.FindPath(1, 1, sz.w-2, sz.h-2)
				if path == nil {
					b.Fatal("expected path through maze")
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// A* with random obstacles — shows cost of exploring blocked cells
// ---------------------------------------------------------------------------

func BenchmarkAStar_WithObstacles(b *testing.B) {
	sizes := []struct {
		name string
		w, h int
		pct  float64
	}{
		{"100x100/10pct", 100, 100, 0.10},
		{"100x100/30pct", 100, 100, 0.30},
		{"100x100/50pct", 100, 100, 0.50},
	}
	for _, sz := range sizes {
		b.Run(sz.name, func(b *testing.B) {
			pf := makeOpenGrid(sz.w, sz.h)
			makeRandomObstacles(pf, sz.pct)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				path := pf.FindPath(0, 0, sz.w-1, sz.h-1)
				// Path may or may not exist; just run the algorithm.
				_ = path
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Path smoothing benchmark
// Expected order of magnitude:
//   100x100 smooth: ~50-500 µs/op (same as FindPath plus smoothing pass)
// ---------------------------------------------------------------------------

func BenchmarkNavigationSmooth(b *testing.B) {
	sizes := []struct {
		name string
		w, h int
	}{
		{"10x10", 10, 10},
		{"100x100", 100, 100},
	}
	for _, sz := range sizes {
		b.Run(sz.name, func(b *testing.B) {
			pf := makeMazeGrid(sz.w, sz.h)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				path := pf.FindPathSmooth(1, 1, sz.w-2, sz.h-2)
				if path == nil {
					b.Fatal("expected smooth path through maze")
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// BenchmarkAStar_NoPath — worst case where all cells are explored
// Expected order of magnitude: explores the entire grid
// ---------------------------------------------------------------------------

func BenchmarkAStar_NoPath(b *testing.B) {
	sizes := []int{10, 100}
	for _, n := range sizes {
		name := fmt.Sprintf("NoPath/%dx%d", n, n)
		b.Run(name, func(b *testing.B) {
			pf := NewPathfinder(n, n)
			// Block a column to disconnect start from end.
			mid := n / 2
			for y := 0; y < n; y++ {
				pf.SetWalkable(mid, y, false)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = pf.FindPath(0, 0, n-1, n-1)
			}
		})
	}
}
