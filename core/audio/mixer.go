package audio

import (
	"bytes"
	"io"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

// ---------------------------------------------------------------------------
// AudioGroup — named bus with its own volume
// ---------------------------------------------------------------------------

// AudioGroup identifies a mixer bus.
type AudioGroup int

const (
	GroupMusic AudioGroup = iota
	GroupSFX
	GroupUI
	GroupCount // sentinel — always last
)

func (g AudioGroup) String() string {
	switch g {
	case GroupMusic:
		return "music"
	case GroupSFX:
		return "sfx"
	case GroupUI:
		return "ui"
	default:
		return "unknown"
	}
}

// ---------------------------------------------------------------------------
// Channel — a single active sound instance
// ---------------------------------------------------------------------------

// ChannelID uniquely identifies an active sound channel within the mixer.
type ChannelID uint64

// ChannelState represents the playback state of a channel.
type ChannelState int

const (
	ChannelPlaying ChannelState = iota
	ChannelPaused
	ChannelStopped
)

// ChannelInfo describes an active or recently stopped sound channel.
type ChannelInfo struct {
	ID       ChannelID
	Group    AudioGroup
	Volume   float64 // 0.0–1.0
	Pitch    float64 // 1.0 = normal
	Pan      float64 // -1.0 (left) to 1.0 (right), 0 = centre
	Loop     bool
	State    ChannelState
	SoundRef string // optional identifier for debugging
}

// channel is an internal tracked sound instance.
type channel struct {
	info    ChannelInfo
	player  *audio.Player
	closeFn func() // called when the channel is removed
}

// ---------------------------------------------------------------------------
// Mixer
// ---------------------------------------------------------------------------

// Mixer manages audio groups and active sound channels.
// It sits alongside the global Manager and provides finer-grained control.
type Mixer struct {
	mu          sync.RWMutex
	ctx         *audio.Context
	channels    map[ChannelID]*channel
	groupVols   [GroupCount]float64 // per-group volume multipliers
	nextID      ChannelID
	listenerX   float64 // world-space X of the listener (for spatial audio)
	listenerY   float64
}

// NewMixer creates a Mixer attached to the given audio context.
func NewMixer(ctx *audio.Context) *Mixer {
	m := &Mixer{
		ctx:       ctx,
		channels:  make(map[ChannelID]*channel),
		nextID:    1,
		listenerX: 0,
		listenerY: 0,
	}
	// Default group volumes: all at 1.0.
	for i := range m.groupVols {
		m.groupVols[i] = 1.0
	}
	return m
}

// ---------------------------------------------------------------------------
// Group volume
// ---------------------------------------------------------------------------

// SetGroupVolume sets the volume multiplier for an audio group (0.0–1.0).
func (m *Mixer) SetGroupVolume(group AudioGroup, vol float64) {
	if vol < 0 {
		vol = 0
	}
	if vol > 1 {
		vol = 1
	}
	m.mu.Lock()
	m.groupVols[group] = vol
	m.mu.Unlock()
}

// GroupVolume returns the current volume multiplier for a group.
func (m *Mixer) GroupVolume(group AudioGroup) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.groupVols[group]
}

// ---------------------------------------------------------------------------
// Listener position (for spatial audio)
// ---------------------------------------------------------------------------

// SetChannelPan updates the stereo pan of an active channel (-1.0 left to 1.0 right).
func (m *Mixer) SetChannelPan(id ChannelID, pan float64) {
	if pan < -1 {
		pan = -1
	}
	if pan > 1 {
		pan = 1
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	ch, ok := m.channels[id]
	if !ok {
		return
	}
	ch.info.Pan = pan
}

// SetListenerPosition sets the world-space position of the audio listener
// (usually the camera or player). Used for spatial/positional audio panning.
func (m *Mixer) SetListenerPosition(x, y float64) {
	m.mu.Lock()
	m.listenerX = x
	m.listenerY = y
	m.mu.Unlock()
}

// ListenerPosition returns the current listener position.
func (m *Mixer) ListenerPosition() (float64, float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.listenerX, m.listenerY
}

// ---------------------------------------------------------------------------
// Playback
// ---------------------------------------------------------------------------

// PlayOpts configures a sound playback request.
type PlayOpts struct {
	Group     AudioGroup // which mixer group (default: GroupSFX)
	Volume    float64    // 0.0–1.0 (default: 1.0)
	Pitch     float64    // 1.0 = normal (default: 1.0)
	Pan       float64    // -1.0 (left) to 1.0 (right), 0 = centre
	Loop      bool       // whether to loop
	SoundX    float64    // world-space X for spatial auto-pan (overrides Pan if non-zero)
	SoundY    float64    // world-space Y (reserved for future 3D audio)
	SoundRef  string     // optional label for debugging
}

// defaultPlayOpts returns a PlayOpts with sensible defaults.
func defaultPlayOpts() PlayOpts {
	return PlayOpts{
		Group:  GroupSFX,
		Volume: 1.0,
		Pitch:  1.0,
		Pan:    0,
		Loop:   false,
	}
}

// Play starts playing a sound with the given options and returns a ChannelID
// that can be used to control playback (stop, pause, set volume, etc.).
// Returns 0 if the sound could not be played.
func (m *Mixer) Play(s *Sound, opts PlayOpts) ChannelID {
	if s == nil {
		return 0
	}

	// Apply defaults for zero-value fields.
	if opts.Volume <= 0 {
		opts.Volume = 1.0
	}
	if opts.Pitch <= 0 {
		opts.Pitch = 1.0
	}

	// Spatial audio: auto-pan based on world position vs listener.
	pan := opts.Pan
	if opts.SoundX != 0 || opts.SoundY != 0 {
		pan = m.computeSpatialPan(opts.SoundX)
	}

	var src io.ReadSeeker = bytes.NewReader(s.data)
	if opts.Loop {
		src = audio.NewInfiniteLoop(bytes.NewReader(s.data), int64(len(s.data)))
	}

	p, err := m.ctx.NewPlayer(src)
	if err != nil {
		return 0
	}

	// Apply group volume + individual volume.
	groupVol := m.groupVols[opts.Group]
	p.SetVolume(opts.Volume * groupVol)

	p.Play()

	id := m.nextID
	m.mu.Lock()
	m.nextID++
	m.channels[id] = &channel{
		info: ChannelInfo{
			ID:       id,
			Group:    opts.Group,
			Volume:   opts.Volume,
			Pitch:    opts.Pitch,
			Pan:      pan,
			Loop:     opts.Loop,
			State:    ChannelPlaying,
			SoundRef: opts.SoundRef,
		},
		player:  p,
		closeFn: func() { p.Close() },
	}
	m.mu.Unlock()

	return id
}

// computeSpatialPan calculates a stereo pan value based on the sound's
// world-space X position relative to the listener.
// Returns -1.0 (full left) to 1.0 (full right).
func (m *Mixer) computeSpatialPan(soundX float64) float64 {
	m.mu.RLock()
	listenerX := m.listenerX
	m.mu.RUnlock()

	dx := soundX - listenerX

	// Attenuation: within ±200px, linear pan; beyond that, hard pan.
	panRange := 200.0
	pan := dx / panRange
	if pan < -1 {
		pan = -1
	}
	if pan > 1 {
		pan = 1
	}
	return pan
}

// ---------------------------------------------------------------------------
// Channel control
// ---------------------------------------------------------------------------

// Stop stops a sound channel. If the channel does not exist, this is a no-op.
func (m *Mixer) Stop(id ChannelID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ch, ok := m.channels[id]
	if !ok {
		return
	}
	if ch.player != nil {
		ch.player.Close()
	}
	ch.info.State = ChannelStopped
	delete(m.channels, id)
}

// Pause pauses a sound channel.
func (m *Mixer) Pause(id ChannelID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ch, ok := m.channels[id]
	if !ok {
		return
	}
	if ch.player != nil {
		ch.player.Pause()
	}
	ch.info.State = ChannelPaused
}

// Resume resumes a paused sound channel.
func (m *Mixer) Resume(id ChannelID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ch, ok := m.channels[id]
	if !ok {
		return
	}
	if ch.player != nil {
		ch.player.Play()
	}
	ch.info.State = ChannelPlaying
}

// SetChannelVolume sets the volume of an active channel (0.0–1.0).
func (m *Mixer) SetChannelVolume(id ChannelID, vol float64) {
	if vol < 0 {
		vol = 0
	}
	if vol > 1 {
		vol = 1
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	ch, ok := m.channels[id]
	if !ok {
		return
	}
	ch.info.Volume = vol
	if ch.player != nil {
		groupVol := m.groupVols[ch.info.Group]
		ch.player.SetVolume(vol * groupVol)
	}
}

// ChannelInfo returns information about a channel, or nil if it doesn't exist.
func (m *Mixer) ChannelInfo(id ChannelID) *ChannelInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ch, ok := m.channels[id]
	if !ok {
		return nil
	}
	// Return a copy.
	info := ch.info
	// Check if the player is still playing (non-looping sounds may have finished).
	if ch.player != nil && !ch.player.IsPlaying() && !ch.info.Loop {
		info.State = ChannelStopped
	}
	return &info
}

// ActiveChannels returns info about all currently tracked channels.
func (m *Mixer) ActiveChannels() []ChannelInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]ChannelInfo, 0, len(m.channels))
	for _, ch := range m.channels {
		info := ch.info
		if ch.player != nil && !ch.player.IsPlaying() && !ch.info.Loop {
			info.State = ChannelStopped
		}
		if info.State != ChannelStopped {
			result = append(result, info)
		}
	}
	return result
}

// ---------------------------------------------------------------------------
// Garbage collection — remove finished channels
// ---------------------------------------------------------------------------

// GC removes channels whose playback has finished (non-looping only).
// Call this periodically (e.g., once per frame) to prevent unbounded growth.
func (m *Mixer) GC() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, ch := range m.channels {
		if ch.player == nil {
			delete(m.channels, id)
			continue
		}
		if ch.info.Loop {
			continue // looping channels stay until explicitly stopped
		}
		if !ch.player.IsPlaying() {
			ch.player.Close()
			delete(m.channels, id)
		}
	}
}

// Len returns the number of currently tracked channels.
func (m *Mixer) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.channels)
}

// Clear stops and removes all channels.
func (m *Mixer) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, ch := range m.channels {
		if ch.player != nil {
			ch.player.Close()
		}
		delete(m.channels, id)
	}
}

// ---------------------------------------------------------------------------
// Global Manager integration
// ---------------------------------------------------------------------------

// mixer is the global Mixer instance, created when Init is called.
var globalMixer *Mixer

// Mixer returns the global Mixer instance. May be nil if Init was not called.
func MixerI() *Mixer {
	return globalMixer
}

// initMixer creates the global Mixer from the Manager's context.
// Called by Manager during Init.
func initMixer(ctx *audio.Context) {
	globalMixer = NewMixer(ctx)
}

// convenience functions using the global mixer
var (
	// GlobalMixer is a shortcut for MixerI().
	GlobalMixer = MixerI
)

// ComputeSpatialPan calculates stereo pan for a sound at worldX relative to the
// global listener. Returns -1.0 (full left) to 1.0 (full right).
// Returns 0 (centre) if the mixer is not initialised.
func ComputeSpatialPan(soundX float64) float64 {
	m := MixerI()
	if m == nil {
		return 0
	}
	return m.computeSpatialPan(soundX)
}

// ListenerPosition returns the current listener position from the global mixer.
func ListenerPosition() (float64, float64) {
	m := MixerI()
	if m == nil {
		return 0, 0
	}
	return m.ListenerPosition()
}

// SetListenerPosition sets the listener position on the global mixer.
func SetListenerPosition(x, y float64) {
	m := MixerI()
	if m != nil {
		m.SetListenerPosition(x, y)
	}
}
