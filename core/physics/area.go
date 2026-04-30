package physics

// Area2D is a detection zone without rigid physics response.
// Fires events when bodies/areas enter or exit.
type Area2D struct {
	EntityID int
	Pos      Vec2
	HalfW    float32
	HalfH    float32
	Shape    ShapeType
	Radius   float32

	// Collision filtering
	Layer uint16
	Mask  uint16

	// Monitoring
	MonitorEnabled bool

	// Overlapping bodies and areas
	overlappingBodies []*RigidBody
	overlappingAreas []*Area2D

	// Callbacks (set by KScript)
	OnBodyEntered  func(body *RigidBody)
	OnBodyExited   func(body *RigidBody)
	OnAreaEntered  func(area *Area2D)
	OnAreaExited   func(area *Area2D)
}

// NewArea2D creates a new Area2D.
func NewArea2D(entityID int, x, y, w, h float32) *Area2D {
	return &Area2D{
		EntityID:        entityID,
		Pos:             Vec2{x, y},
		HalfW:           w / 2,
		HalfH:           h / 2,
		Shape:           ShapeRect,
		Layer:           DefaultLayer,
		Mask:            DefaultMask,
		MonitorEnabled:  true,
		overlappingBodies: make([]*RigidBody, 0),
		overlappingAreas:  make([]*Area2D, 0),
	}
}

// AABB returns the bounding box of the area.
func (a *Area2D) AABB() (minX, minY, maxX, maxY float32) {
	if a.Shape == ShapeCircle {
		minX = a.Pos.X - a.Radius
		minY = a.Pos.Y - a.Radius
		maxX = a.Pos.X + a.Radius
		maxY = a.Pos.Y + a.Radius
	} else {
		minX = a.Pos.X - a.HalfW
		minY = a.Pos.Y - a.HalfH
		maxX = a.Pos.X + a.HalfW
		maxY = a.Pos.Y + a.HalfH
	}
	return minX, minY, maxX, maxY
}

// CheckOverlaps detects and fires events for body/area overlaps.
// Should be called each physics step.
func (a *Area2D) CheckOverlaps(world *PhysicsWorld) {
	if !a.MonitorEnabled {
		return
	}

	// Check body overlaps
	a.checkBodyOverlaps(world)

	// Check area overlaps
	a.checkAreaOverlaps(world)
}

// checkBodyOverlaps detects bodies entering/exiting the area.
func (a *Area2D) checkBodyOverlaps(world *PhysicsWorld) {
	minX, minY, maxX, maxY := a.AABB()

	// Bodies that are currently overlapping
	currentOverlaps := make(map[*RigidBody]bool)

	for _, b := range world.bodies {
		if b.EntityID == a.EntityID {
			continue
		}
		if (a.Layer & b.Mask) == 0 || (b.Layer & a.Mask) == 0 {
			continue
		}

		bMinX, bMinY, bMaxX, bMaxY := b.AABB()
		if bMaxX <= minX || bMinX >= maxX || bMaxY <= minY || bMinY >= maxY {
			continue
		}

		// Body is overlapping
		currentOverlaps[b] = true

		// Check if this is a new overlap
		if !a.isBodyOverlapping(b) {
			a.overlappingBodies = append(a.overlappingBodies, b)
			if a.OnBodyEntered != nil {
				a.OnBodyEntered(b)
			}
		}
	}

	// Check for bodies that exited
	var remaining []*RigidBody
	for _, b := range a.overlappingBodies {
		if currentOverlaps[b] {
			remaining = append(remaining, b)
		} else {
			if a.OnBodyExited != nil {
				a.OnBodyExited(b)
			}
		}
	}
	a.overlappingBodies = remaining
}

// checkAreaOverlaps detects areas entering/exiting the area.
func (a *Area2D) checkAreaOverlaps(world *PhysicsWorld) {
	minX, minY, maxX, maxY := a.AABB()

	// Areas that are currently overlapping
	currentOverlaps := make(map[*Area2D]bool)

	for _, other := range world.areas {
		if other.EntityID == a.EntityID {
			continue
		}
		if (a.Layer & other.Mask) == 0 || (other.Layer & a.Mask) == 0 {
			continue
		}

		otherMinX, otherMinY, otherMaxX, otherMaxY := other.AABB()
		if otherMaxX <= minX || otherMinX >= maxX || otherMaxY <= minY || otherMinY >= maxY {
			continue
		}

		// Area is overlapping
		currentOverlaps[other] = true

		// Check if this is a new overlap
		if !a.isAreaOverlapping(other) {
			a.overlappingAreas = append(a.overlappingAreas, other)
			if a.OnAreaEntered != nil {
				a.OnAreaEntered(other)
			}
		}
	}

	// Check for areas that exited
	var remaining []*Area2D
	for _, other := range a.overlappingAreas {
		if currentOverlaps[other] {
			remaining = append(remaining, other)
		} else {
			if a.OnAreaExited != nil {
				a.OnAreaExited(other)
			}
		}
	}
	a.overlappingAreas = remaining
}

// isAreaOverlapping checks if an area is in the overlap list.
func (a *Area2D) isAreaOverlapping(area *Area2D) bool {
	for _, other := range a.overlappingAreas {
		if other == area {
			return true
		}
	}
	return false
}

// isBodyOverlapping checks if a body is in the overlap list.
func (a *Area2D) isBodyOverlapping(body *RigidBody) bool {
	for _, b := range a.overlappingBodies {
		if b == body {
			return true
		}
	}
	return false
}

// GetOverlappingBodies returns all currently overlapping bodies.
func (a *Area2D) GetOverlappingBodies() []*RigidBody {
	result := make([]*RigidBody, len(a.overlappingBodies))
	copy(result, a.overlappingBodies)
	return result
}

// GetOverlappingBodyCount returns the number of overlapping bodies.
func (a *Area2D) GetOverlappingBodyCount() int {
	return len(a.overlappingBodies)
}

// GetOverlappingAreas returns all currently overlapping areas.
func (a *Area2D) GetOverlappingAreas() []*Area2D {
	result := make([]*Area2D, len(a.overlappingAreas))
	copy(result, a.overlappingAreas)
	return result
}

// GetOverlappingAreaCount returns the number of overlapping areas.
func (a *Area2D) GetOverlappingAreaCount() int {
	return len(a.overlappingAreas)
}

// SetMonitorEnabled enables/disables detection.
func (a *Area2D) SetMonitorEnabled(enabled bool) {
	a.MonitorEnabled = enabled
}

// IsMonitorEnabled returns whether detection is enabled.
func (a *Area2D) IsMonitorEnabled() bool {
	return a.MonitorEnabled
}

// KScript API helpers

// GetOverlappingBodiesKS returns overlapping bodies as interface slice for KScript.
func (a *Area2D) GetOverlappingBodiesKS() []interface{} {
	result := make([]interface{}, len(a.overlappingBodies))
	for i, b := range a.overlappingBodies {
		result[i] = b
	}
	return result
}

// GetOverlappingAreasKS returns overlapping areas as interface slice for KScript.
func (a *Area2D) GetOverlappingAreasKS() []interface{} {
	result := make([]interface{}, len(a.overlappingAreas))
	for i, area := range a.overlappingAreas {
		result[i] = area
	}
	return result
}

// SetOnBodyEnteredKS sets the body entered callback from KScript.
func (a *Area2D) SetOnBodyEnteredKS(fn func(body interface{})) {
	a.OnBodyEntered = func(b *RigidBody) {
		fn(b)
	}
}

// SetOnBodyExitedKS sets the body exited callback from KScript.
func (a *Area2D) SetOnBodyExitedKS(fn func(body interface{})) {
	a.OnBodyExited = func(b *RigidBody) {
		fn(b)
	}
}

// SetOnAreaEnteredKS sets the area entered callback from KScript.
func (a *Area2D) SetOnAreaEnteredKS(fn func(area interface{})) {
	a.OnAreaEntered = func(ar *Area2D) {
		fn(ar)
	}
}

// SetOnAreaExitedKS sets the area exited callback from KScript.
func (a *Area2D) SetOnAreaExitedKS(fn func(area interface{})) {
	a.OnAreaExited = func(ar *Area2D) {
		fn(ar)
	}
}

// RegisterArea2DAPI returns the KScript API for Area2D.
func RegisterArea2DAPI() map[string]interface{} {
	return map[string]interface{}{
		"getOverlappingBodies": func(instance *Area2D) []interface{} {
			return instance.GetOverlappingBodiesKS()
		},
		"getOverlappingAreas": func(instance *Area2D) []interface{} {
			return instance.GetOverlappingAreasKS()
		},
		"setMonitorEnabled": func(instance *Area2D, enabled bool) {
			instance.SetMonitorEnabled(enabled)
		},
		"isMonitorEnabled": func(instance *Area2D) bool {
			return instance.IsMonitorEnabled()
		},
		"setOnBodyEntered": func(instance *Area2D, fn func(body interface{})) {
			instance.SetOnBodyEnteredKS(fn)
		},
		"setOnBodyExited": func(instance *Area2D, fn func(body interface{})) {
			instance.SetOnBodyExitedKS(fn)
		},
		"setOnAreaEntered": func(instance *Area2D, fn func(area interface{})) {
			instance.SetOnAreaEnteredKS(fn)
		},
		"setOnAreaExited": func(instance *Area2D, fn func(area interface{})) {
			instance.SetOnAreaExitedKS(fn)
		},
	}
}
