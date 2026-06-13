package editor

import (
	"encoding/json"
	"os"
	"time"
)

// ─── Animation Clip ──────────────────────────────────────────────────────────

// AnimClip is a reusable animation clip that can be played on a Node2D.
// It's the data model behind the animation timeline in the editor.
type AnimClip struct {
	Name       string      `json:"name"`
	Duration   float64     `json:"duration"`   // total duration in seconds
	Loop       bool        `json:"loop"`       // whether to loop
	Speed      float64     `json:"speed"`      // playback speed multiplier (1.0 = normal)
	Tracks     []AnimTrack `json:"tracks"`     // animation tracks
	EntityID   int         `json:"entityId"`   // which entity this anim targets
}

// AnimTrack is a single property track within an animation.
type AnimTrack struct {
	Property string      `json:"property"` // "x", "y", "rotation", "scale_x", "scale_y", "alpha"
	Keyframes []Keyframe `json:"keyframes"`
}

// Keyframe defines a value at a point in time with an easing function.
type Keyframe struct {
	Time   float64 `json:"time"`   // seconds from start
	Value  float64 `json:"value"`  // target value
	Easing string  `json:"easing"` // "linear", "in_quad", "out_quad", "in_out_quad", etc.
}

// ─── Timeline State ──────────────────────────────────────────────────────────

// TimelineState represents the current playback state in the editor.
type TimelineState struct {
	Clip        *AnimClip    `json:"-"`
	IsPlaying   bool         `json:"-"`
	CurrentTime float64      `json:"-"`
	PlaySpeed   float64      `json:"-"` // 0.25, 0.5, 1.0, 2.0
	Loop        bool         `json:"-"`
	Zoom        float64      `json:"-"` // horizontal zoom of the timeline UI
	ScrollX     float64      `json:"-"` // horizontal scroll offset

	startTime time.Time `json:"-"` // when playback started
	pauseTime float64   `json:"-"` // accumulated time when paused
}

// NewTimelineState creates a new timeline state for the given clip.
func NewTimelineState(clip *AnimClip) *TimelineState {
	return &TimelineState{
		Clip:        clip,
		PlaySpeed:   1.0,
		Loop:        clip != nil && clip.Loop,
		Zoom:        1.0,
		CurrentTime: 0,
	}
}

// Play starts playback from the current position.
func (ts *TimelineState) Play() {
	if ts == nil || ts.Clip == nil {
		return
	}
	if ts.IsPlaying {
		return
	}
	ts.IsPlaying = true
	ts.startTime = time.Now()
}

// Pause pauses playback at the current position.
func (ts *TimelineState) Pause() {
	if !ts.IsPlaying {
		return
	}
	ts.IsPlaying = false
	ts.pauseTime = ts.CurrentTime
}

// Stop stops playback and resets to time 0.
func (ts *TimelineState) Stop() {
	ts.IsPlaying = false
	ts.CurrentTime = 0
}

// Seek moves the playhead to the given time.
func (ts *TimelineState) Seek(t float64) {
	if ts.Clip == nil {
		return
	}
	if t < 0 {
		t = 0
	}
	if t > ts.Clip.Duration {
		t = ts.Clip.Duration
	}
	ts.CurrentTime = t
	ts.startTime = time.Now()
	ts.pauseTime = t
}

// Tick advances the timeline by dt seconds. Returns true if the clip
// is still playing, false if it finished (non-looping).
func (ts *TimelineState) Tick(dt float64) bool {
	if ts == nil || ts.Clip == nil || !ts.IsPlaying {
		return ts != nil && ts.IsPlaying
	}

	dt *= ts.PlaySpeed
	ts.CurrentTime += dt

	if ts.CurrentTime >= ts.Clip.Duration {
		if ts.Loop {
			ts.CurrentTime = 0
		} else {
			ts.CurrentTime = ts.Clip.Duration
			ts.IsPlaying = false
			return false
		}
	}

	return true
}

// GetCurrentKeyframes returns all keyframes that are active at the current time,
// grouped by track. Returns the previous and next keyframe for each track.
type ActiveKeyframes struct {
	TrackProperty string
	PrevKey       *Keyframe // keyframe before current time (or nil)
	NextKey       *Keyframe // keyframe after current time (or nil)
	T             float64   // interpolation factor [0,1] between prev and next
}

// GetActiveKeyframes returns the active keyframes for interpolation at the current time.
func (ts *TimelineState) GetActiveKeyframes() []ActiveKeyframes {
	if ts.Clip == nil {
		return nil
	}

	result := make([]ActiveKeyframes, 0, len(ts.Clip.Tracks))
	t := ts.CurrentTime

	for _, track := range ts.Clip.Tracks {
		if len(track.Keyframes) == 0 {
			continue
		}

		ak := ActiveKeyframes{TrackProperty: track.Property}

		// Find previous and next keyframe
		for i, kf := range track.Keyframes {
			if kf.Time <= t {
				ak.PrevKey = &track.Keyframes[i]
			}
			if kf.Time >= t && ak.NextKey == nil {
				ak.NextKey = &track.Keyframes[i]
			}
		}

		// Compute interpolation factor
		if ak.PrevKey != nil && ak.NextKey != nil && ak.NextKey.Time != ak.PrevKey.Time {
			ak.T = (t - ak.PrevKey.Time) / (ak.NextKey.Time - ak.PrevKey.Time)
			if ak.T < 0 {
				ak.T = 0
			}
			if ak.T > 1 {
				ak.T = 1
			}
		}

		result = append(result, ak)
	}

	return result
}

// ─── Clip File I/O ───────────────────────────────────────────────────────────

// SaveAnimClip saves an animation clip to a .kora.anim file.
func SaveAnimClip(clip *AnimClip, path string) error {
	data, err := json.MarshalIndent(clip, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadAnimClip loads an animation clip from a .kora.anim file.
func LoadAnimClip(path string) (*AnimClip, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var clip AnimClip
	if err := json.Unmarshal(data, &clip); err != nil {
		return nil, err
	}
	return &clip, nil
}

// ─── Default Presets ─────────────────────────────────────────────────────────

// NewDefaultClip creates a new clip with default empty tracks.
func NewDefaultClip(name string, duration float64, entityID int) *AnimClip {
	return &AnimClip{
		Name:     name,
		Duration: duration,
		Speed:    1.0,
		Loop:     false,
		EntityID: entityID,
		Tracks: []AnimTrack{
			{Property: "x", Keyframes: []Keyframe{{Time: 0, Value: 0, Easing: "linear"}, {Time: duration, Value: 0, Easing: "linear"}}},
			{Property: "y", Keyframes: []Keyframe{{Time: 0, Value: 0, Easing: "linear"}, {Time: duration, Value: 0, Easing: "linear"}}},
			{Property: "rotation", Keyframes: []Keyframe{{Time: 0, Value: 0, Easing: "linear"}, {Time: duration, Value: 0, Easing: "linear"}}},
			{Property: "alpha", Keyframes: []Keyframe{{Time: 0, Value: 1, Easing: "linear"}, {Time: duration, Value: 1, Easing: "linear"}}},
		},
	}
}

// ─── Easing Name Resolution ──────────────────────────────────────────────────

// EasingNames returns all available easing function names for the UI.
func EasingNames() []string {
	return []string{
		"linear", "in_quad", "out_quad", "in_out_quad",
		"in_cubic", "out_cubic", "in_out_cubic",
		"in_quart", "out_quart", "in_out_quart",
		"in_quint", "out_quint", "in_out_quint",
		"in_sine", "out_sine", "in_out_sine",
		"in_expo", "out_expo", "in_out_expo",
		"in_circ", "out_circ", "in_out_circ",
		"in_elastic", "out_elastic", "in_out_elastic",
		"in_back", "out_back", "in_out_back",
		"in_bounce", "out_bounce", "in_out_bounce",
	}
}
