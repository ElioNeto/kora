package math

import (
	"math"
)

// EasingFn is a function that maps time t [0,1] to an eased value [0,1].
type EasingFn func(t float64) float64

// Easing preset names — all common easing types following Robert Penner's formulas.
const (
	EaseLinear       = "linear"
	EaseInQuad       = "in_quad"
	EaseOutQuad      = "out_quad"
	EaseInOutQuad    = "in_out_quad"
	EaseInCubic      = "in_cubic"
	EaseOutCubic     = "out_cubic"
	EaseInOutCubic   = "in_out_cubic"
	EaseInQuart      = "in_quart"
	EaseOutQuart     = "out_quart"
	EaseInOutQuart   = "in_out_quart"
	EaseInQuint      = "in_quint"
	EaseOutQuint     = "out_quint"
	EaseInOutQuint   = "in_out_quint"
	EaseInSine       = "in_sine"
	EaseOutSine      = "out_sine"
	EaseInOutSine    = "in_out_sine"
	EaseInCirc       = "in_circ"
	EaseOutCirc      = "out_circ"
	EaseInOutCirc    = "in_out_circ"
	EaseInElastic    = "in_elastic"
	EaseOutElastic   = "out_elastic"
	EaseInOutElastic = "in_out_elastic"
	EaseInBack       = "in_back"
	EaseOutBack      = "out_back"
	EaseInOutBack    = "in_out_back"
	EaseInBounce     = "in_bounce"
	EaseOutBounce    = "out_bounce"
	EaseInOutBounce  = "in_out_bounce"
)

// ---------------------------------------------------------------------------
// Linear
// ---------------------------------------------------------------------------

// Linear maps t linearly — no easing.
func Linear(t float64) float64 { return t }

// ---------------------------------------------------------------------------
// Quad — power of 2
// ---------------------------------------------------------------------------

// InQuad accelerates from zero.
func InQuad(t float64) float64 { return t * t }

// OutQuad decelerates to zero.
func OutQuad(t float64) float64 { return t * (2 - t) }

// InOutQuad accelerates then decelerates.
func InOutQuad(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return -1 + (4-2*t)*t
}

// ---------------------------------------------------------------------------
// Cubic — power of 3
// ---------------------------------------------------------------------------

// InCubic accelerates from zero.
func InCubic(t float64) float64 { return t * t * t }

// OutCubic decelerates to zero.
func OutCubic(t float64) float64 { return 1 - math.Pow(1-t, 3) }

// InOutCubic accelerates then decelerates.
func InOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 3)/2
}

// ---------------------------------------------------------------------------
// Quart — power of 4
// ---------------------------------------------------------------------------

// InQuart accelerates from zero.
func InQuart(t float64) float64 { return t * t * t * t }

// OutQuart decelerates to zero.
func OutQuart(t float64) float64 { return 1 - math.Pow(1-t, 4) }

// InOutQuart accelerates then decelerates.
func InOutQuart(t float64) float64 {
	if t < 0.5 {
		return 8 * t * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 4)/2
}

// ---------------------------------------------------------------------------
// Quint — power of 5
// ---------------------------------------------------------------------------

// InQuint accelerates from zero.
func InQuint(t float64) float64 { return t * t * t * t * t }

// OutQuint decelerates to zero.
func OutQuint(t float64) float64 { return 1 - math.Pow(1-t, 5) }

// InOutQuint accelerates then decelerates.
func InOutQuint(t float64) float64 {
	if t < 0.5 {
		return 16 * t * t * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 5)/2
}

// ---------------------------------------------------------------------------
// Sine
// ---------------------------------------------------------------------------

// InSine accelerates from zero.
func InSine(t float64) float64 { return 1 - math.Cos(t*math.Pi/2) }

// OutSine decelerates to zero.
func OutSine(t float64) float64 { return math.Sin(t * math.Pi / 2) }

// InOutSine accelerates then decelerates.
func InOutSine(t float64) float64 { return (1 - math.Cos(t*math.Pi)) / 2 }

// ---------------------------------------------------------------------------
// Circular
// ---------------------------------------------------------------------------

// InCirc accelerates from zero.
func InCirc(t float64) float64 { return 1 - math.Sqrt(1-t*t) }

// OutCirc decelerates to zero.
func OutCirc(t float64) float64 { return math.Sqrt(1 - (t-1)*(t-1)) }

// InOutCirc accelerates then decelerates.
func InOutCirc(t float64) float64 {
	if t < 0.5 {
		return (1 - math.Sqrt(1-4*t*t)) / 2
	}
	return (math.Sqrt(1-(2*t-2)*(2*t-2)) + 1) / 2
}

// ---------------------------------------------------------------------------
// Elastic — overshoot oscillation
// ---------------------------------------------------------------------------

// InElastic starts with an elastic bounce.
func InElastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	return -math.Pow(2, 10*t-10) * math.Sin((t*10-10.75)*(2*math.Pi/3))
}

// OutElastic ends with an elastic bounce.
func OutElastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	return math.Pow(2, -10*t)*math.Sin((t*10-0.75)*(2*math.Pi/3)) + 1
}

// InOutElastic starts and ends with an elastic bounce.
func InOutElastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	t *= 2
	if t < 1 {
		return -0.5 * math.Pow(2, 10*t-10) * math.Sin((t*10-10.75)*(2*math.Pi/3))
	}
	t--
	return 0.5*math.Pow(2, -10*t)*math.Sin((t*10-0.75)*(2*math.Pi/3)) + 1
}

// ---------------------------------------------------------------------------
// Back — overshoots by a configurable amount (standard 1.70158)
// ---------------------------------------------------------------------------

const backAmount = 1.70158

// InBack starts by moving backwards.
func InBack(t float64) float64 {
	return t * t * ((backAmount+1)*t - backAmount)
}

// OutBack overshoots the end.
func OutBack(t float64) float64 {
	t -= 1
	return t*t*((backAmount+1)*t+backAmount) + 1
}

// InOutBack combines InBack and OutBack.
func InOutBack(t float64) float64 {
	if t < 0.5 {
		t *= 2
		return 0.5 * t * t * ((backAmount+1)*t - backAmount)
	}
	t = 2*t - 2
	return 0.5 * (t*t*((backAmount+1)*t+backAmount) + 2)
}

// ---------------------------------------------------------------------------
// Bounce — simulates a bouncing ball
// ---------------------------------------------------------------------------

// OutBounce bounces at the end.
func OutBounce(t float64) float64 {
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

// InBounce bounces at the start.
func InBounce(t float64) float64 {
	return 1 - OutBounce(1-t)
}

// InOutBounce bounces at both ends.
func InOutBounce(t float64) float64 {
	if t < 0.5 {
		return (1 - OutBounce(1-2*t)) / 2
	}
	return (1 + OutBounce(2*t-1)) / 2
}

// ---------------------------------------------------------------------------
// Lookup helpers
// ---------------------------------------------------------------------------

// EasingByName returns the easing function for the given name.
// Supports all constant names (EaseLinear, EaseInQuad, …) as well as
// legacy names for backward compatibility:
//
//	"ease_in"      → InQuad
//	"ease_out"     → OutQuad
//	"ease_in_out"  → InOutQuad
//	"elastic_out"  → OutElastic
//	"bounce_out"   → OutBounce
//
// Returns Linear if the name is not recognised.
func EasingByName(name string) EasingFn {
	switch name {
	case EaseLinear:
		return Linear
	case EaseInQuad, "ease_in":
		return InQuad
	case EaseOutQuad, "ease_out":
		return OutQuad
	case EaseInOutQuad, "ease_in_out":
		return InOutQuad
	case EaseInCubic:
		return InCubic
	case EaseOutCubic:
		return OutCubic
	case EaseInOutCubic:
		return InOutCubic
	case EaseInQuart:
		return InQuart
	case EaseOutQuart:
		return OutQuart
	case EaseInOutQuart:
		return InOutQuart
	case EaseInQuint:
		return InQuint
	case EaseOutQuint:
		return OutQuint
	case EaseInOutQuint:
		return InOutQuint
	case EaseInSine:
		return InSine
	case EaseOutSine:
		return OutSine
	case EaseInOutSine:
		return InOutSine
	case EaseInCirc:
		return InCirc
	case EaseOutCirc:
		return OutCirc
	case EaseInOutCirc:
		return InOutCirc
	case EaseInElastic:
		return InElastic
	case EaseOutElastic, "elastic_out":
		return OutElastic
	case EaseInOutElastic:
		return InOutElastic
	case EaseInBack:
		return InBack
	case EaseOutBack:
		return OutBack
	case EaseInOutBack:
		return InOutBack
	case EaseInBounce:
		return InBounce
	case EaseOutBounce, "bounce_out":
		return OutBounce
	case EaseInOutBounce:
		return InOutBounce
	default:
		return Linear
	}
}

// Names returns all available easing preset names.
func Names() []string {
	return []string{
		EaseLinear,
		EaseInQuad, EaseOutQuad, EaseInOutQuad,
		EaseInCubic, EaseOutCubic, EaseInOutCubic,
		EaseInQuart, EaseOutQuart, EaseInOutQuart,
		EaseInQuint, EaseOutQuint, EaseInOutQuint,
		EaseInSine, EaseOutSine, EaseInOutSine,
		EaseInCirc, EaseOutCirc, EaseInOutCirc,
		EaseInElastic, EaseOutElastic, EaseInOutElastic,
		EaseInBack, EaseOutBack, EaseInOutBack,
		EaseInBounce, EaseOutBounce, EaseInOutBounce,
	}
}
