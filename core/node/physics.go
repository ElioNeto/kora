package node

import (
	stdMath "math"

	"github.com/ElioNeto/kora/core/math"
)

// PhysicsBody2D is a node with physics simulation properties
type PhysicsBody2D struct {
	*Node2D

	// Physics properties
	linearVelX float64
	linearVelY float64
	angularVel float64

	// Physical properties
	mass       float64
	friction   float64
	restitution float64
	gravityScale float64

	// Collision settings
	solid       bool
	trigger     bool
	collisionGroup int

	// Collision callbacks
	onCollision    func(other *Node2D, eventType CollisionType)
	onEnterOverlap func(other *Node2D)
	onExitOverlap  func(other *Node2D)
}

// NewPhysicsBody2D creates a new PhysicsBody2D node
func NewPhysicsBody2D(name string) *PhysicsBody2D {
	node := NewNode2D(name, 0)
	return &PhysicsBody2D{
		Node2D:       node,
		mass:         1.0,
		friction:     0.5,
		restitution:  0.3,
		gravityScale: 1.0,
		solid:        true,
		trigger:      false,
	}
}

// AddChild overrides Node2D to attach physics bodies
func (p *PhysicsBody2D) AddChild(child *Node2D) {
	p.Node2D.AddChild(child)
}

// SetMass sets the body mass (0 = infinite/static)
func (p *PhysicsBody2D) SetMass(mass float64) {
	p.mass = mass
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
}

// GetRestitution returns current restitution
func (p *PhysicsBody2D) GetRestitution() float64 {
	return p.restitution
}

// SetGravityScale sets gravity multiplier
func (p *PhysicsBody2D) SetGravityScale(scale float64) {
	p.gravityScale = scale
}

// GetGravityScale returns gravity scale
func (p *PhysicsBody2D) GetGravityScale() float64 {
	return p.gravityScale
}

// SetLinearVelocity sets the linear velocity
func (p *PhysicsBody2D) SetLinearVelocity(vx, vy float64) {
	p.linearVelX = vx
	p.linearVelY = vy
}

// GetLinearVelocity returns linear velocity
func (p *PhysicsBody2D) GetLinearVelocity() (float64, float64) {
	return p.linearVelX, p.linearVelY
}

// AddLinearVelocity adds to current linear velocity
func (p *PhysicsBody2D) AddLinearVelocity(vx, vy float64) {
	p.linearVelX += vx
	p.linearVelY += vy
}

// SetAngularVelocity sets rotational velocity (radians per second)
func (p *PhysicsBody2D) SetAngularVelocity(omega float64) {
	p.angularVel = omega
}

// GetAngularVelocity returns angular velocity
func (p *PhysicsBody2D) GetAngularVelocity() float64 {
	return p.angularVel
}

// SetSolid determines if body is solid (collidable)
func (p *PhysicsBody2D) SetSolid(solid bool) {
	p.solid = solid
	p.trigger = !solid
}

// IsSolid returns if body is solid
func (p *PhysicsBody2D) IsSolid() bool {
	return p.solid && !p.trigger
}

// SetTrigger sets body as trigger zone (non-solid)
func (p *PhysicsBody2D) SetTrigger(trigger bool) {
	p.trigger = trigger
	p.solid = !trigger
}

// IsTrigger returns if body is a trigger
func (p *PhysicsBody2D) IsTrigger() bool {
	return p.trigger
}

// ApplyForce applies a force at the center of mass
func (p *PhysicsBody2D) ApplyForce(fx, fy float64) {
	p.linearVelX += fx / p.mass
	p.linearVelY += fy / p.mass
}

// ApplyImpulse applies an instantaneous impulse
func (p *PhysicsBody2D) ApplyImpulse(impX, impY float64) {
	p.linearVelX += impX / p.mass
	p.linearVelY += impY / p.mass
}

// ApplyTorque applies rotational torque
func (p *PhysicsBody2D) ApplyTorque(torque float64) {
	p.angularVel += torque / p.mass // Simplified
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

// Update processes physics simulation for this body
func (p *PhysicsBody2D) Update(dt float64) {
	// Apply gravity
	if p.gravityScale > 0 {
		p.linearVelY -= 980.0 * p.gravityScale * dt
	}

	// Apply friction
	p.linearVelX *= 1.0 - p.friction*dt*10
	p.linearVelY *= 1.0 - p.friction*dt*10

	// Update position from velocity
	pos := p.GetPosition()
	pos.X += float32(p.linearVelX * dt)
	pos.Y += float32(p.linearVelY * dt)
	p.SetPosition(pos.X, pos.Y)

	// Propagate to children
	for _, child := range p.children {
		if child != nil {
			child.Update(dt)
		}
	}
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
	onEnter  func(other *Node2D)
	onExit   func(other *Node2D)
	onOverlap func(other *Node2D)
}

// NewArea2D creates a new Area2D
func NewArea2D(name string) *Area2D {
	return &Area2D{
		Node2D:   NewNode2D(name, 0),
		overlaps: make([]*Node2D, 0),
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

// Camera2D controls the camera view
type Camera2D struct {
	*Node2D

	// Target to follow
	target       *Node2D
	targetOffset math.Vector2

	// Smoothing
	smoothingFactor float64
	lerpMode        bool

	// Camera bounds
	minBounds math.Vector2
	maxBounds math.Vector2
}

// NewCamera2D creates a new Camera2D
func NewCamera2D(name string) *Camera2D {
	return &Camera2D{
		Node2D:          NewNode2D(name, 0),
		smoothingFactor: 0.1,
		lerpMode:        true,
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

// Update processes camera follow logic
func (c *Camera2D) Update(dt float64) {
	// Base update
	if c.Node2D != nil {
		c.Node2D.Update(dt)
	}

	// Follow target
	if c.target != nil && c.targetOffset.X != 0 || c.targetOffset.Y != 0 {
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

// Shake shakes camera for visual effect
func (c *Camera2D) Shake(amount float32, duration float64) {
 startPos := c.GetWorldPosition()

 // Simulated shake - in real implementation:
 // loop over duration, add random offset to position
 for t := 0.0; t < duration; t += 0.016 {
  randX := (float32(stdMath.Sin(t*10)*2-1) * amount) / 10
  randY := (float32(stdMath.Cos(t*10)*2-1) * amount) / 10
  c.SetWorldPosition(startPos.X+randX, startPos.Y+randY)
 }

 // Return to original position
 c.SetWorldPosition(startPos.X, startPos.Y)
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
}

// NewAudioPlayer2D creates a new AudioPlayer2D
func NewAudioPlayer2D(name string) *AudioPlayer2D {
	return &AudioPlayer2D{
		Node2D:    NewNode2D(name, 0),
		volume:    1.0,
		isPlaying: false,
	}
}

// SetSound sets the audio sound name
func (a *AudioPlayer2D) SetSound(soundName string) {
	a.soundName = soundName
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

// Play starts audio playback
func (a *AudioPlayer2D) Play() {
	if !a.isPlaying {
		a.isPlaying = true
		a.paused = false
		// Trigger actual audio playback
	}
}

// Pause pauses audio playback
func (a *AudioPlayer2D) Pause() {
	a.paused = true
}

// Resume resumes paused audio
func (a *AudioPlayer2D) Resume() {
	a.paused = false
}

// Stop stops audio playback
func (a *AudioPlayer2D) Stop() {
	a.isPlaying = false
	a.paused = false
	// Stop actual audio
}

// IsPlaying returns playback state
func (a *AudioPlayer2D) IsPlaying() bool {
	return a.isPlaying && !a.paused
}

// IsPaused returns pause state
func (a *AudioPlayer2D) IsPaused() bool {
	return a.paused
}
