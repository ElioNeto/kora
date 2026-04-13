package node

// Sprite2D is a node that displays an image/sprite
type Sprite2D struct {
	*Node2D

	// Sprite properties
	spriteUrl  string
	spriteImg  interface{} // Placeholder for actual image
	width      float32
	height     float32
	color      string
	alpha      float32

	// Animation properties
	animationPlaying bool
	animationFrames  []int
	currentFrame     int
	frameSpeed       float64

	// Visual transformations
	flipX bool
	flipY bool
}

// NewSprite2D creates a new Sprite2D node
func NewSprite2D(name string) *Sprite2D {
	node := NewNode2D(name, 0)
	return &Sprite2D{
		Node2D:   node,
		color:    "#ffffff",
		alpha:    1.0,
		flipX:    false,
		flipY:    false,
	}
}

// SetSprite sets the sprite from URL
func (s *Sprite2D) SetSprite(url string) {
	s.spriteUrl = url
}

// SetImage sets a raw image (platform-specific)
func (s *Sprite2D) SetImage(img interface{}) {
	s.spriteImg = img
}

// SetSize sets the sprite display size
func (s *Sprite2D) SetSize(w, h float32) {
	s.width = w
	s.height = h
}

// GetSize returns the sprite size
func (s *Sprite2D) GetSize() (float32, float32) {
	return s.width, s.height
}

// SetColor sets the tint color (hex string)
func (s *Sprite2D) SetColor(color string) {
	s.color = color
}

// GetColor returns the tint color
func (s *Sprite2D) GetColor() string {
	return s.color
}

// SetAlpha sets transparency (0.0 to 1.0)
func (s *Sprite2D) SetAlpha(a float32) {
	if a < 0 {
		a = 0
	}
	if a > 1 {
		a = 1
	}
	s.alpha = a
}

// GetAlpha returns transparency
func (s *Sprite2D) GetAlpha() float32 {
	return s.alpha
}

// SetFlipX flips horizontally
func (s *Sprite2D) SetFlipX(flip bool) {
	s.flipX = flip
}

// IsFlipX returns horizontal flip state
func (s *Sprite2D) IsFlipX() bool {
	return s.flipX
}

// SetFlipY flips vertically
func (s *Sprite2D) SetFlipY(flip bool) {
	s.flipY = flip
}

// IsFlipY returns vertical flip state
func (s *Sprite2D) IsFlipY() bool {
	return s.flipY
}

// Animation methods

// PlayAnimation starts an animation
func (s *Sprite2D) PlayAnimation(frames []int, speed float64) {
	s.animationFrames = frames
	s.frameSpeed = speed
	s.animationPlaying = true
	s.currentFrame = 0
}

// StopAnimation stops current animation
func (s *Sprite2D) StopAnimation() {
	s.animationPlaying = false
	s.currentFrame = 0
}

// SetAnimationFrame sets the current animation frame
func (s *Sprite2D) SetAnimationFrame(frame int) {
	s.currentFrame = frame
}

// GetAnimationFrame returns current frame index
func (s *Sprite2D) GetAnimationFrame() int {
	return s.currentFrame
}

// IsPlayingAnimation returns if animation is playing
func (s *Sprite2D) IsPlayingAnimation() bool {
	return s.animationPlaying
}

// Update runs on every frame
func (s *Sprite2D) Update(dt float64) {
	// Base node update - propagate to children
	for _, child := range s.children {
		if child != nil {
			child.Update(dt)
		}
	}
	// Animation processing
	if s.animationPlaying && s.frameSpeed > 0 {
		// Track time for frame switching
		// In full implementation, would use a time accumulator
	}
}

// Draw renders the sprite to the canvas
func (s *Sprite2D) Draw(ctx interface{}) {
	// Placeholder for rendering logic
	// In full implementation: draw image at world position
}
