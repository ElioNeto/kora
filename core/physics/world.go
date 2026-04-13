package physics

const (
	// DefaultGravity is in pixels per second squared (down = +Y).
	DefaultGravity float32 = 980
	// TerminalVelocity caps the falling speed.
	TerminalVelocity float32 = 1200
)

// TileQuery is a function the world calls to know whether a tile at
// world position (px, py) is solid. Provided by core/render/tilemap.
type TileQuery func(px, py float32) bool

// PhysicsWorld manages all bodies and drives the simulation.
type PhysicsWorld struct {
	Gravity float32
	bodies  []*RigidBody
	tileQ   TileQuery
}

// NewWorld creates a PhysicsWorld with standard gravity.
// tileQuery may be nil if no tilemap is used.
func NewWorld(tileQuery TileQuery) *PhysicsWorld {
	return &PhysicsWorld{
		Gravity: DefaultGravity,
		tileQ:   tileQuery,
	}
}

// Register adds a body to the simulation.
func (w *PhysicsWorld) Register(b *RigidBody) {
	w.bodies = append(w.bodies, b)
}

// Remove detaches a body from the simulation.
func (w *PhysicsWorld) Remove(entityID int) {
	for i, b := range w.bodies {
		if b.EntityID == entityID {
			w.bodies = append(w.bodies[:i], w.bodies[i+1:]...)
			return
		}
	}
}

// BodyFor returns the RigidBody registered for the given entity, or nil.
func (w *PhysicsWorld) BodyFor(entityID int) *RigidBody {
	for _, b := range w.bodies {
		if b.EntityID == entityID {
			return b
		}
	}
	return nil
}

// Step advances the simulation by dt seconds.
// Call once per game frame (e.g., from runner/game.go Update).
func (w *PhysicsWorld) Step(dt float32) {
	for _, b := range w.bodies {
		if b.Type == BodyStatic {
			continue
		}

		// Reset contact flags
		b.IsGrounded = false
		b.IsTouching = [4]bool{}

		// Apply gravity
		if b.Type == BodyDynamic {
			b.Vel.Y += w.Gravity * b.Gravity * dt
			if b.Vel.Y > TerminalVelocity {
				b.Vel.Y = TerminalVelocity
			}
		}

		// Integrate position
		b.Pos.X += b.Vel.X * dt
		b.Pos.Y += b.Vel.Y * dt
	}

	// Body-body collision (O(n²) — fine for small entity counts)
	// Check all pairs (i, j) where i < j to avoid duplicates
	for i := 0; i < len(w.bodies); i++ {
		for j := i + 1; j < len(w.bodies); j++ {
			a := w.bodies[i]
			b := w.bodies[j]

			// Skip static-static pairs (they never move)
			if a.Type == BodyStatic && b.Type == BodyStatic {
				continue
			}

			// Dynamic-dynamic: resolve both directions
			if a.Type == BodyDynamic && b.Type == BodyDynamic {
				ResolveAABB(a, b)
				ResolveAABB(b, a)
				continue
			}

			// One is dynamic, one is static/kinematic:
			// Only resolve on the dynamic body
			var dynamic, static *RigidBody
			if a.Type == BodyDynamic {
				dynamic, static = a, b
			} else {
				// b is dynamic (a is static or kinematic)
				dynamic, static = b, a
			}
			if dynamic.Type == BodyKinematic {
				// Kinematic doesn't react to collisions
				continue
			}
			ResolveAABB(dynamic, static)
		}
	}

	// Tilemap collision
	if w.tileQ != nil {
		for _, b := range w.bodies {
			if b.Type == BodyStatic {
				continue
			}
			w.resolveTiles(b)
		}
	}
}

// resolveTiles probes corner pixels and pushes the body out of solid tiles.
func (w *PhysicsWorld) resolveTiles(b *RigidBody) {
	const tileSize float32 = 32

	minX, minY, maxX, maxY := b.AABB()

	// Bottom-center probe (landing)
	if w.tileQ(b.Pos.X, maxY+1) {
		// Snap to tile top
		tileTop := float32(int((maxY)/tileSize)) * tileSize
		b.Pos.Y = tileTop - b.HalfH
		if b.Vel.Y > 0 {
			b.Vel.Y = 0
		}
		b.IsGrounded = true
		b.IsTouching[3] = true
	}

	// Top-center probe (ceiling)
	if w.tileQ(b.Pos.X, minY-1) {
		tileBot := float32(int((minY)/tileSize)+1) * tileSize
		b.Pos.Y = tileBot + b.HalfH
		if b.Vel.Y < 0 {
			b.Vel.Y = 0
		}
		b.IsTouching[2] = true
	}

	// Right probe (wall right)
	if w.tileQ(maxX+1, b.Pos.Y) {
		tileLeft := float32(int((maxX)/tileSize)) * tileSize
		b.Pos.X = tileLeft - b.HalfW
		if b.Vel.X > 0 {
			b.Vel.X = 0
		}
		b.IsTouching[1] = true
	}

	// Left probe (wall left)
	if w.tileQ(minX-1, b.Pos.Y) {
		tileRight := float32(int((minX)/tileSize)+1) * tileSize
		b.Pos.X = tileRight + b.HalfW
		if b.Vel.X < 0 {
			b.Vel.X = 0
		}
		b.IsTouching[0] = true
	}

	_ = minX // silence unused warning when only some probes fire
	_ = minY
}
