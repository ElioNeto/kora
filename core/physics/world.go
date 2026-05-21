package physics

const (
	// DefaultGravityX/Y defines standard gravity in px/s² (down = +Y).
	DefaultGravityX float32 = 0
	DefaultGravityY float32 = 980
	// FixedPhysicsStep is the fixed timestep for physics (60 TPS).
	FixedPhysicsStep float32 = 1.0 / 60.0
	// TerminalVelocity caps the falling speed.
	TerminalVelocity float32 = 1200
)

// TileQuery is a function the world calls to know whether a tile at
// world position (px, py) is solid. Provided by core/render/tilemap.
type TileQuery func(px, py float32) bool

// NodeBody is an interface that physics-capable game nodes can implement
// to register themselves with a PhysicsWorld. This allows decoupling between
// the core/node and core/physics packages (no circular imports).
type NodeBody interface {
	GetPos() (float32, float32)
	GetVelocity() (float32, float32)
	GetMass() float32
	GetShape() (ShapeType, float32, float32) // shape type, width, height (or radius for circles)
	GetCollisionLayer() uint16
	GetCollisionMask() uint16
	IsSolid() bool
}

// PhysicsWorld manages all bodies and areas and drives the simulation at fixed 60 TPS.
type PhysicsWorld struct {
	Gravity     Vec2
	accumulator float32
	bodies      []*RigidBody
	areas       []*Area2D
	tileQ       TileQuery
	spatial     *SpatialHash // broad-phase spatial hash (cellSize = 64)
}

// NewWorld creates a PhysicsWorld with standard gravity (0, 980).
// tileQuery may be nil if no tilemap is used.
func NewWorld(tileQuery TileQuery) *PhysicsWorld {
	return &PhysicsWorld{
		Gravity: Vec2{DefaultGravityX, DefaultGravityY},
		bodies:  make([]*RigidBody, 0),
		areas:   make([]*Area2D, 0),
		tileQ:   tileQuery,
		spatial: NewSpatialHash(64),
	}
}

// SetGravity updates the global gravity vector.
func (w *PhysicsWorld) SetGravity(x, y float32) {
	w.Gravity = Vec2{x, y}
}

// OverlapRect returns all bodies overlapping the given rectangle that match the mask.
// rect is defined by minX, minY, maxX, maxY.
func (w *PhysicsWorld) OverlapRect(minX, minY, maxX, maxY float32, mask uint16) []*RigidBody {
	var result []*RigidBody
	for _, b := range w.bodies {
		if (b.Layer & mask) == 0 {
			continue
		}
		bMinX, bMinY, bMaxX, bMaxY := b.AABB()
		// Check AABB overlap
		if bMaxX <= minX || bMinX >= maxX || bMaxY <= minY || bMinY >= maxY {
			continue
		}
		result = append(result, b)
	}
	return result
}

// GetBodies returns all registered bodies (for CharacterBody2D collision checks).
func (w *PhysicsWorld) GetBodies() []*RigidBody {
	return w.bodies
}

// Register adds a body to the simulation.
func (w *PhysicsWorld) Register(b *RigidBody) {
	w.bodies = append(w.bodies, b)
}

// RegisterArea adds an area to the simulation.
func (w *PhysicsWorld) RegisterArea(a *Area2D) {
	w.areas = append(w.areas, a)
}

// RegisterNodeBody registers a node-level body (via the NodeBody interface).
// It creates an internal RigidBody and adds it to the simulation.
// Returns the entity ID assigned to the new body.
func (w *PhysicsWorld) RegisterNodeBody(nb NodeBody) int {
	x, y := nb.GetPos()
	shape, w_, h_ := nb.GetShape()

	bodyType := BodyDynamic
	if !nb.IsSolid() {
		bodyType = BodyKinematic
	}

	entityID := len(w.bodies) + 1
	body := &RigidBody{
		EntityID: entityID,
		Pos:      Vec2{x, y},
		Shape:    shape,
		Mass:     nb.GetMass(),
		Type:     bodyType,
		Layer:    nb.GetCollisionLayer(),
		Mask:     nb.GetCollisionMask(),
		Gravity:  1,
	}

	switch shape {
	case ShapeCircle:
		body.Radius = w_ / 2
	default: // ShapeRect
		body.HalfW = w_ / 2
		body.HalfH = h_ / 2
	}

	w.bodies = append(w.bodies, body)
	return entityID
}

// UnregisterNodeBody removes a node-level body from the simulation by entity ID.
func (w *PhysicsWorld) UnregisterNodeBody(entityID int) {
	w.Remove(entityID)
}

// Remove detaches a body from the simulation.
func (w *PhysicsWorld) Remove(entityID int) {
	for i, b := range w.bodies {
		if b.EntityID == entityID {
			w.bodies = append(w.bodies[:i], w.bodies[i+1:]...)
			return
		}
	}
	
	// Also check areas
	for i, a := range w.areas {
		if a.EntityID == entityID {
			w.areas = append(w.areas[:i], w.areas[i+1:]...)
			return
		}
	}
}

// RemoveArea detaches an area from the simulation.
func (w *PhysicsWorld) RemoveArea(entityID int) {
	for i, a := range w.areas {
		if a.EntityID == entityID {
			w.areas = append(w.areas[:i], w.areas[i+1:]...)
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

// Step accumulates frame dt and runs fixed 60 TPS physics steps.
// Call once per render frame with the actual delta time.
func (w *PhysicsWorld) Step(dt float32) {
	w.accumulator += dt
	for w.accumulator >= FixedPhysicsStep {
		w.accumulator -= FixedPhysicsStep
		w.stepFixed(FixedPhysicsStep)
	}
}

// stepFixed runs one fixed physics step with the given dt (1/60s).
func (w *PhysicsWorld) stepFixed(dt float32) {
	for _, b := range w.bodies {
		if b.Type == BodyStatic {
			continue
		}

		// Reset contact flags
		b.IsGrounded = false
		b.IsTouching = [4]bool{}

		// Apply gravity (2D)
		if b.Type == BodyDynamic {
			b.Vel.X += w.Gravity.X * b.Gravity * dt
			b.Vel.Y += w.Gravity.Y * b.Gravity * dt
			if b.Vel.Y > TerminalVelocity {
				b.Vel.Y = TerminalVelocity
			}
		}

		// Integrate position
		b.Pos.X += b.Vel.X * dt
		b.Pos.Y += b.Vel.Y * dt
	}

	// Broad-phase spatial hash (amortised O(n))
	w.spatial.Clear()
	for _, b := range w.bodies {
		w.spatial.Insert(b.EntityID, b)
	}

	// Build entity ID → index lookup for duplicate-pair avoidance
	idToIdx := make(map[int]int, len(w.bodies))
	for i, b := range w.bodies {
		idToIdx[b.EntityID] = i
	}

	// Narrow-phase collision using candidate pairs from spatial hash
	for i, a := range w.bodies {
		candidates := w.spatial.GetCandidates(a.EntityID, a)
		for _, cid := range candidates {
			j, ok := idToIdx[cid]
			if !ok || j <= i {
				continue
			}
			b := w.bodies[j]

			// Skip if layers don't match masks
			if (a.Layer&b.Mask) == 0 || (b.Layer&a.Mask) == 0 {
				continue
			}

			if a.Type == BodyStatic && b.Type == BodyStatic {
				continue
			}

			if a.Type == BodyDynamic && b.Type == BodyDynamic {
				ResolveCollision(a, b)
				ResolveCollision(b, a)
				continue
			}

			var dynamic, other *RigidBody
			if a.Type == BodyDynamic {
				dynamic, other = a, b
			} else {
				dynamic, other = b, a
			}
			if dynamic.Type == BodyKinematic {
				continue
			}
			ResolveCollision(dynamic, other)
		}
	}

	// Area-body monitoring
	for _, area := range w.areas {
		if area.MonitorEnabled {
			area.CheckOverlaps(w)
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
