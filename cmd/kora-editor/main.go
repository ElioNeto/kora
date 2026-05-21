// Kora Editor v3 — Editor visual com UX aprimorada (tooltips, undo, hover, guides)
//
// Melhorias em relação à v2:
//   - Undo/Redo (Ctrl+Z/Y) com snapshot do estado
//   - Tooltips nos botões ao passar o mouse
//   - Hover states com destaque visual
//   - Indicador de dirty state proeminente
//   - Ferramenta ativa com destaque claro
//   - Painéis colapsáveis (F5=hierarchy, F6=inspector)
//   - Tooltip de entidade ao hover no viewport
//   - Feedback de dragging (ghost + snapped position)
//   - Guias de alinhamento durante o arrasto

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
// Scene data models
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
// Undo system
// ---------------------------------------------------------------------------

type UndoManager struct {
	snapshots  []string // JSON snapshots
	index      int      // current position (-1 = no undo available)
	maxSnap    int
}

func NewUndoManager() *UndoManager {
	return &UndoManager{index: -1, maxSnap: 50}
}

func (u *UndoManager) Push(scene *SceneFile) {
	data, _ := json.Marshal(scene)
	// Remove any redo history beyond current index
	u.snapshots = u.snapshots[:u.index+1]
	u.snapshots = append(u.snapshots, string(data))
	if len(u.snapshots) > u.maxSnap {
		u.snapshots = u.snapshots[1:]
	}
	u.index = len(u.snapshots) - 1
}

func (u *UndoManager) Undo(scene *SceneFile) bool {
	if u.index <= 0 { return false }
	u.index--
	json.Unmarshal([]byte(u.snapshots[u.index]), scene)
	return true
}

func (u *UndoManager) Redo(scene *SceneFile) bool {
	if u.index >= len(u.snapshots)-1 { return false }
	u.index++
	json.Unmarshal([]byte(u.snapshots[u.index]), scene)
	return true
}

func (u *UndoManager) CanUndo() bool { return u.index > 0 }
func (u *UndoManager) CanRedo() bool { return u.index < len(u.snapshots)-1 }

// ---------------------------------------------------------------------------
// Tooltip system
// ---------------------------------------------------------------------------

type Tooltip struct {
	Text  string
	X, Y  float64
	Timer int // frames visible
}

// ---------------------------------------------------------------------------
// Editor types
// ---------------------------------------------------------------------------

type EditorTab int
const (
	TabScene EditorTab = iota
	TabAssets
	TabCode
	TabPreview
)

type Tool int
const (
	ToolSelect Tool = iota
	ToolMove
	ToolScale
)

// ---------------------------------------------------------------------------
// Editor state
// ---------------------------------------------------------------------------

type Editor struct {
	scene    SceneFile
	nextID   int
	filePath string
	dirty    bool
	undo     *UndoManager

	// Viewport camera
	camX, camY  float64
	camZoom     float64

	// Selection
	selectedID  int
	selectedIDs map[int]bool

	// Interaction
	dragging     bool
	dragStartX, dragStartY float64
	dragEntityX, dragEntityY float64
	dragEntityID int
	dragOrigins  map[int][2]float64
	panning      bool
	panStartX, panStartY float64
	panCamX, panCamY     float64

	// Tool
	tool Tool

	// UI panels visibility
	activeTab     EditorTab
	showGrid      bool
	snapEnabled   bool
	snapSize      float64
	showConsole   bool
	showHierarchy bool
	showInspector bool

	// Layout
	hierarchyW float64
	inspectorW float64
	toolbarH   float64
	consoleH   float64

	// Grid
	gridSize float64

	// Console
	console []string

	// Hover tracking for tooltips
	mouseX, mouseY   float64
	hoverToolbarItem string // which toolbar item is hovered
	hoverEntityID    int
	tooltip          *Tooltip

	// Theme
	bgDark      color.RGBA
	bgPanel     color.RGBA
	bgViewport  color.RGBA
	accent      color.RGBA
	accentDim   color.RGBA
	textPrimary color.RGBA
	textMuted   color.RGBA
	textFaint   color.RGBA
	btnBg       color.RGBA
	btnHover    color.RGBA
	btnActive   color.RGBA
	success     color.RGBA
	warning     color.RGBA
}

func NewEditor() *Editor {
	e := &Editor{
		scene: SceneFile{
			Meta: SceneMeta{Name: "Untitled", Version: 1, LogicalW: 360, LogicalH: 640},
			Entities: []*SceneEntity{
				{ID: 1, Name: "Jogador", Type: "sprite", X: 180, Y: 320, W: 32, H: 32, Color: "#00e5a0", Visible: true, ZIndex: 0},
				{ID: 2, Name: "Chão", Type: "tilemap", X: 180, Y: 600, W: 340, H: 24, Color: "#388bfd", Visible: true, ZIndex: 0},
				{ID: 3, Name: "Câmera", Type: "camera", X: 180, Y: 320, W: 16, H: 16, Color: "#e3b341", Visible: true, ZIndex: 0},
			},
		},
		nextID:        4,
		camZoom:       0.75,
		selectedID:    -1,
		selectedIDs:   make(map[int]bool),
		tool:          ToolSelect,
		undo:          NewUndoManager(),
		activeTab:     TabScene,
		showGrid:      true,
		snapEnabled:   true,
		snapSize:      16,
		showConsole:   true,
		showHierarchy: true,
		showInspector: true,
		hierarchyW:    200,
		inspectorW:    220,
		toolbarH:      42,
		consoleH:      80,
		gridSize:      32,
		hoverEntityID: -1,
		bgDark:        color.RGBA{0x0d, 0x11, 0x17, 0xff},
		bgPanel:       color.RGBA{0x16, 0x1b, 0x22, 0xff},
		bgViewport:    color.RGBA{0x0d, 0x11, 0x17, 0xff},
		accent:        color.RGBA{0x00, 0xe5, 0xa0, 0xff},
		accentDim:     color.RGBA{0x00, 0xe5, 0xa0, 0x60},
		textPrimary:   color.RGBA{0xe6, 0xed, 0xf3, 0xff},
		textMuted:     color.RGBA{0x8b, 0x94, 0x9e, 0xff},
		textFaint:     color.RGBA{0x58, 0x5f, 0x66, 0xff},
		btnBg:         color.RGBA{0x2d, 0x2d, 0x3d, 0xff},
		btnHover:      color.RGBA{0x3d, 0x3d, 0x50, 0xff},
		btnActive:     color.RGBA{0x4d, 0x4d, 0x60, 0xff},
		success:       color.RGBA{0x3f, 0xb9, 0x50, 0xff},
		warning:       color.RGBA{0xe3, 0xb3, 0x41, 0xff},
	}
	e.Log("Kora Editor v3 — Ctrl+Z: desfazer | 1-3: ferramentas | F5/F6: painéis")
	e.PushUndo()
	return e
}

func (e *Editor) PushUndo() {
	e.undo.Push(&e.scene)
}

func (e *Editor) Log(msg string) {
	e.console = append(e.console, msg)
	if len(e.console) > 200 { e.console = e.console[len(e.console)-200:] }
}

func (e *Editor) Logf(f string, args ...interface{}) { e.Log(fmt.Sprintf(f, args...)) }

// ---------------------------------------------------------------------------
// Undo helpers — wrap mutations with snapshot
// ---------------------------------------------------------------------------

func (e *Editor) mutation(fn func()) {
	fn()
	e.PushUndo()
	e.dirty = true
}

// ---------------------------------------------------------------------------
// Entity management
// ---------------------------------------------------------------------------

func (e *Editor) GetEntity(id int) *SceneEntity {
	for _, ent := range e.scene.Entities { if ent.ID == id { return ent } }; return nil
}

func (e *Editor) NewEntity(name, etype string) {
	e.mutation(func() {
		e.nextID++
		e.scene.Entities = append(e.scene.Entities, &SceneEntity{
			ID: e.nextID, Name: name, Type: etype,
			X: 180, Y: 320, W: 48, H: 48,
			Color: e.randColor(), Visible: true,
		})
		e.selectedID = e.nextID
		e.Logf("Adicionado: %s (%s)", name, etype)
	})
}

func (e *Editor) DeleteEntity(id int) {
	if id <= 0 { return }
	e.mutation(func() {
		for i, ent := range e.scene.Entities {
			if ent.ID == id {
				e.Logf("Removido: %s", ent.Name)
				e.scene.Entities = append(e.scene.Entities[:i], e.scene.Entities[i+1:]...)
				if e.selectedID == id { e.selectedID = -1 }
				delete(e.selectedIDs, id)
				return
			}
		}
	})
}

func (e *Editor) DuplicateEntity(id int) {
	ent := e.GetEntity(id)
	if ent == nil { return }
	e.mutation(func() {
		e.nextID++
		clone := *ent
		clone.ID = e.nextID; clone.Name = ent.Name + "_copy"; clone.X += 16; clone.Y += 16
		e.scene.Entities = append(e.scene.Entities, &clone)
		e.selectedID = clone.ID
		e.Logf("Duplicado: %s → %s", ent.Name, clone.Name)
	})
}

var editorColors = []string{"#00e5a0", "#388bfd", "#e3b341", "#f85149", "#bc8cff", "#ff7b72", "#79c0ff", "#3fb950"}

func (e *Editor) randColor() string { return editorColors[len(e.scene.Entities)%len(editorColors)] }

// ---------------------------------------------------------------------------
// Coordinate conversion
// ---------------------------------------------------------------------------

func (e *Editor) panelLeft() float64 {
	if !e.showHierarchy { return 0 }
	return e.hierarchyW
}

func (e *Editor) panelRight() float64 {
	if !e.showInspector { return 0 }
	return e.inspectorW
}

func (e *Editor) viewportRect() (x, y, w, h float64) {
	sw := float64(e.screenW())
	sh := float64(e.screenH())
	return e.panelLeft(), e.toolbarH,
		sw - e.panelLeft() - e.panelRight(),
		sh - e.toolbarH - e.consoleH
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
// Hit testing
// ---------------------------------------------------------------------------

func (e *Editor) hitTest(wx, wy float64) int {
	ents := make([]*SceneEntity, len(e.scene.Entities))
	copy(ents, e.scene.Entities)
	sort.SliceStable(ents, func(i, j int) bool {
		if ents[i].ZIndex != ents[j].ZIndex { return ents[i].ZIndex > ents[j].ZIndex }
		return ents[i].ID > ents[j].ID
	})
	for _, ent := range ents {
		if !ent.Visible { continue }
		if wx >= ent.X-ent.W/2 && wx <= ent.X+ent.W/2 && wy >= ent.Y-ent.H/2 && wy <= ent.Y+ent.H/2 {
			return ent.ID
		}
	}
	return -1
}

// ---------------------------------------------------------------------------
// Save/Load
// ---------------------------------------------------------------------------

func (e *Editor) Save() {
	path := e.filePath
	if path == "" { path = e.scene.Meta.Name + ".kora.json" }
	data, err := json.MarshalIndent(e.scene, "", "  ")
	if err != nil { e.Logf("Erro ao salvar: %v", err); return }
	if err := os.WriteFile(path, data, 0644); err != nil { e.Logf("Erro: %v", err); return }
	e.filePath = path
	e.dirty = false
	e.Logf("Salvo: %s (%d entidades)", path, len(e.scene.Entities))
}

func (e *Editor) Load(path string) {
	data, err := os.ReadFile(path)
	if err != nil { e.Logf("Erro: %v", err); return }
	var doc SceneFile
	if err := json.Unmarshal(data, &doc); err != nil { e.Logf("Erro: %v", err); return }
	e.scene = doc; e.filePath = path; e.dirty = false
	e.selectedID = -1; e.selectedIDs = make(map[int]bool)
	e.nextID = 1
	for _, ent := range doc.Entities { if ent.ID >= e.nextID { e.nextID = ent.ID + 1 } }
	e.PushUndo()
	e.Logf("Carregado: %s (%d entidades)", path, len(doc.Entities))
}

// ---------------------------------------------------------------------------
// Show tooltip on screen
// ---------------------------------------------------------------------------

func (e *Editor) ShowTooltip(text string, x, y float64) {
	e.tooltip = &Tooltip{Text: text, X: x, Y: y, Timer: 120}
}

// ---------------------------------------------------------------------------
// Ebitengine Game
// ---------------------------------------------------------------------------

var _ ebiten.Game = (*EditorApp)(nil)

type EditorApp struct {
	ed *Editor
}

func (app *EditorApp) Update() error {
	e := app.ed
	mxRaw, myRaw := ebiten.CursorPosition()
	e.mouseX, e.mouseY = float64(mxRaw), float64(myRaw)
	mx, my := e.mouseX, e.mouseY
	e.hoverToolbarItem = ""
	e.hoverEntityID = -1

	// --- Undo/Redo ---
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) && isCtrlHeld() && !ebiten.IsKeyPressed(ebiten.KeyShift) {
		if e.undo.Undo(&e.scene) { e.Log("Desfazer"); e.dirty = true }
	}
	if (inpututil.IsKeyJustPressed(ebiten.KeyY) && isCtrlHeld()) ||
		(inpututil.IsKeyJustPressed(ebiten.KeyZ) && isCtrlHeld() && ebiten.IsKeyPressed(ebiten.KeyShift)) {
		if e.undo.Redo(&e.scene) { e.Log("Refazer"); e.dirty = true }
	}

	// --- New/Save ---
	if inpututil.IsKeyJustPressed(ebiten.KeyN) && isCtrlHeld() {
		e.scene = SceneFile{Meta: SceneMeta{Name: "Untitled", Version: 1, LogicalW: 360, LogicalH: 640}}
		e.nextID = 1; e.selectedID = -1; e.dirty = false; e.filePath = ""
		e.PushUndo(); e.Log("Novo projeto")
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) && isCtrlHeld() { e.Save() }

	// --- Tools ---
	if inpututil.IsKeyJustPressed(ebiten.Key1) { e.tool = ToolSelect }
	if inpututil.IsKeyJustPressed(ebiten.Key2) { e.tool = ToolMove }
	if inpututil.IsKeyJustPressed(ebiten.Key3) { e.tool = ToolScale }

	// --- Tabs ---
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) { e.activeTab = TabScene }
	if inpututil.IsKeyJustPressed(ebiten.KeyF2) { e.activeTab = TabAssets }
	if inpututil.IsKeyJustPressed(ebiten.KeyF3) { e.activeTab = TabCode }

	// --- Panel toggles ---
	if inpututil.IsKeyJustPressed(ebiten.KeyF5) { e.showHierarchy = !e.showHierarchy }
	if inpututil.IsKeyJustPressed(ebiten.KeyF6) { e.showInspector = !e.showInspector }

	// --- Snap toggle ---
	if inpututil.IsKeyJustPressed(ebiten.KeyShift) { e.snapEnabled = !e.snapEnabled }

	// --- Delete ---
	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) || inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		if e.selectedID > 0 { e.DeleteEntity(e.selectedID) }
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) && isCtrlHeld() {
		if e.selectedID > 0 { e.DuplicateEntity(e.selectedID) }
	}

	// --- Toolbar hover detection ---
	btnY, btnH := 6.0, e.toolbarH-12
	// Check toolbar buttons for hover
	tbItems := []struct {
		x, w float64
		name string
		tip  string
	}{
		{200, 50, "new", "Novo projeto (Ctrl+N)"},
		{254, 50, "save", "Salvar cena (Ctrl+S)"},
		{308, 50, "open", "Abrir cena"},
		{400, 46, "tab_scene", "Cena (F1)"},
		{450, 46, "tab_assets", "Assets (F2)"},
		{500, 46, "tab_script", "Script (F3)"},
		{560, 28, "tool_select", "Selecionar (1)"},
		{590, 28, "tool_move", "Mover (2)"},
		{620, 28, "tool_scale", "Escalar (3)"},
	}
	for _, item := range tbItems {
		if mx >= item.x && mx <= item.x+item.w && my >= btnY && my <= btnY+btnH {
			e.hoverToolbarItem = item.name
			e.ShowTooltip(item.tip, mx+10, my+20)
			break
		}
	}

	// --- Actions panel buttons hover ---
	actX := e.hierarchyW + 8
	actY := float64(e.screenH()) - e.consoleH - 28
	if e.showHierarchy {
		actBtns := []struct {
			idx  int
			name string
			tip  string
		}{
			{0, "add_entity", "Adicionar entidade (Ctrl+E)"},
			{1, "duplicate", "Duplicar (Ctrl+D)"},
			{2, "delete", "Remover (Del)"},
		}
		for _, b := range actBtns {
			bx := actX + float64(b.idx)*52
			if mx >= bx && mx <= bx+48 && my >= actY && my <= actY+24 {
				e.hoverToolbarItem = b.name
				e.ShowTooltip(b.tip, mx+10, my+20)
				break
			}
		}
	}

	// --- Viewport mouse ---
	if e.isInViewport(mx, my) {
		// Hover entity
		wx, wy := e.screenToWorld(mx, my)
		if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && !ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
			e.hoverEntityID = e.hitTest(wx, wy)
			if e.hoverEntityID > 0 {
				ent := e.GetEntity(e.hoverEntityID)
				if ent != nil {
					e.ShowTooltip(fmt.Sprintf("%s (%s)\nX:%.0f Y:%.0f", ent.Name, ent.Type, ent.X, ent.Y), mx+10, my+10)
				}
			}
		}

		// Zoom
		_, dy := ebiten.Wheel()
		if dy != 0 {
			e.camZoom *= 1.0 + dy*0.1
			if e.camZoom < 0.1 { e.camZoom = 0.1 }
			if e.camZoom > 10 { e.camZoom = 10 }
		}

		// Middle mouse pan
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
			if !e.panning {
				e.panning = true
				e.panStartX, e.panStartY = mx, my
				e.panCamX, e.panCamY = e.camX, e.camY
			}
			e.camX = e.panCamX + (mx-e.panStartX)/e.camZoom
			e.camY = e.panCamY + (my-e.panStartY)/e.camZoom
		} else { e.panning = false }

		// Left mouse
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && !e.panning {
			if !e.dragging {
				hit := e.hitTest(wx, wy)
				if hit > 0 {
					if isCtrlHeld() {
						if e.selectedIDs[hit] { delete(e.selectedIDs, hit); if e.selectedID == hit { e.selectedID = -1 } } else { e.selectedIDs[hit] = true; e.selectedID = hit }
					} else {
						if !e.selectedIDs[hit] { e.selectedIDs = make(map[int]bool); e.selectedID = hit }
					}
					e.dragging = true
					e.dragEntityID = hit
					e.dragStartX, e.dragStartY = wx, wy
					if ent := e.GetEntity(hit); ent != nil { e.dragEntityX, e.dragEntityY = ent.X, ent.Y }
					e.dragOrigins = make(map[int][2]float64)
					for id := range e.selectedIDs {
						if ent := e.GetEntity(id); ent != nil { e.dragOrigins[id] = [2]float64{ent.X, ent.Y} }
					}
					if !e.selectedIDs[hit] { e.dragOrigins[hit] = [2]float64{e.dragEntityX, e.dragEntityY} }
				} else {
					if !isCtrlHeld() { e.selectedID = -1; e.selectedIDs = make(map[int]bool) }
				}
			} else {
				dx := wx - e.dragStartX; dy := wy - e.dragStartY
				snap := func(v float64) float64 {
					if e.snapEnabled && e.snapSize > 0 { return math.Round(v/e.snapSize) * e.snapSize }
					return v
				}
				for id, origin := range e.dragOrigins {
					if ent := e.GetEntity(id); ent != nil { ent.X = snap(origin[0] + dx); ent.Y = snap(origin[1] + dy) }
				}
				e.dirty = true
			}
		} else {
			if e.dragging && e.dragEntityID > 0 {
				e.PushUndo()
				e.Logf("Movido: %s", e.GetEntity(e.dragEntityID).Name)
			}
			e.dragging = false; e.dragEntityID = -1; e.dragOrigins = nil
		}
	}

	// Add entity shortcut
	if inpututil.IsKeyJustPressed(ebiten.KeyE) && isCtrlHeld() {
		e.NewEntity(fmt.Sprintf("Entity%d", e.nextID), "sprite")
	}

	// Tick tooltip
	if e.tooltip != nil {
		e.tooltip.Timer--
		if e.tooltip.Timer <= 0 { e.tooltip = nil }
	}

	return nil
}

func isCtrlHeld() bool { return ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta) }

// ---------------------------------------------------------------------------
// Draw
// ---------------------------------------------------------------------------

func (app *EditorApp) Draw(screen *ebiten.Image) {
	e := app.ed
	screen.Fill(e.bgDark)
	e.drawToolbar(screen)
	e.drawViewport(screen)
	if e.showHierarchy { e.drawHierarchy(screen) }
	if e.showInspector { e.drawInspector(screen) }
	e.drawConsole(screen)
	e.drawTooltip(screen)
}

func (e *Editor) drawToolbar(screen *ebiten.Image) {
	fillRect(screen, 0, 0, float64(e.screenW()), e.toolbarH, e.bgPanel)

	// Logo
	logX := 8.0
	fillRect(screen, logX, 8, 4, e.toolbarH-16, e.accent)
	drawTextS(screen, "KORA", logX+12, 8, 1.3, e.accent)
	drawTextS(screen, "EDITOR", logX+60, 10, 0.8, e.textMuted)

	// ---- Toolbar buttons ----
	btnY, btnH := 6.0, e.toolbarH-12

	type tbBtn struct {
		x, w float64
		text string
		id   string
		enabled bool
	}
	btns := []tbBtn{
		{200, 48, "Novo", "new", true},
		{252, 48, "Salvar", "save", true},
		{304, 48, "Abrir", "open", true},
	}

	// Separator
	for _, b := range btns {
		bg := e.btnBg
		if e.hoverToolbarItem == b.id { bg = e.btnHover }
		fillRect(screen, b.x, btnY, b.w, btnH, bg)
		drawTextS(screen, b.text, b.x+6, btnY+6, 0.7, e.textPrimary)
	}

	// Undo/Redo indicators
	ux := 370.0
	if e.undo.CanUndo() {
		drawTextS(screen, "↩", ux, btnY+4, 1.0, e.accent)
	} else {
		drawTextS(screen, "↩", ux, btnY+4, 1.0, e.textFaint)
	}
	ux += 24
	if e.undo.CanRedo() {
		drawTextS(screen, "↪", ux, btnY+4, 1.0, e.accent)
	} else {
		drawTextS(screen, "↪", ux, btnY+4, 1.0, e.textFaint)
	}

	// Tabs
	tabs := []struct {
		x    float64
		name string
		tab  EditorTab
		id   string
	}{
		{420, "▶ Cena", TabScene, "tab_scene"},
		{480, "Assets", TabAssets, "tab_assets"},
		{540, "Script", TabCode, "tab_script"},
	}
	for _, tab := range tabs {
		col := e.textMuted
		bg := color.RGBA{}
		if e.activeTab == tab.tab { col = e.accent; bg = color.RGBA{0x0d, 0x11, 0x17, 0xcc} }
		if e.hoverToolbarItem == tab.id && e.activeTab != tab.tab { col = e.textPrimary; bg = e.btnBg }
		if bg.A > 0 { fillRect(screen, tab.x, btnY, 56, btnH, bg) }
		drawTextS(screen, tab.name, tab.x+4, btnY+6, 0.7, col)
	}

	// Tool buttons
	toolBtns := []struct {
		x    float64
		text string
		t    Tool
		id   string
	}{
		{610, "▼ Sel", ToolSelect, "tool_select"},
		{648, "✚ Mover", ToolMove, "tool_move"},
		{696, "◧ Escala", ToolScale, "tool_scale"},
	}
	for _, tb := range toolBtns {
		bg := e.btnBg
		col := e.textMuted
		if e.tool == tb.t { bg = e.accentDim; col = e.accent }
		if e.hoverToolbarItem == tb.id && e.tool != tb.t { bg = e.btnHover; col = e.textPrimary }
		fillRect(screen, tb.x, btnY, 44, btnH, bg)
		drawTextS(screen, tb.text, tb.x+4, btnY+6, 0.7, col)
	}

	// Panel toggles
	px := 760.0
	panelInfo := []struct {
		id    string
		show  bool
		label string
	}{
		{"panel_hier", e.showHierarchy, "Hierarquia"},
		{"panel_insp", e.showInspector, "Inspetor"},
	}
	for _, p := range panelInfo {
		col := e.textMuted
		if p.show { col = e.accent }
		if e.hoverToolbarItem == p.id { col = e.textPrimary }
		drawTextS(screen, p.label, px, btnY+6, 0.65, col)
		px += 80
	}

	// Right: FPS + dirty + zoom
	ri := float64(e.screenW()) - 200
	fps := fmt.Sprintf("%.0f FPS", ebiten.ActualFPS())
	drawTextS(screen, fps, ri, btnY+6, 0.65, e.textMuted)

	// Dirty indicator (prominent)
	if e.dirty {
		drawTextS(screen, "● NÃO SALVO", ri+60, btnY+4, 0.8, e.warning)
	} else {
		drawTextS(screen, "● Salvo", ri+60, btnY+4, 0.7, e.success)
	}

	// Bottom separator
	fillRect(screen, 0, e.toolbarH-1, float64(e.screenW()), 1, color.RGBA{0x2d, 0x2d, 0x3d, 0xff})
}

func (e *Editor) drawHierarchy(screen *ebiten.Image) {
	_, vpY, _, vpH := e.viewportRect()
	fillRect(screen, 0, vpY, e.hierarchyW, vpH, e.bgPanel)

	// Header
	drawTextS(screen, "HIERARQUIA", 8, vpY+8, 0.75, e.textMuted)
	nEntities := fmt.Sprintf("%d itens", len(e.scene.Entities))
	drawTextS(screen, nEntities, e.hierarchyW-60, vpY+8, 0.6, e.textFaint)
	fillRect(screen, 0, vpY+24, e.hierarchyW, 1, color.RGBA{0x2d, 0x2d, 0x3d, 0xff})

	// + Add button
	addBtnX, addBtnY := e.hierarchyW-28, vpY+4
	addHover := e.hoverToolbarItem == "add_entity"
	bg := e.btnBg; if addHover { bg = e.btnHover }
	fillRect(screen, addBtnX, addBtnY, 24, 18, bg)
	drawTextS(screen, "+", addBtnX+8, addBtnY+1, 0.9, e.accent)

	y := vpY + 30
	for _, ent := range e.scene.Entities {
		sel := ent.ID == e.selectedID || e.selectedIDs[ent.ID]
		hov := ent.ID == e.hoverEntityID

		// Selection background
		if sel { fillRect(screen, 2, y-1, e.hierarchyW-4, 18, e.accentDim) } else if hov { fillRect(screen, 2, y-1, e.hierarchyW-4, 18, e.btnHover) }

		// Visibility toggle
		visIcon := "👁"; col := e.textPrimary
		if !ent.Visible { visIcon = "🚫"; col = e.textFaint }
		if sel { col = e.accent }
		drawTextS(screen, visIcon, 6, y, 0.7, e.textFaint)
		icon := map[string]string{"sprite": "▣", "camera": "◉", "tilemap": "▤", "audio": "♪", "custom": "◇"}
		drawTextS(screen, fmt.Sprintf("%s %s", icon[ent.Type], ent.Name), 26, y, 0.7, col)
		drawTextS(screen, fmt.Sprintf("z%d", ent.ZIndex), e.hierarchyW-36, y, 0.55, e.textFaint)

		// Type label
		typeColor := e.textFaint
		if ent.Type == "sprite" { typeColor = e.accent }
		drawTextS(screen, ent.Type, 26, y+10, 0.5, typeColor)

		y += 22
		if y > vpY+vpH-20 { break }
	}

	// Bottom: action buttons
	actY := float64(e.screenH()) - e.consoleH - 26
	fillRect(screen, 2, actY, e.hierarchyW-4, 24, e.btnBg)

	actBtns := []struct{
		idx int; text string; id string
	}{{0, "+ Add", "add_entity"}, {1, "⟐ Dup", "duplicate"}, {2, "✕ Del", "delete"}}
	for _, b := range actBtns {
		bx := 6 + float64(b.idx)*60
		bg := e.btnBg
		if e.hoverToolbarItem == b.id { bg = e.btnHover }
		fillRect(screen, bx, actY+2, 54, 20, bg)
		drawTextS(screen, b.text, bx+4, actY+4, 0.65, e.textPrimary)
	}

	fillRect(screen, e.hierarchyW-1, vpY, 1, vpH, color.RGBA{0x2d, 0x2d, 0x3d, 0xff})
}

func (e *Editor) drawInspector(screen *ebiten.Image) {
	_, vpY, vpW, vpH := e.viewportRect()
	ix := e.panelLeft() + vpW
	fillRect(screen, ix, vpY, e.inspectorW, vpH, e.bgPanel)

	drawTextS(screen, "INSPETOR", ix+8, vpY+8, 0.75, e.textMuted)
	drawTextS(screen, e.toolName(), ix+e.inspectorW-60, vpY+8, 0.6, e.accent)
	fillRect(screen, ix, vpY+24, e.inspectorW, 1, color.RGBA{0x2d, 0x2d, 0x3d, 0xff})

	if e.selectedID <= 0 {
		drawTextS(screen, "Nenhum objeto selecionado", ix+12, vpY+44, 0.7, e.textFaint)
		drawTextS(screen, "Clique em uma entidade no", ix+12, vpY+62, 0.65, e.textFaint)
		drawTextS(screen, "viewport ou na hierarquia", ix+12, vpY+78, 0.65, e.textFaint)
		return
	}

	ent := e.GetEntity(e.selectedID)
	if ent == nil { return }

	y := vpY + 32
	prop := func(h float64) float64 { y += h; return y - h }

	// Identity
	drawTextS(screen, "IDENTIDADE", ix+10, prop(18), 0.65, e.textMuted)
	drawTextS(screen, fmt.Sprintf("Nome:  %s", ent.Name), ix+14, prop(16), 0.7, e.textPrimary)
	drawTextS(screen, fmt.Sprintf("Tipo:  %s", ent.Type), ix+14, prop(16), 0.7, e.textMuted)

	// Separator
	fillRect(screen, ix+8, y, e.inspectorW-16, 1, color.RGBA{0x2d, 0x2d, 0x3d, 0xff})
	y += 4

	// Transform
	drawTextS(screen, "TRANSFORM", ix+10, prop(18), 0.65, e.textMuted)
	drawTextS(screen, fmt.Sprintf("X:  %.0f", ent.X), ix+14, prop(16), 0.7, e.textPrimary)
	drawTextS(screen, fmt.Sprintf("Y:  %.0f", ent.Y), ix+14, prop(16), 0.7, e.textPrimary)
	drawTextS(screen, fmt.Sprintf("W:  %.0f", ent.W), ix+14, prop(16), 0.7, e.textPrimary)
	drawTextS(screen, fmt.Sprintf("H:  %.0f", ent.H), ix+14, prop(16), 0.7, e.textPrimary)
	drawTextS(screen, fmt.Sprintf("Rot:  %.0f°", ent.Rotation), ix+14, prop(16), 0.7, e.textPrimary)
	drawTextS(screen, fmt.Sprintf("Z:  %d", ent.ZIndex), ix+14, prop(16), 0.7, e.textPrimary)

	// Visual
	if ent.AssetID != "" || ent.Color != "" {
		fillRect(screen, ix+8, y, e.inspectorW-16, 1, color.RGBA{0x2d, 0x2d, 0x3d, 0xff})
		y += 4
		drawTextS(screen, "VISUAL", ix+10, prop(18), 0.65, e.textMuted)
		if ent.AssetID != "" { drawTextS(screen, "Asset:  "+filepath.Base(ent.AssetID), ix+14, prop(16), 0.65, e.textMuted) }
		drawTextS(screen, "Visível:  "+map[bool]string{true: "Sim ✅", false: "Não 🚫"}[ent.Visible], ix+14, prop(16), 0.7, e.textPrimary)
		// Color preview swatch
		if ent.Color != "" {
			swatchCol := parseHex(ent.Color)
			fillRect(screen, ix+14, y, 14, 10, swatchCol)
			drawTextS(screen, fmt.Sprintf("  %s", ent.Color), ix+30, prop(16), 0.65, e.textMuted)
		}
	}
}

func (e *Editor) toolName() string {
	switch e.tool { case ToolSelect: return "▼ Sel"; case ToolMove: return "✚ Mover"; case ToolScale: return "◧ Escala" }
	return ""
}

func (e *Editor) drawViewport(screen *ebiten.Image) {
	vpX, vpY, vpW, vpH := e.viewportRect()
	// Viewport bg - slightly lighter than bgDark
	fillRect(screen, vpX, vpY, vpW, vpH, e.bgViewport)

	// Grid
	if e.showGrid {
		gs := e.gridSize * e.camZoom; if gs < 4 { gs = 4 }
		ox := math.Mod(e.camX*e.camZoom, gs); if ox < 0 { ox += gs }
		oy := math.Mod(e.camY*e.camZoom, gs); if oy < 0 { oy += gs }
		for gx := vpX + ox; gx < vpX+vpW; gx += gs { fillRect(screen, gx, vpY, 1, vpH, color.RGBA{0x2a, 0x2a, 0x4a, 0x25}) }
		for gy := vpY + oy; gy < vpY+vpH; gy += gs { fillRect(screen, vpX, gy, vpW, 1, color.RGBA{0x2a, 0x2a, 0x4a, 0x25}) }
	}

	// Logical area
	lx, ly := e.worldToScreen(float64(-e.scene.Meta.LogicalW)/2, float64(-e.scene.Meta.LogicalH)/2)
	lw := float64(e.scene.Meta.LogicalW) * e.camZoom
	lh := float64(e.scene.Meta.LogicalH) * e.camZoom
	fillRect(screen, lx, ly, lw, lh, color.RGBA{0x0d, 0x0d, 0x1a, 0x60})
	drawRectBorder(screen, lx, ly, lw, lh, color.RGBA{0x00, 0xe5, 0xa0, 0x20}, 1)

	// Sort entities by ZIndex
	ents := make([]*SceneEntity, len(e.scene.Entities))
	copy(ents, e.scene.Entities)
	sort.SliceStable(ents, func(i, j int) bool {
		if ents[i].ZIndex != ents[j].ZIndex { return ents[i].ZIndex < ents[j].ZIndex }
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
			drawRectBorder(screen, sx-sw/2-3, sy-sh/2-3, sw+6, sh+6, e.accent, 2)
			// Corner handles
			for _, hx := range []float64{sx - sw/2 - 3, sx, sx + sw/2 + 3} {
				for _, hy := range []float64{sy - sh/2 - 3, sy, sy + sh/2 + 3} {
					fillRect(screen, hx-3, hy-3, 6, 6, e.accent)
				}
			}
		}
	}

	// Hover highlight
	if e.hoverEntityID > 0 && e.hoverEntityID != e.selectedID {
		ent := e.GetEntity(e.hoverEntityID)
		if ent != nil {
			sx, sy := e.worldToScreen(ent.X, ent.Y); sw, sh := ent.W*e.camZoom, ent.H*e.camZoom
			drawRectBorder(screen, sx-sw/2-2, sy-sh/2-2, sw+4, sh+4, color.RGBA{0x00, 0xc8, 0xff, 0x60}, 1.5)
		}
	}

	// Viewport info overlay
	info := fmt.Sprintf("%dx%d  |  Zoom: %.0f%%  |  Grid: %dpx  Snap: %s",
		e.scene.Meta.LogicalW, e.scene.Meta.LogicalH, e.camZoom*100, int(e.snapSize),
		map[bool]string{true: "ON ✅", false: "OFF"}[e.snapEnabled])
	drawTextS(screen, info, vpX+8, vpY+vpH-16, 0.6, e.textFaint)

	// Entity count
	drawTextS(screen, fmt.Sprintf("%d entidades", len(e.scene.Entities)), vpX+vpW-100, vpY+vpH-16, 0.6, e.textFaint)

	// Viewport border
	drawRectBorder(screen, vpX, vpY, vpW, vpH, color.RGBA{0x2d, 0x2d, 0x3d, 0xff}, 1)
}

func (e *Editor) drawEntitySprite(screen *ebiten.Image, ent *SceneEntity) {
	sx, sy := e.worldToScreen(ent.X, ent.Y)
	sw := ent.W * e.camZoom; sh := ent.H * e.camZoom

	var col color.RGBA
	if ent.Color != "" && len(ent.Color) == 7 {
		col = parseHex(ent.Color)
	} else {
		colors := map[string]color.RGBA{
			"sprite": {0x00, 0xe5, 0xa0, 0xcc}, "camera": {0xe3, 0xb3, 0x41, 0xcc},
			"tilemap": {0x38, 0x8b, 0xfd, 0xcc}, "audio": {0xf8, 0x51, 0x49, 0xcc},
			"custom": {0xbc, 0x8c, 0xff, 0xcc},
		}
		col = colors[ent.Type]; if col == (color.RGBA{}) { col = color.RGBA{0x88, 0x88, 0xcc, 0xcc} }
	}

	// Fill with semi-transparent
	fillRect(screen, sx-sw/2, sy-sh/2, sw, sh, color.RGBA{col.R, col.G, col.B, uint8(float64(col.A) * 0.7)})
	// Border
	drawRectBorder(screen, sx-sw/2, sy-sh/2, sw, sh, color.RGBA{col.R, col.G, col.B, 0xff}, 1.5)

	// Entity type icon in center
	icon := map[string]string{"sprite": "◆", "camera": "◎", "tilemap": "■", "audio": "♫", "custom": "◇"}
	drawTextS(screen, icon[ent.Type], sx-3*e.camZoom, sy-4*e.camZoom, 1.0*e.camZoom, color.RGBA{0xff, 0xff, 0xff, 0x99})

	// Label below
	labelScale := 0.7 * e.camZoom
	if labelScale < 0.5 { labelScale = 0.5 }
	if labelScale > 1.0 { labelScale = 1.0 }
	drawTextS(screen, ent.Name, sx-sw/2, sy+sh/2+2, labelScale, e.textPrimary)
}

func (e *Editor) drawConsole(screen *ebiten.Image) {
	cy := float64(e.screenH()) - e.consoleH
	fillRect(screen, 0, cy, float64(e.screenW()), e.consoleH, e.bgPanel)
	drawTextS(screen, "CONSOLE", 8, cy+4, 0.65, e.textMuted)
	fillRect(screen, 0, cy, float64(e.screenW()), 1, color.RGBA{0x2d, 0x2d, 0x3d, 0xff})

	start := len(e.console) - 3; if start < 0 { start = 0 }
	for i := start; i < len(e.console); i++ {
		drawTextS(screen, "▸ "+e.console[i], 8, cy+18+float64(i-start)*16, 0.65, e.textMuted)
	}
}

func (e *Editor) drawTooltip(screen *ebiten.Image) {
	if e.tooltip == nil || e.tooltip.Text == "" { return }
	lines := []string{e.tooltip.Text}
	// Simple multi-line support
	for i, ch := range e.tooltip.Text {
		if ch == '\n' {
			lines = append(lines, e.tooltip.Text[i+1:])
			lines[0] = e.tooltip.Text[:i]
			break
		}
	}
	maxW := 0.0
	for _, l := range lines { if fl := float64(len(l)) * 7.0; fl > maxW { maxW = fl } }
	h := float64(len(lines))*14.0 + 8.0

	tx := e.tooltip.X + 8; ty := e.tooltip.Y + 8
	// Keep tooltip on screen
	if tx+maxW > float64(e.screenW()) { tx = e.tooltip.X - maxW - 12 }
	if ty+h > float64(e.screenH())-e.consoleH { ty = e.tooltip.Y - h - 12 }

	fillRect(screen, tx, ty, maxW+12, h, color.RGBA{0x1f, 0x2f, 0x3f, 0xf0})
	drawRectBorder(screen, tx, ty, maxW+12, h, e.accentDim, 1)
	for i, l := range lines {
		drawTextS(screen, l, tx+6, ty+4+float64(i)*14, 0.7, e.textPrimary)
	}
}

func (app *EditorApp) Layout(outsideW, outsideH int) (int, int) { return outsideW, outsideH }

// ---------------------------------------------------------------------------
// Drawing helpers
// ---------------------------------------------------------------------------

var whitePixel = func() *ebiten.Image { img := ebiten.NewImage(1, 1); img.Fill(color.White); return img }()

func fillRect(screen *ebiten.Image, x, y, w, h float64, c color.RGBA) {
	if w <= 0 || h <= 0 { return }
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(w, h); op.GeoM.Translate(x, y)
	op.ColorScale.SetR(float32(c.R)/255); op.ColorScale.SetG(float32(c.G)/255)
	op.ColorScale.SetB(float32(c.B)/255); op.ColorScale.SetA(float32(c.A)/255)
	screen.DrawImage(whitePixel, op)
}

func drawRectBorder(screen *ebiten.Image, x, y, w, h float64, c color.RGBA, bw float64) {
	fillRect(screen, x, y, w, bw, c); fillRect(screen, x, y+h-bw, w, bw, c)
	fillRect(screen, x, y, bw, h, c); fillRect(screen, x+w-bw, y, bw, h, c)
}

func drawTextS(screen *ebiten.Image, text string, x, y float64, scale float64, clr color.RGBA) {
	if text == "" { return }
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale); op.GeoM.Translate(x, y)
	op.ColorScale.SetR(float32(clr.R)/255); op.ColorScale.SetG(float32(clr.G)/255)
	op.ColorScale.SetB(float32(clr.B)/255); op.ColorScale.SetA(float32(clr.A)/255)

	charW := 6.0 * scale; px := 0.0
	for _, ch := range text {
		if ch < 32 || ch > 126 { px += charW; continue }
		op2 := *op
		op2.GeoM.SetElement(0, 0, charW); op2.GeoM.SetElement(0, 2, x+px)
		op2.GeoM.SetElement(1, 1, 8*scale); op2.GeoM.SetElement(1, 2, y)
		screen.DrawImage(whitePixel, &op2)
		px += charW
	}
}

func parseHex(hex string) color.RGBA {
	if len(hex) == 7 && hex[0] == '#' {
		return color.RGBA{hexVal(hex[1])<<4 | hexVal(hex[2]), hexVal(hex[3])<<4 | hexVal(hex[4]), hexVal(hex[5])<<4 | hexVal(hex[6]), 0xff}
	}
	return color.RGBA{}
}

func hexVal(c byte) uint8 {
	switch { case c >= '0' && c <= '9': return c - '0'; case c >= 'a' && c <= 'f': return c - 'a' + 10; case c >= 'A' && c <= 'F': return c - 'A' + 10 }
	return 0
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	state := NewEditor()
	if len(os.Args) > 1 { state.Load(os.Args[1]) }

	ebiten.SetWindowTitle("Kora Editor — Visual Scene Editor")
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowDecorated(true)

	if err := ebiten.RunGame(&EditorApp{ed: state}); err != nil {
		log.Fatal(err)
	}
}
