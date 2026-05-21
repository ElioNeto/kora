package node

import (
	"github.com/ElioNeto/kora/core/render"
)

// Sprite2D is a node that displays an image/sprite with optional animation support.
type Sprite2D struct {
	*Node2D

	// Sprite properties
	spriteUrl string
	spriteImg interface{} // Placeholder for actual image
	width     float32
	height    float32
	color     string
	alpha     float32

	// Spritesheet grid (0 = single image, no grid)
	frameW    int
	frameH    int
	totalCols int
	totalRows int

	// Current frame index (used when no Animator is set)
	currentFrame int

	// Animation via render.Animator
	animator *render.Animator

	// Manual frame advance (time accumulator for simple animations)
	frameTimer float64
	frameIndex int
	frameSpeed float64 // seconds per frame

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

// SetFrameSize sets the sprite sheet frame grid dimensions.
// When frameW and frameH are > 0, the sprite is treated as a grid where
// SetFrame / GetFrame select the active cell.
func (s *Sprite2D) SetFrameSize(frameW, frameH int) {
	s.frameW = frameW
	s.frameH = frameH
}

// SetGridSize sets the total number of columns and rows in the sprite sheet.
func (s *Sprite2D) SetGridSize(cols, rows int) {
	s.totalCols = cols
	s.totalRows = rows
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

// SetFrame sets the current frame index directly (manual mode).
func (s *Sprite2D) SetFrame(frame int) {
	s.currentFrame = frame
}

// GetFrame returns the current frame index.
// If an animator is attached, it delegates to the animator's current frame.
func (s *Sprite2D) GetFrame() int {
	if s.animator != nil && s.animator.Current() != nil {
		return s.animator.CurrentFrame()
	}
	return s.currentFrame
}

// ---------------------------------------------------------------------------
// Animation API
// ---------------------------------------------------------------------------

// SetAnimator attaches a render.Animator to this Sprite2D.
// When set, the animator drives frame selection automatically each Update.
func (s *Sprite2D) SetAnimator(a *render.Animator) {
	s.animator = a
}

// Animator returns the attached animator, or nil.
func (s *Sprite2D) Animator() *render.Animator {
	return s.animator
}

// AddAnimation registers a named animation clip on the attached animator.
func (s *Sprite2D) AddAnimation(name string, frames []int, fps float64, loop bool) {
	if s.animator == nil {
		s.animator = render.NewAnimator(nil)
	}
	s.animator.Add(&render.Animation{
		Name:   name,
		Frames: frames,
		FPS:    fps,
		Loop:   loop,
	})
}

// Play starts a named animation. Returns false if the clip does not exist.
func (s *Sprite2D) Play(name string) bool {
	if s.animator == nil {
		return false
	}
	s.animator.Play(name)
	return s.animator.Current() != nil
}

// StopAnimation stops the current animation and resets to frame 0.
func (s *Sprite2D) StopAnimation() {
	s.animator = nil
	s.frameTimer = 0
	s.frameIndex = 0
	s.currentFrame = 0
	s.frameSpeed = 0
}

// IsPlayingAnimation returns true if an animation is active.
func (s *Sprite2D) IsPlayingAnimation() bool {
	return s.animator != nil && s.animator.Current() != nil && !s.animator.Done
}

// PlayAnimation starts a simple frame-based animation (legacy API).
// frames specifies the sequence of frame indices, speed is the duration per frame in seconds.
func (s *Sprite2D) PlayAnimation(frames []int, speed float64) {
	if len(frames) == 0 {
		return
	}
	s.StopAnimation()
	s.frameIndex = 0
	s.frameTimer = 0
	s.frameSpeed = speed
	s.currentFrame = frames[0]
}

// SetAnimationFrame sets the current animation frame (legacy API).
func (s *Sprite2D) SetAnimationFrame(frame int) {
	s.currentFrame = frame
}

// GetAnimationFrame returns current frame index (legacy API).
func (s *Sprite2D) GetAnimationFrame() int {
	return s.GetFrame()
}

// FrameCount returns the total number of frames in the sprite sheet grid.
// Returns 0 if no grid is configured.
func (s *Sprite2D) FrameCount() int {
	return s.totalCols * s.totalRows
}

// ---------------------------------------------------------------------------
// Update / Draw
// ---------------------------------------------------------------------------

// Update runs on every frame
func (s *Sprite2D) Update(dt float64) {
	// Propagate to children
	for _, child := range s.children {
		if child != nil {
			child.Update(dt)
		}
	}

	// Priority 1: Animator-driven animation
	if s.animator != nil {
		s.animator.Update(dt)
		return
	}

	// Priority 2: Simple frame-based animation (legacy)
	if s.frameSpeed > 0 {
		s.frameTimer += dt
		for s.frameTimer >= s.frameSpeed {
			s.frameTimer -= s.frameSpeed
			s.frameIndex++
		}
	}
}

// Draw renders the sprite to the canvas
func (s *Sprite2D) Draw(ctx interface{}) {
	// Placeholder for rendering logic
	// In full implementation: draw image at world position
}
