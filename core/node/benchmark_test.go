package node

import (
	"fmt"
	"testing"
)

// ---------------------------------------------------------------------------
// Node2D AddChild
// Expected order of magnitude:
//   10 children:   ~10-50 ns/op
//   100 children:  ~100-500 ns/op
//   1000 children: ~1-5 µs/op
// ---------------------------------------------------------------------------

func BenchmarkNode2DAddChild(b *testing.B) {
	sizes := []int{10, 100, 1000}
	for _, n := range sizes {
		name := fmt.Sprintf("%d", n)
		b.Run(name, func(b *testing.B) {
			parent := NewNode2D("parent", 1)
			children := make([]*Node2D, n)
			for i := range children {
				children[i] = NewNode2D(fmt.Sprintf("child%d", i), uint64(i+2))
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Clear children each iteration (except first which starts empty).
				if i > 0 {
					parent.children = parent.children[:0]
				}
				for _, child := range children {
					parent.AddChild(child)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Node2D GetNode by path at various depths
// Expected order of magnitude:
//   depth 1:   ~5-20 ns/op
//   depth 5:   ~20-100 ns/op
//   depth 10:  ~50-250 ns/op
// ---------------------------------------------------------------------------

func BenchmarkNode2DGetNode(b *testing.B) {
	// Build a chain of nodes: root -> n0 -> n1 -> ... -> n9
	root := NewNode2D("root", 1)
	current := root
	for i := 0; i < 10; i++ {
		child := NewNode2D(fmt.Sprintf("n%d", i), uint64(i+2))
		current.AddChild(child)
		current = child
	}

	b.Run("depth1", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = root.GetNode("n0")
		}
	})

	b.Run("depth5", func(b *testing.B) {
		path := "n0/n1/n2/n3/n4"
		for i := 0; i < b.N; i++ {
			_ = root.GetNode(path)
		}
	})

	b.Run("depth10", func(b *testing.B) {
		path := "n0/n1/n2/n3/n4/n5/n6/n7/n8/n9"
		for i := 0; i < b.N; i++ {
			_ = root.GetNode(path)
		}
	})
}

// ---------------------------------------------------------------------------
// Node2D Update tree traversal
// Expected order of magnitude:
//   10 nodes:   ~10-100 ns/op
//   1000 nodes: ~1-10 µs/op
// ---------------------------------------------------------------------------

func BenchmarkNode2DUpdateTree(b *testing.B) {
	sizes := []struct {
		name string
		n    int
	}{
		{"small/10", 10},
		{"large/1000", 1000},
	}
	for _, sz := range sizes {
		b.Run(sz.name, func(b *testing.B) {
			root := NewNode2D("root", 1)
			// Build a tree: root has sz.n children, each child is a leaf.
			for i := 0; i < sz.n; i++ {
				child := NewNode2D(fmt.Sprintf("child%d", i), uint64(i+2))
				root.AddChild(child)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				root.Update(0.016)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Sprite2D Draw — requires ebiten runtime, skip in isolated tests
// ---------------------------------------------------------------------------

func BenchmarkSprite2DDraw(b *testing.B) {
	b.Skip("requires ebiten runtime (screen *ebiten.Image)")
}

// ---------------------------------------------------------------------------
// Particles2D Update
// Expected order of magnitude:
//   100 particles:  ~1-10 µs/op
//   500 particles:  ~5-50 µs/op
//   1000 particles: ~10-100 µs/op
// ---------------------------------------------------------------------------

func BenchmarkParticles2DUpdate(b *testing.B) {
	sizes := []int{100, 500, 1000}
	for _, n := range sizes {
		name := fmt.Sprintf("%d", n)
		b.Run(name, func(b *testing.B) {
			p := NewParticles2D("bench")
			p.SetLifetime(10.0) // long lifetime so particles don't die
			p.SetSpeed(0)
			p.SetSpread(0)
			p.Emit(n)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				p.Update(0.016)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Particles2D Draw — requires ebiten runtime, skip in isolated tests
// ---------------------------------------------------------------------------

func BenchmarkParticles2DDraw(b *testing.B) {
	b.Skip("requires ebiten runtime (screen *ebiten.Image)")
}
