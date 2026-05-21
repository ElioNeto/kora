package physics

import (
	"fmt"
	"testing"
)

// ---------------------------------------------------------------------------
// helpers

func spatialBody(id int, x, y, w, h float32) *RigidBody {
	b := NewBody(id, x, y, w, h, BodyDynamic)
	return b
}

// ---------------------------------------------------------------------------
// NewSpatialHash creates an empty structure

func TestNewSpatialHashEmpty(t *testing.T) {
	sh := NewSpatialHash(64)
	if sh == nil {
		t.Fatal("NewSpatialHash returned nil")
	}
	if sh.cellSize != 64 {
		t.Errorf("expected cellSize 64, got %f", sh.cellSize)
	}
	if len(sh.cells) != 0 {
		t.Errorf("expected empty cells, got %d entries", len(sh.cells))
	}
	if len(sh.bodies) != 0 {
		t.Errorf("expected empty bodies, got %d entries", len(sh.bodies))
	}
}

// ---------------------------------------------------------------------------
// Insert and GetCandidates returns nearby bodies

func TestInsertAndGetCandidates(t *testing.T) {
	sh := NewSpatialHash(64)

	a := spatialBody(1, 0, 0, 16, 16)
	b := spatialBody(2, 30, 0, 16, 16) // same cell
	c := spatialBody(3, 200, 0, 16, 16) // far away, different cell

	sh.Insert(1, a)
	sh.Insert(2, b)
	sh.Insert(3, c)

	candidates := sh.GetCandidates(1, a)
	if len(candidates) == 0 {
		t.Fatal("expected at least one candidate")
	}

	foundB := false
	foundC := false
	for _, id := range candidates {
		if id == 2 {
			foundB = true
		}
		if id == 3 {
			foundC = true
		}
	}
	if !foundB {
		t.Error("expected body 2 (nearby) as candidate")
	}
	if foundC {
		t.Error("did NOT expect body 3 (far away) as candidate")
	}
}

// ---------------------------------------------------------------------------
// Clear removes all bodies

func TestSpatialClear(t *testing.T) {
	sh := NewSpatialHash(64)

	b1 := spatialBody(1, 0, 0, 16, 16)
	b2 := spatialBody(2, 30, 0, 16, 16)
	sh.Insert(1, b1)
	sh.Insert(2, b2)

	if len(sh.bodies) != 2 {
		t.Fatalf("expected 2 bodies before clear, got %d", len(sh.bodies))
	}

	sh.Clear()

	if len(sh.bodies) != 0 {
		t.Errorf("expected 0 bodies after clear, got %d", len(sh.bodies))
	}
	if len(sh.cells) != 0 {
		t.Errorf("expected 0 cells after clear, got %d", len(sh.cells))
	}

	// GetCandidates on a cleared hash should return empty
	candidates := sh.GetCandidates(1, b1)
	if len(candidates) != 0 {
		t.Errorf("expected no candidates after clear, got %d", len(candidates))
	}
}

// ---------------------------------------------------------------------------
// GetCandidates excludes the query body itself

func TestGetCandidatesExcludesSelf(t *testing.T) {
	sh := NewSpatialHash(64)

	a := spatialBody(1, 0, 0, 16, 16)
	b := spatialBody(2, 30, 0, 16, 16)
	sh.Insert(1, a)
	sh.Insert(2, b)

	candidates := sh.GetCandidates(1, a)
	for _, id := range candidates {
		if id == 1 {
			t.Error("GetCandidates should not include the query body itself")
		}
	}
}

// ---------------------------------------------------------------------------
// Bodies in different cells don't appear as candidates

func TestDifferentCellsNoCandidates(t *testing.T) {
	sh := NewSpatialHash(64)

	// cell 0,0
	a := spatialBody(1, 0, 0, 16, 16)
	// cell 4,0 (4*64 = 256 away)
	b := spatialBody(2, 260, 0, 16, 16)

	sh.Insert(1, a)
	sh.Insert(2, b)

	candidates := sh.GetCandidates(1, a)
	for _, id := range candidates {
		if id == 2 {
			t.Error("body in distant cell should not be a candidate")
		}
	}

	candidates = sh.GetCandidates(2, b)
	for _, id := range candidates {
		if id == 1 {
			t.Error("body in distant cell should not be a candidate")
		}
	}
}

// ---------------------------------------------------------------------------
// Body spanning multiple cells is found in each

func TestMultiCellSpan(t *testing.T) {
	sh := NewSpatialHash(64)

	// A large body that spans 3 cells horizontally (x: -10 .. 150 → cells -1, 0, 1, 2)
	big := spatialBody(1, 70, 0, 160, 32)
	// A small body in cell 1 (x=80)
	small := spatialBody(2, 80, 0, 16, 16)

	sh.Insert(1, big)
	sh.Insert(2, small)

	// big spans many cells — small should be a candidate (they overlap in cell 1)
	candidates := sh.GetCandidates(1, big)
	found := false
	for _, id := range candidates {
		if id == 2 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected small body to be a candidate for the large spanning body")
	}

	// Also verify that small has big as candidate
	candidates2 := sh.GetCandidates(2, small)
	found = false
	for _, id := range candidates2 {
		if id == 1 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected large spanning body to be a candidate for the small body")
	}
}

// ---------------------------------------------------------------------------
// Update re-inserts a moved body

func TestSpatialUpdate(t *testing.T) {
	sh := NewSpatialHash(64)

	a := spatialBody(1, 0, 0, 16, 16)
	b := spatialBody(2, 30, 0, 16, 16)

	sh.Insert(1, a)
	sh.Insert(2, b)

	// Move b far away; the spatial hash should be rebuilt each frame
	// (Clear + re-insert), which is the recommended usage pattern.
	sh.Clear()
	b.Pos.X = 300
	sh.Insert(1, a)
	sh.Insert(2, b)

	candidates := sh.GetCandidates(1, a)
	for _, id := range candidates {
		if id == 2 {
			t.Error("after Clear+reinsert, body moved far away should not be a candidate")
		}
	}

	// Verify both bodies exist in hash
	if _, ok := sh.bodies[1]; !ok {
		t.Error("body 1 should exist in hash")
	}
	if _, ok := sh.bodies[2]; !ok {
		t.Error("body 2 should exist in hash")
	}
}

// ---------------------------------------------------------------------------
// Remove removes a body completely

func TestSpatialRemove(t *testing.T) {
	sh := NewSpatialHash(64)

	a := spatialBody(1, 0, 0, 16, 16)
	b := spatialBody(2, 30, 0, 16, 16)

	sh.Insert(1, a)
	sh.Insert(2, b)

	sh.Remove(2, b)

	if _, ok := sh.bodies[2]; ok {
		t.Error("body should be removed from bodies map")
	}

	candidates := sh.GetCandidates(1, a)
	for _, id := range candidates {
		if id == 2 {
			t.Error("removed body should not be a candidate")
		}
	}
}

// ---------------------------------------------------------------------------
// Benchmark: spatial hash with many bodies

func BenchmarkSpatialHashInsert(b *testing.B) {
	sh := NewSpatialHash(64)
	bodies := make([]*RigidBody, 1000)
	for i := range bodies {
		bodies[i] = spatialBody(i, float32(i*10), 0, 16, 16)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sh.Clear()
		for _, body := range bodies {
			sh.Insert(body.EntityID, body)
		}
	}
}

func BenchmarkSpatialHashGetCandidates(b *testing.B) {
	sh := NewSpatialHash(64)
	bodies := make([]*RigidBody, 1000)
	for i := range bodies {
		bodies[i] = spatialBody(i, float32(i*10), 0, 16, 16)
		sh.Insert(bodies[i].EntityID, bodies[i])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, body := range bodies {
			_ = sh.GetCandidates(body.EntityID, body)
		}
	}
}

// BenchmarkBruteForce simulates the O(n²) pair iteration for comparison.
func BenchmarkBruteForce1000(b *testing.B) {
	bodies := make([]*RigidBody, 1000)
	for i := range bodies {
		bodies[i] = spatialBody(i, float32(i*10), 0, 16, 16)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < len(bodies); i++ {
			for j := i + 1; j < len(bodies); j++ {
				// just iterate — no collision resolve in benchmark
				_ = bodies[i].EntityID + bodies[j].EntityID
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Sub-benchmarks grouping by body count (100 and 1000)
// ---------------------------------------------------------------------------

// makeBodies creates n bodies spread across a 1000×1000 area.
func makeBodies(n int) []*RigidBody {
	bodies := make([]*RigidBody, n)
	for i := range bodies {
		x := float32((i * 10) % 1000)
		y := float32(i/100) * 10
		bodies[i] = spatialBody(i, x, y, 16, 16)
	}
	return bodies
}

func BenchmarkSpatialHashBruteForce(b *testing.B) {
	sizes := []int{100, 1000}
	for _, n := range sizes {
		bodies := makeBodies(n)
		name := "BruteForce"
		if n == 100 {
			name = "BruteForce/100"
		} else {
			name = "BruteForce/1000"
		}
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for i := 0; i < len(bodies); i++ {
					for j := i + 1; j < len(bodies); j++ {
						_ = bodies[i].EntityID + bodies[j].EntityID
					}
				}
			}
		})
	}
}

// BenchmarkPhysicsWorldStep measures the World.Step() performance.
// Expected order of magnitude:
//   100 bodies: ~1-5 µs/op
//   1000 bodies: ~50-500 µs/op (broad-phase overhead grows)
func BenchmarkPhysicsWorldStep(b *testing.B) {
	sizes := []int{100, 1000}
	for _, n := range sizes {
		name := fmt.Sprintf("Step/%d", n)
		b.Run(name, func(b *testing.B) {
			w := NewWorld(nil)
			for i := 0; i < n; i++ {
				body := spatialBody(i, float32(i*10), 0, 16, 16)
				w.Register(body)
			}
			dt := float32(1.0 / 60.0)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				w.Step(dt)
			}
		})
	}
}
