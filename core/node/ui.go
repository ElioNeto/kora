// Package node implements the core node system for Kora Engine
package node

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ElioNeto/kora/core/render"
)

// TextAlign defines how text is horizontally aligned relative to the node's position.
type TextAlign int

const (
	TextAlignLeft TextAlign = iota
	TextAlignCenter
	TextAlignRight
)

// ---------------------------------------------------------------------------
// Label
// ---------------------------------------------------------------------------

// Label is a node that renders text using the engine's built-in bitmap font.
type Label struct {
	*Node2D
	Text     string
	FontSize float64                     // scale multiplier (1.0 = 8px)
	Color    struct{ R, G, B, A float32 }
	Align    TextAlign
}

// NewLabel creates a new Label node with the given name and text.
func NewLabel(name, text string) *Label {
	return &Label{
		Node2D:   NewNode2D(name, 0),
		Text:     text,
		FontSize: 1.0,
		Color:    struct{ R, G, B, A float32 }{R: 1, G: 1, B: 1, A: 1},
		Align:    TextAlignLeft,
	}
}

// SetText sets the label text.
func (n *Label) SetText(text string) {
	n.Text = text
}

// SetFontSize sets the font scale multiplier (1.0 = 8px).
func (n *Label) SetFontSize(size float64) {
	n.FontSize = size
}

// SetColor sets the text color with RGBA components in the [0, 1] range.
func (n *Label) SetColor(r, g, b, a float32) {
	n.Color.R = r
	n.Color.G = g
	n.Color.B = b
	n.Color.A = a
}

// SetAlign sets the text alignment.
func (n *Label) SetAlign(align TextAlign) {
	n.Align = align
}

// Update propagates the update to child nodes.
func (n *Label) Update(dt float64) {
	for _, child := range n.children {
		if child != nil {
			child.Update(dt)
		}
	}
}

// Draw renders the label text at the node's world position (screen space).
func (n *Label) Draw(screen *ebiten.Image) {
	if !n.visible || !n.alive || screen == nil {
		return
	}

	// Draw children first
	for _, child := range n.children {
		if child != nil {
			child.Draw(screen)
		}
	}

	if n.Text == "" {
		return
	}

	// Ensure the default font is initialised
	render.DebugTextAt(screen, "", 0, 0)
	if render.DefaultFont == nil {
		return
	}

	scale := n.FontSize
	if scale <= 0 {
		scale = 1.0
	}

	pos := n.GetWorldPosition()
	textW, _ := render.DefaultFont.MeasureText(n.Text)

	var x float64
	switch n.Align {
	case TextAlignCenter:
		x = float64(pos.X) - textW*scale/2
	case TextAlignRight:
		x = float64(pos.X) - textW*scale
	default: // TextAlignLeft
		x = float64(pos.X)
	}
	y := float64(pos.Y)

	cs := &ebiten.ColorScale{}
	cs.Scale(n.Color.R, n.Color.G, n.Color.B, n.Color.A)

	render.DefaultFont.DrawText(screen, n.Text, x, y, scale, cs)
}

// ---------------------------------------------------------------------------
// Button
// ---------------------------------------------------------------------------

// Button is a clickable UI element with a background rectangle and a label.
type Button struct {
	*Node2D
	Label                     *Label
	Width, Height             float32
	NormalColor               struct{ R, G, B, A float32 }
	HoverColor                struct{ R, G, B, A float32 }
	PressedColor              struct{ R, G, B, A float32 }
	disabled                  bool
	onClick                   func()
	hovered                   bool
	pressed                   bool
}

// NewButton creates a new Button node with the given name and label text.
func NewButton(name, text string) *Button {
	btn := &Button{
		Node2D: NewNode2D(name, 0),
		Label:  NewLabel(name+"_Label", text),
		Width:  100,
		Height: 30,
	}
	btn.Label.SetAlign(TextAlignCenter)
	btn.NormalColor = struct{ R, G, B, A float32 }{R: 0.5, G: 0.5, B: 0.5, A: 1.0}
	btn.HoverColor = struct{ R, G, B, A float32 }{R: 0.7, G: 0.7, B: 0.7, A: 1.0}
	btn.PressedColor = struct{ R, G, B, A float32 }{R: 0.3, G: 0.3, B: 0.3, A: 1.0}
	return btn
}

// SetOnClick sets the callback that fires when the button is clicked.
func (n *Button) SetOnClick(fn func()) {
	n.onClick = fn
}

// SetDisabled disables or enables the button. A disabled button does not
// respond to clicks.
func (n *Button) SetDisabled(disabled bool) {
	n.disabled = disabled
}

// Update processes input for hover/click detection and propagates to children.
func (n *Button) Update(dt float64) {
	// Propagate to children
	for _, child := range n.children {
		if child != nil {
			child.Update(dt)
		}
	}

	// Propagate to label (separate node, not in children)
	if n.Label != nil {
		n.Label.Update(dt)
	}

	// Mouse hit test
	mx, my := ebiten.CursorPosition()
	pos := n.GetWorldPosition()

	n.hovered = float32(mx) >= pos.X && float32(mx) <= pos.X+n.Width &&
		float32(my) >= pos.Y && float32(my) <= pos.Y+n.Height

	// Click detection (press → release)
	leftPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	if leftPressed {
		if n.hovered && !n.pressed {
			n.pressed = true
		}
	} else {
		if n.pressed && n.hovered && n.onClick != nil && !n.disabled {
			n.onClick()
		}
		n.pressed = false
	}
}

// Draw renders the button background rect and its label text.
func (n *Button) Draw(screen *ebiten.Image) {
	if !n.visible || !n.alive || screen == nil {
		return
	}

	// Draw children first
	for _, child := range n.children {
		if child != nil {
			child.Draw(screen)
		}
	}

	pos := n.GetWorldPosition()

	if n.Width <= 0 || n.Height <= 0 {
		return
	}

	// Determine background colour based on state
	var bgColor struct{ R, G, B, A float32 }
	if n.disabled {
		bgColor = n.NormalColor
	} else if n.pressed {
		bgColor = n.PressedColor
	} else if n.hovered {
		bgColor = n.HoverColor
	} else {
		bgColor = n.NormalColor
	}

	c := color.RGBA{
		R: uint8(clampF32(bgColor.R) * 255),
		G: uint8(clampF32(bgColor.G) * 255),
		B: uint8(clampF32(bgColor.B) * 255),
		A: uint8(clampF32(bgColor.A) * 255),
	}
	render.DrawRect(screen, float64(pos.X), float64(pos.Y),
		float64(n.Width), float64(n.Height), c)

	// Draw label text centred within the button
	if n.Label == nil || n.Label.Text == "" {
		return
	}

	render.DebugTextAt(screen, "", 0, 0)
	if render.DefaultFont == nil {
		return
	}

	scale := n.Label.FontSize
	if scale <= 0 {
		scale = 1.0
	}

	textW, textH := render.DefaultFont.MeasureText(n.Label.Text)

	var lx float64
	switch n.Label.Align {
	case TextAlignCenter:
		lx = float64(pos.X) + float64(n.Width)/2 - textW*scale/2
	case TextAlignRight:
		lx = float64(pos.X) + float64(n.Width) - textW*scale
	default: // TextAlignLeft
		lx = float64(pos.X)
	}
	ly := float64(pos.Y) + float64(n.Height)/2 - textH*scale/2

	cs := &ebiten.ColorScale{}
	cs.Scale(n.Label.Color.R, n.Label.Color.G, n.Label.Color.B, n.Label.Color.A)

	render.DefaultFont.DrawText(screen, n.Label.Text, lx, ly, scale, cs)
}

// ---------------------------------------------------------------------------
// Panel
// ---------------------------------------------------------------------------

// Panel is a simple filled rectangle node with an optional border.
type Panel struct {
	*Node2D
	Width, Height      float32
	Color              struct{ R, G, B, A float32 }
	BorderColor        struct{ R, G, B, A float32 }
	BorderWidth        float32
}

// NewPanel creates a new Panel node with the given dimensions.
func NewPanel(name string, w, h float32) *Panel {
	return &Panel{
		Node2D:      NewNode2D(name, 0),
		Width:       w,
		Height:      h,
		Color:       struct{ R, G, B, A float32 }{R: 0.2, G: 0.2, B: 0.2, A: 1.0},
		BorderColor: struct{ R, G, B, A float32 }{R: 1, G: 1, B: 1, A: 1.0},
	}
}

// Draw renders the panel's filled rectangle and optional border.
func (n *Panel) Draw(screen *ebiten.Image) {
	if !n.visible || !n.alive || screen == nil {
		return
	}

	// Draw children first
	for _, child := range n.children {
		if child != nil {
			child.Draw(screen)
		}
	}

	if n.Width <= 0 || n.Height <= 0 {
		return
	}

	pos := n.GetWorldPosition()
	x := float64(pos.X)
	y := float64(pos.Y)
	w := float64(n.Width)
	h := float64(n.Height)

	// Fill
	fill := color.RGBA{
		R: uint8(clampF32(n.Color.R) * 255),
		G: uint8(clampF32(n.Color.G) * 255),
		B: uint8(clampF32(n.Color.B) * 255),
		A: uint8(clampF32(n.Color.A) * 255),
	}
	render.DrawRect(screen, x, y, w, h, fill)

	// Border
	if n.BorderWidth > 0 {
		bw := float64(n.BorderWidth)
		border := color.RGBA{
			R: uint8(clampF32(n.BorderColor.R) * 255),
			G: uint8(clampF32(n.BorderColor.G) * 255),
			B: uint8(clampF32(n.BorderColor.B) * 255),
			A: uint8(clampF32(n.BorderColor.A) * 255),
		}
		// Top
		render.DrawRect(screen, x, y, w, bw, border)
		// Bottom
		render.DrawRect(screen, x, y+h-bw, w, bw, border)
		// Left
		render.DrawRect(screen, x, y, bw, h, border)
		// Right
		render.DrawRect(screen, x+w-bw, y, bw, h, border)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// clampF32 clamps a float32 value to the [0, 1] range.
func clampF32(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// Compile-time interface checks
var _ Node = (*Label)(nil)
var _ Node = (*Button)(nil)
var _ Node = (*Panel)(nil)
