package physics

import "math"

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

// PhysicsWorld manages all bodies, joints and areas and drives the simulation at fixed 60 TPS.
type PhysicsWorld struct {
	Gravity      Vec2
	accumulator  float32
	bodies       []*RigidBody
	areas        []*Area2D
	joints       []*Joint
	tileQ        TileQuery
	spatial      *SpatialHash // broad-phase spatial hash (cellSize = 64)
	CCDThreshold float64      // minimum speed for CCD activation (0 = disabled)

	// preMovePos stores each body's position before position integration,
	// used by StepCCD to reconstruct the actual swept path.
	preMovePos map[int]Vec2
}

// NewWorld creates a PhysicsWorld with standard gravity (0, 980).
// tileQuery may be nil if no tilemap is used.
func NewWorld(tileQuery TileQuery) *PhysicsWorld {
	return &PhysicsWorld{
		Gravity:     Vec2{DefaultGravityX, DefaultGravityY},
		bodies:      make([]*RigidBody, 0),
		areas:       make([]*Area2D, 0),
		joints:      make([]*Joint, 0),
		tileQ:       tileQuery,
		spatial:     NewSpatialHash(64),
		preMovePos:  make(map[int]Vec2),
	}
}

// SetGravity updates the global gravity vector.
func (w *PhysicsWorld) SetGravity(x, y float32) {
	w.Gravity = Vec2{x, y}
}

// SetCCDThreshold sets the velocity threshold for CCD activation.
// Bodies with velocity magnitude above this threshold use CCD.
// Set to 0 (default) to disable CCD.
func (w *PhysicsWorld) SetCCDThreshold(threshold float64) {
	w.CCDThreshold = threshold
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

// AddJoint adds a constraint joint to the world.
func (w *PhysicsWorld) AddJoint(j *Joint) {
	w.joints = append(w.joints, j)
}

// RemoveJoint removes a joint from the world.
func (w *PhysicsWorld) RemoveJoint(j *Joint) {
	for i, joint := range w.joints {
		if joint == j {
			w.joints = append(w.joints[:i], w.joints[i+1:]...)
			return
		}
	}
}

// JointCount returns the number of active joints.
func (w *PhysicsWorld) JointCount() int {
	count := 0
	for _, j := range w.joints {
		if j.Active {
			count++
		}
	}
	return count
}

// SolveJoints resolves all active joints.
// Called after collision resolution each fixed step.
func (w *PhysicsWorld) SolveJoints() {
	for _, j := range w.joints {
		j.solve(FixedPhysicsStep)
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

// StepCCD performs continuous collision detection for fast-moving bodies.
// It is called after the regular physics step to detect and resolve tunneling.
func (w *PhysicsWorld) StepCCD(dt float64) {
	if w.CCDThreshold <= 0 {
		return
	}
	for _, a := range w.bodies {
		if a.Type != BodyDynamic {
			continue
		}
		// Reconstruct the actual displacement that occurred during step
		// from the pre-move position saved before position integration.
		prevPos, ok := w.preMovePos[a.EntityID]
		if !ok {
			continue
		}
		dispX := float64(a.Pos.X - prevPos.X)
		dispY := float64(a.Pos.Y - prevPos.Y)
		dispMag := math.Sqrt(dispX*dispX + dispY*dispY)

		// Effective velocity during the step
		velMag := dispMag / dt
		if velMag <= w.CCDThreshold {
			continue
		}

		// Compute minimum dimension (same check as NeedsCCD, but using actual
		// displacement rather than body.Vel which may have been modified by
		// collision resolution).
		var minDim float64
		switch a.Shape {
		case ShapeCircle:
			minDim = float64(a.Radius * 2)
		default:
			w := float64(a.HalfW * 2)
			h := float64(a.HalfH * 2)
			if w < h {
				minDim = w
			} else {
				minDim = h
			}
		}
		if dispMag <= minDim {
			continue
		}

		var prevMinA, prevMaxA Vec2
		switch a.Shape {
		case ShapeCircle:
			prevMinA = Vec2{prevPos.X - a.Radius, prevPos.Y - a.Radius}
			prevMaxA = Vec2{prevPos.X + a.Radius, prevPos.Y + a.Radius}
		default:
			prevMinA = Vec2{prevPos.X - a.HalfW, prevPos.Y - a.HalfH}
			prevMaxA = Vec2{prevPos.X + a.HalfW, prevPos.Y + a.HalfH}
		}

		for _, b := range w.bodies {
			if b == a {
				continue
			}
			// Skip if layers don't match masks
			if (a.Layer&b.Mask) == 0 || (b.Layer&a.Mask) == 0 {
				continue
			}

			minB, minBY, maxB, maxBY := b.AABB()

			result := SweptAABB(prevMinA, prevMaxA, Vec2{minB, minBY}, Vec2{maxB, maxBY}, dispX, dispY)

			if result.Hit && result.Time > 0 && result.Time <= 1 {
				// Reposition the body to the collision point
				a.Pos.X = float32(result.PointX)
				a.Pos.Y = float32(result.PointY)

				// Zero velocity on the colliding axis and set contact flags
				if result.NormalX != 0 {
					a.Vel.X = 0
					if result.NormalX < 0 {
						a.IsTouching[1] = true // Right
					} else {
						a.IsTouching[0] = true // Left
					}
				}
				if result.NormalY != 0 {
					a.Vel.Y = 0
					if result.NormalY < 0 {
						a.IsTouching[2] = true // Top (grounded)
						a.IsGrounded = true
					} else {
						a.IsTouching[3] = true // Bottom (ceiling)
					}
				}

				// Fire callback
				if a.OnCollision != nil {
					a.OnCollision(b, Vec2{float32(result.NormalX), float32(result.NormalY)})
				}
				break
			}
		}
	}
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
	// Clear pre-move positions from previous step
	for k := range w.preMovePos {
		delete(w.preMovePos, k)
	}

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

		// Save pre-move position for CCD (before position integration)
		w.preMovePos[b.EntityID] = b.Pos

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

	// Joint constraint resolution (after collision, before area events)
	w.SolveJoints()

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

	// Continuous collision detection for fast-moving bodies
	w.StepCCD(float64(dt))
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
