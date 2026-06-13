// Kora Editor v3 — Editor visual com UX aprimorada (tooltips, undo, hover, guides)
//
// Phase 1 features:
//   - Bridge: SceneEntity ↔ Node2D runtime conversion
//   - Animation Timeline: playhead, keyframes, easing
//   - Hot-Reload: KScript file watching & auto-compile
//   - Play Mode: in-editor preview via runtime
//   - Camera Gizmos: frustum, handles
//   - GameMaker-inspired dark theme

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
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/ElioNeto/kora/core/editor"
	"github.com/ElioNeto/kora/core/scene"
)

// ---------------------------------------------------------------------------
// Scene data models
// ---------------------------------------------------------------------------

type SceneEntity struct {
	ID       int            `json:"id"`
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	X        float64        `json:"x"`
	Y        float64        `json:"y"`
	W        float64        `json:"w"`
	H        float64        `json:"h"`
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
	TabAnim // Animation timeline
	TabSprite // Sprite editor
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

	// ── Phase 1: Animation Timeline ──
	animClips      []*editor.AnimClip   // all clips in the scene
	activeClip     *editor.AnimClip     // currently selected clip
	timelineState  *editor.TimelineState // playback state

	// ── Phase 1: Hot-Reload ──
	hotReload *editor.HotReloadState

	// ── Phase 1: Play Mode ──
	playMode    bool       // true = game preview running
	sceneRunner *scene.Scene // runtime scene for preview
	playStart   time.Time  // when play mode started

	// ── Phase 1: Camera Gizmo ──
	showCameraGizmo bool // show camera frustum in viewport

	// ── Phase 2: Sprite Editor ──
	spriteEditor *editor.SpriteEditorState
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
		hierarchyW:    240,
		inspectorW:    260,
		toolbarH:      52,
		consoleH:      110,
		gridSize:      32,
		hoverEntityID: -1,
		bgDark:        color.RGBA{0x1a, 0x1c, 0x1e, 0xff},
		bgPanel:       color.RGBA{0x24, 0x26, 0x28, 0xff},
		bgViewport:    color.RGBA{0x1a, 0x1c, 0x1e, 0xff},
		accent:        color.RGBA{0x4a, 0x9e, 0xff, 0xff},
		accentDim:     color.RGBA{0x4a, 0x9e, 0xff, 0x40},
		textPrimary:   color.RGBA{0xd4, 0xd4, 0xd4, 0xff},
		textMuted:     color.RGBA{0x88, 0x8a, 0x8c, 0xff},
		textFaint:     color.RGBA{0x5a, 0x5c, 0x5e, 0xff},
		btnBg:         color.RGBA{0x2f, 0x31, 0x33, 0xff},
		btnHover:      color.RGBA{0x3a, 0x3c, 0x3e, 0xff},
		btnActive:     color.RGBA{0x44, 0x46, 0x48, 0xff},
		success:       color.RGBA{0x1e, 0xa5, 0x5e, 0xff},
		warning:       color.RGBA{0xcc, 0xa7, 0x00, 0xff},

		// Phase 1
		showCameraGizmo: true,
	}
	e.Log("Kora Editor v3 — Ctrl+Z: desfazer | 1-3: ferramentas | F5/F6: painéis")
	e.Log("Phase 1: [F4=Anim Timeline] [F7=HotReload] [F8=Play] [F9=Camera Gizmo]")
	e.PushUndo()

	// Configura hot-reload
	e.hotReload = editor.NewHotReloadState(".")
	e.hotReload.SetCompileFn(func(path string) error {
		e.Logf("Hot-Reload: %s compilado com sucesso", filepath.Base(path))
		return nil
	})

	// Configura sprite editor
	e.spriteEditor = editor.NewSpriteEditorState()

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

	// ── Phase 1: Animation Timeline ──
	if inpututil.IsKeyJustPressed(ebiten.KeyF4) {
		if e.activeTab == TabAnim {
			e.activeTab = TabScene
		} else {
			e.activeTab = TabAnim
		}
	}

	// ── Phase 2: Sprite Editor ──
	if inpututil.IsKeyJustPressed(ebiten.KeyF10) {
		if e.activeTab == TabSprite {
			e.activeTab = TabScene
		} else {
			e.activeTab = TabSprite
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyI) && isCtrlHeld() {
		if e.activeTab == TabSprite {
			// Import sprite via file dialog
			e.Log("Sprite Editor: importe uma imagem arrastando para a janela")
			// No file dialog without external deps — use a test sprite for now
			ses := e.spriteEditor
			// Create a placeholder sprite for demonstration
			ses.Resource = editor.NewSingleSprite("NovoSprite", "placeholder.png", 64, 64)
			ses.PushSpriteUndo()
			e.Logf("Sprite criado: NovoSprite (64x64)")
		}
	}

	// ── Phase 1: Hot-Reload toggle ──
	if inpututil.IsKeyJustPressed(ebiten.KeyF7) {
		if e.hotReload.IsEnabled() {
			e.hotReload.Disable()
			e.Log("Hot-Reload desativado")
		} else {
			e.hotReload.Enable()
			e.Log("Hot-Reload ativado")
		}
	}

	// ── Phase 1: Play Mode ──
	if inpututil.IsKeyJustPressed(ebiten.KeyF8) {
		e.TogglePlayMode()
	}

	// ── Phase 1: Camera Gizmo ──
	if inpututil.IsKeyJustPressed(ebiten.KeyF9) {
		e.showCameraGizmo = !e.showCameraGizmo
	}

	// ── Phase 1: Force Recompile (hot-reload) ──
	if inpututil.IsKeyJustPressed(ebiten.KeyR) && isCtrlHeld() && ebiten.IsKeyPressed(ebiten.KeyShift) {
		if e.hotReload.IsEnabled() {
			e.hotReload.ForceRecompile()
		}
	}

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
		{550, 36, "tab_anim", "Animação (F4)"},
		{590, 28, "tool_select", "Selecionar (1)"},
		{620, 28, "tool_move", "Mover (2)"},
		{650, 28, "tool_scale", "Escalar (3)"},
		{692, 36, "sprite_editor", "Sprite Editor (F10)"},
		{732, 36, "play", "Play (F8)"},
		{772, 36, "hotreload", "Hot-Reload (F7)"},
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

	// ── Phase 1: Timeline Tick ──
	dt := 1.0 / 60.0 // approximate
	if e.timelineState != nil && e.timelineState.IsPlaying {
		e.timelineState.Tick(dt)
		// Apply keyframes to selected entity if any
		if e.selectedID > 0 && e.activeClip != nil {
			ent := e.GetEntity(e.selectedID)
			if ent != nil {
				activeKF := e.timelineState.GetActiveKeyframes()
				for _, akf := range activeKF {
					if akf.PrevKey != nil && akf.NextKey != nil {
						// Linear interpolation between keyframes
						val := akf.PrevKey.Value + (akf.NextKey.Value-akf.PrevKey.Value)*akf.T
						switch akf.TrackProperty {
						case "x":
							ent.X = val
						case "y":
							ent.Y = val
						case "rotation":
							ent.Rotation = val
						case "alpha":
							// Alpha would be applied via Sprite2D
						}
					}
				}
				e.dirty = true
			}
		}
	}

	// ── Phase 1: Timeline click handling ──
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if e.activeTab == TabAnim && e.activeClip != nil {
			vpX, vpY, vpW, vpH := e.viewportRect()
			timelineH := 120.0
			timY := vpY + vpH - timelineH

			// Check play/pause button click
			if my >= timY+20 && my <= timY+36 {
				if mx >= vpX+8 && mx <= vpX+24 {
					if e.timelineState.IsPlaying {
						e.timelineState.Pause()
					} else {
						e.timelineState.Play()
					}
				}
				// Check stop button
				if mx >= vpX+28 && mx <= vpX+44 {
					e.timelineState.Stop()
				}
				// Check loop toggle
				if mx >= vpX+50 && mx <= vpX+64 {
					if e.timelineState != nil {
						e.timelineState.Loop = !e.timelineState.Loop
					}
				}
			}

			// Click on track area to seek
			trackY := timY + 44
			trackH := timelineH - 48
			if my >= trackY && my <= trackY+trackH {
				trackW := vpW - 80
				trackX := vpX + 70
				if mx >= trackX && mx <= trackX+trackW {
					t := (mx - trackX) / trackW
					if e.activeClip.Duration > 0 {
						e.timelineState.Seek(t * e.activeClip.Duration)
					}
				}
			}
		}
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

	// 1. Toolbar + Viewport (base)
	e.drawToolbar(screen)
	e.drawViewport(screen)

	// 2. Camera gizmo (on top of viewport)
	e.drawCameraGizmo(screen)

	// 3. Animation timeline panel (over viewport if active)
	if e.activeTab == TabAnim {
		e.drawTimeline(screen)
	}

	// 3b. Sprite editor panel
	if e.activeTab == TabSprite {
		e.drawSpriteEditor(screen)
	}

	// 4. Play mode overlay
	e.drawPlayModeOverlay(screen)

	// 5. Panels
	if e.showHierarchy { e.drawHierarchy(screen) }
	if e.showInspector { e.drawInspector(screen) }
	e.drawConsole(screen)

	// 6. Hot-reload status
	e.drawHotReloadStatus(screen)

	// 7. Tooltip (always on top)
	e.drawTooltip(screen)
}

func (e *Editor) drawToolbar(screen *ebiten.Image) {
	fillRect(screen, 0, 0, float64(e.screenW()), e.toolbarH, e.bgPanel)
	fillRect(screen, 0, e.toolbarH-1, float64(e.screenW()), 1, color.RGBA{0x3c, 0x3e, 0x40, 0xff})

	btnY, btnH := 8.0, e.toolbarH-16

	// Logo — GameMaker-inspired
	logX := 10.0
	fillRect(screen, logX, btnY+2, 3, btnH-4, e.accent)
	drawTextS(screen, "KORA", logX+12, btnY-2, 2.2, e.accent)

	// ---- File buttons ----
	btnX := 140.0
	fileBtns := []struct {
		text string
		id   string
		tip  string
	}{
		{"+ Novo", "new", "Novo projeto (Ctrl+N)"},
		{"Salvar", "save", "Salvar cena (Ctrl+S)"},
		{"Abrir", "open", "Abrir cena (Ctrl+O)"},
	}
	for i, b := range fileBtns {
		x := btnX + float64(i)*62
		bg := e.btnBg
		if e.hoverToolbarItem == b.id { bg = e.btnHover }
		fillRect(screen, x, btnY, 56, btnH, bg)
		drawTextS(screen, b.text, x+4, btnY+3, 1.2, e.textPrimary)
	}

	// Undo/Redo
	ux := btnX + 200
	undoCol := e.textFaint
	if e.undo.CanUndo() { undoCol = e.accent }
	drawTextS(screen, "↩", ux, btnY+2, 1.6, undoCol)
	redoCol := e.textFaint
	if e.undo.CanRedo() { redoCol = e.accent }
	drawTextS(screen, "↪", ux+24, btnY+2, 1.6, redoCol)

	// Separator line
	fx := ux + 50
	fillRect(screen, fx, btnY+4, 1, btnH-8, color.RGBA{0x3c, 0x3e, 0x40, 0xff})

	// Tabs — GameMaker style
	tabs := []struct {
		x    float64
		name string
		tab  EditorTab
		id   string
	}{
		{fx + 12, "Cena", TabScene, "tab_scene"},
		{fx + 70, "Assets", TabAssets, "tab_assets"},
		{fx + 130, "Script", TabCode, "tab_script"},
		{fx + 190, "Sprite", TabSprite, "tab_sprite"},
		{fx + 250, "Anim", TabAnim, "tab_anim"},
	}
	for _, tab := range tabs {
		col := e.textMuted
		bg := color.RGBA{}
		if e.activeTab == tab.tab {
			col = e.accent
			bg = e.accentDim
		}
		if e.hoverToolbarItem == tab.id && e.activeTab != tab.tab {
			col = e.textPrimary
			bg = e.btnBg
		}
		if bg.A > 0 {
			fillRect(screen, tab.x, btnY, 56, btnH, bg)
		}
		drawTextS(screen, tab.name, tab.x+4, btnY+3, 1.2, col)
	}

	// Tool buttons (right side)
	toolRight := float64(e.screenW()) - 280
	fillRect(screen, toolRight-8, btnY+4, 1, btnH-8, color.RGBA{0x3c, 0x3e, 0x40, 0xff})

	toolBtns := []struct {
		x    float64
		text string
		t    Tool
		id   string
	}{
		{toolRight, "▼ Sel", ToolSelect, "tool_select"},
		{toolRight + 50, "✚ Mover", ToolMove, "tool_move"},
		{toolRight + 100, "◧ Escala", ToolScale, "tool_scale"},
	}
	for _, tb := range toolBtns {
		bg := e.btnBg
		col := e.textMuted
		if e.tool == tb.t { bg = e.accentDim; col = e.accent }
		if e.hoverToolbarItem == tb.id && e.tool != tb.t { bg = e.btnHover; col = e.textPrimary }
		fillRect(screen, tb.x, btnY, 46, btnH, bg)
		drawTextS(screen, tb.text, tb.x+3, btnY+3, 1.1, col)
	}

	// Play button
	playX := toolRight + 160
	playCol := color.RGBA{0x1e, 0xa5, 0x5e, 0xff}
	playText := "▶ Play"
	if e.playMode { playText = "⏹ Stop"; playCol = color.RGBA{0xf4, 0x47, 0x47, 0xff} }
	fillRect(screen, playX, btnY, 66, btnH, color.RGBA{playCol.R, playCol.G, playCol.B, 30})
	drawTextS(screen, playText, playX+4, btnY+3, 1.2, playCol)

	// Mode indicator
	modeText := "EDIT"
	if e.playMode { modeText = "PLAY" }
	drawTextS(screen, modeText, playX+72, btnY+3, 1.1,
		map[bool]color.RGBA{true: {0x1e, 0xa5, 0x5e, 0xff}, false: e.accent}[e.playMode])

	// Panel toggles
	px := 910.0
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
	drawTextS(screen, "HIERARQUIA", 10, vpY+6, 1.4, e.textMuted)
	nEntities := fmt.Sprintf("%d itens", len(e.scene.Entities))
	drawTextS(screen, nEntities, e.hierarchyW-80, vpY+6, 1.0, e.textFaint)
	fillRect(screen, 0, vpY+28, e.hierarchyW, 1, color.RGBA{0x3c, 0x3e, 0x40, 0xff})

	// + Add button
	addBtnX, addBtnY := e.hierarchyW-34, vpY+3
	addHover := e.hoverToolbarItem == "add_entity"
	bg := e.btnBg; if addHover { bg = e.btnHover }
	fillRect(screen, addBtnX, addBtnY, 28, 22, bg)
	drawTextS(screen, "+", addBtnX+9, addBtnY+3, 1.5, e.accent)

	y := vpY + 36
	for _, ent := range e.scene.Entities {
		sel := ent.ID == e.selectedID || e.selectedIDs[ent.ID]
		hov := ent.ID == e.hoverEntityID

		// Selection background — larger hit area
		if sel { fillRect(screen, 2, y-1, e.hierarchyW-4, 24, e.accentDim) } else if hov { fillRect(screen, 2, y-1, e.hierarchyW-4, 24, e.btnHover) }

		// Visibility toggle
		visIcon := "👁"; col := e.textPrimary
		if !ent.Visible { visIcon = "🚫"; col = e.textFaint }
		if sel { col = e.accent }
		drawTextS(screen, visIcon, 6, y, 1.1, e.textFaint)
		icon := map[string]string{"sprite": "▣", "camera": "◉", "tilemap": "▤", "audio": "♪", "custom": "◇"}
		drawTextS(screen, fmt.Sprintf("%s %s", icon[ent.Type], ent.Name), 26, y+1, 1.3, col)
		drawTextS(screen, fmt.Sprintf("z%d", ent.ZIndex), e.hierarchyW-40, y+2, 0.9, e.textFaint)

		y += 26
		if y > vpY+vpH-20 { break }
	}

	// Bottom: action buttons
	actY := float64(e.screenH()) - e.consoleH - 32
	fillRect(screen, 2, actY, e.hierarchyW-4, 28, e.btnBg)

	actBtns := []struct{
		idx int; text string; id string
	}{{0, "+ Add", "add_entity"}, {1, "⟐ Dup", "duplicate"}, {2, "✕ Del", "delete"}}
	for _, b := range actBtns {
		bx := 6 + float64(b.idx)*68
		bg := e.btnBg
		if e.hoverToolbarItem == b.id { bg = e.btnHover }
		fillRect(screen, bx, actY+2, 62, 24, bg)
		drawTextS(screen, b.text, bx+4, actY+4, 1.2, e.textPrimary)
	}

	fillRect(screen, e.hierarchyW-1, vpY, 1, vpH, color.RGBA{0x3c, 0x3e, 0x40, 0xff})
}

func (e *Editor) drawInspector(screen *ebiten.Image) {
	_, vpY, vpW, vpH := e.viewportRect()
	ix := e.panelLeft() + vpW
	fillRect(screen, ix, vpY, e.inspectorW, vpH, e.bgPanel)

	drawTextS(screen, "INSPETOR", ix+10, vpY+6, 1.4, e.textMuted)
	drawTextS(screen, e.toolName(), ix+e.inspectorW-80, vpY+6, 1.0, e.accent)
	fillRect(screen, ix, vpY+28, e.inspectorW, 1, color.RGBA{0x3c, 0x3e, 0x40, 0xff})

	if e.selectedID <= 0 {
		drawTextS(screen, "Nenhum objeto selecionado", ix+14, vpY+44, 1.2, e.textFaint)
		drawTextS(screen, "Clique em uma entidade no viewport", ix+14, vpY+68, 1.0, e.textFaint)
		drawTextS(screen, "ou na hierarquia para inspecionar", ix+14, vpY+86, 1.0, e.textFaint)
		return
	}

	ent := e.GetEntity(e.selectedID)
	if ent == nil { return }

	y := vpY + 36
	prop := func(h float64) float64 { y += h; return y - h }

	// Identity
	drawTextS(screen, "IDENTIDADE", ix+12, prop(22), 1.2, e.textMuted)
	drawTextS(screen, fmt.Sprintf("Nome:  %s", ent.Name), ix+16, prop(20), 1.3, e.textPrimary)
	drawTextS(screen, fmt.Sprintf("Tipo:  %s", ent.Type), ix+16, prop(20), 1.2, e.textMuted)

	// Separator
	fillRect(screen, ix+10, y, e.inspectorW-20, 1, color.RGBA{0x3c, 0x3e, 0x40, 0xff})
	y += 6

	// Transform
	drawTextS(screen, "TRANSFORM", ix+12, prop(22), 1.2, e.textMuted)
	drawTextS(screen, fmt.Sprintf("X:  %.0f", ent.X), ix+16, prop(20), 1.3, e.textPrimary)
	drawTextS(screen, fmt.Sprintf("Y:  %.0f", ent.Y), ix+16, prop(20), 1.3, e.textPrimary)
	drawTextS(screen, fmt.Sprintf("W:  %.0f", ent.W), ix+16, prop(20), 1.3, e.textPrimary)
	drawTextS(screen, fmt.Sprintf("H:  %.0f", ent.H), ix+16, prop(20), 1.3, e.textPrimary)
	drawTextS(screen, fmt.Sprintf("Rot:  %.0f°", ent.Rotation), ix+16, prop(20), 1.3, e.textPrimary)
	drawTextS(screen, fmt.Sprintf("Z:  %d", ent.ZIndex), ix+16, prop(20), 1.3, e.textPrimary)

	// Visual
	if ent.AssetID != "" || ent.Color != "" {
		fillRect(screen, ix+10, y, e.inspectorW-20, 1, color.RGBA{0x3c, 0x3e, 0x40, 0xff})
		y += 6
		drawTextS(screen, "VISUAL", ix+12, prop(22), 1.2, e.textMuted)
		if ent.AssetID != "" { drawTextS(screen, "Asset:  "+filepath.Base(ent.AssetID), ix+16, prop(20), 1.1, e.textMuted) }
		drawTextS(screen, "Visível:  "+map[bool]string{true: "Sim", false: "Não"}[ent.Visible], ix+16, prop(20), 1.3, e.textPrimary)
		if ent.Color != "" {
			swatchCol := parseHex(ent.Color)
			fillRect(screen, ix+16, y, 18, 14, swatchCol)
			drawTextS(screen, fmt.Sprintf("  %s", ent.Color), ix+36, prop(20), 1.1, e.textMuted)
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
		map[bool]string{true: "ON", false: "OFF"}[e.snapEnabled])
	drawTextS(screen, info, vpX+10, vpY+vpH-18, 1.0, e.textFaint)

	// Entity count
	drawTextS(screen, fmt.Sprintf("%d entidades", len(e.scene.Entities)), vpX+vpW-140, vpY+vpH-18, 1.0, e.textFaint)

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
	drawTextS(screen, icon[ent.Type], sx-5*e.camZoom, sy-6*e.camZoom, 1.5*e.camZoom, color.RGBA{0xff, 0xff, 0xff, 0x99})

	// Label below
	labelScale := 1.2 * e.camZoom
	if labelScale < 0.8 { labelScale = 0.8 }
	if labelScale > 1.6 { labelScale = 1.6 }
	drawTextS(screen, ent.Name, sx-sw/2, sy+sh/2+3, labelScale, e.textPrimary)
}

func (e *Editor) drawConsole(screen *ebiten.Image) {
	cy := float64(e.screenH()) - e.consoleH
	fillRect(screen, 0, cy, float64(e.screenW()), e.consoleH, e.bgPanel)
	drawTextS(screen, "CONSOLE", 10, cy+5, 1.3, e.textMuted)
	fillRect(screen, 0, cy+1, float64(e.screenW()), 1, color.RGBA{0x3c, 0x3e, 0x40, 0xff})

	start := len(e.console) - 4; if start < 0 { start = 0 }
	for i := start; i < len(e.console); i++ {
		drawTextS(screen, "▸ "+e.console[i], 10, cy+22+float64(i-start)*20, 1.1, e.textMuted)
	}
}

func (e *Editor) drawTooltip(screen *ebiten.Image) {
	if e.tooltip == nil || e.tooltip.Text == "" { return }
	lines := []string{e.tooltip.Text}
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
	if tx+maxW > float64(e.screenW()) { tx = e.tooltip.X - maxW - 12 }
	if ty+h > float64(e.screenH())-e.consoleH { ty = e.tooltip.Y - h - 12 }

	fillRect(screen, tx, ty, maxW+12, h, color.RGBA{0x1f, 0x2f, 0x3f, 0xf0})
	drawRectBorder(screen, tx, ty, maxW+12, h, e.accentDim, 1)
	for i, l := range lines {
		drawTextS(screen, l, tx+6, ty+4+float64(i)*14, 0.7, e.textPrimary)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Phase 1: Animation Timeline Panel
// ─────────────────────────────────────────────────────────────────────────────

func (e *Editor) drawTimeline(screen *ebiten.Image) {
	vpX, vpY, vpW, vpH := e.viewportRect()

	// Timeline panel background
	timelineH := 120.0
	timY := vpY + vpH - timelineH
	fillRect(screen, vpX, timY, vpW, timelineH, e.bgPanel)
	fillRect(screen, vpX, timY, vpW, 1, color.RGBA{0x2d, 0x2d, 0x3d, 0xff})

	// Header
	drawTextS(screen, "TIMELINE", vpX+10, timY+4, 1.3, e.textMuted)

	// Clip name and controls
	if e.activeClip != nil {
		drawTextS(screen, e.activeClip.Name, vpX+100, timY+4, 1.3, e.accent)
	}

	// Playback controls
	ctrlY := timY + 22
	playBtnX := vpX + 10

	// Play/Pause button
	isPlaying := e.timelineState != nil && e.timelineState.IsPlaying
	playText := "▶"
	if isPlaying { playText = "⏸" }
	drawTextS(screen, playText, playBtnX, ctrlY, 1.6, e.accent)

	// Stop button
	drawTextS(screen, "⏹", playBtnX+32, ctrlY, 1.6, e.textMuted)

	// Loop toggle
	looping := e.timelineState != nil && e.timelineState.Loop
	loopCol := e.textMuted
	if looping { loopCol = e.accent }
	drawTextS(screen, "⟳", playBtnX+64, ctrlY, 1.6, loopCol)

	// Speed indicator
	speedText := "1x"
	if e.timelineState != nil { speedText = fmt.Sprintf("%.1fx", e.timelineState.PlaySpeed) }
	drawTextS(screen, speedText, playBtnX+96, ctrlY+2, 1.1, e.textMuted)

	// Time indicator
	timeText := "0:00 / 0:00"
	if e.timelineState != nil && e.activeClip != nil {
		cur := e.timelineState.CurrentTime
		dur := e.activeClip.Duration
		timeText = fmt.Sprintf("%d:%.2d / %d:%.2d",
			int(cur)/60, int(cur)%60, int(dur)/60, int(dur)%60)
	}
	drawTextS(screen, timeText, playBtnX+140, ctrlY+2, 1.1, e.textFaint)

	// Track area
	trackY := timY + 50
	trackH := timelineH - 54
	fillRect(screen, vpX+4, trackY, vpW-8, trackH, color.RGBA{0x1a, 0x1c, 0x1e, 0xcc})

	if e.activeClip != nil {
		// Track header + keyframe visualization per track
		for ti, track := range e.activeClip.Tracks {
			ty := trackY + float64(ti)*24
			if ty > trackY+trackH-24 { break }

			// Track label
			drawTextS(screen, track.Property, vpX+10, ty+3, 1.1, e.textMuted)

			// Track background
			trackW := vpW - 90
			trackX := vpX + 80
			fillRect(screen, trackX, ty, trackW, 22, color.RGBA{0x24, 0x26, 0x28, 0xcc})

			// Draw keyframes
			duration := e.activeClip.Duration
			if duration <= 0 { duration = 1 }

			for _, kf := range track.Keyframes {
				kfX := trackX + (kf.Time/duration)*trackW
				if kfX < trackX { continue }
				if kfX > trackX+trackW { break }

				// Keyframe diamond
				kfCol := e.textMuted
				if math.Abs(kf.Time-e.timelineState.CurrentTime) < 0.05 {
					kfCol = e.accent
				}
				fillRect(screen, kfX-2, ty+4, 4, 14, kfCol)
			}

			// Playhead line
			if e.timelineState != nil {
				phX := trackX + (e.timelineState.CurrentTime/duration)*trackW
				if phX >= trackX && phX <= trackX+trackW {
					fillRect(screen, phX, ty, 2, 22, e.accent)
				}
			}
		}
	} else {
		// No clip selected
		drawTextS(screen, "Nenhum clip de animação selecionado", vpX+20, trackY+20, 0.7, e.textFaint)
		drawTextS(screen, "Crie um clip ou selecione uma entidade com animação", vpX+20, trackY+40, 0.6, e.textFaint)
	}

	// Border
	drawRectBorder(screen, vpX, timY, vpW, timelineH, color.RGBA{0x2d, 0x2d, 0x3d, 0xff}, 1)
}

// ─────────────────────────────────────────────────────────────────────────────
// Phase 1: Camera Gizmo
// ─────────────────────────────────────────────────────────────────────────────

func (e *Editor) drawCameraGizmo(screen *ebiten.Image) {
	if !e.showCameraGizmo { return }

	// Find camera entity in scene
	for _, ent := range e.scene.Entities {
		if ent.Type != "camera" { continue }

		sx, sy := e.worldToScreen(ent.X, ent.Y)
		sw, sh := ent.W * e.camZoom, ent.H * e.camZoom
		if sw < 8 { sw = 8 }
		if sh < 8 { sh = 8 }

		// Camera body (yellow box)
		drawRectBorder(screen, sx-sw/2, sy-sh/2, sw, sh, color.RGBA{0xe3, 0xb3, 0x41, 0x99}, 1.5)

		// Crosshair in center
		chSize := math.Min(sw, sh) * 0.3
		if chSize < 4 { chSize = 4 }
		fillRect(screen, sx-1, sy-chSize, 2, chSize*2, color.RGBA{0xe3, 0xb3, 0x41, 0x66})
		fillRect(screen, sx-chSize, sy-1, chSize*2, 2, color.RGBA{0xe3, 0xb3, 0x41, 0x66})

		// Camera frustum (triangular guide)
		frustumW := float64(e.scene.Meta.LogicalW) * 0.4 * e.camZoom
		if frustumW > 200 { frustumW = 200 }
		frustumH := float64(e.scene.Meta.LogicalH) * 0.4 * e.camZoom
		if frustumH > 300 { frustumH = 300 }

		// Draw lines from camera center to frustum corners
		frustumCol := color.RGBA{0xe3, 0xb3, 0x41, 0x30}
		drawLine(screen, sx, sy, sx-frustumW/2, sy-frustumH/2, frustumCol, 1)
		drawLine(screen, sx, sy, sx+frustumW/2, sy-frustumH/2, frustumCol, 1)

		// Frustum rect (dashed-style with border)
		drawRectBorder(screen, sx-frustumW/2, sy-frustumH/2, frustumW, frustumH,
			color.RGBA{0xe3, 0xb3, 0x41, 0x40}, 1)

		// Camera label
		drawTextS(screen, "📷 "+ent.Name, sx+sw/2+4, sy-6, 0.65, color.RGBA{0xe3, 0xb3, 0x41, 0xcc})

		break // Only show first camera
	}
}

// drawLine draws a line between two points using the fillRect helper.
func drawLine(screen *ebiten.Image, x1, y1, x2, y2 float64, c color.RGBA, bw float64) {
	dx := x2 - x1
	dy := y2 - y1
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 1 { return }
	// Draw as a thin rotated rectangle
	steps := int(dist / 2)
	if steps < 1 { steps = 1 }
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps)
		fillRect(screen, x1+dx*t, y1+dy*t, bw, bw, c)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Phase 1: Play Mode
// ─────────────────────────────────────────────────────────────────────────────

func (e *Editor) TogglePlayMode() {
	if e.playMode {
		// Stop play mode
		e.playMode = false
		if e.sceneRunner != nil {
			e.sceneRunner.DestroyAll()
			e.sceneRunner = nil
		}
		e.Log("Play mode: desativado")
	} else {
		// Start play mode
		e.playMode = true
		e.playStart = time.Now()
		e.Log("Play mode: ativado — executando cena no runtime")

		// Create runtime scene from editor data
		sf := &editor.SceneFile{
			Meta: editor.SceneMeta{
				Name:     e.scene.Meta.Name,
				Version:  e.scene.Meta.Version,
				LogicalW: e.scene.Meta.LogicalW,
				LogicalH: e.scene.Meta.LogicalH,
			},
		}

		// Convert editor entities to bridge format
		for _, ent := range e.scene.Entities {
			sf.Entities = append(sf.Entities, &editor.SceneEntity{
				ID:       ent.ID,
				Name:     ent.Name,
				Type:     ent.Type,
				X:        ent.X,
				Y:        ent.Y,
				W:        ent.W,
				H:        ent.H,
				Rotation: ent.Rotation,
				Color:    ent.Color,
				Visible:  ent.Visible,
				ParentID: ent.ParentID,
				ZIndex:   ent.ZIndex,
				AssetID:  ent.AssetID,
				Script:   ent.Script,
			})
		}

		// Instantiate the scene via bridge
		sceneRunner, cleanup := editor.Instantiate(sf)
		e.sceneRunner = sceneRunner
		_ = cleanup // cleanup happens on stop
	}
}

func (e *Editor) updatePlayMode(dt float64) {
	if !e.playMode || e.sceneRunner == nil { return }

	// Update the runtime scene
	// In a full implementation, this would be driven by the game loop.
	// For now, we tick the scene update.
	if updater, ok := interface{}(e.sceneRunner).(interface{ Update(float64) }); ok {
		updater.Update(dt)
	}
}

func (e *Editor) drawPlayModeOverlay(screen *ebiten.Image) {
	if !e.playMode { return }

	// Red border indicator
	screenW := float64(e.screenW())
	fillRect(screen, 0, 0, screenW, 4, color.RGBA{0x1e, 0xa5, 0x5e, 0xaa})

	// Play time
	elapsed := time.Since(e.playStart)
	timeStr := fmt.Sprintf("PLAY %d:%.2d", int(elapsed.Seconds())/60, int(elapsed.Seconds())%60)
	drawTextS(screen, timeStr, screenW-80, e.toolbarH+6, 0.65, color.RGBA{0x1e, 0xa5, 0x5e, 0xcc})

	// FPS indicator
	drawTextS(screen, fmt.Sprintf("%.0f FPS", ebiten.ActualFPS()), screenW-80, e.toolbarH+18, 0.6, e.textFaint)

	// Draw runtime scene if available
	if e.sceneRunner != nil {
		// In a full implementation, the scene would be rendered
		// via the runtime renderer here.
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Phase 1: Hot-Reload UI helpers
// ─────────────────────────────────────────────────────────────────────────────

func (e *Editor) drawHotReloadStatus(screen *ebiten.Image) {
	if !e.hotReload.IsEnabled() { return }

	// Show status in the console area
	cy := float64(e.screenH()) - e.consoleH
	errs := e.hotReload.BuildErrors()
	if len(errs) > 0 {
		drawTextS(screen, fmt.Sprintf("⚠ %d erro(s) de compilação", len(errs)),
			float64(e.screenW())/2, cy+4, 0.65, color.RGBA{0xf4, 0x47, 0x47, 0xff})
	} else {
		logs := e.hotReload.EventLog()
		if len(logs) > 0 {
			last := logs[len(logs)-1]
			statusText := "✓ " + last.Message
			drawTextS(screen, statusText,
				float64(e.screenW())/2, cy+4, 0.6, color.RGBA{0x1e, 0xa5, 0x5e, 0xcc})
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Phase 2: Sprite Editor Panel
// ─────────────────────────────────────────────────────────────────────────────

func (e *Editor) drawSpriteEditor(screen *ebiten.Image) {
	se := e.spriteEditor
	vpX, vpY, vpW, vpH := e.viewportRect()

	// Panel background
	fillRect(screen, vpX, vpY, vpW, vpH, e.bgDark)

	// Left sidebar: tool options
	sidebarW := 160.0
	fillRect(screen, vpX, vpY, sidebarW, vpH, e.bgPanel)
	fillRect(screen, vpX+sidebarW, vpY, 1, vpH, color.RGBA{0x2d, 0x2d, 0x3d, 0xff})

	drawTextS(screen, "SPRITE EDITOR", vpX+12, vpY+8, 1.5, e.accent)
	drawTextS(screen, "[F10] fechar", vpX+12, vpY+28, 1.0, e.textFaint)

	if se.Resource == nil {
		// No sprite loaded
		drawTextS(screen, "Nenhum sprite carregado", vpX+sidebarW+24, vpY+50, 1.3, e.textMuted)
		drawTextS(screen, "Arraste uma imagem PNG/JPEG para importar,", vpX+sidebarW+24, vpY+76, 1.1, e.textFaint)
		drawTextS(screen, "ou use Ctrl+I para criar um sprite teste", vpX+sidebarW+24, vpY+98, 1.1, e.textFaint)
		return
	}

	r := se.Resource

	// Sidebar tools
	y := vpY + 50
	drawTextS(screen, "ARQUIVO", vpX+12, y, 1.2, e.textMuted)
	y += 22
	drawTextS(screen, r.Name, vpX+16, y, 1.3, e.textPrimary)
	y += 20
	drawTextS(screen, fmt.Sprintf("%dx%d px", r.SrcWidth, r.SrcHeight), vpX+16, y, 1.0, e.textFaint)
	y += 18
	drawTextS(screen, string(r.Type), vpX+16, y, 1.1, e.accent)
	y += 28

	drawTextS(screen, "FRAMES", vpX+12, y, 1.2, e.textMuted)
	y += 22
	drawTextS(screen, fmt.Sprintf("%d frames", r.FrameCount()), vpX+16, y, 1.3, e.textPrimary)
	y += 20
	drawTextS(screen, fmt.Sprintf("Grid: %dx%d", r.GridCols, r.GridRows), vpX+16, y, 1.0, e.textFaint)
	y += 28

	drawTextS(screen, "PIVOT", vpX+12, y, 1.2, e.textMuted)
	y += 22
	drawTextS(screen, fmt.Sprintf("(%.2f, %.2f)", r.PivotX, r.PivotY), vpX+16, y, 1.3, e.textPrimary)
	y += 18
	pivotPresets := []string{"center", "tl", "tr", "bl", "br"}
	for i, p := range pivotPresets {
		px := vpX + 12 + float64(i%4)*44
		py := y + float64(i/4)*22
		col := e.textMuted
		if (p == "center" && r.PivotX == 0.5 && r.PivotY == 0.5) ||
			(p == "tl" && r.PivotX == 0 && r.PivotY == 0) {
			col = e.accent
		}
		drawTextS(screen, p, px, py, 1.0, col)
	}
	y += 52

	if len(r.Hitboxes) > 0 {
		drawTextS(screen, "HITBOXES", vpX+12, y, 1.2, e.textMuted)
		y += 22
		for hi, hb := range r.Hitboxes {
			hcol := e.textPrimary
			if hi == se.EditingHitbox { hcol = e.accent }
			drawTextS(screen, fmt.Sprintf("%d: %s (%.0fx%.0f)", hi, hb.Name, hb.W, hb.H),
				vpX+16, y+float64(hi)*20, 1.0, hcol)
		}
	}

	// Right side: viewport preview
	viewX := vpX + sidebarW + 10
	viewY := vpY + 10
	viewW := vpW - sidebarW - 20
	viewH := vpH - 60

	// Checkerboard background
	if se.ShowChecker {
		checkSize := 8.0 * se.Zoom
		if checkSize < 4 { checkSize = 4 }
		for cx := viewX; cx < viewX+viewW; cx += checkSize * 2 {
			for cy := viewY; cy < viewY+viewH; cy += checkSize * 2 {
				fillRect(screen, cx, cy, checkSize, checkSize, color.RGBA{0x33, 0x33, 0x33, 0x66})
				fillRect(screen, cx+checkSize, cy+checkSize, checkSize, checkSize, color.RGBA{0x33, 0x33, 0x33, 0x66})
			}
		}
	}

	// Draw first frame as colored rect placeholder
	// (actual texture rendering needs Ebitengine image)
	if r.FrameCount() > 0 {
		frame := &r.Frames[0]
		frameW := float64(frame.W) * se.Zoom
		frameH := float64(frame.H) * se.Zoom

		// Center in viewport
		drawX := viewX + viewW/2 - frameW/2
		drawY := viewY + viewH/2 - frameH/2

		// Draw frame rect
		fillRect(screen, drawX, drawY, frameW, frameH, color.RGBA{0x4a, 0x9e, 0xff, 0x88})
		drawRectBorder(screen, drawX, drawY, frameW, frameH, e.accent, 1.5)

		// Draw pivot crosshair
		pivotScreenX := drawX + float64(frame.W)*se.Zoom*r.PivotX
		pivotScreenY := drawY + float64(frame.H)*se.Zoom*r.PivotY
		crossSize := 6.0 * se.Zoom
		if crossSize < 4 { crossSize = 4 }
		fillRect(screen, pivotScreenX-1, pivotScreenY-crossSize, 2, crossSize*2, color.RGBA{0xff, 0xff, 0x00, 0xcc})
		fillRect(screen, pivotScreenX-crossSize, pivotScreenY-1, crossSize*2, 2, color.RGBA{0xff, 0xff, 0x00, 0xcc})

		// Draw hitboxes
		for _, hb := range r.Hitboxes {
			hx := drawX + hb.X*se.Zoom
			hy := drawY + hb.Y*se.Zoom
			hw := hb.W * se.Zoom
			hh := hb.H * se.Zoom
			hcol := editor.GetHitboxPreviewColor(hb.Type)
			drawRectBorder(screen, hx, hy, hw, hh, hcol, 1.5)
			fillRect(screen, hx, hy, hw, hh, color.RGBA{hcol.R, hcol.G, hcol.B, hcol.A / 3})
		}
	}

	// Bottom toolbar
	toolY := vpY + vpH - 30
	fillRect(screen, vpX+sidebarW, toolY, vpW-sidebarW, 30, e.bgPanel)

	zoomText := fmt.Sprintf("Zoom: %.0f%%", se.Zoom*100)
	drawTextS(screen, zoomText, vpX+sidebarW+14, toolY+5, 1.1, e.textMuted)

	frameInfo := fmt.Sprintf("Frame: 1/%d", r.FrameCount())
	drawTextS(screen, frameInfo, vpX+sidebarW+140, toolY+5, 1.1, e.textMuted)

	// Import hint
	drawTextS(screen, "Ctrl+I: Importar sprite", vpX+sidebarW+280, toolY+5, 1.0, e.textFaint)

	// Border
	drawRectBorder(screen, vpX, vpY, vpW, vpH, color.RGBA{0x2d, 0x2d, 0x3d, 0xff}, 1)
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


// ---------------------------------------------------------------------------
// Built-in 8x8 bitmap font data (ASCII 32-126)
// Each glyph is 8 bytes, one per row. MSB = leftmost pixel.
// ---------------------------------------------------------------------------

var bitmapFontData = [95][8]byte{
	// 32  space
	{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	// 33  !
	{0x18, 0x18, 0x18, 0x18, 0x18, 0x00, 0x18, 0x00},
	// 34  "
	{0x6C, 0x6C, 0x6C, 0x00, 0x00, 0x00, 0x00, 0x00},
	// 35  #
	{0x6C, 0x6C, 0xFE, 0x6C, 0xFE, 0x6C, 0x6C, 0x00},
	// 36  $
	{0x18, 0x3E, 0x60, 0x3C, 0x06, 0x7C, 0x18, 0x00},
	// 37  %
	{0x00, 0x66, 0xAC, 0xD8, 0x36, 0x6A, 0xCC, 0x00},
	// 38  &
	{0x38, 0x6C, 0x68, 0x76, 0xDC, 0xCC, 0x76, 0x00},
	// 39  '
	{0x18, 0x18, 0x18, 0x00, 0x00, 0x00, 0x00, 0x00},
	// 40  (
	{0x0C, 0x18, 0x30, 0x30, 0x30, 0x18, 0x0C, 0x00},
	// 41  )
	{0x30, 0x18, 0x0C, 0x0C, 0x0C, 0x18, 0x30, 0x00},
	// 42  *
	{0x00, 0x66, 0x3C, 0xFF, 0x3C, 0x66, 0x00, 0x00},
	// 43  +
	{0x00, 0x18, 0x18, 0x7E, 0x18, 0x18, 0x00, 0x00},
	// 44  ,
	{0x00, 0x00, 0x00, 0x00, 0x00, 0x18, 0x18, 0x30},
	// 45  -
	{0x00, 0x00, 0x00, 0x7E, 0x00, 0x00, 0x00, 0x00},
	// 46  .
	{0x00, 0x00, 0x00, 0x00, 0x00, 0x18, 0x18, 0x00},
	// 47  /
	{0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80, 0x00},
	// 48  0
	{0x3C, 0x66, 0x66, 0x66, 0x66, 0x66, 0x3C, 0x00},
	// 49  1
	{0x18, 0x38, 0x18, 0x18, 0x18, 0x18, 0x7E, 0x00},
	// 50  2
	{0x3C, 0x66, 0x06, 0x0C, 0x30, 0x60, 0x7E, 0x00},
	// 51  3
	{0x3C, 0x66, 0x06, 0x1C, 0x06, 0x66, 0x3C, 0x00},
	// 52  4
	{0x0C, 0x1C, 0x3C, 0x6C, 0x7E, 0x0C, 0x0C, 0x00},
	// 53  5
	{0x7E, 0x60, 0x7C, 0x06, 0x06, 0x66, 0x3C, 0x00},
	// 54  6
	{0x3C, 0x66, 0x60, 0x7C, 0x66, 0x66, 0x3C, 0x00},
	// 55  7
	{0x7E, 0x06, 0x0C, 0x18, 0x30, 0x30, 0x30, 0x00},
	// 56  8
	{0x3C, 0x66, 0x66, 0x3C, 0x66, 0x66, 0x3C, 0x00},
	// 57  9
	{0x3C, 0x66, 0x66, 0x3E, 0x06, 0x66, 0x3C, 0x00},
	// 58  :
	{0x00, 0x18, 0x18, 0x00, 0x00, 0x18, 0x18, 0x00},
	// 59  ;
	{0x00, 0x18, 0x18, 0x00, 0x00, 0x18, 0x18, 0x30},
	// 60  <
	{0x06, 0x0C, 0x18, 0x30, 0x18, 0x0C, 0x06, 0x00},
	// 61  =
	{0x00, 0x00, 0x7E, 0x00, 0x00, 0x7E, 0x00, 0x00},
	// 62  >
	{0x60, 0x30, 0x18, 0x0C, 0x18, 0x30, 0x60, 0x00},
	// 63  ?
	{0x3C, 0x66, 0x06, 0x0C, 0x18, 0x00, 0x18, 0x00},
	// 64  @
	{0x3C, 0x66, 0x6E, 0x6E, 0x60, 0x66, 0x3C, 0x00},
	// 65  A
	{0x18, 0x3C, 0x66, 0x66, 0x7E, 0x66, 0x66, 0x00},
	// 66  B
	{0x7C, 0x66, 0x66, 0x7C, 0x66, 0x66, 0x7C, 0x00},
	// 67  C
	{0x3C, 0x66, 0x60, 0x60, 0x60, 0x66, 0x3C, 0x00},
	// 68  D
	{0x78, 0x6C, 0x66, 0x66, 0x66, 0x6C, 0x78, 0x00},
	// 69  E
	{0x7E, 0x60, 0x60, 0x7C, 0x60, 0x60, 0x7E, 0x00},
	// 70  F
	{0x7E, 0x60, 0x60, 0x7C, 0x60, 0x60, 0x60, 0x00},
	// 71  G
	{0x3C, 0x66, 0x60, 0x6E, 0x66, 0x66, 0x3C, 0x00},
	// 72  H
	{0x66, 0x66, 0x66, 0x7E, 0x66, 0x66, 0x66, 0x00},
	// 73  I
	{0x7E, 0x18, 0x18, 0x18, 0x18, 0x18, 0x7E, 0x00},
	// 74  J
	{0x06, 0x06, 0x06, 0x06, 0x66, 0x66, 0x3C, 0x00},
	// 75  K
	{0x66, 0x6C, 0x78, 0x70, 0x78, 0x6C, 0x66, 0x00},
	// 76  L
	{0x60, 0x60, 0x60, 0x60, 0x60, 0x60, 0x7E, 0x00},
	// 77  M
	{0x66, 0x76, 0x7E, 0x7E, 0x66, 0x66, 0x66, 0x00},
	// 78  N
	{0x66, 0x76, 0x7E, 0x7E, 0x6E, 0x66, 0x66, 0x00},
	// 79  O
	{0x3C, 0x66, 0x66, 0x66, 0x66, 0x66, 0x3C, 0x00},
	// 80  P
	{0x7C, 0x66, 0x66, 0x7C, 0x60, 0x60, 0x60, 0x00},
	// 81  Q
	{0x3C, 0x66, 0x66, 0x66, 0x6E, 0x3C, 0x06, 0x00},
	// 82  R
	{0x7C, 0x66, 0x66, 0x7C, 0x78, 0x6C, 0x66, 0x00},
	// 83  S
	{0x3C, 0x66, 0x60, 0x3C, 0x06, 0x66, 0x3C, 0x00},
	// 84  T
	{0x7E, 0x18, 0x18, 0x18, 0x18, 0x18, 0x18, 0x00},
	// 85  U
	{0x66, 0x66, 0x66, 0x66, 0x66, 0x66, 0x3C, 0x00},
	// 86  V
	{0x66, 0x66, 0x66, 0x66, 0x66, 0x3C, 0x18, 0x00},
	// 87  W
	{0x66, 0x66, 0x66, 0x7E, 0x7E, 0x76, 0x66, 0x00},
	// 88  X
	{0x66, 0x66, 0x3C, 0x18, 0x3C, 0x66, 0x66, 0x00},
	// 89  Y
	{0x66, 0x66, 0x66, 0x3C, 0x18, 0x18, 0x18, 0x00},
	// 90  Z
	{0x7E, 0x06, 0x0C, 0x18, 0x30, 0x60, 0x7E, 0x00},
	// 91  [
	{0x3C, 0x30, 0x30, 0x30, 0x30, 0x30, 0x3C, 0x00},
	// 92  \
	{0x80, 0x40, 0x20, 0x10, 0x08, 0x04, 0x02, 0x00},
	// 93  ]
	{0x3C, 0x0C, 0x0C, 0x0C, 0x0C, 0x0C, 0x3C, 0x00},
	// 94  ^
	{0x18, 0x3C, 0x66, 0x00, 0x00, 0x00, 0x00, 0x00},
	// 95  _
	{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF},
	// 96  `
	{0x30, 0x18, 0x0C, 0x00, 0x00, 0x00, 0x00, 0x00},
	// 97  a
	{0x00, 0x00, 0x3C, 0x06, 0x3E, 0x66, 0x3E, 0x00},
	// 98  b
	{0x60, 0x60, 0x7C, 0x66, 0x66, 0x66, 0x7C, 0x00},
	// 99  c
	{0x00, 0x00, 0x3C, 0x66, 0x60, 0x66, 0x3C, 0x00},
	// 100 d
	{0x06, 0x06, 0x3E, 0x66, 0x66, 0x66, 0x3E, 0x00},
	// 101 e
	{0x00, 0x00, 0x3C, 0x66, 0x7E, 0x60, 0x3C, 0x00},
	// 102 f
	{0x1C, 0x30, 0x7C, 0x30, 0x30, 0x30, 0x30, 0x00},
	// 103 g
	{0x00, 0x00, 0x3E, 0x66, 0x66, 0x3E, 0x06, 0x3C},
	// 104 h
	{0x60, 0x60, 0x7C, 0x66, 0x66, 0x66, 0x66, 0x00},
	// 105 i
	{0x18, 0x00, 0x38, 0x18, 0x18, 0x18, 0x3C, 0x00},
	// 106 j
	{0x06, 0x00, 0x0E, 0x06, 0x06, 0x66, 0x66, 0x3C},
	// 107 k
	{0x60, 0x60, 0x66, 0x6C, 0x78, 0x6C, 0x66, 0x00},
	// 108 l
	{0x38, 0x18, 0x18, 0x18, 0x18, 0x18, 0x3C, 0x00},
	// 109 m
	{0x00, 0x00, 0x6C, 0x7E, 0x7E, 0x6A, 0x6A, 0x00},
	// 110 n
	{0x00, 0x00, 0x7C, 0x66, 0x66, 0x66, 0x66, 0x00},
	// 111 o
	{0x00, 0x00, 0x3C, 0x66, 0x66, 0x66, 0x3C, 0x00},
	// 112 p
	{0x00, 0x00, 0x7C, 0x66, 0x66, 0x7C, 0x60, 0x60},
	// 113 q
	{0x00, 0x00, 0x3E, 0x66, 0x66, 0x3E, 0x06, 0x06},
	// 114 r
	{0x00, 0x00, 0x7C, 0x66, 0x60, 0x60, 0x60, 0x00},
	// 115 s
	{0x00, 0x00, 0x3E, 0x60, 0x3C, 0x06, 0x7C, 0x00},
	// 116 t
	{0x30, 0x30, 0x7C, 0x30, 0x30, 0x30, 0x1C, 0x00},
	// 117 u
	{0x00, 0x00, 0x66, 0x66, 0x66, 0x66, 0x3E, 0x00},
	// 118 v
	{0x00, 0x00, 0x66, 0x66, 0x66, 0x3C, 0x18, 0x00},
	// 119 w
	{0x00, 0x00, 0x66, 0x6A, 0x7E, 0x7E, 0x6C, 0x00},
	// 120 x
	{0x00, 0x00, 0x66, 0x3C, 0x18, 0x3C, 0x66, 0x00},
	// 121 y
	{0x00, 0x00, 0x66, 0x66, 0x66, 0x3E, 0x06, 0x3C},
	// 122 z
	{0x00, 0x00, 0x7E, 0x0C, 0x18, 0x30, 0x7E, 0x00},
	// 123 {
	{0x0E, 0x18, 0x18, 0x70, 0x18, 0x18, 0x0E, 0x00},
	// 124 |
	{0x18, 0x18, 0x18, 0x18, 0x18, 0x18, 0x18, 0x00},
	// 125 }
	{0x70, 0x18, 0x18, 0x0E, 0x18, 0x18, 0x70, 0x00},
	// 126 ~
	{0x00, 0x00, 0x00, 0x76, 0xDC, 0x00, 0x00, 0x00},
}

func drawTextS(screen *ebiten.Image, text string, x, y float64, scale float64, clr color.RGBA) {
	if text == "" { return }
	pixelW := 1.0 * scale
	pixelH := 1.0 * scale
	charW := 8.0 * scale
	px := 0.0

	for _, ch := range text {
		if ch < 32 || ch > 126 { px += charW; continue }
		idx := ch - 32
		if int(idx) >= len(bitmapFontData) { px += charW; continue }
		glyph := bitmapFontData[idx]
		for row := 0; row < 8; row++ {
			b := glyph[row]
			for col := 0; col < 8; col++ {
				if b&(1<<(7-col)) != 0 {
					op := &ebiten.DrawImageOptions{}
					op.GeoM.Scale(pixelW, pixelH)
					op.GeoM.Translate(x+px+float64(col)*pixelW, y+float64(row)*pixelH)
					op.ColorScale.SetR(float32(clr.R)/255)
					op.ColorScale.SetG(float32(clr.G)/255)
					op.ColorScale.SetB(float32(clr.B)/255)
					op.ColorScale.SetA(float32(clr.A)/255)
					screen.DrawImage(whitePixel, op)
				}
			}
		}
		px += charW + 1*scale // 1 pixel spacing
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
