package node

import (
	"image/color"
	"math"
	"math/rand"

	kmath "github.com/ElioNeto/kora/core/math"
	"github.com/hajimehoshi/ebiten/v2"
)

// whitePixel is a 1×1 white image used as the base texture for drawing
// particles when no custom texture is set. It is lazily initialised.
var whitePixel *ebiten.Image

// MaxParticles is the maximum number of particles that can be active at once.
const MaxParticles = 1000

// Particle represents a single particle in the particle system.
// All fields are public to allow direct inspection and serialization.
type Particle struct {
	// Position is the current world position of the particle.
	Position kmath.Vector2
	// Velocity is the current velocity vector of the particle.
	Velocity kmath.Vector2
	// Lifetime is the remaining lifetime in seconds.
	Lifetime float64
	// MaxLifetime is the initial lifetime set when the particle was spawned.
	MaxLifetime float64
	// Size is the current interpolated size of the particle.
	Size float32
	// StartSize is the size of the particle at spawn.
	StartSize float32
	// EndSize is the size of the particle when it dies.
	EndSize float32
	// Color is the current interpolated color (RGBA 0–1).
	Color struct{ R, G, B, A float32 }
	// StartColor is the color at spawn (RGBA 0–1).
	StartColor struct{ R, G, B, A float32 }
	// EndColor is the color when the particle dies (RGBA 0–1).
	EndColor struct{ R, G, B, A float32 }
	// Rotation is the current rotation in degrees.
	Rotation float32
	// AngularVelocity is the rotation speed in degrees per second.
	AngularVelocity float32
	// Active indicates whether the particle is currently alive.
	Active bool
}

// BlendMode defines how particles are blended with the background.
type BlendMode int

const (
	// BlendModeNormal uses standard alpha blending (source-over).
	BlendModeNormal BlendMode = iota
	// BlendModeAdditive uses additive blending (lighter).
	BlendModeAdditive
	// BlendModeMultiply uses multiplicative blending.
	BlendModeMultiply
)

// Particles2D is a node that emits, updates, and renders a CPU-based 2D
// particle system. Particles are drawn as coloured squares and can be
// configured with emission parameters, gravity, colour/size interpolation
// over lifetime, and optional burst (one-shot) mode.
type Particles2D struct {
	*Node2D

	// Emission parameters
	emitting    bool
	amount      int
	lifetime    float64
	speed       float32
	speedRandom float32
	direction   float32
	spread      float32
	gravity     kmath.Vector2
	startSize   float32
	endSize     float32
	startColor  struct{ R, G, B, A float32 }
	endColor    struct{ R, G, B, A float32 }

	// Burst support
	oneShot    bool
	preprocess int

	// Draw
	texture   string
	blendMode BlendMode

	// Internal state
	particles     []Particle
	particleCount int
	emitAccum     float64
	emitRate      float64
	time          float64
}

// NewParticles2D creates a new Particles2D node with default parameters.
func NewParticles2D(name string) *Particles2D {
	p := &Particles2D{
		Node2D:    NewNode2D(name, 0),
		amount:    10,
		lifetime:  1.0,
		speed:     100.0,
		direction: 270.0, // upward (compatible with engine convention: 0=right, 90=down)
		spread:    30.0,
		startSize: 4.0,
		endSize:   4.0,
		startColor: struct{ R, G, B, A float32 }{
			R: 1, G: 1, B: 1, A: 1,
		},
		endColor: struct{ R, G, B, A float32 }{
			R: 1, G: 1, B: 1, A: 0,
		},
		particles: make([]Particle, 0, MaxParticles),
		emitRate:  10.0,
	}
	p.recalcEmitRate()
	return p
}

// ---------------------------------------------------------------------------
// Emission control
// ---------------------------------------------------------------------------

// Emit spawns count particles at the emitter position. It is safe to call
// regardless of whether the system is currently emitting.
func (p *Particles2D) Emit(count int) {
	for i := 0; i < count; i++ {
		p.emitOne()
	}
}

// Start begins continuous emission. In one-shot mode, all particles are
// emitted immediately and emission stops.
func (p *Particles2D) Start() {
	if p.oneShot {
		p.Emit(p.amount)
		// In one-shot mode we do not set emitting to true — particles are
		// emitted in a single burst and then the system stays idle.
		return
	}
	p.emitting = true
}

// Stop stops continuous emission. Existing particles continue to live and
// fade naturally.
func (p *Particles2D) Stop() {
	p.emitting = false
}

// Restart clears all active particles and resets emission state. If
// preprocess > 0, the system is pre-simulated by that many particles.
func (p *Particles2D) Restart() {
	p.particles = make([]Particle, 0, MaxParticles)
	p.particleCount = 0
	p.emitAccum = 0
	p.time = 0
	p.emitting = false

	if p.oneShot {
		p.Emit(p.amount)
	} else if p.preprocess > 0 {
		// Pre-simulate by emitting particles and advancing their state
		// by a random fraction of their lifetime.
		for i := 0; i < p.preprocess; i++ {
			p.emitOne()
			idx := len(p.particles) - 1
			if idx >= 0 {
				// Advance the particle by a random fraction of its lifetime
				// so pre-processed particles are at different stages.
				advance := rand.Float64() * p.lifetime
				p.simulateParticle(&p.particles[idx], advance)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Configuration setters
// ---------------------------------------------------------------------------

// SetOneShot configures one-shot burst mode. When true, Start() emits all
// particles at once and stops. When false, Start() begins continuous emission.
func (p *Particles2D) SetOneShot(oneShot bool) {
	p.oneShot = oneShot
}

// SetAmount sets the number of particles emitted per burst (one-shot mode)
// or the steady-state particle count target (continuous mode).
func (p *Particles2D) SetAmount(amount int) {
	if amount < 0 {
		amount = 0
	}
	if amount > MaxParticles {
		amount = MaxParticles
	}
	p.amount = amount
	p.recalcEmitRate()
}

// SetSpeed sets the initial particle speed in pixels per second.
func (p *Particles2D) SetSpeed(speed float32) {
	p.speed = speed
}

// SetSpeedRandom sets the random speed variation added to the base speed.
// The actual speed for each particle is speed ± random variation.
func (p *Particles2D) SetSpeedRandom(variation float32) {
	if variation < 0 {
		variation = 0
	}
	p.speedRandom = variation
}

// SetDirection sets the base emission direction in degrees.
// 0 = right, 90 = down, 180 = left, 270 = up.
func (p *Particles2D) SetDirection(degrees float32) {
	p.direction = degrees
}

// SetSpread sets the spread angle in degrees. Particles are emitted in a
// cone from direction-spread/2 to direction+spread/2.
func (p *Particles2D) SetSpread(degrees float32) {
	if degrees < 0 {
		degrees = 0
	}
	p.spread = degrees
}

// SetGravity sets the gravity vector affecting all particles (pixels/s²).
func (p *Particles2D) SetGravity(x, y float32) {
	p.gravity = kmath.NewVector2(x, y)
}

// SetLifetime sets the particle lifetime in seconds.
func (p *Particles2D) SetLifetime(seconds float64) {
	if seconds < 0 {
		seconds = 0
	}
	p.lifetime = seconds
	p.recalcEmitRate()
}

// SetStartSize sets the starting particle size in pixels.
func (p *Particles2D) SetStartSize(size float32) {
	if size < 0 {
		size = 0
	}
	p.startSize = size
}

// SetEndSize sets the ending particle size in pixels.
func (p *Particles2D) SetEndSize(size float32) {
	if size < 0 {
		size = 0
	}
	p.endSize = size
}

// SetStartColor sets the starting colour (RGBA components in range 0–1).
func (p *Particles2D) SetStartColor(r, g, b, a float32) {
	p.startColor.R = clampColor(r)
	p.startColor.G = clampColor(g)
	p.startColor.B = clampColor(b)
	p.startColor.A = clampColor(a)
}

// SetEndColor sets the ending colour (RGBA components in range 0–1).
func (p *Particles2D) SetEndColor(r, g, b, a float32) {
	p.endColor.R = clampColor(r)
	p.endColor.G = clampColor(g)
	p.endColor.B = clampColor(b)
	p.endColor.A = clampColor(a)
}

// SetTexture sets the particle texture by file path. An empty string clears
// the texture and particles are drawn as coloured rectangles.
func (p *Particles2D) SetTexture(path string) {
	p.texture = path
}

// SetBlendMode sets the blend mode for particle rendering.
//
//	0 = normal (alpha blending)
//	1 = additive (lighter)
//	2 = multiply
func (p *Particles2D) SetBlendMode(mode int) {
	switch mode {
	case 0:
		p.blendMode = BlendModeNormal
	case 1:
		p.blendMode = BlendModeAdditive
	case 2:
		p.blendMode = BlendModeMultiply
	default:
		p.blendMode = BlendModeNormal
	}
}

// ---------------------------------------------------------------------------
// Queries
// ---------------------------------------------------------------------------

// GetParticleCount returns the number of currently active particles.
func (p *Particles2D) GetParticleCount() int {
	return p.particleCount
}

// IsEmitting returns whether the system is currently emitting particles
// continuously. One-shot bursts are not considered continuous emission.
func (p *Particles2D) IsEmitting() bool {
	return p.emitting
}

// ---------------------------------------------------------------------------
// Update / Draw
// ---------------------------------------------------------------------------

// Update processes the particle system each frame: emits new particles in
// continuous mode, updates existing particles, and removes dead ones.
func (p *Particles2D) Update(dt float64) {
	if !p.alive {
		return
	}

	p.time += dt

	// 1. Emit new particles in continuous mode
	if p.emitting && !p.oneShot && p.emitRate > 0 {
		p.emitAccum += p.emitRate * dt
		for p.emitAccum >= 1.0 {
			p.emitOne()
			p.emitAccum -= 1.0
		}
	}

	// 2. Update existing particles
	for i := 0; i < len(p.particles); i++ {
		part := &p.particles[i]
		if !part.Active {
			continue
		}

		// Decrease lifetime
		part.Lifetime -= dt
		if part.Lifetime <= 0 {
			part.Active = false
			p.particleCount--
			continue
		}

		// Apply angular velocity
		part.Rotation += part.AngularVelocity * float32(dt)

		// Apply velocity and gravity
		part.Velocity = part.Velocity.Add(p.gravity.Mul(float32(dt)))
		part.Position = part.Position.Add(part.Velocity.Mul(float32(dt)))

		// Interpolate size
		t := 1.0 - part.Lifetime/part.MaxLifetime
		part.Size = lerpFloat32(part.StartSize, part.EndSize, float32(t))

		// Interpolate colour
		part.Color.R = lerpFloat32(part.StartColor.R, part.EndColor.R, float32(t))
		part.Color.G = lerpFloat32(part.StartColor.G, part.EndColor.G, float32(t))
		part.Color.B = lerpFloat32(part.StartColor.B, part.EndColor.B, float32(t))
		part.Color.A = lerpFloat32(part.StartColor.A, part.EndColor.A, float32(t))
	}

	// 3. Compact the particle list (optional — run periodically or when many
	//    inactive slots accumulate). We compact every frame for determinism.
	p.compact()

	// Propagate to children
	p.Node2D.Update(dt)
}

// Draw renders active particles to the screen. Each particle is drawn as a
// coloured rectangle (or tinted sprite if a texture is set).
func (p *Particles2D) Draw(screen *ebiten.Image) {
	if !p.visible || !p.alive {
		return
	}

	// Draw children first (below particles)
	for _, child := range p.children {
		if child != nil {
			child.Draw(screen)
		}
	}

	if p.particleCount == 0 {
		return
	}

	// Lazily initialise the white pixel texture.
	if whitePixel == nil {
		whitePixel = ebiten.NewImage(1, 1)
		whitePixel.Fill(color.White)
	}

	// Resolve blend mode
	compositeMode := compositeModeFromBlend(p.blendMode)

	// Draw each active particle
	for i := range p.particles {
		part := &p.particles[i]
		if !part.Active {
			continue
		}

		s := float64(part.Size)
		half := s * 0.5
		px := float64(part.Position.X)
		py := float64(part.Position.Y)

		// Build the transformation matrix manually so that the 1×1 white pixel
		// is scaled to `size×size`, centred at (px, py), and rotated around its
		// centre when part.Rotation != 0.
		var geo ebiten.GeoM

		if part.Rotation != 0 {
			rad := float64(part.Rotation) * math.Pi / 180
			cos := math.Cos(rad)
			sin := math.Sin(rad)

			// Transform:
			//   x' = px + (u-0.5)*s*cosθ - (v-0.5)*s*sinθ
			//   y' = py + (u-0.5)*s*sinθ + (v-0.5)*s*cosθ
			//
			// In matrix form (u,v in [0,1]):
			//   |x'|   |s*cosθ   -s*sinθ   px - half*cosθ + half*sinθ| |u|
			//   |y'| = |s*sinθ    s*cosθ   py - half*sinθ - half*cosθ| |v|
			//   |1 |   |   0         0                  1           | |1|
			geo.SetElement(0, 0, s*cos)
			geo.SetElement(0, 1, -s*sin)
			geo.SetElement(0, 2, px-half*cos+half*sin)
			geo.SetElement(1, 0, s*sin)
			geo.SetElement(1, 1, s*cos)
			geo.SetElement(1, 2, py-half*sin-half*cos)
		} else {
			geo.SetElement(0, 0, s)
			geo.SetElement(0, 1, 0)
			geo.SetElement(0, 2, px-half)
			geo.SetElement(1, 0, 0)
			geo.SetElement(1, 1, s)
			geo.SetElement(1, 2, py-half)
		}

		opts := &ebiten.DrawImageOptions{
			GeoM:          geo,
			CompositeMode: compositeMode,
		}
		opts.ColorScale.SetR(part.Color.R)
		opts.ColorScale.SetG(part.Color.G)
		opts.ColorScale.SetB(part.Color.B)
		opts.ColorScale.SetA(part.Color.A)

		screen.DrawImage(whitePixel, opts)
	}
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// emitOne spawns a single particle at the emitter's world position. If the
// pool is full, the particle is silently dropped.
func (p *Particles2D) emitOne() {
	if p.particleCount >= MaxParticles {
		return
	}

	// Compute emission angle: direction + random spread
	angle := p.direction
	if p.spread > 0 {
		halfSpread := p.spread * 0.5
		angle += (rand.Float32()*2 - 1) * halfSpread
	}

	// Convert degrees to radians
	rad := float64(angle) * math.Pi / 180

	// Compute speed with random variation
	speed := p.speed
	if p.speedRandom > 0 {
		speed += (rand.Float32()*2 - 1) * p.speedRandom
	}
	if speed < 0 {
		speed = 0
	}

	// Velocity from angle and speed
	vel := kmath.NewVector2(
		float32(math.Cos(rad))*speed,
		float32(math.Sin(rad))*speed,
	)

	// World position of the emitter
	pos := p.GetWorldPosition()

	// Lifetime with small random variation (±10%)
	lifetime := p.lifetime
	if lifetime > 0 {
		lifetime *= 0.9 + rand.Float64()*0.2
	}

	// Size randomisation (±20%)
	sizeVariation := 0.8 + rand.Float32()*0.4
	startSize := p.startSize * sizeVariation
	endSize := p.endSize * sizeVariation

	part := Particle{
		Position:        pos,
		Velocity:        vel,
		Lifetime:        lifetime,
		MaxLifetime:     lifetime,
		Size:            startSize,
		StartSize:       startSize,
		EndSize:         endSize,
		Rotation:        0,
		AngularVelocity: (rand.Float32()*2 - 1) * 360, // ±360°/s
		Active:          true,
	}
	part.StartColor = p.startColor
	part.EndColor = p.endColor
	part.Color = p.startColor

	p.particles = append(p.particles, part)
	p.particleCount++
}

// simulateParticle advances a particle by dt seconds without going through
// the full emission pipeline. Used for preprocess.
func (p *Particles2D) simulateParticle(part *Particle, dt float64) {
	if !part.Active {
		return
	}

	part.Lifetime -= dt
	if part.Lifetime <= 0 {
		part.Active = false
		p.particleCount--
		return
	}

	part.Velocity = part.Velocity.Add(p.gravity.Mul(float32(dt)))
	part.Position = part.Position.Add(part.Velocity.Mul(float32(dt)))
	part.Rotation += part.AngularVelocity * float32(dt)

	t := 1.0 - part.Lifetime/part.MaxLifetime
	part.Size = lerpFloat32(part.StartSize, part.EndSize, float32(t))
	part.Color.R = lerpFloat32(part.StartColor.R, part.EndColor.R, float32(t))
	part.Color.G = lerpFloat32(part.StartColor.G, part.EndColor.G, float32(t))
	part.Color.B = lerpFloat32(part.StartColor.B, part.EndColor.B, float32(t))
	part.Color.A = lerpFloat32(part.StartColor.A, part.EndColor.A, float32(t))
}

// compact removes inactive particles from the slice to keep it tight.
func (p *Particles2D) compact() {
	if len(p.particles) == 0 {
		return
	}

	// Move active particles to the front (in-place filter)
	j := 0
	for i := range p.particles {
		if p.particles[i].Active {
			if i != j {
				p.particles[j] = p.particles[i]
			}
			j++
		}
	}
	// Trim trailing inactive slots
	p.particles = p.particles[:j]
}

// recalcEmitRate recomputes the continuous emission rate based on amount
// and lifetime, targeting approximately `amount` particles alive at steady
// state.
func (p *Particles2D) recalcEmitRate() {
	if p.lifetime > 0 {
		p.emitRate = float64(p.amount) / p.lifetime
	} else {
		p.emitRate = float64(p.amount)
	}
}

// lerpFloat32 linearly interpolates between a and b by t [0,1].
func lerpFloat32(a, b, t float32) float32 {
	return a + (b-a)*t
}

// clampColor clamps a colour component to [0,1].
func clampColor(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// compositeModeFromBlend translates a BlendMode to an ebiten CompositeMode.
func compositeModeFromBlend(mode BlendMode) ebiten.CompositeMode {
	switch mode {
	case BlendModeAdditive:
		return ebiten.CompositeModeLighter
	case BlendModeMultiply:
		return ebiten.CompositeModeMultiply
	default:
		return ebiten.CompositeModeSourceOver
	}
}

// ---------------------------------------------------------------------------
// Compile-time interface checks
// ---------------------------------------------------------------------------

var _ Node = (*Particles2D)(nil)
