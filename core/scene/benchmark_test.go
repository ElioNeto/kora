package scene_test

import (
	"fmt"
	"testing"

	"github.com/ElioNeto/kora/core/scene"
)

// benchEntity is a minimal entity for benchmark use.
type benchEntity struct {
	scene.BaseEntity
	alive   bool
}

func newBenchEntity() *benchEntity { return &benchEntity{alive: true} }
func (e *benchEntity) IsAlive() bool  { return e.alive }
func (e *benchEntity) Destroy()       { e.alive = false }
func (e *benchEntity) Update(_ float64) {}

// ---------------------------------------------------------------------------
// Scene Spawn
// Expected order of magnitude:
//   10 entities:   ~100-500 ns/op
//   100 entities:  ~1-5 µs/op
//   1000 entities: ~10-50 µs/op
// ---------------------------------------------------------------------------

func BenchmarkSceneSpawn(b *testing.B) {
	sizes := []int{10, 100, 1000}
	for _, n := range sizes {
		name := fmt.Sprintf("Spawn/%d", n)
		b.Run(name, func(b *testing.B) {
			s := scene.New()
			entities := make([]*benchEntity, n)
			for i := range entities {
				entities[i] = newBenchEntity()
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Create a fresh scene every iteration to avoid growing debt.
				if i > 0 {
					s = scene.New()
				}
				for _, e := range entities {
					s.Spawn(e)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Scene Update
// Expected order of magnitude:
//   10 entities:   ~50-200 ns/op
//   100 entities:  ~500-2000 ns/op
//   1000 entities: ~5-20 µs/op
// ---------------------------------------------------------------------------

func BenchmarkSceneUpdate(b *testing.B) {
	sizes := []int{10, 100, 1000}
	for _, n := range sizes {
		name := fmt.Sprintf("Update/%d", n)
		b.Run(name, func(b *testing.B) {
			s := scene.New()
			for i := 0; i < n; i++ {
				e := newBenchEntity()
				s.Spawn(e)
			}
			s.Update(0.016) // flush pending spawns

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s.Update(0.016)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Scene Find (by group)
// Expected order of magnitude:
//   10 in group:  ~10-50 ns/op
//   100 in group: ~50-200 ns/op
//   1000 in group:~500-2000 ns/op
// ---------------------------------------------------------------------------

func BenchmarkSceneFind(b *testing.B) {
	sizes := []int{10, 100, 1000}
	for _, n := range sizes {
		name := fmt.Sprintf("Find/%d", n)
		b.Run(name, func(b *testing.B) {
			s := scene.New()
			for i := 0; i < n; i++ {
				e := newBenchEntity()
				s.SpawnInGroup("enemies", e)
			}
			s.Update(0.016) // flush pending spawns

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = s.Find("enemies")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Scene FindAll (by group)
// Expected order of magnitude:
//   10 in group:  ~10-50 ns/op
//   100 in group: ~50-200 ns/op
//   1000 in group:~500-2000 ns/op
// ---------------------------------------------------------------------------

func BenchmarkSceneFindAll(b *testing.B) {
	sizes := []int{10, 100, 1000}
	for _, n := range sizes {
		name := fmt.Sprintf("FindAll/%d", n)
		b.Run(name, func(b *testing.B) {
			s := scene.New()
			for i := 0; i < n; i++ {
				e := newBenchEntity()
				s.SpawnInGroup("enemies", e)
			}
			s.Update(0.016) // flush pending spawns

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = s.FindAll("enemies")
			}
		})
	}
}
