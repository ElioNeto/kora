package async

// ----------------------------------------------------------------------------
// Tween — interpolate values over time
// ----------------------------------------------------------------------------

// EaseFunc computes an easing value from 0 to 1.
type EaseFunc func(t float64) float64

// EaseNone is a linear interpolation (no easing).
func EaseNone(t float64) float64 {
	return t
}

// EaseInQuad eases in with a quadratic curve.
func EaseInQuad(t float64) float64 {
	return t * t
}

// EaseOutQuad eases out with a quadratic curve.
func EaseOutQuad(t float64) float64 {
	return t * (2 - t)
}

// EaseInOutQuad eases in and out with a quadratic curve.
func EaseInOutQuad(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return 1 - (-2*t + 2)*(-2*t + 2) / 2
}

// Tween interpolates a value from `from` to `to` over `duration` using `ease`.
type Tween struct {
	from     float64
	to       float64
	duration float64
	elapsed  float64
	ease     EaseFunc
}

// Tween creates a tween that will interpolate from `from` to `to` over `duration` seconds.
func Tween(from, to, duration float64, ease EaseFunc) *Tween {
	return &Tween{
		from:     from,
		to:       to,
		duration: duration,
		elapsed:  0,
		ease:     ease,
	}
}

// Tick advances the tween by `dt` seconds and returns its current value.
func (t *Tween) Tick(dt float64) float64 {
	t.elapsed += dt
	if t.elapsed >= t.duration {
		return t.to
	}
	return t.from + (t.to-t.from)*t.ease(t.elapsed/t.duration)
}
