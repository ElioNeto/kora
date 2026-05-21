package render

// Current returns the currently playing animation clip, or nil.
func (a *Animator) Current() *Animation {
	return a.current
}

// CurrentFrame returns the current frame index within the active clip.
func (a *Animator) CurrentFrame() int {
	return a.frame
}
