package async

import (
	kmath "github.com/ElioNeto/kora/core/math"
)

// Easing functions — all map t∈[0,1] → value∈[0,1].

// EasingFn is a function that maps time t [0,1] to an eased value [0,1].
type EasingFn = kmath.EasingFn

// Easing presets — delegates to the full set in kmath.
var (
	Linear     EasingFn = kmath.Linear
	EaseIn     EasingFn = kmath.InQuad
	EaseOut    EasingFn = kmath.OutQuad
	EaseInOut  EasingFn = kmath.InOutQuad
	ElasticOut EasingFn = kmath.OutElastic
	BounceOut  EasingFn = kmath.OutBounce
)

// ----------------------------------------------------------------------------
// Float tween
// ----------------------------------------------------------------------------

type tweenFloatTask struct {
	ptr      *float64
	start    float64
	target   float64
	duration float64
	elapsed  float64
	easing   EasingFn
}

// TweenFloat animates *ptr from its current value to target over duration seconds.
func TweenFloat(ptr *float64, target, duration float64, easing EasingFn) Task {
	if easing == nil {
		easing = EaseInOut
	}
	return &tweenFloatTask{
		ptr:      ptr,
		start:    *ptr,
		target:   target,
		duration: duration,
		easing:   easing,
	}
}

func (t *tweenFloatTask) Tick(dt float64) Status {
	t.elapsed += dt
	progress := t.elapsed / t.duration
	if progress >= 1 {
		*t.ptr = t.target
		return Done
	}
	*t.ptr = t.start + (t.target-t.start)*t.easing(progress)
	return Running
}

// ----------------------------------------------------------------------------
// Float32 tween (for Alpha, scale, etc.)
// ----------------------------------------------------------------------------

type tweenFloat32Task struct {
	ptr      *float32
	start    float32
	target   float32
	duration float64
	elapsed  float64
	easing   EasingFn
}

// TweenFloat32 animates a float32 pointer.
func TweenFloat32(ptr *float32, target float32, duration float64, easing EasingFn) Task {
	if easing == nil {
		easing = EaseInOut
	}
	return &tweenFloat32Task{
		ptr: ptr, start: *ptr, target: target, duration: duration, easing: easing,
	}
}

func (t *tweenFloat32Task) Tick(dt float64) Status {
	t.elapsed += dt
	progress := t.elapsed / t.duration
	if progress >= 1 {
		*t.ptr = t.target
		return Done
	}
	f := float32(t.easing(progress))
	*t.ptr = t.start + (t.target-t.start)*f
	return Running
}
