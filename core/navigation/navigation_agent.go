package navigation

import (
	stdMath "math"

	"github.com/ElioNeto/kora/core/math"
	"github.com/ElioNeto/kora/core/node"
)

// NavigationAgent2D is an AI agent that follows paths computed by a NavigationRegion2D.
// It embeds *Node2D so it can be added to the scene tree and moved along the path.
type NavigationAgent2D struct {
	*node.Node2D
	region    *NavigationRegion2D
	path      []math.Vector2
	pathIndex int
	speed     float64
	targetPos math.Vector2
	hasTarget bool
	arrived   bool
}

// NewNavigationAgent2D creates a navigation agent linked to the given region.
func NewNavigationAgent2D(name string, region *NavigationRegion2D) *NavigationAgent2D {
	n := node.NewNode2D(name, 0)
	return &NavigationAgent2D{
		Node2D:    n,
		region:    region,
		speed:     100.0, // default speed: 100 world units per second
		pathIndex: 0,
		arrived:   true,
	}
}

// SetTarget sets a destination for the agent in world coordinates.
// The agent will compute a path from its current position to the target.
func (na *NavigationAgent2D) SetTarget(x, y float64) {
	na.targetPos = math.NewVector2(float32(x), float32(y))
	na.hasTarget = true
	na.arrived = false
	na.pathIndex = 0

	pos := na.GetWorldPosition()
	na.path = na.region.GetPath(float64(pos.X), float64(pos.Y), x, y)
	if na.path == nil {
		na.arrived = true
	}
}

// SetTargetSmooth sets a destination and computes a smoothed path.
func (na *NavigationAgent2D) SetTargetSmooth(x, y float64) {
	na.targetPos = math.NewVector2(float32(x), float32(y))
	na.hasTarget = true
	na.arrived = false
	na.pathIndex = 0

	pos := na.GetWorldPosition()
	na.path = na.region.GetPathSmooth(float64(pos.X), float64(pos.Y), x, y)
	if na.path == nil {
		na.arrived = true
	}
}

// GetNextPoint returns the next point on the path the agent should move toward.
// Returns (0, 0, false) if there is no path or the agent has arrived.
func (na *NavigationAgent2D) GetNextPoint() (float64, float64, bool) {
	if na.arrived || na.path == nil || na.pathIndex >= len(na.path) {
		return 0, 0, false
	}
	p := na.path[na.pathIndex]
	return float64(p.X), float64(p.Y), true
}

// Update moves the agent along the path. Called every frame with the delta time.
func (na *NavigationAgent2D) Update(dt float64) {
	// Propagate to children first
	na.Node2D.Update(dt)

	if na.arrived || !na.hasTarget || na.path == nil {
		return
	}

	if na.pathIndex >= len(na.path) {
		na.arrived = true
		return
	}

	// Get current position and target waypoint
	pos := na.GetWorldPosition()
	waypoint := na.path[na.pathIndex]

	dx := float64(waypoint.X) - float64(pos.X)
	dy := float64(waypoint.Y) - float64(pos.Y)
	dist := stdMath.Sqrt(dx*dx + dy*dy)

	// Arrived at waypoint threshold (2 world units)
	if dist < 2.0 {
		na.pathIndex++
		if na.pathIndex >= len(na.path) {
			na.arrived = true
		}
		return
	}

	// Move toward waypoint
	step := na.speed * dt
	if step > dist {
		step = dist
	}

	ratio := step / dist
	newX := float64(pos.X) + dx*ratio
	newY := float64(pos.Y) + dy*ratio
	na.SetWorldPosition(float32(newX), float32(newY))
}

// HasArrived returns whether the agent reached its destination.
func (na *NavigationAgent2D) HasArrived() bool {
	return na.arrived
}

// GetPath returns the current path the agent is following.
func (na *NavigationAgent2D) GetPath() []math.Vector2 {
	return na.path
}

// IsTargetReachable returns whether the current target is reachable.
// Returns false if no target has been set or no path could be found.
func (na *NavigationAgent2D) IsTargetReachable() bool {
	if !na.hasTarget {
		return false
	}
	return na.path != nil
}

// SetSpeed sets the movement speed in world units per second.
func (na *NavigationAgent2D) SetSpeed(speed float64) {
	na.speed = speed
}

// GetSpeed returns the agent's movement speed.
func (na *NavigationAgent2D) GetSpeed() float64 {
	return na.speed
}

// Compile-time interface check: NavigationAgent2D must satisfy node.Node.
var _ node.Node = (*NavigationAgent2D)(nil)
