// Kora Editor — Visual scene editor built with Go + Ebitengine
//
// Interface inspirada no GameMaker Studio e Godot:
//   - Viewport com grid e entidades
//   - Painel de hierarquia
//   - Painel de inspetor
//   - Toolbar com operações de arquivo
//
// Uso:
//   go run ./cmd/kora-editor
//   make editor

package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ---------------------------------------------------------------------------
// Core data structures (compatible with .kora.json format)
// ---------------------------------------------------------------------------

type EditorEntity struct {
	ID       int              `json:"id"`
	Name     string           `json:"name"`
	Type     string           `json:"type"`
	X, Y     float64          `json:"x,omitempty"`
	W, H     float64          `json:"w,omitempty"`
	Rotation float64          `json:"rotation,omitempty"`
	Color    string           `json:"color,omitempty"`
	Visible  bool             `json:"visible"`
	ParentID int              `json:"parentId,omitempty"`
	Children []*EditorEntity  `json:"children,omitempty"`
	AssetID  string           `json:"assetId,omitempty"`
	ZIndex   int              `json:"zIndex,omitempty"`
}

type SceneMeta struct {
	Name     string `json:"name"`
	Version  int    `json:"version"`
	LogicalW int    `json:"logicalW"`
	LogicalH int    `json:"logicalH"`
}

type SceneDoc struct {
	Meta     SceneMeta       `json:"meta"`
	Entities []*EditorEntity `json:"entities"`
}

// ---------------------------------------------------------------------------
// Editor state
// ---------------------------------------------------------------------------

type EditorTool int
const (
	ToolSelect EditorTool = iota
	ToolMove
	ToolAdd
)

type EditorState struct {
	// Scene data
	scene    SceneDoc
	nextID   int
	filePath string
	dirty    bool

	// Viewport camera
	camX, camY   float64
	camZoom      float64

	// Selection
	selectedID   int
	hoveredID    int

	// Interaction
	dragging     bool
	dragStartX, dragStartY float64
	dragEntityX, dragEntityY float64
	dragEntityID int
	tool         EditorTool

	// UI layout
	screenW, screenH float64
	panelW           float64
	toolbarH         float64
	consoleH         float64

	// Console
	consoleLines []string

	// Grid
	gridSize     float64
	snapEnabled  bool
	snapSize     float64

	// State flags
	showGrid     bool
	showHierarchy bool
	showInspector bool
}

func NewEditor() *EditorState {
	return &EditorState{
		scene: SceneDoc{
			Meta: SceneMeta{Name: "Untitled", Version: 1, LogicalW: 640, LogicalH: 360},
			Entities: []*EditorEntity{
				{ID: 1, Name: "Player", Type: "sprite", X: 320, Y: 180, W: 32, H: 32, Color: "#00e5a0", Visible: true},
				{ID: 2, Name: "Camera", Type: "camera", X: 320, Y: 180, W: 16, H: 16, Color: "#e3b341", Visible: true},
			},
		},
		nextID:        3,
		camZoom:       1.0,
		selectedID:    -1,
		hoveredID:     -1,
		tool:          ToolSelect,
		panelW:        200,
		toolbarH:      36,
		consoleH:      24,
		gridSize:      32,
		snapEnabled:   true,
		snapSize:      16,
		showGrid:      true,
		showHierarchy: true,
		showInspector: true,
		consoleLines:  []string{"Kora Editor v1.0 — Go native"},
	}
}

func (e *EditorState) Log(msg string) {
	e.consoleLines = append(e.consoleLines, msg)
	if len(e.consoleLines) > 100 {
		e.consoleLines = e.consoleLines[len(e.consoleLines)-100:]
	}
}

func (e *EditorState) Logf(format string, args ...interface{}) {
	e.Log(fmt.Sprintf(format, args...))
}

// ---------------------------------------------------------------------------
// Entity operations
// ---------------------------------------------------------------------------

func (e *EditorState) NewEntity(name, etype string) *EditorEntity {
	e.nextID++
	ent := &EditorEntity{
		ID: e.nextID, Name: name, Type: etype,
		X: 320, Y: 180, W: 32, H: 32,
		Color: "#00e5a0", Visible: true,
	}
	e.scene.Entities = append(e.scene.Entities, ent)
	e.dirty = true
	e.Logf("Created %s: %s", etype, name)
	return ent
}

func (e *EditorState) GetEntity(id int) *EditorEntity {
	for _, ent := range e.scene.Entities {
		if ent.ID == id {
			return ent
		}
	}
	return nil
}

func (e *EditorState) DeleteEntity(id int) {
	if id < 0 { return }
	idx := -1
	for i, ent := range e.scene.Entities {
		if ent.ID == id {
			idx = i
			break
		}
	}
	if idx >= 0 {
		name := e.scene.Entities[idx].Name
		e.scene.Entities = append(e.scene.Entities[:idx], e.scene.Entities[idx+1:]...)
		if e.selectedID == id {
			e.selectedID = -1
		}
		e.dirty = true
		e.Logf("Deleted: %s", name)
	}
}

func (e *EditorState) DuplicateEntity(id int) {
	ent := e.GetEntity(id)
	if ent == nil { return }
	e.nextID++
	clone := *ent
	clone.ID = e.nextID
	clone.Name = ent.Name + "_copy"
	clone.X += 16
	clone.Y += 16
	e.scene.Entities = append(e.scene.Entities, &clone)
	e.selectedID = clone.ID
	e.dirty = true
	e.Logf("Duplicated: %s → %s", ent.Name, clone.Name)
}

// ---------------------------------------------------------------------------
// World/Screen conversion
// ---------------------------------------------------------------------------

func (e *EditorState) screenToWorld(sx, sy float64) (float64, float64) {
	vpX := float64(e.panelW)
	vpY := float64(e.toolbarH)
	vpW := float64(e.screenW) - float64(e.panelW)*2
	vpH := float64(e.screenH) - float64(e.toolbarH) - float64(e.consoleH)
	cx := vpX + vpW/2 + e.camX*e.camZoom
	cy := vpY + vpH/2 + e.camY*e.camZoom
	return (sx - cx) / e.camZoom, (sy - cy) / e.camZoom
}

func (e *EditorState) worldToScreen(wx, wy float64) (float64, float64) {
	vpX := float64(e.panelW)
	vpY := float64(e.toolbarH)
	vpW := float64(e.screenW) - float64(e.panelW)*2
	vpH := float64(e.screenH) - float64(e.toolbarH) - float64(e.consoleH)
	cx := vpX + vpW/2 + e.camX*e.camZoom
	cy := vpY + vpH/2 + e.camY*e.camZoom
	return cx + wx*e.camZoom, cy + wy*e.camZoom
}

func (e *EditorState) isInViewport(sx, sy float64) bool {
	return sx >= float64(e.panelW) && sx < float64(e.screenW-e.panelW) &&
		sy >= float64(e.toolbarH) && sy < float64(e.screenH-e.consoleH)
}

// ---------------------------------------------------------------------------
// Load/Save
// ---------------------------------------------------------------------------

func (e *EditorState) Save(path string) error {
	data, err := json.MarshalIndent(e.scene, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	e.filePath = path
	e.dirty = false
	e.Logf("Saved: %s (%d entities)", path, len(e.scene.Entities))
	return nil
}

func (e *EditorState) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	var doc SceneDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	e.scene = doc
	e.filePath = path
	e.dirty = false
	e.selectedID = -1
	e.nextID = 0
	for _, ent := range doc.Entities {
		if ent.ID > e.nextID {
			e.nextID = ent.ID
		}
	}
	e.nextID++
	e.Logf("Loaded: %s (%d entities)", path, len(doc.Entities))
	return nil
}

// ---------------------------------------------------------------------------
// Scene entity tree
// ---------------------------------------------------------------------------

func allEntities(ents []*EditorEntity) []*EditorEntity {
	var result []*EditorEntity
	for _, e := range ents {
		result = append(result, e)
		result = append(result, allEntities(e.Children)...)
	}
	return result
}

// ---------------------------------------------------------------------------
// Ebitengine game interface
// ---------------------------------------------------------------------------

type EditorApp struct {
	state *EditorState
}

func (app *EditorApp) Update() error {
	e := app.state
	w, h := ebiten.WindowSize(); e.screenW, e.screenH = float64(w), float64(h)

	// --- Keyboard shortcuts ---
	if inpututil.IsKeyJustPressed(ebiten.KeyN) && ebiten.IsKeyPressed(ebiten.KeyControl) {
		e.scene = SceneDoc{Meta: SceneMeta{Name: "Untitled", Version: 1, LogicalW: 640, LogicalH: 360}}
		e.nextID = 1
		e.selectedID = -1
		e.dirty = false
		e.filePath = ""
		e.Log("New scene created")
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) && ebiten.IsKeyPressed(ebiten.KeyControl) {
		path := e.filePath
		if path == "" {
			path = "untitled.kora.json"
		}
		e.Save(path)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyO) && ebiten.IsKeyPressed(ebiten.KeyControl) {
		// For now, just try to open a dialog via console
		e.Log("Use: make editor-load FILE=scene.kora.json")
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
		e.showGrid = !e.showGrid
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF5) {
		e.Log("Play mode - not yet implemented")
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) || inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		if e.selectedID > 0 {
			e.DeleteEntity(e.selectedID)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) && ebiten.IsKeyPressed(ebiten.KeyControl) {
		if e.selectedID > 0 {
			e.DuplicateEntity(e.selectedID)
		}
	}

	// --- Tool switching ---
	if inpututil.IsKeyJustPressed(ebiten.Key1) { e.tool = ToolSelect }
	if inpututil.IsKeyJustPressed(ebiten.Key2) { e.tool = ToolMove }
	if inpututil.IsKeyJustPressed(ebiten.Key3) { e.tool = ToolAdd }

	// --- Mouse ---
	mx, my := ebiten.CursorPosition()
	mxf, myf := float64(mx), float64(my)

	if e.isInViewport(mxf, myf) {
		wx, wy := e.screenToWorld(mxf, myf)

		// Scroll to zoom
		_, dy := ebiten.Wheel()
		if dy != 0 {
			e.camZoom *= 1.0 + dy*0.1
			if e.camZoom < 0.1 { e.camZoom = 0.1 }
			if e.camZoom > 10 { e.camZoom = 10 }
		}

		// Middle button drag to pan
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
			if !e.dragging {
				e.dragging = true
				e.dragStartX, e.dragStartY = mxf, myf
				e.dragEntityX, e.dragEntityY = e.camX, e.camY
			}
			e.camX = e.dragEntityX + (mxf - e.dragStartX) / e.camZoom
			e.camY = e.dragEntityY + (myf - e.dragStartY) / e.camZoom
		} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			if !e.dragging {
				// Hit test: check entities in reverse draw order
				hitID := -1
				ents := e.scene.Entities
				for i := len(ents) - 1; i >= 0; i-- {
					ent := ents[i]
					if !ent.Visible { continue }
					if wx >= ent.X - ent.W/2 && wx <= ent.X + ent.W/2 &&
						wy >= ent.Y - ent.H/2 && wy <= ent.Y + ent.H/2 {
						hitID = ent.ID
						break
					}
				}
				if hitID >= 0 {
					e.selectedID = hitID
					e.dragging = true
					e.dragEntityID = hitID
					ent := e.GetEntity(hitID)
					if ent != nil {
						e.dragStartX, e.dragStartY = wx, wy
						e.dragEntityX, e.dragEntityY = ent.X, ent.Y
					}
				} else {
					e.selectedID = -1
				}
			} else {
				// Dragging an entity
				if e.dragEntityID > 0 {
					ent := e.GetEntity(e.dragEntityID)
					if ent != nil {
						dx := wx - e.dragStartX
						dy := wy - e.dragStartY
						if e.snapEnabled && e.snapSize > 0 {
							ent.X = math.Round((e.dragEntityX+dx)/e.snapSize) * e.snapSize
							ent.Y = math.Round((e.dragEntityY+dy)/e.snapSize) * e.snapSize
						} else {
							ent.X = e.dragEntityX + dx
							ent.Y = e.dragEntityY + dy
						}
						e.dirty = true
					}
				}
			}
		} else {
			e.dragging = false
			e.dragEntityID = -1

			// Hover detection
			e.hoveredID = -1
			ents := e.scene.Entities
			for i := len(ents) - 1; i >= 0; i-- {
				ent := ents[i]
				if !ent.Visible { continue }
				if wx >= ent.X - ent.W/2 && wx <= ent.X + ent.W/2 &&
					wy >= ent.Y - ent.H/2 && wy <= ent.Y + ent.H/2 {
					e.hoveredID = ent.ID
					break
				}
			}
		}
	}

	return nil
}

func (app *EditorApp) Draw(screen *ebiten.Image) {
	e := app.state

	// Background
	screen.Fill(color.RGBA{0x1a, 0x1a, 0x2e, 0xff})

	// --- Toolbar ---
	drawRect(screen, 0, 0, e.screenW, e.toolbarH, color.RGBA{0x25, 0x25, 0x3d, 0xff})
	tbY := float64(e.toolbarH) / 2

	// Title
	title := fmt.Sprintf("Kora Editor — %s%s", e.scene.Meta.Name, map[bool]string{true: " ●", false: ""}[e.dirty])
	drawText(screen, title, 8, tbY-6, 1.0, color.White)

	// Tool buttons
	btnX := 300.0
	tools := []string{"Sel", "Move", "Add"}
	curTool := int(e.tool)
	for i, name := range tools {
		bg := color.RGBA{0x3a, 0x3a, 0x55, 0xff}
		if i == curTool {
			bg = color.RGBA{0x00, 0xe5, 0xa0, 0xaa}
		}
		drawRect(screen, btnX, 4, 48, e.toolbarH-8, bg)
		drawText(screen, name, btnX+16, tbY-6, 0.8, color.White)
		btnX += 54
	}

	// FPS
	fps := fmt.Sprintf("FPS: %.0f  Zoom: %.0f%%", ebiten.ActualFPS(), e.camZoom*100)
	drawText(screen, fps, float64(e.screenW)-160, tbY-6, 0.8, color.RGBA{0x88, 0x88, 0xaa, 0xff})

	// Snap toggle
	snapText := fmt.Sprintf("Snap: %d", int(e.snapSize))
	snapColor := color.RGBA{0x88, 0x88, 0xaa, 0xff}
	if e.snapEnabled { snapColor = color.RGBA{0x00, 0xe5, 0xa0, 0xff} }
	drawText(screen, snapText, float64(e.screenW)-280, tbY-6, 0.8, snapColor)

	// --- Viewport ---
	vpX := float64(e.panelW)
	vpY := float64(e.toolbarH)
	vpW := float64(e.screenW) - float64(e.panelW)*2
	vpH := float64(e.screenH) - float64(e.toolbarH) - float64(e.consoleH)
	drawRect(screen, vpX, vpY, vpW, vpH, color.RGBA{0x0d, 0x0d, 0x1a, 0xff})

	// Grid
	if e.showGrid {
		drawGrid(screen, e)
	}

	// Entities
	drawRect(screen, vpX, vpY, vpW, vpH, color.RGBA{0, 0, 0, 0}, 2) // viewport border
	for _, ent := range e.scene.Entities {
		if !ent.Visible { continue }
		drawEntity(screen, e, ent)
	}

	// Selection outline
	if e.selectedID > 0 {
		ent := e.GetEntity(e.selectedID)
		if ent != nil {
			sx, sy := e.worldToScreen(ent.X, ent.Y)
			sw, sh := ent.W*e.camZoom, ent.H*e.camZoom
			drawRect(screen, sx-sw/2-2, sy-sh/2-2, sw+4, sh+4, color.RGBA{0x00, 0xe5, 0xa0, 0xff}, 2)
		}
	}

	// --- Hierarchy panel ---
	if e.showHierarchy {
		drawRect(screen, 0, vpY, vpX, vpH, color.RGBA{0x1f, 0x1f, 0x33, 0xff})
		drawText(screen, "Hierarquia", 6, vpY+4, 0.9, color.RGBA{0x88, 0x88, 0xaa, 0xff})
		y := vpY + 22
		for _, ent := range e.scene.Entities {
			col := color.RGBA{0xff, 0xff, 0xff, 0xff}
			if ent.ID == e.selectedID { col = color.RGBA{0x00, 0xe5, 0xa0, 0xff} }
			if ent.ID == e.hoveredID { col = color.RGBA{0x00, 0xc8, 0xff, 0xff} }
			icon := map[string]string{"sprite": "▣", "camera": "◉", "tilemap": "▤", "audio": "♪", "custom": "◇"}[ent.Type]
			drawText(screen, fmt.Sprintf("%s %s", icon, ent.Name), 8, y, 0.8, col)
			y += 18
			if y > vpY+vpH-10 { break }
		}
	}

	// --- Inspector panel ---
	if e.showInspector {
		ix := float64(e.screenW - e.panelW)
		drawRect(screen, ix, vpY, float64(e.panelW), vpH, color.RGBA{0x1f, 0x1f, 0x33, 0xff})
		drawText(screen, "Inspetor", ix+6, vpY+4, 0.9, color.RGBA{0x88, 0x88, 0xaa, 0xff})

		if e.selectedID > 0 {
			ent := e.GetEntity(e.selectedID)
			if ent != nil {
				y := vpY + 24
				drawText(screen, "Nome: "+ent.Name, ix+8, y, 0.8, color.White); y += 16
				drawText(screen, fmt.Sprintf("Tipo: %s", ent.Type), ix+8, y, 0.8, color.RGBA{0xaa, 0xaa, 0xcc, 0xff}); y += 16
				drawText(screen, fmt.Sprintf("X: %.0f", ent.X), ix+8, y, 0.8, color.White); y += 16
				drawText(screen, fmt.Sprintf("Y: %.0f", ent.Y), ix+8, y, 0.8, color.White); y += 16
				drawText(screen, fmt.Sprintf("W: %.0f", ent.W), ix+8, y, 0.8, color.White); y += 16
				drawText(screen, fmt.Sprintf("H: %.0f", ent.H), ix+8, y, 0.8, color.White); y += 16
				if ent.Rotation != 0 {
					drawText(screen, fmt.Sprintf("Rot: %.1f°", ent.Rotation), ix+8, y, 0.8, color.White); y += 16
				}
				drawText(screen, fmt.Sprintf("Visível: %t", ent.Visible), ix+8, y, 0.8, color.RGBA{0xaa, 0xaa, 0xcc, 0xff}); y += 16
				if ent.AssetID != "" {
					drawText(screen, "Asset: "+filepath.Base(ent.AssetID), ix+8, y, 0.8, color.RGBA{0xaa, 0xaa, 0xcc, 0xff})
				}
			}
		} else {
			drawText(screen, "Selecione uma entidade", ix+8, vpY+24, 0.8, color.RGBA{0x66, 0x66, 0x88, 0xff})
		}
	}

	// --- Console ---
	cy := float64(e.screenH - e.consoleH)
	drawRect(screen, 0, cy, float64(e.screenW), float64(e.consoleH), color.RGBA{0x15, 0x15, 0x25, 0xff})
	if len(e.consoleLines) > 0 {
		last := e.consoleLines[len(e.consoleLines)-1]
		drawText(screen, "❯ "+last, 8, cy+6, 0.8, color.RGBA{0x88, 0xcc, 0xaa, 0xff})
	}
}

func (app *EditorApp) Layout(outsideW, outsideH int) (int, int) {
	return outsideW, outsideH
}

// ---------------------------------------------------------------------------
// Drawing helpers
// ---------------------------------------------------------------------------

var whitePixel = ebiten.NewImage(1, 1)

func init() {
	whitePixel.Fill(color.White)
}

func drawRect(screen *ebiten.Image, x, y, w, h float64, c color.Color, borderWidth ...float64) {
	if len(borderWidth) > 0 && borderWidth[0] > 0 {
		bw := borderWidth[0]
		// Draw as border: 4 rectangles
		fillRect(screen, x, y, w, bw, c)        // top
		fillRect(screen, x, y+h-bw, w, bw, c)    // bottom
		fillRect(screen, x, y, bw, h, c)          // left
		fillRect(screen, x+w-bw, y, bw, h, c)     // right
		return
	}
	fillRect(screen, x, y, w, h, c)
}

func fillRect(screen *ebiten.Image, x, y, w, h float64, c color.Color) {
	if w <= 0 || h <= 0 { return }
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(w, h)
	op.GeoM.Translate(x, y)
	op.ColorScale.SetR(float32(colorComponent(c, 0)))
	op.ColorScale.SetG(float32(colorComponent(c, 1)))
	op.ColorScale.SetB(float32(colorComponent(c, 2)))
	op.ColorScale.SetA(float32(colorComponent(c, 3)))
	screen.DrawImage(whitePixel, op)
}

func colorComponent(c color.Color, idx int) uint8 {
	rgba := color.RGBAModel.Convert(c).(color.RGBA)
	switch idx {
	case 0: return rgba.R
	case 1: return rgba.G
	case 2: return rgba.B
	case 3: return rgba.A
	}
	return 255
}

func drawText(screen *ebiten.Image, text string, x, y float64, scale float64, clr color.Color) {
	if text == "" { return }
	// Uses the built-in bitmap font from render package
	// For now, use a simple approach: draw with color
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x, y)
	rgba := color.RGBAModel.Convert(clr).(color.RGBA)
	op.ColorScale.SetR(float32(rgba.R) / 255)
	op.ColorScale.SetG(float32(rgba.G) / 255)
	op.ColorScale.SetB(float32(rgba.B) / 255)
	op.ColorScale.SetA(float32(rgba.A) / 255)

	// Simple monochrome text rendering using the font system
	// Each character is drawn as a small rect from the font atlas
	_ = op
	_ = screen
	// TODO: Wire up render.DefaultFont when available
	// For now, log is sufficient — font rendering requires GPU context
}

func drawEntity(screen *ebiten.Image, e *EditorState, ent *EditorEntity) {
	sx, sy := e.worldToScreen(ent.X, ent.Y)
	sw := ent.W * e.camZoom
	sh := ent.H * e.camZoom

	// Parse color
	col := parseHexColor(ent.Color)
	if ent.Color == "" || (col.R == 0 && col.G == 0 && col.B == 0 && col.A == 0) {
		colors := map[string]color.RGBA{
			"sprite":  {0x00, 0xe5, 0xa0, 0xcc},
			"camera":  {0xe3, 0xb3, 0x41, 0xcc},
			"tilemap": {0x38, 0x8b, 0xfd, 0xcc},
			"audio":   {0xf8, 0x51, 0x49, 0xcc},
			"custom":  {0xbc, 0x8c, 0xff, 0xcc},
		}
		col = colors[ent.Type]
		if (col.R == 0 && col.G == 0 && col.B == 0) { col = color.RGBA{0x88, 0x88, 0xcc, 0xcc} }
	}

	// Entity rect
	drawRect(screen, sx-sw/2, sy-sh/2, sw, sh, col)

	// Entity name
	nameScale := 0.7 * e.camZoom
	if nameScale < 0.5 { nameScale = 0.5 }
	if nameScale > 1.2 { nameScale = 1.2 }
	drawText(screen, ent.Name, sx-sw/2+2, sy-sh/2-14*nameScale, nameScale, color.RGBA{0xff, 0xff, 0xff, 0xcc})
}

func drawGrid(screen *ebiten.Image, e *EditorState) {
	gs := e.gridSize * e.camZoom
	if gs < 4 { gs = 4 }

	vpX := float64(e.panelW)
	vpY := float64(e.toolbarH)
	vpW := float64(e.screenW) - float64(e.panelW)*2
	vpH := float64(e.screenH) - float64(e.toolbarH) - float64(e.consoleH)

	ox := math.Mod(e.camX*e.camZoom, gs)
	oy := math.Mod(e.camY*e.camZoom, gs)
	if ox < 0 { ox += gs }
	if oy < 0 { oy += gs }

	gridCol := color.RGBA{0x2a, 0x2a, 0x4a, 0x40}

	for gx := vpX + ox; gx < vpX+vpW; gx += gs {
		drawRect(screen, gx, vpY, 1, vpH, gridCol)
	}
	for gy := vpY + oy; gy < vpY+vpH; gy += gs {
		drawRect(screen, vpX, gy, vpW, 1, gridCol)
	}
}

func parseHexColor(hex string) color.RGBA {
	if len(hex) == 7 && hex[0] == '#' {
		r := hexByte(hex[1])<<4 | hexByte(hex[2])
		g := hexByte(hex[3])<<4 | hexByte(hex[4])
		b := hexByte(hex[5])<<4 | hexByte(hex[6])
		return color.RGBA{r, g, b, 255}
	}
	return color.RGBA{}
}

func hexByte(c byte) uint8 {
	switch {
	case c >= '0' && c <= '9': return c - '0'
	case c >= 'a' && c <= 'f': return c - 'a' + 10
	case c >= 'A' && c <= 'F': return c - 'A' + 10
	}
	return 0
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	state := NewEditor()

	// Load scene from command-line arg if provided
	if len(os.Args) > 1 {
		path := os.Args[1]
		if err := state.Load(path); err != nil {
			log.Fatalf("Failed to load scene %q: %v", path, err)
		}
	}

	ebiten.SetWindowTitle("Kora Editor")
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	app := &EditorApp{state: state}
	state.Log("Editor ready — Ctrl+N: new, Ctrl+S: save, F3: grid, 1-3: tools")
	state.Log("Middle-click: pan, Scroll: zoom, Click to select, Drag to move")

	if err := ebiten.RunGame(app); err != nil {
		log.Fatal(err)
	}
}
