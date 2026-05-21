package node

import (
	"github.com/ElioNeto/kora/core/math"
)

// ---------------------------------------------------------------------------
// Action types
// ---------------------------------------------------------------------------

// CutsceneActionType defines what the action does.
type CutsceneActionType int

const (
	ActionWait CutsceneActionType = iota
	ActionMoveTo
	ActionFadeIn
	ActionFadeOut
	ActionShowText
	ActionHideText
	ActionPlayAnimation
	ActionSetVisible
	ActionRunScript
	ActionSetCamera
	ActionPlaySound
	ActionWaitForSignal
)

// ---------------------------------------------------------------------------
// Action definition
// ---------------------------------------------------------------------------

// CutsceneAction is a single step in a cutscene timeline.
type CutsceneAction struct {
	Type     CutsceneActionType
	Duration float64 // seconds (for actions that take time)
	Target   string  // node path target

	FromX, FromY, ToX, ToY float32
	TargetX, TargetY       float32

	Text       string
	Value      float64
	BoolVal    bool
	Signal     string
	Script     string
	EasingType string // "linear", "ease_in", "ease_out", "ease_in_out"
}

// ---------------------------------------------------------------------------
// Scene interface
// ---------------------------------------------------------------------------

// Scene is used by CutscenePlayer to find nodes and check signals.
type Scene interface {
	// FindNode returns a node by path. The return type is interface{} so that
	// concrete node types (e.g. *Sprite2D) can be returned even if they don't
	// satisfy the Node interface.
	FindNode(path string) interface{}
	SignalFired(name string) bool
}

// ---------------------------------------------------------------------------
// Internal action state helpers
// ---------------------------------------------------------------------------

// moveToState stores the starting position for MoveTo interpolation.
type moveToState struct {
	startPos math.Vector2
}

// fadeState stores the starting alpha for FadeIn/FadeOut interpolation.
type fadeState struct {
	startAlpha float64
}

// ---------------------------------------------------------------------------
// CutscenePlayer
// ---------------------------------------------------------------------------

// CutscenePlayer plays through a sequence of cutscene actions in order.
//
// Usage:
//
//	player := NewCutscenePlayer("intro")
//	player.AddAction(WaitAction(1.0))
//	player.AddAction(MoveToAction("Player", 100, 200, 2.0, "ease_out"))
//	player.Play()
type CutscenePlayer struct {
	*Node2D

	actions     []CutsceneAction
	currentIdx  int
	elapsed     float64
	isPlaying   bool
	paused      bool
	loop        bool
	onComplete  func()
	scene       Scene
	actionState map[int]interface{} // per-action state for interpolation
}

// NewCutscenePlayer creates a new CutscenePlayer node.
func NewCutscenePlayer(name string) *CutscenePlayer {
	return &CutscenePlayer{
		Node2D:      NewNode2D(name, 0),
		actionState: make(map[int]interface{}),
	}
}

// AddAction adds an action to the cutscene timeline.
func (cp *CutscenePlayer) AddAction(action CutsceneAction) {
	cp.actions = append(cp.actions, action)
}

// Play starts the cutscene from the beginning.
func (cp *CutscenePlayer) Play() {
	cp.currentIdx = 0
	cp.elapsed = 0
	cp.isPlaying = true
	cp.paused = false
	cp.actionState = make(map[int]interface{})
}

// Stop stops playback and resets to the beginning.
func (cp *CutscenePlayer) Stop() {
	cp.isPlaying = false
	cp.paused = false
	cp.currentIdx = 0
	cp.elapsed = 0
	cp.actionState = make(map[int]interface{})
}

// Pause pauses playback at the current position.
func (cp *CutscenePlayer) Pause() {
	if cp.isPlaying {
		cp.paused = true
	}
}

// Resume resumes from the paused position.
func (cp *CutscenePlayer) Resume() {
	if cp.paused {
		cp.paused = false
	}
}

// IsPlaying returns whether the cutscene is currently active and not paused.
func (cp *CutscenePlayer) IsPlaying() bool {
	return cp.isPlaying && !cp.paused
}

// SetLoop sets whether the cutscene loops back to the beginning when finished.
func (cp *CutscenePlayer) SetLoop(loop bool) {
	cp.loop = loop
}

// SetOnComplete sets a callback that fires when the cutscene finishes
// (not called in loop mode).
func (cp *CutscenePlayer) SetOnComplete(fn func()) {
	cp.onComplete = fn
}

// SetScene sets the scene reference used for finding nodes by path and
// checking signals.
func (cp *CutscenePlayer) SetScene(s Scene) {
	cp.scene = s
}

// GetCurrentAction returns the index of the current action being processed.
func (cp *CutscenePlayer) GetCurrentAction() int {
	return cp.currentIdx
}

// GetTotalActions returns the total number of actions in the timeline.
func (cp *CutscenePlayer) GetTotalActions() int {
	return len(cp.actions)
}

// GetProgress returns the cutscene progress as a value between 0.0 and 1.0.
// Returns 1.0 when there are no actions.
func (cp *CutscenePlayer) GetProgress() float64 {
	if len(cp.actions) == 0 {
		return 1.0
	}
	return float64(cp.currentIdx) / float64(len(cp.actions))
}

// Clear removes all actions and resets the player to its initial state.
func (cp *CutscenePlayer) Clear() {
	cp.actions = nil
	cp.currentIdx = 0
	cp.elapsed = 0
	cp.isPlaying = false
	cp.paused = false
	cp.actionState = make(map[int]interface{})
}

// SkipToEnd jumps to the end of the cutscene, applying the final state of
// all timed actions, and firing the onComplete callback.
func (cp *CutscenePlayer) SkipToEnd() {
	if len(cp.actions) == 0 {
		return
	}

	// Apply final state of every action that modifies node properties.
	for i, action := range cp.actions {
		switch action.Type {
		case ActionMoveTo, ActionFadeIn, ActionFadeOut:
			cp.currentIdx = i
			if _, exists := cp.actionState[i]; !exists {
				cp.initActionState(action)
			}
			cp.applyActionAt(action, 1.0)
		case ActionSetVisible:
			if n := cp.findTarget(action.Target); n != nil {
				n.SetVisible(action.BoolVal)
			}
		default:
			// Wait, ShowText, HideText, RunScript, PlaySound,
			// PlayAnimation, SetCamera, WaitForSignal are skipped.
		}
	}

	cp.currentIdx = len(cp.actions)
	cp.finish()
}

// ---------------------------------------------------------------------------
// Update — drives the cutscene timeline
// ---------------------------------------------------------------------------

// Update processes the cutscene actions. It implements the Node interface
// and propagates to children first.
func (cp *CutscenePlayer) Update(dt float64) {
	// Propagate to children first
	if cp.Node2D != nil {
		cp.Node2D.Update(dt)
	}

	if !cp.isPlaying || cp.paused {
		return
	}

	// Handle end-of-list: loop back or finish.
	if cp.tryHandleEnd() {
		return
	}

	action := cp.actions[cp.currentIdx]

	// Initialize per-action state if this is the first time we see this index.
	if _, exists := cp.actionState[cp.currentIdx]; !exists {
		cp.initActionState(action)
	}

	// Instant actions fire once and advance immediately.
	if cp.isInstantAction(action.Type) {
		cp.applyActionAt(action, 1.0)
		cp.advance()
		cp.tryHandleEnd()
		return
	}

	// WaitForSignal: block until the scene reports the signal as fired.
	if action.Type == ActionWaitForSignal {
		if cp.scene != nil && cp.scene.SignalFired(action.Signal) {
			cp.advance()
			cp.tryHandleEnd()
		}
		return
	}

	// Timed actions (Wait, MoveTo, FadeIn, FadeOut)
	cp.elapsed += dt

	if cp.elapsed >= action.Duration {
		cp.applyActionAt(action, 1.0)
		cp.advance()
		cp.tryHandleEnd()
		return
	}

	t := cp.elapsed / action.Duration
	cp.applyActionAt(action, t)
}

// tryHandleEnd checks whether the cutscene has reached the end of its action
// list. In loop mode it resets to the beginning; otherwise it calls finish().
// Returns true when the caller should return from Update (no more processing
// on this frame).
func (cp *CutscenePlayer) tryHandleEnd() bool {
	if cp.currentIdx >= len(cp.actions) {
		if cp.loop {
			cp.currentIdx = 0
			cp.elapsed = 0
			cp.actionState = make(map[int]interface{})
		} else {
			cp.finish()
		}
		return true
	}
	return false
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// finish stops playback and fires the completion callback.
func (cp *CutscenePlayer) finish() {
	cp.isPlaying = false
	cp.elapsed = 0
	if cp.onComplete != nil {
		cp.onComplete()
	}
}

// advance moves to the next action in the timeline.
func (cp *CutscenePlayer) advance() {
	cp.currentIdx++
	cp.elapsed = 0
}

// isInstantAction returns true for actions that fire once and advance
// immediately without waiting for a duration.
func (cp *CutscenePlayer) isInstantAction(actionType CutsceneActionType) bool {
	switch actionType {
	case ActionShowText, ActionHideText, ActionSetVisible,
		ActionRunScript, ActionPlaySound, ActionPlayAnimation, ActionSetCamera:
		return true
	default:
		return false
	}
}

// initActionState captures the necessary starting values for an action
// that requires interpolation over time.
func (cp *CutscenePlayer) initActionState(action CutsceneAction) {
	switch action.Type {
	case ActionMoveTo:
		state := &moveToState{}
		if n := cp.findTarget(action.Target); n != nil {
			state.startPos = n.GetPosition()
		} else {
			state.startPos = math.NewVector2(action.FromX, action.FromY)
		}
		cp.actionState[cp.currentIdx] = state

	case ActionFadeIn:
		startAlpha := float64(0)
		if n := cp.findTarget(action.Target); n != nil {
			// Alpha is tracked per-action; reading from Node2D is placeholder
			// — full implementation requires Sprite2D scene integration.
			_ = n
		}
		cp.actionState[cp.currentIdx] = &fadeState{startAlpha: startAlpha}

	case ActionFadeOut:
		startAlpha := float64(1.0)
		if n := cp.findTarget(action.Target); n != nil {
			_ = n
		}
		cp.actionState[cp.currentIdx] = &fadeState{startAlpha: startAlpha}
	}
}

// applyActionAt applies the effect of an action at normalised time t ∈ [0,1].
func (cp *CutscenePlayer) applyActionAt(action CutsceneAction, t float64) {
	easing := EasingByName(action.EasingType)
	easedT := easing(t)

	switch action.Type {
	case ActionMoveTo:
		state, ok := cp.actionState[cp.currentIdx].(*moveToState)
		if !ok {
			return
		}
		x := float64(state.startPos.X) + (float64(action.ToX)-float64(state.startPos.X))*easedT
		y := float64(state.startPos.Y) + (float64(action.ToY)-float64(state.startPos.Y))*easedT
		if n := cp.findTarget(action.Target); n != nil {
			n.SetPosition(float32(x), float32(y))
		}

	case ActionFadeIn:
		state, ok := cp.actionState[cp.currentIdx].(*fadeState)
		if !ok {
			return
		}
		alpha := state.startAlpha + (1.0-state.startAlpha)*easedT
		cp.setAlpha(action.Target, float32(alpha))

	case ActionFadeOut:
		state, ok := cp.actionState[cp.currentIdx].(*fadeState)
		if !ok {
			return
		}
		alpha := state.startAlpha + (0.0-state.startAlpha)*easedT
		cp.setAlpha(action.Target, float32(alpha))

	case ActionSetVisible:
		if n := cp.findTarget(action.Target); n != nil {
			n.SetVisible(action.BoolVal)
		}

	case ActionWait:
		// Wait does not modify any node property.

	default:
		// ShowText, HideText, PlayAnimation, RunScript, SetCamera,
		// PlaySound, WaitForSignal are applied as no-ops at this level;
		// the scene / game code is expected to handle them via the
		// Scene interface or external callbacks.
	}
}

// findTarget resolves a node path to a *Node2D, first attempting the Scene
// interface and then falling back to the node tree (parent lookup).
func (cp *CutscenePlayer) findTarget(path string) *Node2D {
	v := cp.findTargetNode(path)
	if v == nil {
		return nil
	}
	// Try direct Node assertion (works for *Node2D, PhysicsBody2D, etc.)
	if n, ok := v.(Node); ok {
		return extractNode2D(n)
	}
	// Try node2DProvider interface (works for *Sprite2D, which embeds *Node2D
	// but does not satisfy Node due to its Draw signature)
	if p, ok := v.(interface{ GetNode2D() *Node2D }); ok {
		return p.GetNode2D()
	}
	return nil
}

// findTargetNode resolves a node path to an interface{} value, preserving
// the original concrete type. Callers can type-assert as needed.
func (cp *CutscenePlayer) findTargetNode(path string) interface{} {
	if path == "" {
		return nil
	}

	// Try the Scene interface first — the Scene may return typed nodes
	// such as *Sprite2D directly, preserving the concrete type.
	if cp.scene != nil {
		if n := cp.scene.FindNode(path); n != nil {
			return n
		}
	}

	// Fall back to node tree lookup relative to the parent.
	// Note: the tree stores nodes as *Node2D, so concrete types like
	// *Sprite2D are NOT preserved through this path.
	parent := cp.GetParent()
	if parent == nil {
		return nil
	}
	if child := parent.GetChild(path); child != nil {
		return child
	}
	if n := parent.GetNode(path); n != nil {
		return n
	}
	return nil
}

// setAlpha sets the alpha on a target identified by path.
// The alpha is tracked per-action; actual sprite alpha application
// requires scene-level integration and is a placeholder for now.
func (cp *CutscenePlayer) setAlpha(path string, alpha float32) {
	_ = path
	_ = alpha
}

// ---------------------------------------------------------------------------
// Helper functions for creating common actions
// ---------------------------------------------------------------------------

// WaitAction creates an action that waits for the given duration in seconds.
func WaitAction(duration float64) CutsceneAction {
	return CutsceneAction{
		Type:     ActionWait,
		Duration: duration,
	}
}

// MoveToAction creates an action that moves a node to (x, y) over duration.
func MoveToAction(target string, x, y float32, duration float64, easing string) CutsceneAction {
	return CutsceneAction{
		Type:       ActionMoveTo,
		Target:     target,
		ToX:        x,
		ToY:        y,
		Duration:   duration,
		EasingType: easing,
	}
}

// FadeInAction creates an action that fades a sprite in over duration.
func FadeInAction(target string, duration float64) CutsceneAction {
	return CutsceneAction{
		Type:       ActionFadeIn,
		Target:     target,
		Duration:   duration,
		EasingType: "linear",
	}
}

// FadeOutAction creates an action that fades a sprite out over duration.
func FadeOutAction(target string, duration float64) CutsceneAction {
	return CutsceneAction{
		Type:       ActionFadeOut,
		Target:     target,
		Duration:   duration,
		EasingType: "linear",
	}
}

// ShowTextAction creates an action that shows text.
func ShowTextAction(text string, duration float64) CutsceneAction {
	return CutsceneAction{
		Type:     ActionShowText,
		Text:     text,
		Duration: duration,
	}
}

// HideTextAction creates an action that hides the current text.
func HideTextAction() CutsceneAction {
	return CutsceneAction{
		Type: ActionHideText,
	}
}

// PlayAnimationAction creates an action that plays an animation on a target node.
func PlayAnimationAction(target string, animationName string) CutsceneAction {
	return CutsceneAction{
		Type:   ActionPlayAnimation,
		Target: target,
		Text:   animationName,
	}
}

// SetVisibleAction creates an action that sets a node's visibility.
func SetVisibleAction(target string, visible bool) CutsceneAction {
	return CutsceneAction{
		Type:    ActionSetVisible,
		Target:  target,
		BoolVal: visible,
	}
}

// RunScriptAction creates an action that executes a script.
func RunScriptAction(script string) CutsceneAction {
	return CutsceneAction{
		Type:   ActionRunScript,
		Script: script,
	}
}

// PlaySoundAction creates an action that plays a sound.
func PlaySoundAction(soundName string) CutsceneAction {
	return CutsceneAction{
		Type: ActionPlaySound,
		Text: soundName,
	}
}

// WaitForSignalAction creates an action that waits until a signal fires.
func WaitForSignalAction(signal string) CutsceneAction {
	return CutsceneAction{
		Type:   ActionWaitForSignal,
		Signal: signal,
	}
}

// Compile-time interface check: CutscenePlayer must satisfy Node.
var _ Node = (*CutscenePlayer)(nil)
