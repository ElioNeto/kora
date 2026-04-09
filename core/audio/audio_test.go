package audio_test

import (
	"testing"

	"github.com/ElioNeto/kora/core/audio"
)

// Ebitengine’s audio context requires a real audio device which is not
// available in headless CI. We test only safe, device-free code paths.

func TestPlaySFXNilManager(t *testing.T) {
	// audio.Init not called — PlaySFX should be a no-op, not panic.
	err := audio.PlaySFX(nil, 1.0)
	if err != nil {
		t.Errorf("expected nil error when manager not initialised, got %v", err)
	}
}

func TestPlayBGMNilManager(t *testing.T) {
	err := audio.PlayBGM(nil, false)
	if err != nil {
		t.Errorf("expected nil error when manager not initialised, got %v", err)
	}
}

func TestStopBGMNilManager(t *testing.T) {
	audio.StopBGM() // must not panic
}

func TestSetVolumesNilManager(t *testing.T) {
	audio.SetBGMVolume(0.5) // must not panic
	audio.SetSFXVolume(0.5) // must not panic
}
