package async

import "math"

// Easing functions — all map t∈[0,1] → value∈[0,1].

type EasingFn func(t float64) float64

var (
	Linear    EasingFn = func(t float64) float64 { return t }
	EaseIn    EasingFn = func(t float64) float64 { return t * t }
	EaseOut   EasingFn = func(t float64) float64 { return t * (2 - t) }
	EaseInOut EasingFn = func(t float64) float64 {
		if t < 0.5 {
			return 2 * t * t
		}
		return -1 + (4-2*t)*t
	}
	ElasticOut EasingFn = func(t float64) float64 {
		if t == 0 || t == 1 {
			return t
		}
		return math.Pow(2, -10*t)*math.Sin((t*10-0.75)*(2*math.Pi/3)) + 1
	}
	BounceOut EasingFn = func(t float64) float64 {
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
