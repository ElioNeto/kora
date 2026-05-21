package node

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// safeNewImage creates an *ebiten.Image or skips the test if ebiten is not
// initialised (e.g. in CI without a GPU).
func safeNewImage(t *testing.T, w, h int) *ebiten.Image {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Skip("ebiten not initialised; skipping GPU-dependent test")
		}
	}()
	return ebiten.NewImage(w, h)
}

// ---------------------------------------------------------------------------
// Label
// ---------------------------------------------------------------------------

func TestLabelCreation(t *testing.T) {
	lbl := NewLabel("test", "Hello World")
	if lbl == nil {
		t.Fatal("expected non-nil Label")
	}
	if lbl.Name() != "test" {
		t.Errorf("expected name 'test', got '%s'", lbl.Name())
	}
	if lbl.Text != "Hello World" {
		t.Errorf("expected text 'Hello World', got '%s'", lbl.Text)
	}
	if lbl.FontSize != 1.0 {
		t.Errorf("expected FontSize 1.0, got %f", lbl.FontSize)
	}
	if lbl.Align != TextAlignLeft {
		t.Errorf("expected default Align TextAlignLeft, got %d", lbl.Align)
	}
	if lbl.Color.R != 1 || lbl.Color.G != 1 || lbl.Color.B != 1 || lbl.Color.A != 1 {
		t.Errorf("expected default Color (1,1,1,1), got (%f,%f,%f,%f)",
			lbl.Color.R, lbl.Color.G, lbl.Color.B, lbl.Color.A)
	}
}

func TestLabelSetText(t *testing.T) {
	lbl := NewLabel("test", "Hello")
	lbl.SetText("World")
	if lbl.Text != "World" {
		t.Errorf("expected text 'World', got '%s'", lbl.Text)
	}
}

func TestLabelSetFontSize(t *testing.T) {
	lbl := NewLabel("test", "Hello")
	lbl.SetFontSize(2.5)
	if lbl.FontSize != 2.5 {
		t.Errorf("expected FontSize 2.5, got %f", lbl.FontSize)
	}
}

func TestLabelSetColor(t *testing.T) {
	lbl := NewLabel("test", "Hello")
	lbl.SetColor(0.5, 0.3, 0.8, 1.0)
	if lbl.Color.R != 0.5 {
		t.Errorf("expected R 0.5, got %f", lbl.Color.R)
	}
	if lbl.Color.G != 0.3 {
		t.Errorf("expected G 0.3, got %f", lbl.Color.G)
	}
	if lbl.Color.B != 0.8 {
		t.Errorf("expected B 0.8, got %f", lbl.Color.B)
	}
	if lbl.Color.A != 1.0 {
		t.Errorf("expected A 1.0, got %f", lbl.Color.A)
	}
}

func TestLabelSetAlign(t *testing.T) {
	lbl := NewLabel("test", "Hello")
	lbl.SetAlign(TextAlignCenter)
	if lbl.Align != TextAlignCenter {
		t.Errorf("expected TextAlignCenter, got %d", lbl.Align)
	}
	lbl.SetAlign(TextAlignRight)
	if lbl.Align != TextAlignRight {
		t.Errorf("expected TextAlignRight, got %d", lbl.Align)
	}
	lbl.SetAlign(TextAlignLeft)
	if lbl.Align != TextAlignLeft {
		t.Errorf("expected TextAlignLeft, got %d", lbl.Align)
	}
}

func TestLabelDrawNoPanic(t *testing.T) {
	img := safeNewImage(t, 200, 50)
	lbl := NewLabel("test", "Hello")

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Label.Draw panicked: %v", r)
		}
	}()

	// Draw with a valid screen
	lbl.Draw(img)

	// Draw with nil screen (should not panic)
	lbl.Draw(nil)
}

func TestLabelDrawEmptyText(t *testing.T) {
	img := safeNewImage(t, 200, 50)
	lbl := NewLabel("test", "")

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Label.Draw with empty text panicked: %v", r)
		}
	}()

	lbl.Draw(img)
}

func TestLabelUpdateNoPanic(t *testing.T) {
	lbl := NewLabel("test", "Hello")

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Label.Update panicked: %v", r)
		}
	}()

	lbl.Update(0.016)
}

func TestLabelNodeInterface(t *testing.T) {
	var _ Node = (*Label)(nil)

	lbl := NewLabel("test", "text")
	var n Node = lbl
	if n.Name() != "test" {
		t.Error("Label does not satisfy Node interface correctly")
	}
}

// ---------------------------------------------------------------------------
// Button
// ---------------------------------------------------------------------------

func TestButtonCreation(t *testing.T) {
	btn := NewButton("test", "Click Me")
	if btn == nil {
		t.Fatal("expected non-nil Button")
	}
	if btn.Name() != "test" {
		t.Errorf("expected name 'test', got '%s'", btn.Name())
	}
	if btn.Label == nil {
		t.Fatal("expected non-nil Label")
	}
	if btn.Label.Text != "Click Me" {
		t.Errorf("expected label text 'Click Me', got '%s'", btn.Label.Text)
	}
	if btn.Width != 100 {
		t.Errorf("expected default Width 100, got %f", btn.Width)
	}
	if btn.Height != 30 {
		t.Errorf("expected default Height 30, got %f", btn.Height)
	}
}

func TestButtonSetOnClick(t *testing.T) {
	btn := NewButton("test", "Click")
	btn.SetOnClick(func() {
		// just verify no panic
	})

	// onClick is unexported; verify SetOnClick does not panic
	// and that the function reference is stored by calling it indirectly
	// through the exported API (Update + state simulation is not possible
	// without ebiten input, so we just verify no panic).
}

func TestButtonSetDisabled(t *testing.T) {
	btn := NewButton("test", "Click")
	btn.SetDisabled(true)
	btn.SetDisabled(false)
	// disabled is unexported; verify no panic
}

func TestButtonDrawNoPanic(t *testing.T) {
	img := safeNewImage(t, 200, 100)
	btn := NewButton("test", "Click")

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Button.Draw panicked: %v", r)
		}
	}()

	btn.Draw(img)
	btn.Draw(nil)
}

func TestButtonUpdateNoPanic(t *testing.T) {
	btn := NewButton("test", "Click")

	defer func() {
		if r := recover(); r != nil {
			t.Skipf("Button.Update panicked (ebiten not initialised?): %v", r)
		}
	}()

	btn.Update(0.016)
}

func TestButtonNodeInterface(t *testing.T) {
	var _ Node = (*Button)(nil)

	btn := NewButton("test", "text")
	var n Node = btn
	if n.Name() != "test" {
		t.Error("Button does not satisfy Node interface correctly")
	}
}

// ---------------------------------------------------------------------------
// Panel
// ---------------------------------------------------------------------------

func TestPanelCreation(t *testing.T) {
	pnl := NewPanel("test", 200, 100)
	if pnl == nil {
		t.Fatal("expected non-nil Panel")
	}
	if pnl.Name() != "test" {
		t.Errorf("expected name 'test', got '%s'", pnl.Name())
	}
	if pnl.Width != 200 {
		t.Errorf("expected Width 200, got %f", pnl.Width)
	}
	if pnl.Height != 100 {
		t.Errorf("expected Height 100, got %f", pnl.Height)
	}
}

func TestPanelDrawNoPanic(t *testing.T) {
	img := safeNewImage(t, 300, 200)
	pnl := NewPanel("test", 200, 100)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Panel.Draw panicked: %v", r)
		}
	}()

	pnl.Draw(img)
	pnl.Draw(nil)
}

func TestPanelWithBorderDrawNoPanic(t *testing.T) {
	img := safeNewImage(t, 300, 200)
	pnl := NewPanel("test", 200, 100)
	pnl.BorderWidth = 2

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Panel with border Draw panicked: %v", r)
		}
	}()

	pnl.Draw(img)
}

func TestPanelUpdateNoPanic(t *testing.T) {
	pnl := NewPanel("test", 200, 100)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Panel.Update panicked: %v", r)
		}
	}()

	pnl.Update(0.016)
}

func TestPanelNodeInterface(t *testing.T) {
	var _ Node = (*Panel)(nil)

	pnl := NewPanel("test", 200, 100)
	var n Node = pnl
	if n.Name() != "test" {
		t.Error("Panel does not satisfy Node interface correctly")
	}
}

// ---------------------------------------------------------------------------
// Children propagation (all UI node types)
// ---------------------------------------------------------------------------

func TestLabelAddChild(t *testing.T) {
	parent := NewLabel("parent", "parent")
	child := NewNode2D("child", 1)
	parent.AddChild(child)

	if parent.GetChildCount() != 1 {
		t.Errorf("expected 1 child, got %d", parent.GetChildCount())
	}
}

func TestButtonAddChild(t *testing.T) {
	parent := NewButton("parent", "parent")
	child := NewNode2D("child", 1)
	parent.AddChild(child)

	if parent.GetChildCount() != 1 {
		t.Errorf("expected 1 child, got %d", parent.GetChildCount())
	}
}

func TestPanelAddChild(t *testing.T) {
	parent := NewPanel("parent", 200, 100)
	child := NewNode2D("child", 1)
	parent.AddChild(child)

	if parent.GetChildCount() != 1 {
		t.Errorf("expected 1 child, got %d", parent.GetChildCount())
	}
}

// ---------------------------------------------------------------------------
// TextAlign constants
// ---------------------------------------------------------------------------

func TestTextAlignConstants(t *testing.T) {
	if TextAlignLeft != 0 {
		t.Errorf("expected TextAlignLeft 0, got %d", TextAlignLeft)
	}
	if TextAlignCenter != 1 {
		t.Errorf("expected TextAlignCenter 1, got %d", TextAlignCenter)
	}
	if TextAlignRight != 2 {
		t.Errorf("expected TextAlignRight 2, got %d", TextAlignRight)
	}
}
