package render

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// ----------------------------------------------------------------------------
// Sprite
// ----------------------------------------------------------------------------

// Sprite is a rectangular region of an *ebiten.Image atlas.
type Sprite struct {
	image  *ebiten.Image
	Bounds image.Rectangle
}

// NewSprite wraps an existing *ebiten.Image as a full-image Sprite.
func NewSprite(img *ebiten.Image) *Sprite {
	return &Sprite{image: img, Bounds: img.Bounds()}
}

// Sub returns a Sprite pointing to a sub-region of the atlas.
// x, y are the top-left corner; w, h are the size in pixels.
func (s *Sprite) Sub(x, y, w, h int) *Sprite {
	return &Sprite{
		image:  s.image,
		Bounds: image.Rect(x, y, x+w, y+h),
	}
}

// Width returns the sprite's pixel width.
func (s *Sprite) Width() int { return s.Bounds.Dx() }

// Height returns the sprite's pixel height.
func (s *Sprite) Height() int { return s.Bounds.Dy() }

// ----------------------------------------------------------------------------
// SpriteSheet — uniform grid atlas
// ----------------------------------------------------------------------------

// SpriteSheet slices an atlas into a uniform grid of frames.
type SpriteSheet struct {
	atlas      *ebiten.Image
	FrameW     int
	FrameH     int
	Cols       int
	frames     []*Sprite
}

// NewSpriteSheet creates a SpriteSheet by slicing atlas into frameW×frameH cells.
func NewSpriteSheet(atlas *ebiten.Image, frameW, frameH int) *SpriteSheet {
	w, h := atlas.Bounds().Dx(), atlas.Bounds().Dy()
	cols := w / frameW
	rows := h / frameH
	ss := &SpriteSheet{atlas: atlas, FrameW: frameW, FrameH: frameH, Cols: cols}
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			ss.frames = append(ss.frames, &Sprite{
				image:  atlas,
				Bounds: image.Rect(col*frameW, row*frameH, (col+1)*frameW, (row+1)*frameH),
			})
		}
	}
	return ss
}

// Frame returns the sprite at the given index.
func (ss *SpriteSheet) Frame(idx int) *Sprite {
	if idx < 0 || idx >= len(ss.frames) {
		return nil
	}
	return ss.frames[idx]
}

// Len returns the total number of frames.
func (ss *SpriteSheet) Len() int { return len(ss.frames) }

// ----------------------------------------------------------------------------
// Animator — drives frame animation
// ----------------------------------------------------------------------------

// Animation describes a named clip: a slice of frame indices and playback speed.
type Animation struct {
	Name   string
	Frames []int
	FPS    float64
	Loop   bool
}

// Animator plays sprite animations from a SpriteSheet.
type Animator struct {
	sheet    *SpriteSheet
	clips    map[string]*Animation
	current  *Animation
	frame    int     // index inside current.Frames
	elapsed  float64 // seconds since last frame change
	Done     bool    // true when a non-looping animation finishes
}

// NewAnimator creates an Animator for the given sheet.
func NewAnimator(sheet *SpriteSheet) *Animator {
	return &Animator{
		sheet: sheet,
		clips: make(map[string]*Animation),
	}
}

// Add registers a named animation clip.
func (a *Animator) Add(anim *Animation) {
	a.clips[anim.Name] = anim
}

// Play switches to the named clip. If already playing, does nothing.
func (a *Animator) Play(name string) {
	clip, ok := a.clips[name]
	if !ok {
		return
	}
	if a.current == clip {
		return
	}
	a.current = clip
	a.frame = 0
	a.elapsed = 0
	a.Done = false
}

// Update advances the animation by dt seconds.
func (a *Animator) Update(dt float64) {
	if a.current == nil || a.Done {
		return
	}
	a.elapsed += dt
	secondsPerFrame := 1.0 / a.current.FPS
	for a.elapsed >= secondsPerFrame {
		a.elapsed -= secondsPerFrame
		a.frame++
		if a.frame >= len(a.current.Frames) {
			if a.current.Loop {
				a.frame = 0
			} else {
				a.frame = len(a.current.Frames) - 1
				a.Done = true
				return
			}
		}
	}
}

// CurrentSprite returns the sprite for the current animation frame.
func (a *Animator) CurrentSprite() *Sprite {
	if a.current == nil || len(a.current.Frames) == 0 {
		return nil
	}
	return a.sheet.Frame(a.current.Frames[a.frame])
}
