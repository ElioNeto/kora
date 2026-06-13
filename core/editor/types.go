// Package editor provides shared types and conversion between the
// editor data model (SceneEntity) and the runtime Node2D tree.
//
// This is the bridge that enables:
//   - Loading editor scenes in the runtime (preview/play)
//   - Saving runtime scenes back to editor format
//   - Bidirectional synchronisation between editor and runtime
package editor

// SceneEntity represents a single entity in the editor's scene graph.
// Must match the JSON schema of .kora.json files.
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

// SceneMeta holds metadata for the scene file.
type SceneMeta struct {
	Name     string `json:"name"`
	Version  int    `json:"version"`
	LogicalW int    `json:"logicalW"`
	LogicalH int    `json:"logicalH"`
}

// SceneFile is the top-level container for a .kora.json scene.
type SceneFile struct {
	Meta     SceneMeta      `json:"meta"`
	Entities []*SceneEntity `json:"entities"`
}

// EditorTab represents which tab is active in the editor UI.
type EditorTab int

const (
	TabScene   EditorTab = iota // Scene viewport
	TabAssets                   // Asset browser
	TabCode                     // KScript editor
	TabPreview                  // Game preview
	TabAnim                     // Animation timeline
)

// EditorTool represents the active tool in the editor.
type EditorTool int

const (
	ToolSelect EditorTool = iota
	ToolMove
	ToolScale
)

// UndoEntry is a single entry in the undo history.
type UndoEntry struct {
	Snapshot string // JSON snapshot of the scene
}
