// Package node implements the core node system for Kora Engine
package node

import (
	"fmt"
	"image/color"
	stdMath "math"

	"github.com/ElioNeto/kora/core/math"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// GizmoColor stores RGBA color values for gizmo rendering.
// Each component is in the range [0, 1].
type GizmoColor struct {
	R, G, B, A float32
}

// BonePose holds the local transform data for a bone at a specific pose.
type BonePose struct {
	Name     string
	Position math.Vector2 // local position relative to parent
	Rotation float32      // local rotation in degrees
	ScaleX   float32
	ScaleY   float32
}

// Bone2D is a single bone in a skeleton hierarchy.
type Bone2D struct {
	*Node2D

	// Bone dimensions
	Length        float32 // bone length in pixels
	DefaultLength float32 // rest length

	// IK (Inverse Kinematics) settings
	IKMode       bool // whether this bone participates in IK solving
	IKSegments   int  // number of segments (for CCD solver)
	BendDirection int  // +1 or -1 for IK bend direction

	// Skin attachment
	SpriteName string // optional sprite attached to this bone

	// Visual (for gizmo drawing)
	ShowGizmo  bool
	GizmoColor GizmoColor
}

// Skeleton2D is a parent node managing a tree of Bone2D children.
type Skeleton2D struct {
	*Node2D

	// All bones in the skeleton (collected from children)
	bones       []*Bone2D
	bonesByName map[string]*Bone2D

	// Rest pose
	restPose    map[string]BonePose
	currentPose map[string]BonePose

	// IK target (world position for IK solver)
	ikTargetX, ikTargetY float64
	ikActive             bool

	// Animation
	animator *AnimationPlayer
}

// ---------------------------------------------------------------------------
// Bone2D constructor & methods
// ---------------------------------------------------------------------------

// NewBone2D creates a new Bone2D with the given name.
func NewBone2D(name string) *Bone2D {
	return &Bone2D{
		Node2D:        NewNode2D(name, 0),
		Length:        32.0,
		DefaultLength: 32.0,
		BendDirection: 0,
		ShowGizmo:     false,
		GizmoColor:    GizmoColor{R: 1, G: 1, B: 1, A: 1},
	}
}

// SetLength sets the bone length in pixels.
func (b *Bone2D) SetLength(length float32) {
	b.Length = length
}

// GetLength returns the bone length in pixels.
func (b *Bone2D) GetLength() float32 {
	return b.Length
}

// GetEndPosition returns the world position of the bone tip.
// The end position is computed as: worldPos + rotated_direction * length.
func (b *Bone2D) GetEndPosition() math.Vector2 {
	worldPos := b.GetWorldPosition()
	worldRot := b.GetWorldRotation()

	// Convert degrees to radians
	rad := float64(worldRot) * stdMath.Pi / 180.0

	dirX := float32(stdMath.Cos(rad))
	dirY := float32(stdMath.Sin(rad))

	return math.Vector2{
		X: worldPos.X + dirX*b.Length,
		Y: worldPos.Y + dirY*b.Length,
	}
}

// SetIKMode enables or disables this bone's participation in IK solving.
func (b *Bone2D) SetIKMode(enabled bool) {
	b.IKMode = enabled
}

// SetBendDirection sets the IK bend direction constraint.
// +1 means bend only in the positive direction, -1 in the negative, 0 for no constraint.
func (b *Bone2D) SetBendDirection(dir int) {
	if dir < -1 {
		dir = -1
	}
	if dir > 1 {
		dir = 1
	}
	b.BendDirection = dir
}

// ---------------------------------------------------------------------------
// Skeleton2D constructor & methods
// ---------------------------------------------------------------------------

// NewSkeleton2D creates a new Skeleton2D node.
func NewSkeleton2D(name string) *Skeleton2D {
	return &Skeleton2D{
		Node2D:      NewNode2D(name, 0),
		bonesByName: make(map[string]*Bone2D),
		restPose:    make(map[string]BonePose),
		currentPose: make(map[string]BonePose),
	}
}

// Update processes the skeleton each frame: collects bones, applies IK if active,
// updates the animation player, and propagates to children.
func (s *Skeleton2D) Update(dt float64) {
	// Collect bones from registered hierarchy
	s.collectBones()

	// Process IK if active
	if s.ikActive && len(s.bones) > 0 {
		// Find the deepest chain and solve IK toward the target
		endBone := s.findDeepestBone()
		if endBone != nil {
			s.SolveIKChain(endBone.GetName(), s.ikTargetX, s.ikTargetY)
		}
	}

	// Update animation player
	if s.animator != nil {
		s.animator.Update(dt)
	}

	// Propagate to children (base Node2D update)
	for _, child := range s.children {
		if child != nil {
			child.Update(dt)
		}
	}
}

// Draw renders the skeleton's bone gizmos to the screen.
// Only bones with ShowGizmo enabled are drawn.
func (s *Skeleton2D) Draw(screen *ebiten.Image) {
	if !s.visible || !s.alive {
		return
	}

	// Draw gizmos for all bones with ShowGizmo enabled
	for _, bone := range s.bones {
		if bone.ShowGizmo {
			s.drawBoneGizmo(bone, screen)
		}
	}

	// Propagate to children
	for _, child := range s.children {
		if child != nil {
			child.Draw(screen)
		}
	}
}

// drawBoneGizmo draws a single bone's gizmo: line from joint to tip, plus a circle at the joint.
func (s *Skeleton2D) drawBoneGizmo(bone *Bone2D, screen *ebiten.Image) {
	startPos := bone.GetWorldPosition()
	endPos := bone.GetEndPosition()

	// Determine color
	clr := s.gizmoColorToRGBA(bone.GizmoColor)

	// Draw bone line
	ebitenutil.DrawLine(screen,
		float64(startPos.X), float64(startPos.Y),
		float64(endPos.X), float64(endPos.Y),
		clr,
	)

	// Draw joint circle at start position
	ebitenutil.DrawCircle(screen,
		float64(startPos.X), float64(startPos.Y),
		3.0, // radius
		clr,
	)

	// If this bone has children, also draw a small circle at the end position
	if len(bone.children) > 0 {
		ebitenutil.DrawCircle(screen,
			float64(endPos.X), float64(endPos.Y),
			2.0,
			clr,
		)
	}
}

// gizmoColorToRGBA converts a GizmoColor to color.RGBA.
// If the color is zero-valued (all components 0), default white is used.
func (s *Skeleton2D) gizmoColorToRGBA(gc GizmoColor) color.RGBA {
	if gc.R == 0 && gc.G == 0 && gc.B == 0 && gc.A == 0 {
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	return color.RGBA{
		R: uint8(clamp01(gc.R) * 255),
		G: uint8(clamp01(gc.G) * 255),
		B: uint8(clamp01(gc.B) * 255),
		A: uint8(clamp01(gc.A) * 255),
	}
}

// clamp01 clamps a float32 value to the [0, 1] range.
func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// ---------------------------------------------------------------------------
// Bone management
// ---------------------------------------------------------------------------

// GetBone finds a bone by name. Returns nil if not found.
func (s *Skeleton2D) GetBone(name string) *Bone2D {
	return s.bonesByName[name]
}

// GetAllBones returns all bones in the skeleton as a flat slice.
func (s *Skeleton2D) GetAllBones() []*Bone2D {
	return s.bones
}

// AddBone adds a bone under the named parent.
// If parentName is empty, the bone is added as a direct child of the skeleton.
// The parent must already exist in the skeleton.
func (s *Skeleton2D) AddBone(parentName string, bone *Bone2D) {
	if bone == nil {
		return
	}

	if parentName == "" {
		// Add as direct child of skeleton
		s.Node2D.AddChild(bone)
	} else {
		// Find parent and add as its child
		parent := s.GetBone(parentName)
		if parent != nil {
			parent.AddChild(bone)
		} else {
			panic(fmt.Sprintf("Skeleton2D.AddBone: parent bone '%s' not found", parentName))
		}
	}

	// Register bone
	if _, exists := s.bonesByName[bone.GetName()]; !exists {
		s.bonesByName[bone.GetName()] = bone
		s.bones = append(s.bones, bone)
	}
}

// collectBones rebuilds the flat bones slice from the name map.
// Called each frame in Update to ensure consistency.
func (s *Skeleton2D) collectBones() {
	s.bones = make([]*Bone2D, 0, len(s.bonesByName))
	for _, bone := range s.bonesByName {
		s.bones = append(s.bones, bone)
	}
}

// findDeepestBone returns the bone with the deepest hierarchy depth.
// This is used as the default end effector for IK.
func (s *Skeleton2D) findDeepestBone() *Bone2D {
	var deepest *Bone2D
	var maxDepth int

	var walk func(n *Node2D, depth int)
	walk = func(n *Node2D, depth int) {
		if n == nil {
			return
		}
		// Check if this Node2D corresponds to a registered bone
		if bone, ok := s.bonesByName[n.GetName()]; ok {
			if depth > maxDepth {
				maxDepth = depth
				deepest = bone
			}
		}
		for _, child := range n.GetChildren() {
			walk(child, depth+1)
		}
	}

	walk(s.Node2D, 0)
	return deepest
}

// ---------------------------------------------------------------------------
// Rest pose
// ---------------------------------------------------------------------------

// StoreRestPose saves the current bone transforms as the rest pose.
func (s *Skeleton2D) StoreRestPose() {
	s.restPose = make(map[string]BonePose, len(s.bones))
	for _, bone := range s.bones {
		s.restPose[bone.GetName()] = BonePose{
			Name:     bone.GetName(),
			Position: bone.GetPosition(),
			Rotation: bone.GetRotation(),
			ScaleX:   bone.GetScaleX(),
			ScaleY:   bone.GetScaleY(),
		}
	}
}

// ApplyRestPose resets all bones to their stored rest pose.
func (s *Skeleton2D) ApplyRestPose() {
	if len(s.restPose) == 0 {
		return
	}
	for _, bone := range s.bones {
		if pose, ok := s.restPose[bone.GetName()]; ok {
			bone.SetPosition(pose.Position.X, pose.Position.Y)
			bone.SetRotation(pose.Rotation)
			bone.SetScaleX(pose.ScaleX)
			bone.SetScaleY(pose.ScaleY)
		}
	}
}

// ---------------------------------------------------------------------------
// IK - Inverse Kinematics (CCD solver)
// ---------------------------------------------------------------------------

// SetIKTarget sets the world position target for IK solving and activates IK.
func (s *Skeleton2D) SetIKTarget(x, y float64) {
	s.ikTargetX = x
	s.ikTargetY = y
	s.ikActive = true
}

// ClearIKTarget deactivates the IK target.
func (s *Skeleton2D) ClearIKTarget() {
	s.ikActive = false
}

// SolveIK applies the Cyclic Coordinate Descent (CCD) IK solver to the given
// bone chain. The chain must be ordered from root to end effector.
// The solver iterates multiple times for convergence.
func (s *Skeleton2D) SolveIK(chain []*Bone2D, targetX, targetY float64) {
	if len(chain) == 0 {
		return
	}

	iterations := 20
	if chain[0].IKSegments > 0 {
		iterations = chain[0].IKSegments
	}

	for iter := 0; iter < iterations; iter++ {
		for i := len(chain) - 1; i >= 0; i-- {
			bone := chain[i]

			// Get bone start position in world space
			startPos := bone.GetWorldPosition()

			// Get current end effector position (tip of the last bone in the chain)
			endBone := chain[len(chain)-1]
			endEffector := endBone.GetEndPosition()

			// Calculate angle from bone start to target
			toTarget := stdMath.Atan2(targetY-float64(startPos.Y), targetX-float64(startPos.X))

			// Calculate angle from bone start to current end effector
			toEnd := stdMath.Atan2(
				float64(endEffector.Y-startPos.Y),
				float64(endEffector.X-startPos.X),
			)

			// Rotation difference (in radians)
			delta := toTarget - toEnd

			// Normalize delta to [-pi, pi]
			for delta > stdMath.Pi {
				delta -= 2 * stdMath.Pi
			}
			for delta < -stdMath.Pi {
				delta += 2 * stdMath.Pi
			}

			// Apply bend direction constraint
			if bone.BendDirection > 0 && delta < 0 {
				delta = 0
			} else if bone.BendDirection < 0 && delta > 0 {
				delta = 0
			}

			// Apply rotation (convert radians to degrees)
			rotationDeg := delta * 180.0 / stdMath.Pi
			bone.SetRotation(bone.GetRotation() + float32(rotationDeg))
		}
	}
}

// SolveIKChain automatically builds a bone chain from the named end bone to the
// skeleton root and solves IK toward the given target position.
func (s *Skeleton2D) SolveIKChain(endBoneName string, targetX, targetY float64) {
	endBone := s.GetBone(endBoneName)
	if endBone == nil {
		return
	}

	// Build chain from end bone to root
	chain := s.buildChain(endBone)
	if len(chain) == 0 {
		return
	}

	s.SolveIK(chain, targetX, targetY)
}

// buildChain builds a bone chain from the given end bone up to the skeleton root.
// The resulting chain is ordered from root to end effector.
func (s *Skeleton2D) buildChain(endBone *Bone2D) []*Bone2D {
	chain := make([]*Bone2D, 0)

	// Walk from end bone up to root via parent hierarchy
	current := endBone
	for current != nil {
		chain = append(chain, current)

		// Get parent as *Node2D
		parent := current.GetParent()
		if parent == nil {
			break
		}

		// Check if parent is a registered bone in this skeleton
		if parentBone, ok := s.bonesByName[parent.GetName()]; ok {
			current = parentBone
		} else {
			break
		}
	}

	// Reverse chain so root is first
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}

	return chain
}

// ---------------------------------------------------------------------------
// Animation
// ---------------------------------------------------------------------------

// SetAnimator attaches an AnimationPlayer to this skeleton.
func (s *Skeleton2D) SetAnimator(ap *AnimationPlayer) {
	s.animator = ap
}

// Animator returns the attached AnimationPlayer, or nil.
func (s *Skeleton2D) Animator() *AnimationPlayer {
	return s.animator
}

// PlayAnimation starts a named animation on the attached AnimationPlayer.
// Returns false if no animator is attached or the clip is not found.
func (s *Skeleton2D) PlayAnimation(name string) bool {
	if s.animator == nil {
		return false
	}
	return s.animator.Play(name)
}

// StopAnimation stops the current animation.
func (s *Skeleton2D) StopAnimation() {
	if s.animator != nil {
		s.animator.Stop()
	}
}

// ---------------------------------------------------------------------------
// Compile-time interface checks
// ---------------------------------------------------------------------------

var (
	_ Node = (*Bone2D)(nil)
	_ Node = (*Skeleton2D)(nil)
)
