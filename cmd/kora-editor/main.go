// Kora Editor v2 — Editor visual nativo em Go, inspirado no GameMaker/Godot
//
// Layout:
//   ┌──────────────────────────────────────────────────────────┐
//   │  Toolbar: [New] [Save] [Open]  Tabs: Scene | Assets     │
//   ├──────────┬───────────────────────────────┬───────────────┤
//   │          │                               │  Inspector    │
//   │Hierarchy │        Viewport               │  Nome: Player│
//   │  Player  │    ┌──────┐   ┌──────┐       │  X: 320      │
//   │  Ground  │    │Player│   │Camera│       │  Y: 180      │
//   │  Camera  │    └──────┘   └──────┘       │  W: 32       │
//   │          │                               │  H: 32       │
//   ├──────────┴───────────────────────────────┴───────────────┤
//   │  Console: [Info] Entity created                          │
//   └──────────────────────────────────────────────────────────┘

package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ---------------------------------------------------------------------------
// Scene data models (.kora.json compatible)
// ---------------------------------------------------------------------------

type SceneEntity struct {
	ID       int            `json:"id"`
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	X, Y     float64        `json:"x"`
	W, H     float64        `json:"w"`
	Rotation float64        `json:"rotation,omitempty"`
	Color    string         `json:"color,omitempty"`
	Visible  bool           `json:"visible"`
	ParentID int            `json:"parentId,omitempty"`
	Children []*SceneEntity `json:"children,omitempty"`
	AssetID  string         `json:"assetId,omitempty"`
	ZIndex   int            `json:"zIndex,omitempty"`
	Script   string         `json:"script,omitempty"`
}

type SceneMeta struct {
	Name     string `json:"name"`
	Version  int    `json:"version"`
	LogicalW int    `json:"logicalW"`
	LogicalH int    `json:"logicalH"`
}

type SceneFile struct {
	Meta     SceneMeta      `json:"meta"`
	Entities []*SceneEntity `json:"entities"`
}

// ---------------------------------------------------------------------------
// Tab types for the editor
// ---------------------------------------------------------------------------

type EditorTab int
const (
	TabScene EditorTab = iota
	TabAssets
	TabCode
	TabPreview
)

// ---------------------------------------------------------------------------
// Editor state — mirrors the web editor's state model
// ---------------------------------------------------------------------------

type Tool int
const (
	ToolSelect Tool = iota
	ToolMove
	ToolScale
)

type Editor struct {
	// Scene
	scene   SceneFile
	nextID  int
	filePath string
	dirty   bool

	// Viewport camera
	camX, camY   float64
	camZoom      float64

	// Selection
	selectedID   int
	selectedIDs  map[int]bool // multi-select

	// Interaction
	dragging     bool
	dragStartX, dragStartY float64
	dragEntityX, dragEntityY float64
	dragEntityID int
	dragOrigins  map[int][2]float64 // original positions for multi-drag
	panning      bool
	panStartX, panStartY float64
	panCamX, panCamY     float64

	// Tools
	tool     Tool

	// UI state
	activeTab   EditorTab
	showGrid    bool
	snapEnabled bool
	snapSize    float64
	showConsole bool

	// Layout
	viewportW, viewportH float64
	hierarchyW           float64
	inspectorW           float64
	toolbarH             float64
	consoleH             float64

	// Grid
	gridSize     float64

	// Console
	console     []string

	// Theme colors
	bgDark      color.RGBA
	bgPanel     color.RGBA
	bgViewport  color.RGBA
	accent      color.RGBA
	textPrimary color.RGBA
	textMuted   color.RGBA
	textFaint   color.RGBA
}

func NewEditor() *Editor {
	e := &Editor{
		scene: SceneFile{
			Meta: SceneMeta{Name: "Untitled", Version: 1, LogicalW: 360, LogicalH: 640},
			Entities: []*SceneEntity{
				{ID: 1, Name: "Player", Type: "sprite", X: 180, Y: 320, W: 32, H: 32, Color: "#00e5a0", Visible: true, ZIndex: 0},
				{ID: 2, Name: "Ground", Type: "tilemap", X: 180, Y: 600, W: 340, H: 24, Color: "#388bfd", Visible: true, ZIndex: 0},
				{ID: 3, Name: "MainCamera", Type: "camera", X: 180, Y: 320, W: 16, H: 16, Color: "#e3b341", Visible: true, ZIndex: 0},
			},
		},
		nextID:       4,
		camZoom:      0.75,
		selectedID:   -1,
		selectedIDs:  make(map[int]bool),
		tool:         ToolSelect,
		activeTab:    TabScene,
		showGrid:     true,
		snapEnabled:  true,
		snapSize:     16,
		showConsole:  true,
		hierarchyW:   220,
		inspectorW:   240,
		toolbarH:     40,
		consoleH:     80,
		filePath:     "",
		bgDark:       color.RGBA{0x0d, 0x11, 0x17, 0xff},
		bgPanel:      color.RGBA{0x16, 0x1b, 0x22, 0xff},
		bgViewport:   color.RGBA{0x0d, 0x11, 0x17, 0xff},
		accent:       color.RGBA{0x00, 0xe5, 0xa0, 0xff},
		textPrimary:  color.RGBA{0xe6, 0xed, 0xf3, 0xff},
		textMuted:    color.RGBA{0x7d, 0x85, 0x90, 0xff},
		textFaint:    color.RGBA{0x48, 0x4f, 0x58, 0xff},
	}
	e.Log("Kora Editor v2 — Go native")
	e.Log("Ctrl+N: novo | Ctrl+S: salvar | 1-3: ferramentas | F3: grid")
	return e
}

func (e *Editor) Log(msg string) {
	e.console = append(e.console, msg)
	if len(e.console) > 200 {
		e.console = e.console[len(e.console)-200:]
	}
}

func (e *Editor) Logf(f string, args ...interface{}) {
	e.Log(fmt.Sprintf(f, args...))
}

// ---------------------------------------------------------------------------
// Entity management
// ---------------------------------------------------------------------------

func (e *Editor) NewEntity(name, etype string) *SceneEntity {
	e.nextID++
	ent := &SceneEntity{
		ID: e.nextID, Name: name, Type: etype,
		X: 180, Y: 320, W: 48, H: 48,
		Color: e.randColor(), Visible: true, ZIndex: 0,
	}
	e.scene.Entities = append(e.scene.Entities, ent)
	e.selectedID = ent.ID
	e.dirty = true
	e.Logf("Added: %s (%s)", name, etype)
	return ent
}

func (e *Editor) GetEntity(id int) *SceneEntity {
	for _, ent := range e.scene.Entities {
		if ent.ID == id { return ent }
	}
	return nil
}

func (e *Editor) DeleteEntity(id int) {
	if id <= 0 { return }
	for i, ent := range e.scene.Entities {
		if ent.ID == id {
			e.Logf("Deleted: %s", ent.Name)
			e.scene.Entities = append(e.scene.Entities[:i], e.scene.Entities[i+1:]...)
			if e.selectedID == id { e.selectedID = -1 }
			delete(e.selectedIDs, id)
			e.dirty = true
			return
		}
	}
}

func (e *Editor) DuplicateEntity(id int) {
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

var editorColors = []string{
	"#00e5a0", "#388bfd", "#e3b341", "#f85149",
	"#bc8cff", "#ff7b72", "#79c0ff", "#3fb950",
}

func (e *Editor) randColor() string {
	return editorColors[len(e.scene.Entities)%len(editorColors)]
}

// ---------------------------------------------------------------------------
// Coordinate conversion
// ---------------------------------------------------------------------------

func (e *Editor) viewportRect() (x, y, w, h float64) {
	return e.hierarchyW, e.toolbarH,
		float64(e.screenW()) - e.hierarchyW - e.inspectorW,
		float64(e.screenH()) - e.toolbarH - e.consoleH
}

func (e *Editor) screenW() int { w, _ := ebiten.WindowSize(); return w }
func (e *Editor) screenH() int { _, h := ebiten.WindowSize(); return h }

func (e *Editor) screenToWorld(sx, sy float64) (float64, float64) {
	vpX, vpY, vpW, vpH := e.viewportRect()
	cx := vpX + vpW/2 + e.camX*e.camZoom
	cy := vpY + vpH/2 + e.camY*e.camZoom
	return (sx - cx) / e.camZoom, (sy - cy) / e.camZoom
}

func (e *Editor) worldToScreen(wx, wy float64) (float64, float64) {
	vpX, vpY, vpW, vpH := e.viewportRect()
	cx := vpX + vpW/2 + e.camX*e.camZoom
	cy := vpY + vpH/2 + e.camY*e.camZoom
	return cx + wx*e.camZoom, cy + wy*e.camZoom
}

func (e *Editor) isInViewport(sx, sy float64) bool {
	vpX, vpY, vpW, vpH := e.viewportRect()
	return sx >= vpX && sx < vpX+vpW && sy >= vpY && sy < vpY+vpH
}

// ---------------------------------------------------------------------------
// Save/Load
// ---------------------------------------------------------------------------

func (e *Editor) Save() {
	path := e.filePath
	if path == "" {
		path = e.scene.Meta.Name + ".kora.json"
	}
	data, err := json.MarshalIndent(e.scene, "", "  ")
	if err != nil {
		e.Logf("Error saving: %v", err)
		return
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		e.Logf("Error saving: %v", err)
		return
	}
	e.filePath = path
	e.dirty = false
	e.Logf("Saved: %s (%d entities)", path, len(e.scene.Entities))
}

func (e *Editor) Load(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		e.Logf("Error loading: %v", err)
		return
	}
	var doc SceneFile
	if err := json.Unmarshal(data, &doc); err != nil {
		e.Logf("Error parsing: %v", err)
		return
	}
	e.scene = doc
	e.filePath = path
	e.dirty = false
	e.selectedID = -1
	e.selectedIDs = make(map[int]bool)
	e.nextID = 1
	for _, ent := range doc.Entities {
		if ent.ID >= e.nextID { e.nextID = ent.ID + 1 }
	}
	e.Logf("Loaded: %s (%d entities)", path, len(doc.Entities))
}

// ---------------------------------------------------------------------------
// Hit testing
// ---------------------------------------------------------------------------

func (e *Editor) hitTest(wx, wy float64) int {
	// Sort by ZIndex descending, then by ID descending (later = on top)
	ents := make([]*SceneEntity, len(e.scene.Entities))
	copy(ents, e.scene.Entities)
	sort.SliceStable(ents, func(i, j int) bool {
		if ents[i].ZIndex != ents[j].ZIndex {
			return ents[i].ZIndex > ents[j].ZIndex
		}
		return ents[i].ID > ents[j].ID
	})
	for _, ent := range ents {
		if !ent.Visible { continue }
		if wx >= ent.X-ent.W/2 && wx <= ent.X+ent.W/2 &&
			wy >= ent.Y-ent.H/2 && wy <= ent.Y+ent.H/2 {
			return ent.ID
		}
	}
	return -1
}

// ---------------------------------------------------------------------------
// Ebitengine Game implementation
// ---------------------------------------------------------------------------

var _ ebiten.Game = (*EditorApp)(nil)

type EditorApp struct {
	ed   *Editor
	font *FontRenderer
}

type FontRenderer struct {
	// Simple pixel-based font rendering as fallback
}

func NewFontRenderer() *FontRenderer {
	return &FontRenderer{}
}

// DrawString draws text at screen position using simple pixel rectangles
// as a fallback — will be replaced by render.DefaultFont later
func (f *FontRenderer) DrawString(screen *ebiten.Image, s string, x, y float64, scale float64, clr color.RGBA) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x, y)
	op.ColorScale.SetR(float32(clr.R)/255)
	op.ColorScale.SetG(float32(clr.G)/255)
	op.ColorScale.SetB(float32(clr.B)/255)
	op.ColorScale.SetA(float32(clr.A)/255)

	px := 0.0
	for _, ch := range s {
		if ch == '\n' {
			px = 0
			y += 10 * scale
			continue
		}
		// Draw each character as a small rect (bitmap-style)
		if ch >= 32 && ch <= 126 {
			charW := 6.0 * scale
			charH := 8.0 * scale
			_ = charH
			op2 := *op
			op2.GeoM.SetElement(0, 0, charW)
			op2.GeoM.SetElement(0, 2, x+px)
			op2.GeoM.SetElement(1, 1, charH)
			op2.GeoM.SetElement(1, 2, y)
			screen.DrawImage(whitePixel, &op2)
		}
		px += 6.0 * scale
	}
}

var whitePixel = func() *ebiten.Image {
	img := ebiten.NewImage(1, 1)
	img.Fill(color.White)
	return img
}()

func (app *EditorApp) Update() error {
	e := app.ed

	// --- Keyboard shortcuts (global) ---
	if inpututil.IsKeyJustPressed(ebiten.KeyN) && isCtrlHeld() {
		e.scene = SceneFile{Meta: SceneMeta{Name: "Untitled", Version: 1, LogicalW: 360, LogicalH: 640}}
		e.nextID = 1; e.selectedID = -1; e.dirty = false; e.filePath = ""
		e.Log("New scene")
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) && isCtrlHeld() {
		e.Save()
	}

	// Tool switching
	if inpututil.IsKeyJustPressed(ebiten.Key1) { e.tool = ToolSelect }
	if inpututil.IsKeyJustPressed(ebiten.Key2) { e.tool = ToolMove }
	if inpututil.IsKeyJustPressed(ebiten.Key3) { e.tool = ToolScale }

	// Tab switching via F-keys
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) { e.activeTab = TabScene }
	if inpututil.IsKeyJustPressed(ebiten.KeyF2) { e.activeTab = TabAssets }
	if inpututil.IsKeyJustPressed(ebiten.KeyF3) { e.activeTab = TabCode }
	if inpututil.IsKeyJustPressed(ebiten.KeyF4) { e.activeTab = TabPreview }

	// Delete selected
	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) || inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		if e.selectedID > 0 { e.DeleteEntity(e.selectedID) }
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) && isCtrlHeld() {
		if e.selectedID > 0 { e.DuplicateEntity(e.selectedID) }
	}

	// Grid toggle
	if inpututil.IsKeyJustPressed(ebiten.KeyG) && isCtrlHeld() {
		e.showGrid = !e.showGrid
	}

	// Snap toggle
	if inpututil.IsKeyJustPressed(ebiten.KeyShift) {
		e.snapEnabled = !e.snapEnabled
	}

	// --- Mouse handling ---
	mx, myRaw := ebiten.CursorPosition()
	mxf := float64(mx)
	my := float64(myRaw)

	// Scroll wheel for zoom
	_, dy := ebiten.Wheel()
	if dy != 0 {
		if e.isInViewport(mxf, my) {
			e.camZoom *= 1.0 + dy*0.1
			if e.camZoom < 0.1 { e.camZoom = 0.1 }
			if e.camZoom > 10 { e.camZoom = 10 }
		}
	}

	// Middle mouse for panning
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		if !e.panning {
			e.panning = true
			e.panStartX, e.panStartY = mxf, my
			e.panCamX, e.panCamY = e.camX, e.camY
		}
		e.camX = e.panCamX + (mxf-e.panStartX)/e.camZoom
		e.camY = e.panCamY + (my-e.panStartY)/e.camZoom
	} else {
		e.panning = false
	}

	// Left mouse in viewport
	if e.isInViewport(mxf, my) && !e.panning {
		wx, wy := e.screenToWorld(mxf, my)

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			if !e.dragging {
				hit := e.hitTest(wx, wy)
				if hit > 0 {
					// Ctrl+click for multi-select
					if isCtrlHeld() {
						if e.selectedIDs[hit] {
							delete(e.selectedIDs, hit)
							if e.selectedID == hit {
								e.selectedID = -1
							}
						} else {
							e.selectedIDs[hit] = true
							e.selectedID = hit
						}
					} else {
						if !e.selectedIDs[hit] {
							e.selectedIDs = make(map[int]bool)
							e.selectedID = hit
						}
					}
					e.dragging = true
					e.dragEntityID = hit
					e.dragStartX, e.dragStartY = wx, wy
					ent := e.GetEntity(hit)
					if ent != nil {
						e.dragEntityX, e.dragEntityY = ent.X, ent.Y
					}
					// Store original positions for multi-drag
					e.dragOrigins = make(map[int][2]float64)
					for id := range e.selectedIDs {
						if ent := e.GetEntity(id); ent != nil {
							e.dragOrigins[id] = [2]float64{ent.X, ent.Y}
						}
					}
					if !e.selectedIDs[hit] {
						e.dragOrigins[hit] = [2]float64{e.dragEntityX, e.dragEntityY}
					}
				} else {
					if !isCtrlHeld() {
						e.selectedID = -1
						e.selectedIDs = make(map[int]bool)
					}
				}
			} else {
				// Dragging
				dx := wx - e.dragStartX
				dy := wy - e.dragStartY
				snap := func(v float64) float64 {
					if e.snapEnabled && e.snapSize > 0 {
						return math.Round(v/e.snapSize) * e.snapSize
					}
					return v
				}
				for id, origin := range e.dragOrigins {
					if ent := e.GetEntity(id); ent != nil {
						ent.X = snap(origin[0] + dx)
						ent.Y = snap(origin[1] + dy)
					}
				}
			}
		} else {
			e.dragging = false
			e.dragEntityID = -1
			e.dragOrigins = nil
		}
	} else {
		e.dragging = false
		e.dragEntityID = -1
		e.dragOrigins = nil
	}

	return nil
}

func isCtrlHeld() bool {
	return ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
}

func (app *EditorApp) Draw(screen *ebiten.Image) {
	e := app.ed

	// --- Background ---
	screen.Fill(e.bgDark)

	// --- Toolbar ---
	e.drawToolbar(screen)

	// --- Hierarchy panel ---
	e.drawHierarchy(screen)

	// --- Inspector panel ---
	e.drawInspector(screen)

	// --- Viewport ---
	e.drawViewport(screen)

	// --- Console ---
	e.drawConsole(screen)

	// --- Side borders ---
	_, vpY, vpW, vpH := e.viewportRect()
	sep := color.RGBA{0x21, 0x26, 0x2d, 0xff}
	fillRect(screen, e.hierarchyW, vpY, 1, vpH, sep)
	fillRect(screen, e.hierarchyW+vpW, vpY, 1, vpH, sep)
}

func (e *Editor) drawToolbar(screen *ebiten.Image) {
	w := e.screenW()
	// Background
	fillRect(screen, 0, 0, float64(w), e.toolbarH, e.bgPanel)

	// Logo / Title
	logoX := 8.0
	fillRect(screen, logoX, 8, 20, e.toolbarH-16, e.accent) // accent bar
	drawTextS(screen, "Kora", logoX+28, 10, 1.2, e.accent)
	drawTextS(screen, "Editor", logoX+62, 12, 0.8, e.textMuted)

	// File buttons
	btnY := 6.0
	btnH := e.toolbarH - 12

	buttons := []struct {
		x, w float64
		text string
		key  string
		fn   func()
	}{
		{200, 50, "Novo", "N", func() {
			e.scene = SceneFile{Meta: SceneMeta{Name: "Untitled", Version: 1, LogicalW: 360, LogicalH: 640}}
			e.nextID = 1; e.selectedID = -1; e.dirty = false; e.filePath = ""
			e.Log("New scene")
		}},
		{254, 50, "Salvar", "S", e.Save},
		{308, 50, "Abrir", "O", func() { e.Log("Use: make editor FILE=scene.kora.json") }},
	}
	for _, btn := range buttons {
		fillRect(screen, btn.x, btnY, btn.w, btnH, color.RGBA{0x21, 0x26, 0x2d, 0xff})
		drawTextS(screen, btn.text, btn.x+4, btnY+4, 0.7, e.textPrimary)
		_ = btn.key
	}

	// Tab buttons
	tabs := []struct {
		x    float64
		name string
		tab  EditorTab
	}{
		{400, "Scene", TabScene},
		{450, "Assets", TabAssets},
		{500, "Script", TabCode},
	}
	for _, tab := range tabs {
		col := e.textMuted
		bg := color.RGBA{}
		if e.activeTab == tab.tab {
			col = e.accent
			bg = color.RGBA{0x0d, 0x11, 0x17, 0xaa}
		}
		if bg.A > 0 {
			fillRect(screen, tab.x, btnY, 46, btnH, bg)
		}
		drawTextS(screen, "▶", tab.x+2, btnY+4, 0.6, col)
		drawTextS(screen, tab.name, tab.x+16, btnY+4, 0.7, col)
	}

	// Tool buttons
	toolBtns := []struct {
		x    float64
		name string
		t    Tool
	}{
		{560, "▼", ToolSelect},
		{590, "✚", ToolMove},
		{620, "◧", ToolScale},
	}
	for _, tb := range toolBtns {
		bg := color.RGBA{0x21, 0x26, 0x2d, 0xff}
		if e.tool == tb.t { bg = color.RGBA{0x00, 0xe5, 0xa0, 0x30} }
		fillRect(screen, tb.x, btnY, 28, btnH, bg)
		drawTextS(screen, tb.name, tb.x+6, btnY+4, 0.7, e.textPrimary)
	}

	// Right side info
	info := fmt.Sprintf("%dx%d  Zoom: %.0f%%",
		e.scene.Meta.LogicalW, e.scene.Meta.LogicalH, e.camZoom*100)
	drawTextS(screen, info, float64(w)-150, 12, 0.7, e.textMuted)

	// Dirty indicator
	if e.dirty {
		drawTextS(screen, "●", float64(w)-172, 10, 0.9, e.accent)
	}

	// Bottom border
	fillRect(screen, 0, e.toolbarH-1, float64(w), 1, color.RGBA{0x21, 0x26, 0x2d, 0xff})
}

func (e *Editor) drawHierarchy(screen *ebiten.Image) {
	_, vpY, _, vpH := e.viewportRect()
	// Background
	fillRect(screen, 0, vpY, e.hierarchyW, vpH, e.bgPanel)

	// Header
	drawTextS(screen, "Hierarquia", 8, vpY+6, 0.8, e.textMuted)
	fillRect(screen, 0, vpY+24, e.hierarchyW, 1, color.RGBA{0x21, 0x26, 0x2d, 0xff})

	// Entity list
	y := vpY + 30
	for _, ent := range e.scene.Entities {
		// Background for selected
		if ent.ID == e.selectedID || e.selectedIDs[ent.ID] {
			fillRect(screen, 2, y-1, e.hierarchyW-4, 18, color.RGBA{0x00, 0xe5, 0xa0, 0x15})
		}

		// Icon
		icon := map[string]string{"sprite": "▣", "camera": "◉", "tilemap": "▤", "audio": "♪", "custom": "◇"}
		ic := icon[ent.Type]
		if ic == "" { ic = "⬡" }

		// Name
		col := e.textPrimary
		if ent.ID == e.selectedID { col = e.accent }

		// Visibility toggle
		visIcon := "👁"
		if !ent.Visible { visIcon = "🚫" }

		drawTextS(screen, visIcon, 6, y, 0.7, e.textMuted)
		drawTextS(screen, ic+" "+ent.Name, 24, y, 0.7, col)

		// Z-index badge
		drawTextS(screen, fmt.Sprintf("z%d", ent.ZIndex), e.hierarchyW-40, y, 0.6, e.textFaint)
		y += 18
		if y > vpY+vpH-20 { break }
	}

	// Right border
	fillRect(screen, e.hierarchyW-1, vpY, 1, vpH, color.RGBA{0x21, 0x26, 0x2d, 0xff})
}

func (e *Editor) drawInspector(screen *ebiten.Image) {
	_, vpY, vpW, vpH := e.viewportRect()
	ix := e.hierarchyW + vpW

	// Background
	fillRect(screen, ix, vpY, e.inspectorW, vpH, e.bgPanel)

	// Header
	drawTextS(screen, "Inspetor", ix+8, vpY+6, 0.8, e.textMuted)
	fillRect(screen, ix, vpY+24, e.inspectorW, 1, color.RGBA{0x21, 0x26, 0x2d, 0xff})

	if e.selectedID <= 0 {
		drawTextS(screen, "Nenhuma entidade", ix+8, vpY+36, 0.7, e.textFaint)
		return
	}

	ent := e.GetEntity(e.selectedID)
	if ent == nil { return }

	y := vpY + 32
	propY := func() float64 { y += 18; return y - 18 }

	// Identity section
	drawTextS(screen, "Identidade", ix+8, propY(), 0.7, e.textMuted)
	drawTextS(screen, "Nome: "+ent.Name, ix+12, propY(), 0.7, e.textPrimary)
	drawTextS(screen, "Tipo: "+ent.Type, ix+12, propY(), 0.7, e.textMuted)

	// Transform section
	y += 4
	drawTextS(screen, "Transform", ix+8, propY(), 0.7, e.textMuted)
	drawTextS(screen, fmt.Sprintf("X: %.0f", ent.X), ix+12, propY(), 0.7, e.textPrimary)
	drawTextS(screen, fmt.Sprintf("Y: %.0f", ent.Y), ix+12, propY(), 0.7, e.textPrimary)
	drawTextS(screen, fmt.Sprintf("W: %.0f", ent.W), ix+12, propY(), 0.7, e.textPrimary)
	drawTextS(screen, fmt.Sprintf("H: %.0f", ent.H), ix+12, propY(), 0.7, e.textPrimary)
	drawTextS(screen, fmt.Sprintf("Rot: %.0f°", ent.Rotation), ix+12, propY(), 0.7, e.textPrimary)
	drawTextS(screen, fmt.Sprintf("Z-Index: %d", ent.ZIndex), ix+12, propY(), 0.7, e.textPrimary)

	// Visual section
	y += 4
	drawTextS(screen, "Visual", ix+8, propY(), 0.7, e.textMuted)
	drawTextS(screen, "Visível: "+map[bool]string{true: "Sim", false: "Não"}[ent.Visible], ix+12, propY(), 0.7, e.textPrimary)
	if ent.AssetID != "" {
		drawTextS(screen, "Asset: "+filepath.Base(ent.AssetID), ix+12, propY(), 0.7, e.textMuted)
	}
}

func (e *Editor) drawViewport(screen *ebiten.Image) {
	vpX, vpY, vpW, vpH := e.viewportRect()

	// Background
	fillRect(screen, vpX, vpY, vpW, vpH, e.bgViewport)

	// Grid
	if e.showGrid {
		e.drawGrid(screen)
	}

	// Logical area highlight
	lx, ly := e.worldToScreen(float64(-e.scene.Meta.LogicalW)/2, float64(-e.scene.Meta.LogicalH)/2)
	lw := float64(e.scene.Meta.LogicalW) * e.camZoom
	lh := float64(e.scene.Meta.LogicalH) * e.camZoom
	fillRect(screen, lx, ly, lw, lh, color.RGBA{0, 0, 0, 0x40})
	drawRectBorder(screen, lx, ly, lw, lh, color.RGBA{0x00, 0xe5, 0xa0, 0x30}, 1)

	// Sort entities by ZIndex for drawing
	ents := make([]*SceneEntity, len(e.scene.Entities))
	copy(ents, e.scene.Entities)
	sort.SliceStable(ents, func(i, j int) bool {
		if ents[i].ZIndex != ents[j].ZIndex {
			return ents[i].ZIndex < ents[j].ZIndex
		}
		return ents[i].ID < ents[j].ID
	})

	// Draw entities
	for _, ent := range ents {
		if !ent.Visible { continue }
		e.drawEntitySprite(screen, ent)
	}

	// Selection outline
	if e.selectedID > 0 {
		ent := e.GetEntity(e.selectedID)
		if ent != nil {
			sx, sy := e.worldToScreen(ent.X, ent.Y)
			sw, sh := ent.W*e.camZoom, ent.H*e.camZoom
			drawRectBorder(screen, sx-sw/2-2, sy-sh/2-2, sw+4, sh+4, e.accent, 2)
		}
	}

	// Viewport border
	drawRectBorder(screen, vpX, vpY, vpW, vpH, color.RGBA{0x21, 0x26, 0x2d, 0xff}, 1)
}

func (e *Editor) drawGrid(screen *ebiten.Image) {
	vpX, vpY, vpW, vpH := e.viewportRect()
	gs := e.gridSize * e.camZoom
	if gs < 4 { gs = 4 }

	ox := math.Mod(e.camX*e.camZoom, gs)
	oy := math.Mod(e.camY*e.camZoom, gs)
	if ox < 0 { ox += gs }
	if oy < 0 { oy += gs }

	gridCol := color.RGBA{0x2a, 0x2a, 0x4a, 0x30}

	for gx := vpX + ox; gx < vpX+vpW; gx += gs {
		fillRect(screen, gx, vpY, 1, vpH, gridCol)
	}
	for gy := vpY + oy; gy < vpY+vpH; gy += gs {
		fillRect(screen, vpX, gy, vpW, 1, gridCol)
	}
}

func (e *Editor) drawEntitySprite(screen *ebiten.Image, ent *SceneEntity) {
	sx, sy := e.worldToScreen(ent.X, ent.Y)
	sw := ent.W * e.camZoom
	sh := ent.H * e.camZoom

	// Determine color
	var col color.RGBA
	if ent.Color != "" && len(ent.Color) == 7 {
		col = parseHex(ent.Color)
	} else {
		colors := map[string]color.RGBA{
			"sprite": {0x00, 0xe5, 0xa0, 0xaa},
			"camera": {0xe3, 0xb3, 0x41, 0xaa},
			"tilemap": {0x38, 0x8b, 0xfd, 0xaa},
			"audio": {0xf8, 0x51, 0x49, 0xaa},
			"custom": {0xbc, 0x8c, 0xff, 0xaa},
		}
		col = colors[ent.Type]
		if col == (color.RGBA{}) { col = color.RGBA{0x88, 0x88, 0xcc, 0xaa} }
	}

	// Fill
	fillRect(screen, sx-sw/2, sy-sh/2, sw, sh, col)

	// Border
	drawRectBorder(screen, sx-sw/2, sy-sh/2, sw, sh,
		color.RGBA{col.R, col.G, col.B, 0xff}, 1.5)

	// Name label
	labelScale := 0.7 * e.camZoom
	if labelScale < 0.5 { labelScale = 0.5 }
	if labelScale > 1.2 { labelScale = 1.2 }
	drawTextS(screen, ent.Name, sx-sw/2+2, sy-sh/2-14*labelScale, labelScale, e.textPrimary)
}

func (e *Editor) drawConsole(screen *ebiten.Image) {
	w := e.screenW()
	cy := float64(e.screenH()) - e.consoleH

	// Background
	fillRect(screen, 0, cy, float64(w), e.consoleH, e.bgPanel)

	// Header
	drawTextS(screen, "Console", 8, cy+4, 0.7, e.textMuted)
	fillRect(screen, 0, cy, float64(w), 1, color.RGBA{0x21, 0x26, 0x2d, 0xff})

	// Messages (show last ~3 lines)
	start := len(e.console) - 3
	if start < 0 { start = 0 }
	for i := start; i < len(e.console); i++ {
		lineY := cy + 18 + float64(i-start)*16
		drawTextS(screen, "❯ "+e.console[i], 8, lineY, 0.65, e.textMuted)
	}
}

func (app *EditorApp) Layout(outsideW, outsideH int) (int, int) {
	return outsideW, outsideH
}

// ---------------------------------------------------------------------------
// Drawing helpers
// ---------------------------------------------------------------------------

func fillRect(screen *ebiten.Image, x, y, w, h float64, c color.RGBA) {
	if w <= 0 || h <= 0 { return }
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(w, h)
	op.GeoM.Translate(x, y)
	alpha := float32(c.A) / 255
	if alpha > 0 {
		op.ColorScale.SetR(float32(c.R)/255)
		op.ColorScale.SetG(float32(c.G)/255)
		op.ColorScale.SetB(float32(c.B)/255)
		op.ColorScale.SetA(alpha)
		screen.DrawImage(whitePixel, op)
	}
}

func drawRectBorder(screen *ebiten.Image, x, y, w, h float64, c color.RGBA, bw float64) {
	fillRect(screen, x, y, w, bw, c)       // top
	fillRect(screen, x, y+h-bw, w, bw, c)   // bottom
	fillRect(screen, x, y, bw, h, c)         // left
	fillRect(screen, x+w-bw, y, bw, h, c)    // right
}

func drawTextS(screen *ebiten.Image, text string, x, y float64, scale float64, clr color.RGBA) {
	if text == "" { return }
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x, y)
	op.ColorScale.SetR(float32(clr.R) / 255)
	op.ColorScale.SetG(float32(clr.G) / 255)
	op.ColorScale.SetB(float32(clr.B) / 255)
	op.ColorScale.SetA(float32(clr.A) / 255)

	px := 0.0
	charW := 6.0 * scale
	charH := 8.0 * scale

	for _, ch := range text {
		if ch == '\n' {
			px = 0
			y += charH + 2*scale
			continue
		}
		if ch >= 32 && ch <= 126 {
			op2 := *op
			op2.GeoM.SetElement(0, 0, charW)
			op2.GeoM.SetElement(0, 2, x+px)
			op2.GeoM.SetElement(1, 1, charH)
			op2.GeoM.SetElement(1, 2, y)
			screen.DrawImage(whitePixel, &op2)
		}
		px += charW
	}
}

func parseHex(hex string) color.RGBA {
	if len(hex) == 7 && hex[0] == '#' {
		return color.RGBA{
			hexVal(hex[1])<<4 | hexVal(hex[2]),
			hexVal(hex[3])<<4 | hexVal(hex[4]),
			hexVal(hex[5])<<4 | hexVal(hex[6]),
			0xff,
		}
	}
	return color.RGBA{}
}

func hexVal(c byte) uint8 {
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

	// Load scene from arg if provided
	if len(os.Args) > 1 {
		state.Load(os.Args[1])
	}

	ebiten.SetWindowTitle("Kora Editor — Go Native")
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowDecorated(true)

	app := &EditorApp{ed: state}

	if err := ebiten.RunGame(app); err != nil {
		log.Fatal(err)
	}
}
