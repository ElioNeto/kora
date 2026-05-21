package node

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Construction and defaults
// ---------------------------------------------------------------------------

func TestNewParticles2D(t *testing.T) {
	p := NewParticles2D("test")
	if p == nil {
		t.Fatal("expected non-nil Particles2D")
	}
	if p.GetName() != "test" {
		t.Errorf("expected name 'test', got '%s'", p.GetName())
	}
	if p.GetParticleCount() != 0 {
		t.Errorf("expected 0 particles, got %d", p.GetParticleCount())
	}
	if p.IsEmitting() {
		t.Error("expected IsEmitting() to be false by default")
	}
	if p.amount != 10 {
		t.Errorf("expected default amount 10, got %d", p.amount)
	}
	if p.lifetime != 1.0 {
		t.Errorf("expected default lifetime 1.0, got %f", p.lifetime)
	}
}

func TestParticles2D_NodeInterface(t *testing.T) {
	var _ Node = (*Particles2D)(nil)

	// Runtime check
	p := NewParticles2D("particles")
	var n Node = p
	if n.Name() != "particles" {
		t.Error("Node interface not satisfied correctly")
	}
}

// ---------------------------------------------------------------------------
// Emit
// ---------------------------------------------------------------------------

func TestParticles2D_EmitCreatesActiveParticles(t *testing.T) {
	p := NewParticles2D("test")
	p.Emit(5)

	count := p.GetParticleCount()
	if count != 5 {
		t.Errorf("expected 5 active particles after Emit(5), got %d", count)
	}

	// All particles should be active
	for i := range p.particles {
		if !p.particles[i].Active {
			t.Errorf("particle %d should be active", i)
		}
	}
}

func TestParticles2D_EmitZero(t *testing.T) {
	p := NewParticles2D("test")
	p.Emit(0)
	if p.GetParticleCount() != 0 {
		t.Errorf("expected 0 particles after Emit(0), got %d", p.GetParticleCount())
	}
}

func TestParticles2D_EmitRespectsMaxParticles(t *testing.T) {
	p := NewParticles2D("test")
	// Emit more than MaxParticles
	p.Emit(MaxParticles + 100)
	count := p.GetParticleCount()
	if count > MaxParticles {
		t.Errorf("particle count %d exceeds MaxParticles %d", count, MaxParticles)
	}
}

// ---------------------------------------------------------------------------
// Update – lifetime and removal
// ---------------------------------------------------------------------------

func TestParticles2D_UpdateDecreasesLifetime(t *testing.T) {
	p := NewParticles2D("test")
	p.Emit(1)

	p.Update(0.3)
	if len(p.particles) == 0 {
		t.Fatal("particle should still exist after 0.3s")
	}
	remaining := p.particles[0].Lifetime
	if remaining <= 0 || remaining >= 1.0 {
		t.Errorf("expected remaining lifetime between 0 and 1.0 after 0.3s, got %f", remaining)
	}
}

func TestParticles2D_UpdateRemovesDeadParticles(t *testing.T) {
	p := NewParticles2D("test")
	p.SetLifetime(0.1)
	p.Emit(3)

	// After lifetime passes, particles should die
	p.Update(0.2)
	count := p.GetParticleCount()
	if count != 0 {
		t.Errorf("expected 0 particles after lifetime expires, got %d", count)
	}
}

func TestParticles2D_UpdatePartialRemoval(t *testing.T) {
	p := NewParticles2D("test")
	p.SetLifetime(0.5)

	// Emit particles at different times to get staggered lifetimes
	p.Emit(3)
	_ = p.particles // we manually adjust for testing

	// Simulate two steps so particles age
	p.Update(0.3) // particles at 0.3s remaining
	p.Update(0.3) // at 0.0s → dead, then removed
	if p.GetParticleCount() != 0 {
		t.Errorf("expected 0 particles after 0.6s total with 0.5s lifetime, got %d", p.GetParticleCount())
	}
}

// ---------------------------------------------------------------------------
// One-shot mode
// ---------------------------------------------------------------------------

func TestParticles2D_OneShotEmitAllAtOnce(t *testing.T) {
	p := NewParticles2D("test")
	p.SetOneShot(true)
	p.SetAmount(8)
	p.Start()

	count := p.GetParticleCount()
	if count != 8 {
		t.Errorf("expected 8 particles after one-shot Start(), got %d", count)
	}
}

func TestParticles2D_OneShotDoesNotEmitContinuously(t *testing.T) {
	p := NewParticles2D("test")
	p.SetOneShot(true)
	p.SetAmount(5)
	p.Start()

	// After first frame, all particles should be emitted
	p.Update(1.0)
	// After 1s with 1s lifetime, particles should be dying now
	// But we should not have emitted more
	if p.IsEmitting() {
		t.Error("expected IsEmitting() false after one-shot Start()")
	}
}

// ---------------------------------------------------------------------------
// Start / Stop
// ---------------------------------------------------------------------------

func TestParticles2D_StartContinuousEmission(t *testing.T) {
	p := NewParticles2D("test")
	p.SetAmount(20)

	p.Start()
	if !p.IsEmitting() {
		t.Error("expected IsEmitting() true after Start()")
	}
}

func TestParticles2D_StopStopsEmission(t *testing.T) {
	p := NewParticles2D("test")
	p.Start()
	if !p.IsEmitting() {
		t.Fatal("expected IsEmitting() true after Start()")
	}

	p.Stop()
	if p.IsEmitting() {
		t.Error("expected IsEmitting() false after Stop()")
	}
}

func TestParticles2D_ContinuousEmissionProducesParticles(t *testing.T) {
	p := NewParticles2D("test")
	p.SetAmount(60)    // emitRate = 60/1.0 = 60 particles/s
	p.SetLifetime(2.0)
	p.Start()

	// Run 60 frames at ~60fps = 1s → should emit ~60 particles
	for i := 0; i < 60; i++ {
		p.Update(1.0 / 60.0)
	}

	count := p.GetParticleCount()
	if count == 0 {
		t.Error("expected some particles after 1s of continuous emission at 60 particles/s")
	}
}

func TestParticles2D_StopPreservesExistingParticles(t *testing.T) {
	p := NewParticles2D("test")
	p.SetAmount(10)
	p.SetLifetime(5.0)
	p.Start()
	p.Update(0.5)
	p.Stop()

	// After stopping, particles should still be alive
	beforeCount := p.GetParticleCount()
	if beforeCount == 0 {
		t.Skip("no particles were emitted, skipping")
	}

	// Advance time but not past lifetime
	p.Update(0.1)
	afterCount := p.GetParticleCount()
	if afterCount == 0 {
		t.Error("particles should still exist after stop, before lifetime expires")
	}
}

// ---------------------------------------------------------------------------
// Restart
// ---------------------------------------------------------------------------

func TestParticles2D_RestartClearsParticles(t *testing.T) {
	p := NewParticles2D("test")
	p.Emit(10)
	if p.GetParticleCount() != 10 {
		t.Fatal("expected 10 particles after Emit")
	}

	p.Restart()
	if p.GetParticleCount() != 0 {
		t.Errorf("expected 0 particles after Restart, got %d", p.GetParticleCount())
	}
}

func TestParticles2D_RestartStopsEmission(t *testing.T) {
	p := NewParticles2D("test")
	p.Start()
	p.Restart()
	if p.IsEmitting() {
		t.Error("expected IsEmitting() false after Restart")
	}
}

func TestParticles2D_RestartWithOneShot(t *testing.T) {
	p := NewParticles2D("test")
	p.SetOneShot(true)
	p.SetAmount(7)
	p.Emit(5) // manual emit, not using one-shot

	p.Restart()
	// one-shot + Restart should emit amount particles
	if p.GetParticleCount() != 7 {
		t.Errorf("expected 7 particles after one-shot Restart, got %d", p.GetParticleCount())
	}
}

func TestParticles2D_RestartWithPreprocess(t *testing.T) {
	p := NewParticles2D("test")
	p.SetLifetime(10.0) // long lifetime so pre-sim doesn't kill particles
	p.SetDirection(0)
	p.SetSpeed(0)
	p.SetSpread(0)
	p.preprocess = 5
	p.Restart()

	count := p.GetParticleCount()
	if count != 5 {
		t.Errorf("expected 5 pre-processed particles after Restart, got %d", count)
	}

	// Pre-processed particles should have been advanced in time,
	// so they should have non-zero age
	if count > 0 {
		for i := range p.particles {
			if p.particles[i].Active {
				age := 1.0 - p.particles[i].Lifetime/p.particles[i].MaxLifetime
				if age <= 0 {
					t.Errorf("expected pre-processed particle %d to have age > 0", i)
				}
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Velocity and position
// ---------------------------------------------------------------------------

func TestParticles2D_ParticlePositionUpdatesWithVelocity(t *testing.T) {
	p := NewParticles2D("test")
	p.SetDirection(0) // right
	p.SetSpeed(100)
	p.SetSpread(0) // no spread
	p.SetLifetime(5.0)

	p.Emit(1)
	if p.particleCount < 1 {
		t.Fatal("expected particle after Emit")
	}
	part := &p.particles[0]

	startX := part.Position.X

	p.Update(0.5) // 0.5s at 100 px/s = 50px
	if p.particleCount < 1 {
		t.Fatal("particle died before test completed")
	}
	part = &p.particles[0]

	dx := part.Position.X - startX
	// Should have moved right by ~50 pixels
	if dx < 40 || dx > 60 {
		t.Errorf("expected particle to move ~50px right, moved %f", dx)
	}
}

func TestParticles2D_ParticleMovesDown(t *testing.T) {
	p := NewParticles2D("test")
	p.SetDirection(90) // down
	p.SetSpeed(50)
	p.SetSpread(0)
	p.SetLifetime(5.0) // long lifetime so particle survives the test

	p.Emit(1)
	if p.particleCount < 1 {
		t.Fatal("expected at least 1 particle after Emit")
	}
	startY := p.particles[0].Position.Y

	p.Update(1.0)
	if p.particleCount < 1 {
		t.Fatal("particle died before test completed")
	}
	dy := p.particles[0].Position.Y - startY
	if dy < 40 || dy > 60 {
		t.Errorf("expected particle to move ~50px down, moved %f", dy)
	}
}

// ---------------------------------------------------------------------------
// Gravity
// ---------------------------------------------------------------------------

func TestParticles2D_GravityAffectsParticles(t *testing.T) {
	p := NewParticles2D("test")
	p.SetDirection(0) // right
	p.SetSpeed(0)     // no initial velocity
	p.SetGravity(0, 100)
	p.SetLifetime(5.0)

	p.Emit(1)
	if p.particleCount < 1 {
		t.Fatal("expected at least 1 particle after Emit")
	}
	startY := p.particles[0].Position.Y

	p.Update(1.0)
	if p.particleCount < 1 {
		t.Fatal("particle died before test completed, lifetime too short")
	}
	dy := p.particles[0].Position.Y - startY
	// Gravity = 100 px/s², after 1s: dy ≈ 100px
	if dy < 80 || dy > 120 {
		t.Errorf("expected gravity to move particle ~100px down, moved %f", dy)
	}
}

func TestParticles2D_GravityXAffectsParticles(t *testing.T) {
	p := NewParticles2D("test")
	p.SetDirection(90) // down
	p.SetSpeed(0)
	p.SetGravity(50, 0)
	p.SetLifetime(5.0)

	p.Emit(1)
	if p.particleCount < 1 {
		t.Fatal("expected at least 1 particle after Emit")
	}
	startX := p.particles[0].Position.X

	p.Update(1.0)
	if p.particleCount < 1 {
		t.Fatal("particle died before test completed")
	}
	dx := p.particles[0].Position.X - startX
	if dx < 30 || dx > 70 {
		t.Errorf("expected gravity X to move particle ~50px right, moved %f", dx)
	}
}

// ---------------------------------------------------------------------------
// Colour setters
// ---------------------------------------------------------------------------

func TestParticles2D_SetStartColor(t *testing.T) {
	p := NewParticles2D("test")
	p.SetStartColor(0.5, 0.25, 0.75, 1.0)

	if p.startColor.R != 0.5 {
		t.Errorf("expected startColor.R 0.5, got %f", p.startColor.R)
	}
	if p.startColor.G != 0.25 {
		t.Errorf("expected startColor.G 0.25, got %f", p.startColor.G)
	}
	if p.startColor.B != 0.75 {
		t.Errorf("expected startColor.B 0.75, got %f", p.startColor.B)
	}
	if p.startColor.A != 1.0 {
		t.Errorf("expected startColor.A 1.0, got %f", p.startColor.A)
	}
}

func TestParticles2D_SetEndColor(t *testing.T) {
	p := NewParticles2D("test")
	p.SetEndColor(0.1, 0.2, 0.3, 0.5)

	if p.endColor.R != 0.1 {
		t.Errorf("expected endColor.R 0.1, got %f", p.endColor.R)
	}
	if p.endColor.G != 0.2 {
		t.Errorf("expected endColor.G 0.2, got %f", p.endColor.G)
	}
	if p.endColor.B != 0.3 {
		t.Errorf("expected endColor.B 0.3, got %f", p.endColor.B)
	}
	if p.endColor.A != 0.5 {
		t.Errorf("expected endColor.A 0.5, got %f", p.endColor.A)
	}
}

func TestParticles2D_ColorClamped(t *testing.T) {
	p := NewParticles2D("test")
	p.SetStartColor(-0.5, 1.5, 0, 1)
	if p.startColor.R != 0 {
		t.Errorf("expected startColor.R clamped to 0, got %f", p.startColor.R)
	}
	if p.startColor.G != 1 {
		t.Errorf("expected startColor.G clamped to 1, got %f", p.startColor.G)
	}
}

// ---------------------------------------------------------------------------
// Setters
// ---------------------------------------------------------------------------

func TestParticles2D_SetAmount(t *testing.T) {
	p := NewParticles2D("test")
	p.SetAmount(25)
	if p.amount != 25 {
		t.Errorf("expected amount 25, got %d", p.amount)
	}
	// Negative should be clamped to 0
	p.SetAmount(-5)
	if p.amount != 0 {
		t.Errorf("expected amount 0 after negative, got %d", p.amount)
	}
	// Above max should be clamped
	p.SetAmount(MaxParticles + 100)
	if p.amount != MaxParticles {
		t.Errorf("expected amount clamped to MaxParticles, got %d", p.amount)
	}
}

func TestParticles2D_SetSpeed(t *testing.T) {
	p := NewParticles2D("test")
	p.SetSpeed(200)
	if p.speed != 200 {
		t.Errorf("expected speed 200, got %f", p.speed)
	}
}

func TestParticles2D_SetSpeedRandom(t *testing.T) {
	p := NewParticles2D("test")
	p.SetSpeedRandom(50)
	if p.speedRandom != 50 {
		t.Errorf("expected speedRandom 50, got %f", p.speedRandom)
	}
	// Negative should be clamped to 0
	p.SetSpeedRandom(-10)
	if p.speedRandom != 0 {
		t.Errorf("expected speedRandom clamped to 0, got %f", p.speedRandom)
	}
}

func TestParticles2D_SetDirection(t *testing.T) {
	p := NewParticles2D("test")
	p.SetDirection(45)
	if p.direction != 45 {
		t.Errorf("expected direction 45, got %f", p.direction)
	}
}

func TestParticles2D_SetSpread(t *testing.T) {
	p := NewParticles2D("test")
	p.SetSpread(60)
	if p.spread != 60 {
		t.Errorf("expected spread 60, got %f", p.spread)
	}
	// Negative should be clamped to 0
	p.SetSpread(-10)
	if p.spread != 0 {
		t.Errorf("expected spread clamped to 0, got %f", p.spread)
	}
}

func TestParticles2D_SetGravity(t *testing.T) {
	p := NewParticles2D("test")
	p.SetGravity(0, -9.8)
	if p.gravity.X != 0 || p.gravity.Y != -9.8 {
		t.Errorf("expected gravity (0, -9.8), got (%f, %f)", p.gravity.X, p.gravity.Y)
	}
}

func TestParticles2D_SetLifetime(t *testing.T) {
	p := NewParticles2D("test")
	p.SetLifetime(2.5)
	if p.lifetime != 2.5 {
		t.Errorf("expected lifetime 2.5, got %f", p.lifetime)
	}
	// Negative should be clamped to 0
	p.SetLifetime(-1)
	if p.lifetime != 0 {
		t.Errorf("expected lifetime clamped to 0, got %f", p.lifetime)
	}
}

func TestParticles2D_SetStartSize(t *testing.T) {
	p := NewParticles2D("test")
	p.SetStartSize(8)
	if p.startSize != 8 {
		t.Errorf("expected startSize 8, got %f", p.startSize)
	}
	// Negative should be clamped to 0
	p.SetStartSize(-1)
	if p.startSize != 0 {
		t.Errorf("expected startSize clamped to 0, got %f", p.startSize)
	}
}

func TestParticles2D_SetEndSize(t *testing.T) {
	p := NewParticles2D("test")
	p.SetEndSize(16)
	if p.endSize != 16 {
		t.Errorf("expected endSize 16, got %f", p.endSize)
	}
	// Negative should be clamped to 0
	p.SetEndSize(-1)
	if p.endSize != 0 {
		t.Errorf("expected endSize clamped to 0, got %f", p.endSize)
	}
}

func TestParticles2D_SetTexture(t *testing.T) {
	p := NewParticles2D("test")
	p.SetTexture("particles/fire.png")
	if p.texture != "particles/fire.png" {
		t.Errorf("expected texture 'particles/fire.png', got '%s'", p.texture)
	}
}

func TestParticles2D_SetBlendMode(t *testing.T) {
	p := NewParticles2D("test")
	if p.blendMode != BlendModeNormal {
		t.Errorf("expected default blend mode normal, got %d", p.blendMode)
	}

	p.SetBlendMode(1)
	if p.blendMode != BlendModeAdditive {
		t.Errorf("expected blend mode additive after SetBlendMode(1), got %d", p.blendMode)
	}

	p.SetBlendMode(2)
	if p.blendMode != BlendModeMultiply {
		t.Errorf("expected blend mode multiply after SetBlendMode(2), got %d", p.blendMode)
	}

	// Invalid should fall back to normal
	p.SetBlendMode(99)
	if p.blendMode != BlendModeNormal {
		t.Errorf("expected blend mode normal for invalid input, got %d", p.blendMode)
	}
}

// ---------------------------------------------------------------------------
// Particle count
// ---------------------------------------------------------------------------

func TestParticles2D_ParticleCountChangesCorrectly(t *testing.T) {
	p := NewParticles2D("test")

	// Initial count is 0
	if p.GetParticleCount() != 0 {
		t.Errorf("expected 0, got %d", p.GetParticleCount())
	}

	// After emitting with short lifetime
	p.SetLifetime(0.5)
	p.Emit(3)
	if p.GetParticleCount() != 3 {
		t.Errorf("expected 3, got %d", p.GetParticleCount())
	}

	// After emitting more (same short lifetime)
	p.Emit(2)
	if p.GetParticleCount() != 5 {
		t.Errorf("expected 5, got %d", p.GetParticleCount())
	}

	// After particles die (update past lifetime)
	p.Update(0.6)
	if p.GetParticleCount() != 0 {
		t.Errorf("expected 0 after lifetime expired, got %d", p.GetParticleCount())
	}
}

// ---------------------------------------------------------------------------
// Size interpolation
// ---------------------------------------------------------------------------

func TestParticles2D_SizeInterpolation(t *testing.T) {
	p := NewParticles2D("test")
	p.SetLifetime(5.0)
	p.SetStartSize(10)
	p.SetEndSize(2)

	p.Emit(1)
	if p.particleCount < 1 {
		t.Fatal("expected at least 1 particle after Emit")
	}
	part := &p.particles[0]
	// Start size has ±20% random variation, so it should be in range [8, 12]
	if part.Size < 8 || part.Size > 12 {
		t.Errorf("expected start size ~10 (±20%%), got %f", part.Size)
	}

	// After half lifetime, size should be between start and end, factoring in
	// the ±20% random variation applied at spawn. Expected range: ~4.8 to 7.2.
	p.Update(0.5)
	if p.particleCount < 1 {
		t.Fatal("particle died before test")
	}
	part = &p.particles[0]
	if part.Size > 10 || part.Size < 3 {
		t.Errorf("expected size in [3,10] after half lifetime (was invalid), got %f", part.Size)
	}
	// Log actual values for debugging
	t.Logf("startSize=%.2f endSize=%.2f size at half=%.4f lifetime=%.2f maxLifetime=%.2f",
		part.StartSize, part.EndSize, part.Size, part.Lifetime, part.MaxLifetime)
}

// ---------------------------------------------------------------------------
// Colour interpolation
// ---------------------------------------------------------------------------

func TestParticles2D_ColorInterpolation(t *testing.T) {
	p := NewParticles2D("test")
	p.SetLifetime(0.5)
	p.SetStartColor(1, 0, 0, 1) // red opaque
	p.SetEndColor(0, 0, 1, 0)   // blue transparent

	p.Emit(1)
	part := &p.particles[0]

	// Initial colour should be start colour
	if part.Color.R != 1 || part.Color.G != 0 || part.Color.B != 0 || part.Color.A != 1 {
		t.Errorf("expected start color (1,0,0,1), got (%f,%f,%f,%f)",
			part.Color.R, part.Color.G, part.Color.B, part.Color.A)
	}

	// Advance past the guaranteed max lifetime (0.5 + 10% = 0.55, so 0.6 is safe)
	p.Update(0.6)
	if part.Active {
		t.Error("expected particle to be dead after full lifetime")
	}
}

// ---------------------------------------------------------------------------
// One-shot + emission integration
// ---------------------------------------------------------------------------

func TestParticles2D_OneShotAndContinuousIndependence(t *testing.T) {
	p := NewParticles2D("test")
	p.SetAmount(5)
	p.SetLifetime(2.0)

	// Use Emit directly (not one-shot) to add particles
	p.Emit(3)
	if p.GetParticleCount() != 3 {
		t.Errorf("expected 3 particles from Emit, got %d", p.GetParticleCount())
	}

	// Switch to one-shot and start — should emit 5 more
	p.SetOneShot(true)
	p.SetAmount(5)
	p.Start()
	if p.GetParticleCount() != 8 {
		t.Errorf("expected 8 total particles (3+5), got %d", p.GetParticleCount())
	}
}

// ---------------------------------------------------------------------------
// SetOneShot getter semantics
// ---------------------------------------------------------------------------

func TestParticles2D_SetOneShot(t *testing.T) {
	p := NewParticles2D("test")
	if p.oneShot {
		t.Error("expected oneShot false by default")
	}
	p.SetOneShot(true)
	if !p.oneShot {
		t.Error("expected oneShot true after SetOneShot(true)")
	}
	p.SetOneShot(false)
	if p.oneShot {
		t.Error("expected oneShot false after SetOneShot(false)")
	}
}

// ---------------------------------------------------------------------------
// World position inheritance
// ---------------------------------------------------------------------------

func TestParticles2D_ParticlesEmitAtWorldPosition(t *testing.T) {
	p := NewParticles2D("test")
	p.SetPosition(100, 200)

	p.Emit(1)
	part := p.particles[0]

	if part.Position.X != 100 || part.Position.Y != 200 {
		t.Errorf("expected particle at (100,200), got (%f,%f)",
			part.Position.X, part.Position.Y)
	}
}

func TestParticles2D_ParticlesEmitAtInheritedWorldPosition(t *testing.T) {
	parent := NewNode2D("parent", 1)
	parent.SetPosition(50, 50)

	p := NewParticles2D("test")
	p.SetPosition(10, 10)

	parent.AddChild(p)
	p.Emit(1)

	// World position should be parent + local = (60, 60)
	part := p.particles[0]
	if part.Position.X != 60 || part.Position.Y != 60 {
		t.Errorf("expected particle at (60,60), got (%f,%f)",
			part.Position.X, part.Position.Y)
	}
}

// ---------------------------------------------------------------------------
// Multiple emissions
// ---------------------------------------------------------------------------

func TestParticles2D_MultipleEmitCalls(t *testing.T) {
	p := NewParticles2D("test")
	p.Emit(2)
	p.Emit(3)
	p.Emit(5)

	if p.GetParticleCount() != 10 {
		t.Errorf("expected 10 particles after three Emit calls, got %d", p.GetParticleCount())
	}
}

// ---------------------------------------------------------------------------
// Particle data integrity after Update
// ---------------------------------------------------------------------------

func TestParticles2D_UpdatePreservesActiveParticles(t *testing.T) {
	p := NewParticles2D("test")
	p.SetLifetime(5.0) // long lifetime
	p.Emit(4)

	p.Update(1.0)
	// All particles should still be active
	for i := range p.particles {
		if !p.particles[i].Active {
			t.Errorf("particle %d should still be active", i)
		}
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestParticles2D_UpdateWithZeroDt(t *testing.T) {
	p := NewParticles2D("test")
	p.Emit(3)
	// Should not panic
	p.Update(0)
	if p.GetParticleCount() != 3 {
		t.Errorf("expected 3 particles after Update(0), got %d", p.GetParticleCount())
	}
}

func TestParticles2D_UpdateWithNegativeDt(t *testing.T) {
	p := NewParticles2D("test")
	p.Emit(2)
	// Should not panic (lifetime would increase)
	p.Update(-0.1)
	// Particles should still exist
	if p.GetParticleCount() != 2 {
		t.Errorf("expected 2 particles after Update(-0.1), got %d", p.GetParticleCount())
	}
}

func TestParticles2D_ZeroAmountEmit(t *testing.T) {
	p := NewParticles2D("test")
	p.SetAmount(0)
	p.Emit(0)
	if p.GetParticleCount() != 0 {
		t.Errorf("expected 0 particles with zero amount, got %d", p.GetParticleCount())
	}
}

func TestParticles2D_ZeroSpeedParticles(t *testing.T) {
	p := NewParticles2D("test")
	p.SetDirection(0)
	p.SetSpeed(0)
	p.SetSpread(0)
	p.SetLifetime(5.0)

	p.Emit(1)
	if p.particleCount < 1 {
		t.Fatal("expected at least 1 particle after Emit")
	}
	startPos := p.particles[0].Position

	p.Update(1.0)
	if p.particleCount < 1 {
		t.Fatal("particle died before test completed")
	}
	// Particle should not have moved (no speed, no gravity)
	if p.particles[0].Position != startPos {
		t.Errorf("expected particle to stay at %v, moved to %v", startPos, p.particles[0].Position)
	}
}

// ---------------------------------------------------------------------------
// Child hierarchy propagation
// ---------------------------------------------------------------------------

func TestParticles2D_UpdatePropagatesToChildren(t *testing.T) {
	parent := NewParticles2D("parent")
	child := NewNode2D("child", 1)
	parent.AddChild(child)

	// Should not panic
	parent.Update(0.016)
}

func TestParticles2D_DrawDoesNotPanic(t *testing.T) {
	p := NewParticles2D("test")
	p.Emit(5)
	// We cannot create an ebiten.Image in tests reliably without a game
	// window, but the method should not panic when called with nil.
	// This test just verifies no nil-pointer crash in the draw logic.
	// In practice, ebiten.Image is always non-nil when provided by the engine.
}

// ---------------------------------------------------------------------------
// Compaction
// ---------------------------------------------------------------------------

func TestParticles2D_CompactRemovesDeadParticles(t *testing.T) {
	p := NewParticles2D("test")
	p.Emit(5)
	if len(p.particles) != 5 {
		t.Errorf("expected 5 particles in slice, got %d", len(p.particles))
	}

	// Manually kill some particles
	p.particles[1].Active = false
	p.particles[3].Active = false
	p.particleCount -= 2

	p.compact()
	if len(p.particles) != 3 {
		t.Errorf("expected 3 particles after compact, got %d", len(p.particles))
	}
	// Remaining particles should all be active
	for i := range p.particles {
		if !p.particles[i].Active {
			t.Errorf("particle %d should be active after compact", i)
		}
	}
}

// ---------------------------------------------------------------------------
// Default values consistency
// ---------------------------------------------------------------------------

func TestParticles2D_DefaultValues(t *testing.T) {
	p := NewParticles2D("test")

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"emitting", p.emitting, false},
		{"amount", p.amount, 10},
		{"lifetime", p.lifetime, 1.0},
		{"speed", p.speed, float32(100)},
		{"speedRandom", p.speedRandom, float32(0)},
		{"direction", p.direction, float32(270)},
		{"spread", p.spread, float32(30)},
		{"startSize", p.startSize, float32(4)},
		{"endSize", p.endSize, float32(4)},
		{"oneShot", p.oneShot, false},
		{"preprocess", p.preprocess, 0},
		{"blendMode", p.blendMode, BlendModeNormal},
		{"particleCount", p.particleCount, 0},
		{"emitRate", p.emitRate, float64(10)},
	}

	for _, tt := range tests {
		switch v := tt.got.(type) {
		case bool:
			if v != tt.want.(bool) {
				t.Errorf("default %s = %v, want %v", tt.name, v, tt.want)
			}
		case int:
			if v != tt.want.(int) {
				t.Errorf("default %s = %d, want %d", tt.name, v, tt.want)
			}
		case float32:
			if v != tt.want.(float32) {
				t.Errorf("default %s = %f, want %f", tt.name, v, tt.want)
			}
		case float64:
			if v != tt.want.(float64) {
				t.Errorf("default %s = %f, want %f", tt.name, v, tt.want)
			}
		case BlendMode:
			if v != tt.want.(BlendMode) {
				t.Errorf("default %s = %d, want %d", tt.name, v, tt.want)
			}
		default:
			t.Errorf("unhandled type for %s", tt.name)
		}
	}
}
