package node

import (
	stdMath "math"
	"sync/atomic"

	"github.com/ElioNeto/kora/core/audio"
	"github.com/ElioNeto/kora/core/math"
	"github.com/ElioNeto/kora/core/physics"
	"github.com/hajimehoshi/ebiten/v2"
)

// nextNodeBodyID provides unique entity IDs for PhysicsBody2D bodies.
var nextNodeBodyID int32 = 1000

func allocNodeBodyID() int {
	return int(atomic.AddInt32(&nextNodeBodyID, 1))
}

// PhysicsBody2D is a node that wraps core/physics types.
// It implements scene.PhysicsNode for SceneTree integration.
type PhysicsBody2D struct {
	*Node2D
	physicsBody interface{} // Holds the core/physics body type

	// Common physics properties
	mass         float64
	friction     float64
	restitution  float64
	gravityScale float64
	linearVelX   float64
	linearVelY   float64
	angularVel   float64
	solid        bool
	trigger      bool
	collisionGroup int

	// Physics world integration (optional — nil means no physics simulation)
	physicsWorld *physics.PhysicsWorld
	physBody     *physics.RigidBody // The actual body registered in PhysicsWorld
	bodyID       int                // Entity ID in the physics world

	// Shape dimensions (for creating the physics body)
	width  float64
	height float64

	// Collision filtering
	collisionLayer uint16
	collisionMask  uint16

	// Callbacks
	onCollision          func(other *Node2D, eventType CollisionType)
	onEnterOverlap       func(other *Node2D)
	onExitOverlap        func(other *Node2D)
}

// NewPhysicsBody2D creates a new PhysicsBody2D node
func NewPhysicsBody2D(name string) *PhysicsBody2D {
	node := NewNode2D(name, 0)
	return &PhysicsBody2D{
		Node2D:         node,
		mass:           1.0,           // Default mass
		friction:       0.1,           // Default friction
		restitution:    0.0,           // Default restitution (no bounce)
		gravityScale:   1.0,           // Default gravity scale
		linearVelX:     0.0,           // Initial velocity zero
		linearVelY:     0.0,
		angularVel:     0.0,           // Initial angular velocity zero
		solid:          true,          // Default to solid (collidable)
		trigger:        false,         // Not a trigger by default
		collisionGroup: 0,             // Default collision group
		width:          32,            // Default shape size
		height:         32,
		collisionLayer: uint16(physics.DefaultLayer),
		collisionMask:  uint16(physics.DefaultMask),
		bodyID:         -1,            // No physics body yet
		onCollision:          nil,
		onEnterOverlap:       nil,
		onExitOverlap:        nil,
	}
}

// AddChild satisfies Node interface
func (p *PhysicsBody2D) AddChild(child Node) {
	p.Node2D.AddChild(child)
}

// SetMass sets the body mass (0 = infinite/static)
func (p *PhysicsBody2D) SetMass(mass float64) {
	p.mass = mass
	if p.physicsWorld != nil {
		p.SyncToWorld()
	}
}

// GetMass returns current mass
func (p *PhysicsBody2D) GetMass() float64 {
	return p.mass
}

// SetFriction sets the friction coefficient (0 to 1)
func (p *PhysicsBody2D) SetFriction(friction float64) {
	if friction < 0 {
		friction = 0
	}
	if friction > 1 {
		friction = 1
	}
	p.friction = friction
	if p.physicsWorld != nil {
		p.SyncToWorld()
	}
}

// GetFriction returns current friction
func (p *PhysicsBody2D) GetFriction() float64 {
	return p.friction
}

// SetRestitution sets the bounciness (0 to 1)
func (p *PhysicsBody2D) SetRestitution(restitution float64) {
	if restitution < 0 {
		restitution = 0
	}
	if restitution > 1 {
		restitution = 1
	}
	p.restitution = restitution
	if p.physicsWorld != nil {
		p.SyncToWorld()
	}
}

// GetRestitution returns current restitution
func (p *PhysicsBody2D) GetRestitution() float64 {
	return p.restitution
}

// SetGravityScale sets gravity multiplier
func (p *PhysicsBody2D) SetGravityScale(scale float64) {
	p.gravityScale = scale
	if p.physicsWorld != nil {
		p.SyncToWorld()
	}
}

// GetGravityScale returns gravity scale
func (p *PhysicsBody2D) GetGravityScale() float64 {
	return p.gravityScale
}

// SetLinearVelocity sets the linear velocity
func (p *PhysicsBody2D) SetLinearVelocity(vx, vy float64) {
	p.linearVelX = vx
	p.linearVelY = vy
	if p.physicsWorld != nil {
		p.SyncToWorld()
	}
}

// GetLinearVelocity returns linear velocity
func (p *PhysicsBody2D) GetLinearVelocity() (float64, float64) {
	return p.linearVelX, p.linearVelY
}

// AddLinearVelocity adds to current linear velocity
func (p *PhysicsBody2D) AddLinearVelocity(vx, vy float64) {
	p.linearVelX += vx
	p.linearVelY += vy
	if p.physicsWorld != nil {
		p.SyncToWorld()
	}
}

// SetAngularVelocity sets rotational velocity (radians per second)
func (p *PhysicsBody2D) SetAngularVelocity(omega float64) {
	p.angularVel = omega
	if p.physicsWorld != nil {
		p.SyncToWorld()
	}
}

// GetAngularVelocity returns angular velocity
func (p *PhysicsBody2D) GetAngularVelocity() float64 {
	return p.angularVel
}

// SetSolid determines if body is solid (collidable)
func (p *PhysicsBody2D) SetSolid(solid bool) {
	p.solid = solid
	p.trigger = !solid
	if p.physicsWorld != nil {
		p.SyncToWorld()
	}
}

// IsSolid returns if body is solid
func (p *PhysicsBody2D) IsSolid() bool {
	return p.solid && !p.trigger
}

// SetTrigger sets body as trigger zone (non-solid)
func (p *PhysicsBody2D) SetTrigger(trigger bool) {
	p.trigger = trigger
	p.solid = !trigger
	if p.physicsWorld != nil {
		p.SyncToWorld()
	}
}

// IsTrigger returns if body is a trigger
func (p *PhysicsBody2D) IsTrigger() bool {
	return p.trigger
}

// ApplyForce applies a force at the center of mass
func (p *PhysicsBody2D) ApplyForce(fx, fy float64) {
	p.linearVelX += fx / p.mass
	p.linearVelY += fy / p.mass
	if p.physicsWorld != nil {
		p.SyncToWorld()
	}
}

// ApplyImpulse applies an instantaneous impulse
func (p *PhysicsBody2D) ApplyImpulse(impX, impY float64) {
	p.linearVelX += impX / p.mass
	p.linearVelY += impY / p.mass
	if p.physicsWorld != nil {
		p.SyncToWorld()
	}
}

// ApplyTorque applies rotational torque
func (p *PhysicsBody2D) ApplyTorque(torque float64) {
	p.angularVel += torque / p.mass // Simplified
	if p.physicsWorld != nil {
		p.SyncToWorld()
	}
}

// SetCollisionGroup sets the collision group
func (p *PhysicsBody2D) SetCollisionGroup(group int) {
	p.collisionGroup = group
}

// GetCollisionGroup returns collision group
func (p *PhysicsBody2D) GetCollisionGroup() int {
	return p.collisionGroup
}

// SetOnCollision sets the collision callback
func (p *PhysicsBody2D) SetOnCollision(fn func(other *Node2D, eventType CollisionType)) {
	p.onCollision = fn
}

// SetOnEnterOverlap sets the overlap enter callback
func (p *PhysicsBody2D) SetOnEnterOverlap(fn func(other *Node2D)) {
	p.onEnterOverlap = fn
}

// SetOnExitOverlap sets the overlap exit callback
func (p *PhysicsBody2D) SetOnExitOverlap(fn func(other *Node2D)) {
	p.onExitOverlap = fn
}

// TriggerCollision manually triggers collision event
func (p *PhysicsBody2D) TriggerCollision(other *Node2D, eventType CollisionType) {
	if p.onCollision != nil {
		p.onCollision(other, eventType)
	}
}

// TriggerOverlapEnter manually triggers overlap enter
func (p *PhysicsBody2D) TriggerOverlapEnter(other *Node2D) {
	if p.onEnterOverlap != nil {
		p.onEnterOverlap(other)
	}
}

// TriggerOverlapExit manually triggers overlap exit
func (p *PhysicsBody2D) TriggerOverlapExit(other *Node2D) {
	if p.onExitOverlap != nil {
		p.onExitOverlap(other)
	}
}

// Update processes physics simulation for this body.
// If the body is attached to a PhysicsWorld, it syncs position and velocity
// from the physics engine (no self-integration).
func (p *PhysicsBody2D) Update(dt float64) {
	// Sync from physics world (if attached)
	if p.physicsWorld != nil && p.physBody != nil {
		p.SyncFromWorld()
	}

	// Propagate to children
	for _, child := range p.children {
		if child != nil {
			child.Update(dt)
		}
	}
}

// ---------------------------------------------------------------------------
// PhysicsWorld integration
// ---------------------------------------------------------------------------

// SetPhysicsWorld attaches (or detaches) this body to/from a PhysicsWorld.
// If world is non-nil, a RigidBody is created in the physics world and the
// node's current properties are synced to it.
// If world is nil, the body is removed from its previous world.
func (p *PhysicsBody2D) SetPhysicsWorld(world *physics.PhysicsWorld) {
	// Unregister from old world
	if p.physicsWorld != nil && p.physBody != nil {
		p.physicsWorld.Remove(p.bodyID)
	}

	p.physicsWorld = world

	if world != nil {
		// Generate a unique entity ID if not already set
		if p.bodyID < 0 {
			p.bodyID = allocNodeBodyID()
		}
		p.SyncToWorld()
	} else {
		p.physBody = nil
		p.bodyID = -1
	}
}

// SyncToWorld copies the node's current properties to the physics body.
// If no physics body exists yet, one is created and registered.
func (p *PhysicsBody2D) SyncToWorld() {
	if p.physicsWorld == nil {
		return
	}

	pos := p.GetPosition()

	// Determine body type
	bodyType := physics.BodyDynamic
	if !p.solid || p.trigger || p.mass == 0 {
		bodyType = physics.BodyStatic
	}

	// Create physics body if it doesn't exist
	if p.physBody == nil {
		p.physBody = physics.NewBody(p.bodyID, pos.X, pos.Y, float32(p.width), float32(p.height), bodyType)
		p.physBody.NodeRef = p.Node2D

		// Bridge physics collision callback to node-level callback.
		// When the physics engine detects a collision, it fires the
		// physics.RigidBody.OnCollision callback, which we translate
		// to the node's onCollision callback with the correct Node2D reference.
		p.physBody.OnCollision = func(other *physics.RigidBody, normal physics.Vec2) {
			if other != nil && other.NodeRef != nil {
				if otherNode, ok := other.NodeRef.(*Node2D); ok {
					if p.onCollision != nil {
						p.onCollision(otherNode, CollisionTypeCollide)
					}
				}
			}
		}
		p.physicsWorld.Register(p.physBody)
	}

	// Sync all properties
	p.physBody.Pos = physics.Vec2{X: pos.X, Y: pos.Y}
	p.physBody.Vel = physics.Vec2{X: float32(p.linearVelX), Y: float32(p.linearVelY)}
	p.physBody.Mass = float32(p.mass)
	p.physBody.Gravity = float32(p.gravityScale)
	p.physBody.Type = bodyType
	p.physBody.Layer = p.collisionLayer
	p.physBody.Mask = p.collisionMask
	p.physBody.HalfW = float32(p.width) / 2
	p.physBody.HalfH = float32(p.height) / 2

	// Ensure NodeRef is set (in case body was created elsewhere)
	if p.physBody.NodeRef == nil {
		p.physBody.NodeRef = p.Node2D
	}
}

// SyncFromWorld copies the physics body's state back to the node.
// Called every frame in Update() when attached to a PhysicsWorld.
func (p *PhysicsBody2D) SyncFromWorld() {
	if p.physBody == nil {
		return
	}

	// Copy position from physics body to node
	p.SetPosition(p.physBody.Pos.X, p.physBody.Pos.Y)

	// Copy velocity
	p.linearVelX = float64(p.physBody.Vel.X)
	p.linearVelY = float64(p.physBody.Vel.Y)
}

// ---------------------------------------------------------------------------
// Shape dimensions
// ---------------------------------------------------------------------------

// SetSize sets the collision shape dimensions for this body.
func (p *PhysicsBody2D) SetSize(w, h float64) {
	p.width = w
	p.height = h
	if p.physicsWorld != nil {
		p.SyncToWorld()
	}
}

// GetSize returns the collision shape dimensions.
func (p *PhysicsBody2D) GetSize() (float64, float64) {
	return p.width, p.height
}

// ---------------------------------------------------------------------------
// Collision filtering
// ---------------------------------------------------------------------------

// SetCollisionLayer sets the collision layer bitmask.
func (p *PhysicsBody2D) SetCollisionLayer(layer uint16) {
	p.collisionLayer = layer
	if p.physicsWorld != nil {
		p.SyncToWorld()
	}
}

// GetCollisionLayer returns the collision layer bitmask.
func (p *PhysicsBody2D) GetCollisionLayer() uint16 {
	return p.collisionLayer
}

// SetCollisionMask sets the collision mask bitmask (which layers to collide with).
func (p *PhysicsBody2D) SetCollisionMask(mask uint16) {
	p.collisionMask = mask
	if p.physicsWorld != nil {
		p.SyncToWorld()
	}
}

// GetCollisionMask returns the collision mask bitmask.
func (p *PhysicsBody2D) GetCollisionMask() uint16 {
	return p.collisionMask
}

// RigidBody2D is a dynamic physics body
type RigidBody2D struct {
	*PhysicsBody2D
}

// NewRigidBody2D creates a new RigidBody2D
func NewRigidBody2D(name string) *RigidBody2D {
	pb := NewPhysicsBody2D(name)
	return &RigidBody2D{
		PhysicsBody2D: pb,
	}
}

// StaticBody2D is a stationary physics body
type StaticBody2D struct {
	*PhysicsBody2D
}

// NewStaticBody2D creates a new StaticBody2D
func NewStaticBody2D(name string) *StaticBody2D {
	pb := NewPhysicsBody2D(name)
	pb.SetMass(0) // Static bodies have infinite mass
	return &StaticBody2D{
		PhysicsBody2D: pb,
	}
}

// Area2D is a sensor/trigger area for collision detection
type Area2D struct {
	*Node2D

	// Overlapping bodies
	overlaps []*Node2D

	// Callbacks
	onEnter   func(other *Node2D)
	onExit    func(other *Node2D)
	onOverlap func(other *Node2D)

	// Physics world integration (optional — delegates overlap detection to PhysicsWorld)
	physicsWorld   *physics.PhysicsWorld
	width          float64
	height         float64
	collisionLayer uint16
	collisionMask  uint16
}

// NewArea2D creates a new Area2D
func NewArea2D(name string) *Area2D {
	return &Area2D{
		Node2D:         NewNode2D(name, 0),
		overlaps:       make([]*Node2D, 0),
		width:          32,
		height:         32,
		collisionLayer: uint16(physics.DefaultLayer),
		collisionMask:  uint16(physics.DefaultMask),
	}
}

// GetOverlaps returns all overlapping bodies
func (a *Area2D) GetOverlaps() []*Node2D {
	return a.overlaps
}

// GetOverlapCount returns number of overlaps
func (a *Area2D) GetOverlapCount() int {
	return len(a.overlaps)
}

// OnEnter sets the enter callback
func (a *Area2D) OnEnter(fn func(other *Node2D)) {
	a.onEnter = fn
}

// OnExit sets the exit callback
func (a *Area2D) OnExit(fn func(other *Node2D)) {
	a.onExit = fn
}

// OnOverlap sets the overlap callback
func (a *Area2D) OnOverlap(fn func(other *Node2D)) {
	a.onOverlap = fn
}

// AddOverlap adds an overlapping body
func (a *Area2D) AddOverlap(body *Node2D) {
	// Check already in list
	for _, o := range a.overlaps {
		if o == body {
			return
		}
	}

	a.overlaps = append(a.overlaps, body)

	if a.onEnter != nil {
		a.onEnter(body)
	}
	if a.onOverlap != nil {
		a.onOverlap(body)
	}
}

// RemoveOverlap removes an overlapping body
func (a *Area2D) RemoveOverlap(body *Node2D) {
	for i, o := range a.overlaps {
		if o == body {
			a.overlaps = append(a.overlaps[:i], a.overlaps[i+1:]...)
			if a.onExit != nil {
				a.onExit(body)
			}
			return
		}
	}
}

// ClearOverlaps removes all overlaps
func (a *Area2D) ClearOverlaps() {
	for _, body := range a.overlaps {
		if a.onExit != nil {
			a.onExit(body)
		}
	}
	a.overlaps = make([]*Node2D, 0)
}

// ---------------------------------------------------------------------------
// Area2D PhysicsWorld integration
// ---------------------------------------------------------------------------

// SetPhysicsWorld attaches this area to a PhysicsWorld for overlap detection.
func (a *Area2D) SetPhysicsWorld(world *physics.PhysicsWorld) {
	a.physicsWorld = world
}

// SetSize sets the area's detection zone dimensions.
func (a *Area2D) SetSize(w, h float64) {
	a.width = w
	a.height = h
}

// GetSize returns the area's detection zone dimensions.
func (a *Area2D) GetSize() (float64, float64) {
	return a.width, a.height
}

// SetCollisionLayer sets the collision layer for overlap detection.
func (a *Area2D) SetCollisionLayer(layer uint16) {
	a.collisionLayer = layer
}

// GetCollisionLayer returns the collision layer.
func (a *Area2D) GetCollisionLayer() uint16 {
	return a.collisionLayer
}

// SetCollisionMask sets which layers this area detects overlaps with.
func (a *Area2D) SetCollisionMask(mask uint16) {
	a.collisionMask = mask
}

// GetCollisionMask returns the collision mask.
func (a *Area2D) GetCollisionMask() uint16 {
	return a.collisionMask
}

// SyncWithPhysics uses the attached PhysicsWorld to detect overlapping bodies.
// It replaces the manual overlap detection with the physics engine's AABB checks.
// Should be called once per frame (e.g., in Update or after PhysicsWorld.Step).
func (a *Area2D) SyncWithPhysics() {
	if a.physicsWorld == nil {
		return
	}

	pos := a.GetPosition()
	halfW := float32(a.width * 0.5)
	halfH := float32(a.height * 0.5)
	minX := float32(pos.X) - halfW
	minY := float32(pos.Y) - halfH
	maxX := float32(pos.X) + halfW
	maxY := float32(pos.Y) + halfH

	// Query physics world for overlapping bodies (by area's collision mask)
	bodies := a.physicsWorld.OverlapRect(minX, minY, maxX, maxY, a.collisionMask)

	// Filter by reverse direction: the body must also accept collisions with
	// the area's layer (matching physics.Area2D bidirectional filter).
	var filtered []*physics.RigidBody
	for _, b := range bodies {
		if (a.collisionLayer & b.Mask) != 0 {
			filtered = append(filtered, b)
		}
	}

	// Build set of currently overlapping Node2D references
	currentNodes := make(map[*Node2D]bool)
	for _, b := range filtered {
		if ref, ok := b.NodeRef.(*Node2D); ok {
			currentNodes[ref] = true
		}
	}

	// Find new overlaps (in physics but not in our list)
	for node := range currentNodes {
		found := false
		for _, existing := range a.overlaps {
			if existing == node {
				found = true
				break
			}
		}
		if !found {
			a.overlaps = append(a.overlaps, node)
			if a.onEnter != nil {
				a.onEnter(node)
			}
			if a.onOverlap != nil {
				a.onOverlap(node)
			}
		}
	}

	// Find removed overlaps (in our list but not in physics)
	var remaining []*Node2D
	for _, existing := range a.overlaps {
		if currentNodes[existing] {
			remaining = append(remaining, existing)
		} else {
			if a.onExit != nil {
				a.onExit(existing)
			}
		}
	}
	a.overlaps = remaining
}

// Camera2D controls the camera view
type Camera2D struct {
	*Node2D

	// Zoom level (1.0 = normal)
	Zoom float64

	// Target to follow
	target       *Node2D
	targetOffset math.Vector2

	// Smoothing
	smoothingFactor float64
	lerpMode        bool

	// Camera bounds
	minBounds math.Vector2
	maxBounds math.Vector2

	// Shake state (non-blocking, applied each frame in Update)
	shaking        bool
	shakeIntensity float64
	shakeDuration  float64
	shakeElapsed   float64
	shakeOffset    math.Vector2
	
	// Viewport size (set by runner)
	viewportW float64
	viewportH float64
}

// NewCamera2D creates a new Camera2D
func NewCamera2D(name string) *Camera2D {
	return &Camera2D{
		Node2D:          NewNode2D(name, 0),
		Zoom:            1.0,
		smoothingFactor: 0.1,
		lerpMode:        true,
		viewportW:       360,
		viewportH:       640,
	}
}

// SetTarget sets the target to follow
func (c *Camera2D) SetTarget(target *Node2D, offset math.Vector2) {
	c.target = target
	c.targetOffset = offset
}

// GetTarget returns current target
func (c *Camera2D) GetTarget() *Node2D {
	return c.target
}

// SetTargetOffset sets the offset from target
func (c *Camera2D) SetTargetOffset(offset math.Vector2) {
	c.targetOffset = offset
}

// GetTargetOffset returns current offset
func (c *Camera2D) GetTargetOffset() math.Vector2 {
	return c.targetOffset
}

// SetSmoothing sets the smoothing factor (0 to 1)
func (c *Camera2D) SetSmoothing(factor float64) {
	if factor < 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}
	c.smoothingFactor = factor
}

// GetSmoothing returns smoothing factor
func (c *Camera2D) GetSmoothing() float64 {
	return c.smoothingFactor
}

// SetBounds sets camera bounds
func (c *Camera2D) SetBounds(min, max math.Vector2) {
	c.minBounds = min
	c.maxBounds = max
}

// GetBounds returns current bounds
func (c *Camera2D) GetBounds() (math.Vector2, math.Vector2) {
	return c.minBounds, c.maxBounds
}

// OnScreenRect returns the visible rectangle in world coordinates
func (c *Camera2D) OnScreenRect() math.Rect {
	// Placeholder - would need viewport dimensions
	return math.Rect{}
}

// Update processes camera follow logic and applies shake.
func (c *Camera2D) Update(dt float64) {
	// Base update
	if c.Node2D != nil {
		c.Node2D.Update(dt)
	}

	// Follow target
	if c.target != nil {
		targetPos := c.target.GetWorldPosition()
		targetPos = targetPos.Add(c.targetOffset)
		currentPos := c.GetWorldPosition()

		if c.lerpMode && c.smoothingFactor > 0 {
			// Smooth interpolation
			pos := math.Lerp(currentPos, targetPos, c.smoothingFactor)
			c.SetWorldPosition(pos.X, pos.Y)
		} else {
			// Direct follow
			c.SetWorldPosition(targetPos.X, targetPos.Y)
		}

		// Apply bounds
		c.clampToBounds()
	}

	// Apply shake (non-blocking, decay over time)
	if c.shaking {
		c.shakeElapsed += dt
		if c.shakeElapsed >= c.shakeDuration {
			// Shake finished — reset position
			c.shaking = false
			c.shakeOffset = math.Vector2{}
			c.shakeIntensity = 0
			c.shakeElapsed = 0
		} else {
			// Decay intensity over time
			progress := c.shakeElapsed / c.shakeDuration
			currentIntensity := c.shakeIntensity * (1 - progress)

			// Apply pseudo-random offset based on elapsed time
			// Using sin/cos with different frequencies to create a shake pattern
			randX := stdMath.Sin(c.shakeElapsed*50)*currentIntensity - currentIntensity*0.5
			randY := stdMath.Cos(c.shakeElapsed*47)*currentIntensity - currentIntensity*0.5

			c.shakeOffset = math.Vector2{X: float32(randX), Y: float32(randY)}
		}
	}
}

// clampToBounds ensures camera stays within bounds
func (c *Camera2D) clampToBounds() {
	pos := c.GetWorldPosition()

	if c.minBounds.X > 0 && pos.X < c.minBounds.X {
		pos.X = c.minBounds.X
	}
	if c.minBounds.Y > 0 && pos.Y < c.minBounds.Y {
		pos.Y = c.minBounds.Y
	}
	if c.maxBounds.X > 0 && pos.X > c.maxBounds.X {
		pos.X = c.maxBounds.X
	}
	if c.maxBounds.Y > 0 && pos.Y > c.maxBounds.Y {
		pos.Y = c.maxBounds.Y
	}

	c.SetWorldPosition(pos.X, pos.Y)
}

// Shake starts a non-blocking camera shake effect.
// amount controls the maximum displacement in world units; duration is in seconds.
// The shake is applied each frame in Update() and decays to zero over duration.
func (c *Camera2D) Shake(amount float32, duration float64) {
	if duration <= 0 || amount <= 0 {
		return
	}
	c.shaking = true
	c.shakeIntensity = float64(amount)
	c.shakeDuration = duration
	c.shakeElapsed = 0
	c.shakeOffset = math.Vector2{}
}

// ---------------------------------------------------------------------------
// Camera transformation API
// ---------------------------------------------------------------------------

// SetViewport sets the logical viewport dimensions (screen size).
// Called by the runner during initialization.
func (c *Camera2D) SetViewport(w, h float64) {
	c.viewportW = w
	c.viewportH = h
}

// GetViewport returns the current viewport dimensions.
func (c *Camera2D) GetViewport() (float64, float64) {
	return c.viewportW, c.viewportH
}

// GetTransform returns the camera's view transformation matrix (world-to-screen).
// Apply this to ebiten.DrawImageOptions.GeoM before drawing game objects.
//
// The matrix:
//  1. Translates so the camera centre is at (0,0)
//  2. Scales by zoom
//  3. Translates to screen centre
//  4. Applies shake offset
func (c *Camera2D) GetTransform() ebiten.GeoM {
	var m ebiten.GeoM

	pos := c.GetWorldPosition()
	shakeOffset := c.GetShakeOffset()

	// 1. Translate world so camera centre is at origin
	m.Translate(float64(-pos.X-shakeOffset.X), float64(-pos.Y-shakeOffset.Y))

	// 2. Scale by zoom
	if c.Zoom != 0 {
		m.Scale(c.Zoom, c.Zoom)
	}

	// 3. Translate to screen centre
	m.Translate(c.viewportW/2, c.viewportH/2)

	return m
}

// GetShakeOffset returns the current shake offset vector.
func (c *Camera2D) GetShakeOffset() math.Vector2 {
	return c.shakeOffset
}

// WorldToScreen converts a world-space coordinate to screen-space pixel position.
func (c *Camera2D) WorldToScreen(wx, wy float64) (float64, float64) {
	m := c.GetTransform()
	sx, sy := m.Apply(wx, wy)
	return sx, sy
}

// ScreenToWorld converts a screen-space pixel position to world-space coordinate.
func (c *Camera2D) ScreenToWorld(sx, sy float64) (float64, float64) {
	m := c.GetTransform()
	if m.IsInvertible() {
		m.Invert()
		wx, wy := m.Apply(sx, sy)
		return wx, wy
	}
	// If matrix is not invertible (e.g. zoom=0), return origin
	return 0, 0
}

// AudioPlayer2D plays audio
type AudioPlayer2D struct {
	*Node2D

	// Audio properties
	soundName    string
	volume       float32
	isPlaying    bool
	looping      bool
	paused       bool

	// Connection to the real audio system
	channelID int // mixer channel ID (0 = not playing)
}

// NewAudioPlayer2D creates a new AudioPlayer2D
func NewAudioPlayer2D(name string) *AudioPlayer2D {
	return &AudioPlayer2D{
		Node2D:    NewNode2D(name, 0),
		volume:    1.0,
		isPlaying: false,
		channelID: 0,
	}
}

// SetSound sets the audio sound name and pre-loads the sound via the audio manager.
// If the audio manager is not initialised, the sound name is stored but not loaded.
func (a *AudioPlayer2D) SetSound(soundName string) {
	// Stop any currently playing sound
	if a.channelID != 0 {
		a.Stop()
	}
	a.soundName = soundName
	// Pre-load the sound via the audio manager (nil-safe)
	_ = audio.PreloadSound(soundName)
}

// GetSound returns current sound name
func (a *AudioPlayer2D) GetSound() string {
	return a.soundName
}

// SetVolume sets playback volume (0 to 1)
func (a *AudioPlayer2D) SetVolume(volume float32) {
	if volume < 0 {
		volume = 0
	}
	if volume > 1 {
		volume = 1
	}
	a.volume = volume
}

// GetVolume returns current volume
func (a *AudioPlayer2D) GetVolume() float32 {
	return a.volume
}

// SetLooping sets whether audio should loop
func (a *AudioPlayer2D) SetLooping(looping bool) {
	a.looping = looping
}

// IsLooping returns loop state
func (a *AudioPlayer2D) IsLooping() bool {
	return a.looping
}

// Play starts audio playback via the real audio system.
// If the audio manager is not initialised, only the flags are set (no-op).
func (a *AudioPlayer2D) Play() {
	if a.isPlaying && !a.paused {
		return // already playing
	}
	if a.soundName == "" {
		return
	}

	// If we have a paused channel, resume it instead of restarting
	if a.paused && a.channelID != 0 {
		if mixer := audio.GlobalMixer(); mixer != nil {
			mixer.Resume(audio.ChannelID(a.channelID))
		}
		a.paused = false
		a.isPlaying = true
		return
	}

	// Compute spatial pan from node position relative to listener
	pos := a.GetWorldPosition()
	pan := audio.ComputeSpatialPan(float64(pos.X))

	// Play via the global audio manager (nil-safe -> returns 0)
	id := audio.PlayNodeSound(a.soundName, float64(a.volume), a.looping, pan)
	if id != 0 {
		// Stop any previous channel before assigning the new one
		if a.channelID != 0 {
			if mixer := audio.GlobalMixer(); mixer != nil {
				mixer.Stop(audio.ChannelID(a.channelID))
			}
		}
		a.channelID = id
		a.isPlaying = true
		a.paused = false
	} else {
		// Audio not available — just set flags (backward compatible)
		a.isPlaying = true
		a.paused = false
	}
}

// Pause pauses the audio playback on the audio channel.
func (a *AudioPlayer2D) Pause() {
	if a.channelID != 0 {
		if mixer := audio.GlobalMixer(); mixer != nil {
			mixer.Pause(audio.ChannelID(a.channelID))
		}
	}
	a.paused = true
}

// Resume resumes the paused audio on the audio channel.
func (a *AudioPlayer2D) Resume() {
	if a.channelID != 0 {
		if mixer := audio.GlobalMixer(); mixer != nil {
			mixer.Resume(audio.ChannelID(a.channelID))
		}
	}
	a.paused = false
	a.isPlaying = true
}

// Stop stops audio playback on the audio channel.
func (a *AudioPlayer2D) Stop() {
	if a.channelID != 0 {
		if mixer := audio.GlobalMixer(); mixer != nil {
			mixer.Stop(audio.ChannelID(a.channelID))
		}
		a.channelID = 0
	}
	a.isPlaying = false
	a.paused = false
}

// IsPlaying returns playback state
func (a *AudioPlayer2D) IsPlaying() bool {
	return a.isPlaying && !a.paused
}

// IsPaused returns pause state
func (a *AudioPlayer2D) IsPaused() bool {
	return a.paused
}

// Update recomputes spatial audio pan based on node position relative to the listener.
// Should be called every frame when the node or listener moves.
func (a *AudioPlayer2D) Update(dt float64) {
	// Call base Node2D.Update (which propagates to children and runs scripts)
	if a.Node2D != nil {
		a.Node2D.Update(dt)
	}

	// Update spatial pan for the active channel
	if a.channelID != 0 {
		if mixer := audio.GlobalMixer(); mixer != nil {
			pos := a.GetWorldPosition()
			pan := audio.ComputeSpatialPan(float64(pos.X))
			mixer.SetChannelPan(audio.ChannelID(a.channelID), pan)
		}
	}
}

// Compile-time interface checks
var (
	_ Node = (*PhysicsBody2D)(nil)
	_ Node = (*RigidBody2D)(nil)
	_ Node = (*StaticBody2D)(nil)
	_ Node = (*Camera2D)(nil)
	_ Node = (*Area2D)(nil)
	_ Node = (*AudioPlayer2D)(nil)
)
