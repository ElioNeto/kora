// Package audio provides a simple sound manager for the Kora engine.
//
// Built on top of Ebitengine’s audio package (github.com/hajimehoshi/ebiten/v2/audio).
//
// Supported formats via helper loaders: OGG Vorbis, WAV, MP3.
//
// Usage:
//
//	audio.Init(44100)
//	bgm, _ := audio.LoadOGG("assets/bgm.ogg")
//	audio.PlayBGM(bgm, true)    // looping
//
//	sfx, _ := audio.LoadWAV("assets/jump.wav")
//	audio.PlaySFX(sfx, 1.0)     // volume 0..1
package audio

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

// ----------------------------------------------------------------------------
// Manager singleton
// ----------------------------------------------------------------------------

var mgr *Manager

// Init initialises the global audio manager with the given sample rate.
// Must be called once before any other audio function.
func Init(sampleRate int) error {
	var err error
	mgr, err = NewManager(sampleRate)
	if err == nil {
		initMixer(mgr.ctx)
	}
	return err
}

// Manager wraps an Ebitengine audio.Context and owns all players.
type Manager struct {
	ctx    *audio.Context
	bgm    *audio.Player
	sfxVol float64
	bgmVol float64
	sounds map[string]*Sound // cached sounds loaded by name/path
}

// NewManager creates a Manager with the given sample rate.
func NewManager(sampleRate int) (*Manager, error) {
	ctx := audio.NewContext(sampleRate)
	return &Manager{ctx: ctx, sfxVol: 1.0, bgmVol: 1.0, sounds: make(map[string]*Sound)}, nil
}

// Context returns the underlying Ebitengine audio context.
func (m *Manager) Context() *audio.Context { return m.ctx }

// ----------------------------------------------------------------------------
// Sound data
// ----------------------------------------------------------------------------

// Sound holds decoded PCM bytes ready for playback.
type Sound struct {
	data []byte
	rate int
}

// LoadOGG decodes an OGG Vorbis file from path.
func LoadOGG(path string) (*Sound, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	decoded, err := vorbis.DecodeWithoutResampling(f)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(decoded)
	if err != nil {
		return nil, err
	}
	return &Sound{data: data, rate: mgr.ctx.SampleRate()}, nil
}

// LoadWAV decodes a WAV file from path.
func LoadWAV(path string) (*Sound, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	decoded, err := wav.DecodeWithoutResampling(f)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(decoded)
	if err != nil {
		return nil, err
	}
	return &Sound{data: data, rate: mgr.ctx.SampleRate()}, nil
}

// LoadMP3 decodes an MP3 file from path.
func LoadMP3(path string) (*Sound, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	decoded, err := mp3.DecodeWithoutResampling(f)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(decoded)
	if err != nil {
		return nil, err
	}
	return &Sound{data: data, rate: mgr.ctx.SampleRate()}, nil
}

// LoadOGGBytes decodes OGG from an in-memory byte slice (e.g. embed.FS).
func LoadOGGBytes(b []byte) (*Sound, error) {
	decoded, err := vorbis.DecodeWithoutResampling(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(decoded)
	if err != nil {
		return nil, err
	}
	return &Sound{data: data, rate: mgr.ctx.SampleRate()}, nil
}

// ----------------------------------------------------------------------------
// BGM
// ----------------------------------------------------------------------------

// PlayBGM starts the background music player. If loop is true, it repeats.
// Any previously playing BGM is stopped first.
func PlayBGM(s *Sound, loop bool) error {
	if mgr == nil {
		return nil
	}
	return mgr.PlayBGM(s, loop)
}

func (m *Manager) PlayBGM(s *Sound, loop bool) error {
	if s == nil {
		return nil
	}
	if m.bgm != nil {
		_ = m.bgm.Close()
		m.bgm = nil
	}
	var src io.ReadSeeker = bytes.NewReader(s.data)
	if loop {
		src = audio.NewInfiniteLoop(bytes.NewReader(s.data), int64(len(s.data)))
	}
	p, err := m.ctx.NewPlayer(src)
	if err != nil {
		return err
	}
	p.SetVolume(m.bgmVol)
	p.Play()
	m.bgm = p
	return nil
}

// StopBGM stops the current background music.
func StopBGM() {
	if mgr != nil && mgr.bgm != nil {
		_ = mgr.bgm.Close()
		mgr.bgm = nil
	}
}

// SetBGMVolume sets the BGM volume (0.0 – 1.0).
func SetBGMVolume(v float64) {
	if mgr == nil {
		return
	}
	mgr.bgmVol = v
	if mgr.bgm != nil {
		mgr.bgm.SetVolume(v)
	}
}

// ----------------------------------------------------------------------------
// SFX
// ----------------------------------------------------------------------------

// PlaySFX plays a sound effect at the given volume (0.0 – 1.0).
// Each call creates a short-lived player that is garbage-collected when done.
func PlaySFX(s *Sound, volume float64) error {
	if mgr == nil {
		return nil
	}
	return mgr.PlaySFX(s, volume)
}

func (m *Manager) PlaySFX(s *Sound, volume float64) error {
	if s == nil {
		return nil
	}
	p, err := m.ctx.NewPlayer(bytes.NewReader(s.data))
	if err != nil {
		return err
	}
	p.SetVolume(volume * m.sfxVol)
	p.Play()
	// The player is intentionally not stored — Ebitengine keeps it alive
	// internally until playback finishes.
	return nil
}

// SetSFXVolume sets a global multiplier for all SFX (0.0 – 1.0).
func SetSFXVolume(v float64) {
	if mgr != nil {
		mgr.sfxVol = v
	}
}

// ----------------------------------------------------------------------------
// Node sound integration (for AudioPlayer2D and similar)
// ----------------------------------------------------------------------------

// GetSound returns a previously cached sound by name, or nil.
func (m *Manager) GetSound(name string) *Sound {
	if m.sounds == nil {
		return nil
	}
	return m.sounds[name]
}

// SetSound stores a sound in the manager's cache by name.
func (m *Manager) SetSound(name string, s *Sound) {
	if m.sounds == nil {
		m.sounds = make(map[string]*Sound)
	}
	m.sounds[name] = s
}

// LoadSound loads a sound file from path, caches it, and returns it.
// The file extension determines the decoder (.ogg, .wav, .mp3).
// If the sound is already cached it is returned immediately.
func (m *Manager) LoadSound(path string) (*Sound, error) {
	if s, ok := m.sounds[path]; ok {
		return s, nil
	}
	lower := strings.ToLower(path)
	var s *Sound
	var err error
	switch {
	case strings.HasSuffix(lower, ".ogg"):
		s, err = LoadOGG(path)
	case strings.HasSuffix(lower, ".wav"):
		s, err = LoadWAV(path)
	case strings.HasSuffix(lower, ".mp3"):
		s, err = LoadMP3(path)
	default:
		return nil, fmt.Errorf("unsupported audio format: %s", path)
	}
	if err != nil {
		return nil, err
	}
	m.sounds[path] = s
	return s, nil
}

// PlayNodeSound plays a sound by name with the given settings.
// If the sound is not yet cached, it attempts to load it from the path.
// Returns a channel ID (int) that can be used for control (Stop, Pause, Resume).
// Returns 0 if the sound cannot be found or the mixer is not initialised.
func (m *Manager) PlayNodeSound(name string, volume float64, loop bool, pan float64) int {
	mixer := MixerI()
	if mixer == nil {
		return 0
	}
	s := m.GetSound(name)
	if s == nil {
		// Attempt on-demand load (name may be a file path)
		var err error
		s, err = m.LoadSound(name)
		if err != nil {
			return 0
		}
	}
	id := mixer.Play(s, PlayOpts{
		Group:  GroupSFX,
		Volume: volume,
		Pan:    pan,
		Loop:   loop,
	})
	return int(id)
}

// global helpers (nil-safe)

// PreloadSound loads and caches a sound file via the global manager.
func PreloadSound(path string) error {
	if mgr == nil {
		return nil
	}
	_, err := mgr.LoadSound(path)
	return err
}

// PlayNodeSound plays a cached sound by name via the global manager.
// Returns 0 if not initialised or sound not found.
func PlayNodeSound(name string, volume float64, loop bool, pan float64) int {
	if mgr == nil {
		return 0
	}
	return mgr.PlayNodeSound(name, volume, loop, pan)
}
