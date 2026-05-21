package render

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// ---------------------------------------------------------------------------
// Bitmap font types
// ---------------------------------------------------------------------------

// Glyph represents a single character in a bitmap font atlas.
type Glyph struct {
	X, Y, W, H int     // position and size in the atlas
	Advance    int      // horizontal advance in pixels
	img        *ebiten.Image // cached sub-image (set at font creation)
}

// BitmapFont stores a font atlas and glyph data for rendering monospace text.
type BitmapFont struct {
	atlas      *ebiten.Image
	glyphs     map[rune]Glyph
	lineHeight int
	baseSize   int // reference font size in pixels
}

// NewBitmapFont creates a BitmapFont from an atlas image.
// The atlas is treated as a grid: cols columns × rows rows of charW × charH cells.
// firstChar is the rune of the top-left cell; columns advance the rune first.
func NewBitmapFont(atlas *ebiten.Image, firstChar rune, cols, rows int, charW, charH int) *BitmapFont {
	f := &BitmapFont{
		atlas:      atlas,
		glyphs:     make(map[rune]Glyph, cols*rows),
		lineHeight: charH,
		baseSize:   charH,
	}
	for i := 0; i < cols*rows; i++ {
		r := firstChar + rune(i)
		col := i % cols
		row := i / cols
		rect := image.Rect(col*charW, row*charH, (col+1)*charW, (row+1)*charH)
		f.glyphs[r] = Glyph{
			X:       rect.Min.X,
			Y:       rect.Min.Y,
			W:       charW,
			H:       charH,
			Advance: charW,
			img:     atlas.SubImage(rect).(*ebiten.Image),
		}
	}
	return f
}

// LineHeight returns the font line height in pixels.
func (f *BitmapFont) LineHeight() int {
	return f.lineHeight
}

// MeasureText returns the pixel width and height of text rendered at scale 1.0.
func (f *BitmapFont) MeasureText(text string) (float64, float64) {
	width := 0
	fallback := f.glyphs['?']
	for _, r := range text {
		g, ok := f.glyphs[r]
		if !ok {
			g = fallback
		}
		width += g.Advance
	}
	return float64(width), float64(f.lineHeight)
}

// DrawText renders text at the given screen position.
// scale controls size (1.0 = original pixel size).
// If clr is nil, opaque white is used.
func (f *BitmapFont) DrawText(screen *ebiten.Image, text string, x, y float64, scale float64, clr *ebiten.ColorScale) {
	if screen == nil || f.atlas == nil {
		return
	}
	cx := x
	fallback := f.glyphs['?']
	for _, r := range text {
		g, ok := f.glyphs[r]
		if !ok {
			g = fallback
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(cx, y)
		if clr != nil {
			op.ColorScale = *clr
		}
		screen.DrawImage(g.img, op)
		cx += float64(g.Advance) * scale
	}
}
