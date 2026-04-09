package scene

// TransitionFn is a callback invoked when a scene transition completes.
type TransitionFn func()

// FadeState tracks a fade-in / fade-out animation driven by the game loop.
type FadeState struct {
	Alpha    float32 // 0 = transparent, 1 = opaque
	Duration float64 // seconds for full transition
	elapsed  float64
	dir      int   // +1 fade-in, -1 fade-out
	onDone   TransitionFn
	active   bool
}

// FadeIn starts a fade from opaque to transparent over duration seconds.
func (f *FadeState) FadeIn(duration float64, onDone TransitionFn) {
	f.Alpha = 1
	f.Duration = duration
	f.elapsed = 0
	f.dir = -1
	f.onDone = onDone
	f.active = true
}

// FadeOut starts a fade from transparent to opaque over duration seconds.
func (f *FadeState) FadeOut(duration float64, onDone TransitionFn) {
	f.Alpha = 0
	f.Duration = duration
	f.elapsed = 0
	f.dir = +1
	f.onDone = onDone
	f.active = true
}

// Tick advances the fade animation. Returns true when complete.
func (f *FadeState) Tick(dt float64) bool {
	if !f.active {
		return false
	}
	f.elapsed += dt
	progress := float32(f.elapsed / f.Duration)
	if progress >= 1 {
		progress = 1
		f.active = false
	}
	if f.dir > 0 {
		f.Alpha = progress
	} else {
		f.Alpha = 1 - progress
	}
	if !f.active && f.onDone != nil {
		f.onDone()
	}
	return !f.active
}

// Active reports whether a transition is in progress.
func (f *FadeState) Active() bool { return f.active }
