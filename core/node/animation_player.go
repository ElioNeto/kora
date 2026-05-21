package node

import (
	stdMath "math"
)

// ---------------------------------------------------------------------------
// Easing
// ---------------------------------------------------------------------------

// EasingFunc maps t ∈ [0,1] → eased value ∈ [0,1].
type EasingFunc func(t float64) float64

var (
	EaseLinear    EasingFunc = func(t float64) float64 { return t }
	EaseIn        EasingFunc = func(t float64) float64 { return t * t }
	EaseOut       EasingFunc = func(t float64) float64 { return t * (2 - t) }
	EaseInOut     EasingFunc = func(t float64) float64 {
		if t < 0.5 {
			return 2 * t * t
		}
		return -1 + (4-2*t)*t
	}
	EaseElasticOut EasingFunc = func(t float64) float64 {
		if t == 0 || t == 1 {
			return t
		}
		return stdMath.Pow(2, -10*t)*stdMath.Sin((t*10-0.75)*(2*stdMath.Pi/3)) + 1
	}
	EaseBounceOut EasingFunc = func(t float64) float64 {
		switch {
		case t < 1/2.75:
			return 7.5625 * t * t
		case t < 2/2.75:
			t -= 1.5 / 2.75
			return 7.5625*t*t + 0.75
		case t < 2.5/2.75:
			t -= 2.25 / 2.75
			return 7.5625*t*t + 0.9375
		default:
			t -= 2.625 / 2.75
			return 7.5625*t*t + 0.984375
		}
	}
)

// EasingByName returns the easing function for a given name.
// Supported: "linear", "ease_in", "ease_out", "ease_in_out", "elastic_out", "bounce_out".
// Defaults to EaseLinear if the name is not recognised.
func EasingByName(name string) EasingFunc {
	switch name {
	case "linear":
		return EaseLinear
	case "ease_in":
		return EaseIn
	case "ease_out":
		return EaseOut
	case "ease_in_out":
		return EaseInOut
	case "elastic_out":
		return EaseElasticOut
	case "bounce_out":
		return EaseBounceOut
	default:
		return EaseLinear
	}
}

// ---------------------------------------------------------------------------
// Property type
// ---------------------------------------------------------------------------

// AnimProperty identifies a numeric property on a Node2D.
type AnimProperty int

const (
	AnimPropPosX AnimProperty = iota
	AnimPropPosY
	AnimPropRotation
	AnimPropScaleX
	AnimPropScaleY
	AnimPropAlpha
)

// AnimPropertyFromString parses a property name.
// Supported: "x", "y", "rotation", "scale_x", "scale_y", "alpha".
func AnimPropertyFromString(s string) (AnimProperty, bool) {
	switch s {
	case "x":
		return AnimPropPosX, true
	case "y":
		return AnimPropPosY, true
	case "rotation":
		return AnimPropRotation, true
	case "scale_x":
		return AnimPropScaleX, true
	case "scale_y":
		return AnimPropScaleY, true
	case "alpha":
		return AnimPropAlpha, true
	default:
		return AnimPropPosX, false
	}
}

// ---------------------------------------------------------------------------
// Keyframe
// ---------------------------------------------------------------------------

// Keyframe defines a value at a point in time.
type Keyframe struct {
	Time   float64     // seconds from the start of the clip
	Value  float64     // target value
	Easing EasingFunc  // easing function for the segment leading TO this keyframe
}

// ---------------------------------------------------------------------------
// AnimationClip
// ---------------------------------------------------------------------------

// AnimationClip is a named, reusable animation definition.
type AnimationClip struct {
	Name      string
	Target    string               // node path relative to the AnimationPlayer's parent
	Property  AnimProperty         // which property to animate
	Keyframes []Keyframe           // sorted by Time
	Duration  float64              // total duration in seconds (computed from the last keyframe)
	Loop      bool                 // whether the animation loops
}

// ---------------------------------------------------------------------------
// AnimationPlayer
// ---------------------------------------------------------------------------

// AnimationPlayer is a node that drives property animations on other nodes.
//
// Usage:
//
//	player := NewAnimationPlayer("anim")
//	player.AddClip(&AnimationClip{
//	    Name: "bounce",
//	    Target: "Player",
//	    Property: AnimPropPosY,
//	    Keyframes: []Keyframe{
//	        {Time: 0, Value: 0, Easing: EaseLinear},
//	        {Time: 0.5, Value: -50, Easing: EaseOut},
//	        {Time: 1, Value: 0, Easing: EaseIn},
//	    },
//	    Duration: 1,
//	    Loop: true,
//	})
//	player.Play("bounce")
type AnimationPlayer struct {
	*Node2D

	clips   map[string]*AnimationClip
	current *AnimationClip
	elapsed float64
	playing bool
	paused  bool
	done    bool

	// Callbacks
	onFinish func(name string)
}

// NewAnimationPlayer creates a new AnimationPlayer node.
func NewAnimationPlayer(name string) *AnimationPlayer {
	return &AnimationPlayer{
		Node2D: NewNode2D(name, 0),
		clips:  make(map[string]*AnimationClip),
	}
}

// AddClip registers an animation clip. If a clip with the same name already
// exists, it is replaced.
func (ap *AnimationPlayer) AddClip(clip *AnimationClip) {
	ap.clips[clip.Name] = clip
}

// GetClip returns a registered clip by name, or nil.
func (ap *AnimationPlayer) GetClip(name string) *AnimationClip {
	return ap.clips[name]
}

// ClipNames returns the names of all registered clips.
func (ap *AnimationPlayer) ClipNames() []string {
	names := make([]string, 0, len(ap.clips))
	for n := range ap.clips {
		names = append(names, n)
	}
	return names
}

// RemoveClip removes a clip by name.
func (ap *AnimationPlayer) RemoveClip(name string) {
	delete(ap.clips, name)
}

// Play starts a named animation. Returns false if the clip does not exist.
func (ap *AnimationPlayer) Play(name string) bool {
	clip, ok := ap.clips[name]
	if !ok {
		return false
	}
	ap.current = clip
	ap.elapsed = 0
	ap.playing = true
	ap.paused = false
	ap.done = false
	return true
}

// Stop stops the current animation and resets to the beginning.
func (ap *AnimationPlayer) Stop() {
	ap.playing = false
	ap.paused = false
	ap.elapsed = 0
	ap.done = false
}

// Pause pauses the current animation at the current position.
func (ap *AnimationPlayer) Pause() {
	if ap.playing {
		ap.paused = true
	}
}

// Resume resumes a paused animation.
func (ap *AnimationPlayer) Resume() {
	if ap.paused {
		ap.paused = false
	}
}

// IsPlaying returns true if an animation is currently active and not paused.
func (ap *AnimationPlayer) IsPlaying() bool {
	return ap.playing && !ap.paused
}

// IsPaused returns true if the animation is paused.
func (ap *AnimationPlayer) IsPaused() bool {
	return ap.paused
}

// IsDone returns true if a non-looping animation has finished.
func (ap *AnimationPlayer) IsDone() bool {
	return ap.done
}

// CurrentClip returns the currently playing clip, or nil.
func (ap *AnimationPlayer) CurrentClip() *AnimationClip {
	return ap.current
}

// OnFinish sets a callback that fires when a non-looping animation completes.
func (ap *AnimationPlayer) OnFinish(fn func(name string)) {
	ap.onFinish = fn
}

// CurrentTime returns the current playback time in seconds.
func (ap *AnimationPlayer) CurrentTime() float64 {
	return ap.elapsed
}

// Progress returns the current playback progress as a 0-1 value.
// Returns 0 if no clip is playing.
func (ap *AnimationPlayer) Progress() float64 {
	if ap.current == nil || ap.current.Duration <= 0 {
		return 0
	}
	return ap.elapsed / ap.current.Duration
}

// ---------------------------------------------------------------------------
// Update — applies the current keyframe to the target node
// ---------------------------------------------------------------------------

// Update evaluates the current animation and applies it to the target node.
func (ap *AnimationPlayer) Update(dt float64) {
	// Propagate to children first
	if ap.Node2D != nil {
		ap.Node2D.Update(dt)
	}

	if !ap.playing || ap.paused || ap.current == nil {
		return
	}

	ap.elapsed += dt

	if ap.elapsed >= ap.current.Duration {
		if ap.current.Loop {
			ap.elapsed = stdMath.Mod(ap.elapsed, ap.current.Duration)
		} else {
			ap.elapsed = ap.current.Duration
			ap.playing = false
			ap.done = true
			// Apply final keyframe
			ap.applyAtTime(ap.current.Duration)
			if ap.onFinish != nil {
				ap.onFinish(ap.current.Name)
			}
			return
		}
	}

	ap.applyAtTime(ap.elapsed)
}

// applyAtTime evaluates the animation at time t and applies the value to
// the target node's property.
func (ap *AnimationPlayer) applyAtTime(t float64) {
	clip := ap.current
	if clip == nil || len(clip.Keyframes) == 0 {
		return
	}

	// Find the target node.
	target := ap.resolveTarget()
	if target == nil {
		return
	}

	// Evaluate the value at time t.
	val := ap.evaluate(t)

	// Apply to the target's property.
	switch clip.Property {
	case AnimPropPosX:
		target.SetX(float32(val))
	case AnimPropPosY:
		target.SetY(float32(val))
	case AnimPropRotation:
		target.SetRotation(float32(val))
	case AnimPropScaleX:
		target.SetScaleX(float32(val))
	case AnimPropScaleY:
		target.SetScaleY(float32(val))
	case AnimPropAlpha:
		// Alpha is stored on Sprite2D, not Node2D. For now this is a no-op
		// on plain Node2D; Sprite2D handles it via SetAlpha.
		if sp, ok := interface{}(target).(*Sprite2D); ok {
			sp.SetAlpha(float32(val))
		}
	}
}

// resolveTarget finds the target node by the clip's Target path.
// If Target is empty, the AnimationPlayer's parent is used as the target.
func (ap *AnimationPlayer) resolveTarget() *Node2D {
	clip := ap.current
	if clip == nil {
		return nil
	}

	targetPath := clip.Target
	if targetPath == "" {
		// Default to parent node.
		return ap.GetParent()
	}

	// Resolve relative to the AnimationPlayer's parent.
	parent := ap.GetParent()
	if parent == nil {
		return nil
	}

	// Try direct child lookup first.
	if child := parent.GetChild(targetPath); child != nil {
		return child
	}

	// Try path lookup (e.g., "Player/Sprite").
	n := parent.GetNode(targetPath)
	if n == nil {
		return nil
	}
	// GetNode returns Node interface, extract *Node2D.
	if n2d, ok := n.(*Node2D); ok {
		return n2d
	}
	return nil
}

// evaluate computes the interpolated value at time t using the clip's keyframes.
func (ap *AnimationPlayer) evaluate(t float64) float64 {
	clip := ap.current
	if clip == nil || len(clip.Keyframes) == 0 {
		return 0
	}

	kfs := clip.Keyframes

	// Before or at the first keyframe.
	if t <= kfs[0].Time {
		return kfs[0].Value
	}

	// After or at the last keyframe.
	last := kfs[len(kfs)-1]
	if t >= last.Time {
		return last.Value
	}

	// Find the two keyframes bracketing time t.
	for i := 1; i < len(kfs); i++ {
		if t < kfs[i].Time {
			prev := kfs[i-1]
			next := kfs[i]
			segmentDuration := next.Time - prev.Time
			if segmentDuration <= 0 {
				return next.Value
			}

			// Normalise t within the segment.
			localT := (t - prev.Time) / segmentDuration

			// Apply the easing function of the NEXT keyframe.
			easing := next.Easing
			if easing == nil {
				easing = EaseLinear
			}
			easedT := easing(localT)

			return prev.Value + (next.Value-prev.Value)*easedT
		}
	}

	return last.Value
}

// Compile-time interface checks
var _ Node = (*AnimationPlayer)(nil)
