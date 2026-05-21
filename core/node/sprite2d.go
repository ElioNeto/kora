// Package node implements the core node system for Kora Engine
package node

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/ElioNeto/kora/core/render"
)

// Sprite2D is a node that displays an image/sprite with optional animation support.
type Sprite2D struct {
	*Node2D

	// Sprite properties
	spriteUrl string
	ebitenImg *ebiten.Image // cached ebiten image reference
	width     float32
	height    float32
	color     color.RGBA
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

	// Fallback color (shown when no texture is loaded)
	fallbackColor color.RGBA
}

// NewSprite2D creates a new Sprite2D node
func NewSprite2D(name string) *Sprite2D {
	node := NewNode2D(name, 0)
	return &Sprite2D{
		Node2D:        node,
		alpha:         1.0,
		flipX:         false,
		flipY:         false,
		fallbackColor: color.RGBA{0x00, 0xe5, 0xa0, 0xff},
		width:         32,
		height:        32,
	}
}

// SetSprite sets the sprite from URL and loads the texture
func (s *Sprite2D) SetSprite(url string) {
	s.spriteUrl = url
	// Attempt to load from texture cache
	if img := render.GetTexture(url); img != nil {
		s.ebitenImg = img
		s.width = float32(img.Bounds().Dx())
		s.height = float32(img.Bounds().Dy())
	} else {
		s.ebitenImg = nil
	}
}

// GetSpriteURL returns the sprite URL/path
func (s *Sprite2D) GetSpriteURL() string {
	return s.spriteUrl
}

// SetImage sets a raw ebiten.Image directly
func (s *Sprite2D) SetImage(img *ebiten.Image) {
	s.ebitenImg = img
	if img != nil {
		bounds := img.Bounds()
		s.width = float32(bounds.Dx())
		s.height = float32(bounds.Dy())
	}
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

// SetColorString sets the tint color (hex string like "#ff8800")
func (s *Sprite2D) SetColorString(hex string) {
	if len(hex) == 7 && hex[0] == '#' {
		r := hexToByte(hex[1])<<4 | hexToByte(hex[2])
		g := hexToByte(hex[3])<<4 | hexToByte(hex[4])
		b := hexToByte(hex[5])<<4 | hexToByte(hex[6])
		s.color = color.RGBA{r, g, b, 255}
	}
}

func hexToByte(c byte) uint8 {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}

// SetFallbackColor sets the color shown when no texture is loaded
func (s *Sprite2D) SetFallbackColor(c color.RGBA) {
	s.fallbackColor = c
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

// StopAnimation stops the current animation.
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

// Draw renders the sprite to the screen
func (s *Sprite2D) Draw(screen *ebiten.Image) {
	if !s.visible || !s.alive || screen == nil {
		return
	}

	// Draw children first (behind this sprite)
	for _, child := range s.children {
		if child != nil {
			child.Draw(screen)
		}
	}

	// Get world position and rotation
	pos := s.GetWorldPosition()
	worldRot := s.GetWorldRotation()

	// Determine what to draw
	drawW := float64(s.width)
	drawH := float64(s.height)
	if drawW <= 0 || drawH <= 0 {
		return
	}

	op := &ebiten.DrawImageOptions{}

	// Center pivot
	op.GeoM.Translate(-drawW/2, -drawH/2)

	// Flip
	if s.flipX {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(drawW, 0)
	}
	if s.flipY {
		op.GeoM.Scale(1, -1)
		op.GeoM.Translate(0, drawH)
	}

	// Rotation (degrees -> radians)
	if worldRot != 0 {
		rad := worldRot * math.Pi / 180
		op.GeoM.Rotate(float64(rad))
	}

	// Apply Node2D scale
	scaleX := float64(s.GetScaleX())
	scaleY := float64(s.GetScaleY())
	if scaleX != 1 || scaleY != 1 {
		op.GeoM.Scale(scaleX, scaleY)
	}

	// World position
	op.GeoM.Translate(float64(pos.X), float64(pos.Y))

	// Alpha
	if s.alpha < 1 {
		op.ColorScale.ScaleAlpha(float32(s.alpha))
	}

	if s.ebitenImg != nil {
		// Draw the actual texture
		if s.frameW > 0 && s.totalCols > 0 {
			// Spritesheet: extract the current frame
			frame := s.GetFrame()
			col := frame % s.totalCols
			row := frame / s.totalCols
			sx := col * s.frameW
			sy := row * s.frameH
			subRect := image.Rect(sx, sy, sx+s.frameW, sy+s.frameH)
			subImg := s.ebitenImg.SubImage(subRect).(*ebiten.Image)
			screen.DrawImage(subImg, op)
		} else {
			screen.DrawImage(s.ebitenImg, op)
		}
	} else if s.spriteUrl == "" {
		// No texture and no URL set: draw fallback colored rectangle
		drawFallbackRect(screen, drawW, drawH, s.fallbackColor, op.GeoM)
	}
}

// drawFallbackRect draws a colored rectangle when no sprite texture is available.
func drawFallbackRect(screen *ebiten.Image, w, h float64, c color.RGBA, geo ebiten.GeoM) {
	pixel := ebiten.NewImage(1, 1)
	pixel.Fill(c)
	op := &ebiten.DrawImageOptions{GeoM: geo}
	op.GeoM.Scale(w, h)
	screen.DrawImage(pixel, op)
}

// Compile-time interface check
var _ Node = (*Sprite2D)(nil)
