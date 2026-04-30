package physics

// CharacterBody2D is controlled by KScript code; NOT fully simulated by physics.
// Uses moveAndSlide/moveAndCollide for movement with collision response.
type CharacterBody2D struct {
	*RigidBody

	// Character-specific state
	FloorNormal    Vec2
	OnFloor        bool
	OnWallLeft     bool
	OnWallRight    bool
	OnCeiling      bool

	// Movement properties
	SafeMargin     float32 // Snaps to floor when close
	MaxSlides      int    // Max slide count in moveAndSlide
}

// NewCharacterBody2D creates a new CharacterBody2D.
func NewCharacterBody2D(entityID int, x, y, w, h float32) *CharacterBody2D {
	return &CharacterBody2D{
		RigidBody:  NewBody(entityID, x, y, w, h, BodyDynamic),
		SafeMargin: 0.1,
		MaxSlides:  4,
	}
}

// MoveAndSlide moves the body with velocity, sliding along surfaces.
// Returns the remaining velocity after collision response.
// This is the main movement function for character control.
func (c *CharacterBody2D) MoveAndSlide(velocity Vec2, bodies []*RigidBody) Vec2 {
	c.OnFloor = false
	c.OnWallLeft = false
	c.OnWallRight = false
	c.OnCeiling = false

	// Move in small steps to detect collisions along the path
	stepSize := float32(4.0) // 4 pixels per sub-step
	remaining := velocity
	steps := 0
	maxSteps := 100

	for steps < maxSteps && (remaining.X != 0 || remaining.Y != 0) {
		steps++

		// Calculate this sub-step movement
		stepX := remaining.X
		stepY := remaining.Y
		if stepX > stepSize {
			stepX = stepSize
		} else if stepX < -stepSize {
			stepX = -stepSize
		}
		if stepY > stepSize {
			stepY = stepSize
		} else if stepY < -stepSize {
			stepY = -stepSize
		}

		// Try this sub-step
		nextPos := Vec2{X: c.Pos.X + stepX, Y: c.Pos.Y + stepY}
		oldPos := c.Pos
		c.Pos = nextPos

		collided := false
		for _, b := range bodies {
			if b == c.RigidBody {
				continue
			}
			if (c.Layer & b.Mask) == 0 || (b.Layer & c.Mask) == 0 {
				continue
			}

			ov := TestAABB(c.RigidBody, b)
			if ov.Hit {
				// Restore position and resolve
				c.Pos = oldPos

				// Track surface types
				if ov.DepthY > 0 && ov.NormalY < 0 {
					c.OnFloor = true
					c.FloorNormal = Vec2{0, -1}
				} else if ov.DepthY > 0 && ov.NormalY > 0 {
					c.OnCeiling = true
				}
				if ov.DepthX > 0 {
					if ov.NormalX < 0 {
						c.OnWallLeft = true
					} else {
						c.OnWallRight = true
					}
				}

				// Consume velocity in collision direction
				if ov.DepthX > 0 {
					remaining.X = 0
				}
				if ov.DepthY > 0 {
					remaining.Y = 0
				}

				collided = true
			}
		}

		if !collided {
			// Sub-step succeeded, consume the velocity
			remaining.X -= stepX
			remaining.Y -= stepY
		} else {
			// Collision happened, stop moving in this direction
			// but continue to try other axis
		}

		if remaining.X*remaining.X+remaining.Y*remaining.Y < 1 {
			break
		}
	}

	return remaining
}

// MoveAndCollide moves the body and returns collision info.
// Does NOT slide; stops on collision and returns collision data.
func (c *CharacterBody2D) MoveAndCollide(motion Vec2, bodies []*RigidBody) *CollisionInfo {
	nextPos := Vec2{
		X: c.Pos.X + motion.X,
		Y: c.Pos.Y + motion.Y,
	}

	c.Pos = nextPos

	for _, b := range bodies {
		if b == c.RigidBody {
			continue
		}
		if (c.Layer & b.Mask) == 0 || (b.Layer & c.Mask) == 0 {
			continue
		}

		ov := TestAABB(c.RigidBody, b)
		if ov.Hit {
			return &CollisionInfo{
				Hit:    true,
				Body:   b,
				Normal: Vec2{ov.NormalX, ov.NormalY},
				Pos:    nextPos,
			}
		}
	}

	return &CollisionInfo{Hit: false}
}

// IsOnFloor returns true if character is on floor.
func (c *CharacterBody2D) IsOnFloor() bool {
	return c.OnFloor
}

// IsOnWall returns true if character is touching a wall.
func (c *CharacterBody2D) IsOnWall() bool {
	return c.OnWallLeft || c.OnWallRight
}

// IsOnCeiling returns true if character is touching ceiling.
func (c *CharacterBody2D) IsOnCeiling() bool {
	return c.OnCeiling
}

// GetFloorNormal returns the floor normal vector.
func (c *CharacterBody2D) GetFloorNormal() Vec2 {
	return c.FloorNormal
}

// SetVelocity sets the character's velocity.
func (c *CharacterBody2D) SetVelocity(vel Vec2) {
	c.RigidBody.Vel = vel
}

// GetVelocity returns the character's velocity.
func (c *CharacterBody2D) GetVelocity() Vec2 {
	return c.RigidBody.Vel
}

// KScript API helpers (float64 params)

// MoveAndSlideKS wraps MoveAndSlide for KScript.
func (c *CharacterBody2D) MoveAndSlideKS(vx, vy float64, world *PhysicsWorld) (float64, float64) {
	vel := Vec2{float32(vx), float32(vy)}
	remaining := c.MoveAndSlide(vel, world.GetBodies())
	return float64(remaining.X), float64(remaining.Y)
}

// MoveAndCollideKS wraps MoveAndCollide for KScript.
func (c *CharacterBody2D) MoveAndCollideKS(mx, my float64, world *PhysicsWorld) map[string]interface{} {
	motion := Vec2{float32(mx), float32(my)}
	info := c.MoveAndCollide(motion, world.GetBodies())

	return map[string]interface{}{
		"hit":    info.Hit,
		"normalX": float64(info.Normal.X),
		"normalY": float64(info.Normal.Y),
		"posX":   float64(info.Pos.X),
		"posY":   float64(info.Pos.Y),
	}
}

// CollisionInfo holds collision result data.
type CollisionInfo struct {
	Hit    bool
	Body   *RigidBody
	Normal Vec2
	Pos    Vec2
}
