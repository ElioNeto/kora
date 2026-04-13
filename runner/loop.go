package runner

import (
	"github.com/ElioNeto/kora/core/audio"
)

// init extends the Game struct fields (Go doesn’t allow partial struct defs,
// so we use a build-tag-free extension pattern via embedding-like fields
// declared once here and referenced in game.go via the same package).
func init() {} // keeps the file non-empty for Go tooling

// AudioConfig configures the audio subsystem from within the runner.
type AudioConfig struct {
	SampleRate int // default 44100
}

// InitAudio initialises the global audio manager.
// Call before Run if the game uses sound.
func InitAudio(cfg AudioConfig) error {
	rate := cfg.SampleRate
	if rate == 0 {
		rate = 44100
	}
	return audio.Init(rate)
}
