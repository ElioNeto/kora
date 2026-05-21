package math

import (
	"testing"
)

// allEasingFns returns all easing functions for tests
func allEasingFns() map[string]EasingFn {
	return map[string]EasingFn{
		EaseLinear:       Linear,
		EaseInQuad:       InQuad,
		EaseOutQuad:      OutQuad,
		EaseInOutQuad:    InOutQuad,
		EaseInCubic:      InCubic,
		EaseOutCubic:     OutCubic,
		EaseInOutCubic:   InOutCubic,
		EaseInQuart:      InQuart,
		EaseOutQuart:     OutQuart,
		EaseInOutQuart:   InOutQuart,
		EaseInQuint:      InQuint,
		EaseOutQuint:     OutQuint,
		EaseInOutQuint:   InOutQuint,
		EaseInSine:       InSine,
		EaseOutSine:      OutSine,
		EaseInOutSine:    InOutSine,
		EaseInCirc:       InCirc,
		EaseOutCirc:      OutCirc,
		EaseInOutCirc:    InOutCirc,
		EaseInElastic:    InElastic,
		EaseOutElastic:   OutElastic,
		EaseInOutElastic: InOutElastic,
		EaseInBack:       InBack,
		EaseOutBack:      OutBack,
		EaseInOutBack:    InOutBack,
		EaseInBounce:     InBounce,
		EaseOutBounce:    OutBounce,
		EaseInOutBounce:  InOutBounce,
	}
}

func TestEasingStartsAtZero(t *testing.T) {
	const eps = 1e-10
	for name, fn := range allEasingFns() {
		if got := fn(0); got < -eps || got > eps {
			t.Errorf("%s(0) = %v, want ~0", name, got)
		}
	}
}

func TestEasingEndsAtOne(t *testing.T) {
	const eps = 1e-10
	for name, fn := range allEasingFns() {
		if got := fn(1); got < 1-eps || got > 1+eps {
			t.Errorf("%s(1) = %v, want ~1", name, got)
		}
	}
}

func TestEasingStaysWithinBounds(t *testing.T) {
	points := []float64{0, 0.1, 0.25, 0.5, 0.75, 0.9, 1}
	// Back and Elastic easings naturally overshoot beyond [0,1] — that's
	// their intended behavior (e.g., pulling back before launching, or
	// oscillating like a spring). Widening the tolerance for all easings
	// would miss real bugs, so we apply per-function bounds.
	overshootEasings := map[string]bool{
		EaseInBack: true, EaseOutBack: true, EaseInOutBack: true,
		EaseInElastic: true, EaseOutElastic: true, EaseInOutElastic: true,
	}
	for name, fn := range allEasingFns() {
		lo, hi := -0.01, 1.01
		if overshootEasings[name] {
			lo, hi = -0.3, 1.3
		}
		for _, p := range points {
			v := fn(p)
			if v < lo || v > hi {
				t.Errorf("%s(%v) = %v, out of range [%v, %v]", name, p, v, lo, hi)
			}
		}
	}
}

func TestEasingByName(t *testing.T) {
	for name, wantFn := range allEasingFns() {
		gotFn := EasingByName(name)
		if gotFn == nil {
			t.Errorf("EasingByName(%q) returned nil", name)
			continue
		}
		// Check same function identity
		if gotFn(0.5) != wantFn(0.5) {
			t.Errorf("EasingByName(%q)(0.5) = %v, want %v", name, gotFn(0.5), wantFn(0.5))
		}
	}
}

func TestEasingByNameUnknown(t *testing.T) {
	fn := EasingByName("nonexistent")
	if fn == nil {
		t.Fatal("EasingByName should not return nil for unknown name")
	}
	if fn(0) != 0 || fn(1) != 1 {
		t.Error("default easing (Linear) should satisfy fn(0)=0, fn(1)=1")
	}
}

func TestEasingByLegacyName(t *testing.T) {
	legacy := map[string]string{
		"linear":       EaseLinear,
		"ease_in":      EaseInQuad,
		"ease_out":     EaseOutQuad,
		"ease_in_out":  EaseInOutQuad,
		"elastic_out":  EaseOutElastic,
		"bounce_out":   EaseOutBounce,
	}
	for legacyName, expectedName := range legacy {
		got := EasingByName(legacyName)
		want := EasingByName(expectedName)
		if got(0.5) != want(0.5) {
			t.Errorf("EasingByName(%q)(0.5) = %v, want %v (same as %q)",
				legacyName, got(0.5), want(0.5), expectedName)
		}
	}
}

func TestEasingMonotonic(t *testing.T) {
	// Test that each easing function is non-decreasing over [0,1].
	// Elastic, Back, and Bounce easings may overshoot or oscillate
	// (by design), so they are excluded from monotonic testing.
	// We only test monotonic
	// for easings that are supposed to be monotonic.
	monotonic := map[string]EasingFn{
		EaseLinear:      Linear,
		EaseInQuad:      InQuad,
		EaseOutQuad:     OutQuad,
		EaseInOutQuad:   InOutQuad,
		EaseInCubic:     InCubic,
		EaseOutCubic:    OutCubic,
		EaseInOutCubic:  InOutCubic,
		EaseInQuart:     InQuart,
		EaseOutQuart:    OutQuart,
		EaseInOutQuart:  InOutQuart,
		EaseInQuint:     InQuint,
		EaseOutQuint:    OutQuint,
		EaseInOutQuint:  InOutQuint,
		EaseInSine:      InSine,
		EaseOutSine:     OutSine,
		EaseInOutSine:   InOutSine,
		EaseInCirc:      InCirc,
		EaseOutCirc:     OutCirc,
		EaseInOutCirc:   InOutCirc,
	}

	const steps = 1000
	const dt = 1.0 / steps

	for name, fn := range monotonic {
		var prev float64
		for i := 0; i <= steps; i++ {
			tval := float64(i) * dt
			v := fn(tval)
			if v < prev-1e-12 {
				t.Errorf("%s is not non-decreasing: t=%v, prev=%v, curr=%v", name, tval, prev, v)
				break
			}
			prev = v
		}
	}
}

func TestNames(t *testing.T) {
	names := Names()
	if len(names) != 28 {
		t.Errorf("expected 28 easing names, got %d", len(names))
	}

	seen := make(map[string]bool)
	for _, n := range names {
		if seen[n] {
			t.Errorf("duplicate name %q", n)
		}
		seen[n] = true
		if EasingByName(n) == nil {
			t.Errorf("Names() includes %q but EasingByName returns nil", n)
		}
	}
}

func TestEasingSpecificValues(t *testing.T) {
	tests := []struct {
		name string
		fn   EasingFn
		at25 float64
		at50 float64
		at75 float64
	}{
		{EaseLinear, Linear, 0.25, 0.5, 0.75},
		{EaseInQuad, InQuad, 0.0625, 0.25, 0.5625},
		{EaseOutQuad, OutQuad, 0.4375, 0.75, 0.9375},
		{EaseInOutQuad, InOutQuad, 0.125, 0.5, 0.875},
		{EaseInCubic, InCubic, 0.015625, 0.125, 0.421875},
		{EaseOutCubic, OutCubic, 0.578125, 0.875, 0.984375},
		{EaseInOutCubic, InOutCubic, 0.0625, 0.5, 0.9375},
		{EaseInQuart, InQuart, 0.00390625, 0.0625, 0.31640625},
		{EaseOutQuart, OutQuart, 0.68359375, 0.9375, 0.99609375},
		{EaseInOutQuart, InOutQuart, 0.03125, 0.5, 0.96875},
		{EaseInQuint, InQuint, 0.0009765625, 0.03125, 0.2373046875},
		{EaseOutQuint, OutQuint, 0.7626953125, 0.96875, 0.9990234375},
		{EaseInOutQuint, InOutQuint, 0.015625, 0.5, 0.984375},
		{EaseInSine, InSine, 0.07612046748871326, 0.2928932188134524, 0.6173165676349102},
		{EaseOutSine, OutSine, 0.3826834323650898, 0.7071067811865476, 0.9238795325112867},
		{EaseInOutSine, InOutSine, 0.1464466094067262, 0.5, 0.8535533905932738},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(0.25); !approxEqual(got, tt.at25) {
				t.Errorf("%s(0.25) = %v, want %v", tt.name, got, tt.at25)
			}
			if got := tt.fn(0.5); !approxEqual(got, tt.at50) {
				t.Errorf("%s(0.5) = %v, want %v", tt.name, got, tt.at50)
			}
			if got := tt.fn(0.75); !approxEqual(got, tt.at75) {
				t.Errorf("%s(0.75) = %v, want %v", tt.name, got, tt.at75)
			}
		})
	}
}

func approxEqual(a, b float64) bool {
	const epsilon = 1e-12
	if a == b {
		return true
	}
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= epsilon
}
