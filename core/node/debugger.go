// Package node implements the core node system for Kora Engine.
package node

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/ElioNeto/kora/core/render"
)

// ---------------------------------------------------------------------------
// Layout constants for the debug overlay
// ---------------------------------------------------------------------------

const (
	debugPanelWidth   = 300.0
	debugPanelPadding = 6.0
	debugLineHeight   = 14.0
	debugStartX       = 4.0
	debugStartY       = 4.0
)

// ---------------------------------------------------------------------------
// Color constants
// ---------------------------------------------------------------------------

var (
	debugBgColor = color.RGBA{0, 0, 0, 180}
)

// ---------------------------------------------------------------------------
// DebugConsole
// ---------------------------------------------------------------------------

// DebugConsole is an in-game debug overlay for runtime inspection.
// It displays real-time information such as FPS, entity counts, task counts,
// and the node hierarchy. Visibility is toggled with a configurable key.
//
// Embed DebugConsole into your scene's root node to attach the overlay.
type DebugConsole struct {
	*Node2D
	Visible         bool    // toggle visibility
	ToggleKey       string  // key to toggle (default "F3")
	ShowFPS         bool    // show FPS panel
	ShowEntityCount bool    // show entity count panel
	ShowPhysics     bool    // show physics info panel
	ShowTaskCount   bool    // show async task count panel
	ShowNodeTree    bool    // show node hierarchy tree
	ShowCameraInfo  bool    // show camera info panel
	ShowMemory      bool    // show memory info panel
	Scale           float64 // text scale factor

	// Internal references (set via SetScene / SetSceneTree)
	scene     interface{}
	sceneTree interface{}
	trackedGroups []string // group names for per-group entity breakdown

	// Cached node tree lines (rebuilt each frame when visible)
	nodeTree     []string
	selectedNode string // currently selected node path

	// FPS sampling — circular buffer for min / max / avg over ~1 s
	fpsSamples [120]float64
	fpsIndex   int
	fpsCount   int
}

// NewDebugConsole creates a new DebugConsole with default settings.
// The console starts hidden and toggles with the F3 key.
func NewDebugConsole(name string) *DebugConsole {
	return &DebugConsole{
		Node2D:          NewNode2D(name, 0),
		Visible:         false,
		ToggleKey:       "F3",
		ShowFPS:         true,
		ShowEntityCount: true,
		ShowPhysics:     false,
		ShowTaskCount:   true,
		ShowNodeTree:    false,
		ShowCameraInfo:  false,
		ShowMemory:      false,
		Scale:           1.0,
	}
}

// ---------------------------------------------------------------------------
// Basic accessors / mutators
// ---------------------------------------------------------------------------

// Toggle shows or hides the debug console.
func (dc *DebugConsole) Toggle() {
	dc.Visible = !dc.Visible
}

// IsVisible returns whether the console is currently shown.
func (dc *DebugConsole) IsVisible() bool {
	return dc.Visible
}

// SetScene sets the scene reference used for entity inspection.
// The argument must provide Count() and Find(name) methods.
func (dc *DebugConsole) SetScene(sc interface{ Count() int; Find(name string) interface{} }) {
	dc.scene = sc
}

// SetSceneTree sets the scene-tree reference used for tree-level inspection.
// The argument must provide CurrentScene(), IsPaused() and TPS() methods.
func (dc *DebugConsole) SetSceneTree(st interface {
	CurrentScene() interface{ Count() int }
	IsPaused() bool
	TPS() float64
}) {
	dc.sceneTree = st
}

// ---------------------------------------------------------------------------
// Panel toggles
// ---------------------------------------------------------------------------

// SetShowFPS toggles the FPS display panel.
func (dc *DebugConsole) SetShowFPS(show bool) { dc.ShowFPS = show }

// SetShowEntityCount toggles the entity count panel.
func (dc *DebugConsole) SetShowEntityCount(show bool) { dc.ShowEntityCount = show }

// SetShowPhysics toggles the physics info panel.
func (dc *DebugConsole) SetShowPhysics(show bool) { dc.ShowPhysics = show }

// SetShowTaskCount toggles the async task count panel.
func (dc *DebugConsole) SetShowTaskCount(show bool) { dc.ShowTaskCount = show }

// SetShowNodeTree toggles the node hierarchy tree panel.
func (dc *DebugConsole) SetShowNodeTree(show bool) { dc.ShowNodeTree = show }

// SetShowCameraInfo toggles the camera info panel.
func (dc *DebugConsole) SetShowCameraInfo(show bool) { dc.ShowCameraInfo = show }

// SetShowMemory toggles the memory info panel.
func (dc *DebugConsole) SetShowMemory(show bool) { dc.ShowMemory = show }

// SetTextScale sets the font scale factor for the debug overlay.
func (dc *DebugConsole) SetTextScale(scale float64) { dc.Scale = scale }

// AddTrackedGroup adds a group name to show a per-group entity count
// breakdown in the entity count panel.
func (dc *DebugConsole) AddTrackedGroup(name string) {
	for _, g := range dc.trackedGroups {
		if g == name {
			return
		}
	}
	dc.trackedGroups = append(dc.trackedGroups, name)
}

// ---------------------------------------------------------------------------
// Update — called once per tick
// ---------------------------------------------------------------------------

// Update processes toggle key input and refreshes debug data.
func (dc *DebugConsole) Update(dt float64) {
	dc.Node2D.Update(dt)

	// Toggle on key press (edge-triggered).
	if key := parseKeyName(dc.ToggleKey); key != ebiten.Key(0) {
		if inpututil.IsKeyJustPressed(key) {
			dc.Toggle()
		}
	}

	// Always sample FPS so we have data ready when visible.
	dc.sampleFPS()

	// Refresh node tree when visible.
	if dc.ShowNodeTree {
		dc.refreshNodeTree()
	}
}

// ---------------------------------------------------------------------------
// Draw — called once per frame
// ---------------------------------------------------------------------------

// Draw renders the debug overlay with current information.
// Panels are drawn as semi-transparent rectangles with text lines.
func (dc *DebugConsole) Draw(screen *ebiten.Image) {
	if !dc.Visible {
		return
	}

	lines := dc.buildDebugLines()
	if len(lines) == 0 {
		return
	}

	lh := debugLineHeight * dc.Scale
	pad := debugPanelPadding * dc.Scale
	panelW := debugPanelWidth * dc.Scale
	panelH := float64(len(lines))*lh + pad*2

	// Semi-transparent background covering all panels.
	render.DrawRect(screen, debugStartX, debugStartY, panelW, panelH, debugBgColor)

	// Draw each text line.
	x := debugStartX + pad
	y := debugStartY + pad
	for _, line := range lines {
		render.DebugTextAt(screen, line, int(x), int(y))
		y += lh
	}
}

// ---------------------------------------------------------------------------
// Internal — FPS sampling
// ---------------------------------------------------------------------------

// sampleFPS records the current FPS into the circular buffer.
func (dc *DebugConsole) sampleFPS() {
	if dc.fpsCount < len(dc.fpsSamples) {
		dc.fpsCount++
	}
	dc.fpsSamples[dc.fpsIndex] = ebiten.ActualFPS()
	dc.fpsIndex = (dc.fpsIndex + 1) % len(dc.fpsSamples)
}

// computeFPSStats returns min, max and average over the current sample window.
func (dc *DebugConsole) computeFPSStats() (min, max, avg float64) {
	if dc.fpsCount == 0 {
		return 0, 0, 0
	}
	min = dc.fpsSamples[0]
	max = dc.fpsSamples[0]
	sum := 0.0
	for i := 0; i < dc.fpsCount; i++ {
		v := dc.fpsSamples[i]
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}
	avg = sum / float64(dc.fpsCount)
	return
}

// ---------------------------------------------------------------------------
// Internal — debug line builder
// ---------------------------------------------------------------------------

// sceneWithGroups is an extended interface that some scene types satisfy,
// allowing per-group entity counts without coupling to the scene package.
type sceneWithGroups interface {
	Count() int
	Find(name string) interface{}
	CountInGroup(group string) int
}

// sceneWithScheduler is an extended interface that allows reading the
// scheduler's active task count.
type sceneWithScheduler interface {
	Count() int
	Find(name string) interface{}
	Scheduler() interface{ Len() int }
}

// buildDebugLines constructs the text lines for every enabled panel.
func (dc *DebugConsole) buildDebugLines() []string {
	var lines []string

	// ---- FPS panel ----
	if dc.ShowFPS {
		lines = append(lines, "--- FPS ---")
		cur := ebiten.ActualFPS()
		tps := ebiten.ActualTPS()
		lines = append(lines, fmt.Sprintf("FPS: %.1f  TPS: %.1f", cur, tps))
		if dc.fpsCount > 0 {
			min, max, avg := dc.computeFPSStats()
			lines = append(lines, fmt.Sprintf("Avg: %.1f  Min: %.1f  Max: %.1f", avg, min, max))
		}
	}

	// ---- Entity count panel ----
	if dc.ShowEntityCount && dc.scene != nil {
		lines = append(lines, "--- Entities ---")
		if sc, ok := dc.scene.(interface{ Count() int }); ok {
			lines = append(lines, fmt.Sprintf("Total: %d", sc.Count()))
		}
		// Per-group breakdown for tracked groups.
		for _, group := range dc.trackedGroups {
			if sc, ok := dc.scene.(sceneWithGroups); ok {
				n := sc.CountInGroup(group)
				lines = append(lines, fmt.Sprintf("  %s: %d", group, n))
			}
		}
	}

	// ---- Task count panel ----
	if dc.ShowTaskCount && dc.scene != nil {
		lines = append(lines, "--- Tasks ---")
		if sc, ok := dc.scene.(sceneWithScheduler); ok {
			lines = append(lines, fmt.Sprintf("Active: %d", sc.Scheduler().Len()))
		}
	}

	// ---- Node tree panel ----
	if dc.ShowNodeTree && dc.scene != nil {
		lines = append(lines, "--- Node Tree ---")
		if len(dc.nodeTree) > 0 {
			lines = append(lines, dc.nodeTree...)
		} else {
			lines = append(lines, "(empty)")
		}
	}

	// ---- Physics panel ----
	if dc.ShowPhysics {
		lines = append(lines, "--- Physics ---")
		if dc.scene != nil {
			if sc, ok := dc.scene.(interface{ Count() int }); ok {
				lines = append(lines, fmt.Sprintf("Entities: %d", sc.Count()))
			}
		} else {
			lines = append(lines, "N/A")
		}
	}

	// ---- Camera info panel ----
	if dc.ShowCameraInfo {
		lines = append(lines, "--- Camera ---")
		lines = append(lines, "N/A")
	}

	// ---- Memory panel ----
	if dc.ShowMemory {
		lines = append(lines, "--- Memory ---")
		lines = append(lines, "N/A")
	}

	// ---- Scene tree status (always shown when tree is set) ----
	if dc.sceneTree != nil {
		lines = append(lines, "--- Scene Tree ---")
		if st, ok := dc.sceneTree.(interface {
			CurrentScene() interface{ Count() int }
			IsPaused() bool
			TPS() float64
		}); ok {
			paused := "running"
			if st.IsPaused() {
				paused = "paused"
			}
			cs := st.CurrentScene()
			count := 0
			if cs != nil {
				if c, ok := cs.(interface{ Count() int }); ok {
					count = c.Count()
				}
			}
			lines = append(lines, fmt.Sprintf("State: %s  Entities: %d", paused, count))
		}
	}

	return lines
}

// ---------------------------------------------------------------------------
// Internal — node tree builder
// ---------------------------------------------------------------------------

// refreshNodeTree rebuilds the cached node tree lines from the current scene.
func (dc *DebugConsole) refreshNodeTree() {
	dc.nodeTree = dc.nodeTree[:0]

	if dc.scene == nil {
		dc.nodeTree = append(dc.nodeTree, "(no scene)")
		return
	}

	// Show basic scene info as the root of the tree.
	sc, ok := dc.scene.(interface{ Count() int })
	if !ok {
		dc.nodeTree = append(dc.nodeTree, "(scene)")
		return
	}
	dc.nodeTree = append(dc.nodeTree, fmt.Sprintf("Scene (%d entities)", sc.Count()))

	// Append tracked groups as second-level tree entries.
	for _, group := range dc.trackedGroups {
		if swg, ok := dc.scene.(sceneWithGroups); ok {
			n := swg.CountInGroup(group)
			dc.nodeTree = append(dc.nodeTree, fmt.Sprintf("  ├─ %s (%d)", group, n))
		} else {
			dc.nodeTree = append(dc.nodeTree, fmt.Sprintf("  ├─ %s", group))
		}
	}
}

// ---------------------------------------------------------------------------
// Key name to ebiten.Key mapping
// ---------------------------------------------------------------------------

// keyNameToEbiten maps common key names to ebiten.Key values.
var keyNameToEbiten = map[string]ebiten.Key{
	"F1":  ebiten.KeyF1,
	"F2":  ebiten.KeyF2,
	"F3":  ebiten.KeyF3,
	"F4":  ebiten.KeyF4,
	"F5":  ebiten.KeyF5,
	"F6":  ebiten.KeyF6,
	"F7":  ebiten.KeyF7,
	"F8":  ebiten.KeyF8,
	"F9":  ebiten.KeyF9,
	"F10": ebiten.KeyF10,
	"F11": ebiten.KeyF11,
	"F12": ebiten.KeyF12,
}

// parseKeyName converts a string key name to an ebiten.Key.
// Returns ebiten.Key(0) when the name is not recognised.
func parseKeyName(name string) ebiten.Key {
	if key, ok := keyNameToEbiten[strings.ToUpper(name)]; ok {
		return key
	}
	return ebiten.Key(0)
}

// compile-time interface check: *DebugConsole satisfies Node (via embedded *Node2D)
var _ Node = (*DebugConsole)(nil)
