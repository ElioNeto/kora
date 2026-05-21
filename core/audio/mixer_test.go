package audio

import (
	"sync"
	"testing"
)

// initOnce ensures Init is called only once across all tests.
var initOnce sync.Once

func initMixerForTest() {
	initOnce.Do(func() {
		Init(44100)
	})
}

func TestNewMixer(t *testing.T) {
	initMixerForTest()
	m := MixerI()
	if m == nil {
		t.Fatal("expected non-nil mixer")
	}
	if m.Len() != 0 {
		t.Errorf("expected 0 channels, got %d", m.Len())
	}
}

func TestGroupVolume(t *testing.T) {
	initMixerForTest()
	m := MixerI()

	// Default volumes should be 1.0
	for i := GroupMusic; i < GroupCount; i++ {
		if v := m.GroupVolume(i); v != 1.0 {
			t.Errorf("expected group %d volume 1.0, got %f", i, v)
		}
	}

	// Set and verify
	m.SetGroupVolume(GroupSFX, 0.5)
	if v := m.GroupVolume(GroupSFX); v != 0.5 {
		t.Errorf("expected SFX volume 0.5, got %f", v)
	}

	// Clamp to [0, 1]
	m.SetGroupVolume(GroupSFX, -0.1)
	if v := m.GroupVolume(GroupSFX); v != 0 {
		t.Errorf("expected SFX volume clamped to 0, got %f", v)
	}
	m.SetGroupVolume(GroupSFX, 2.0)
	if v := m.GroupVolume(GroupSFX); v != 1 {
		t.Errorf("expected SFX volume clamped to 1, got %f", v)
	}
}

func TestGroupString(t *testing.T) {
	tests := []struct {
		g   AudioGroup
		str string
	}{
		{GroupMusic, "music"},
		{GroupSFX, "sfx"},
		{GroupUI, "ui"},
		{AudioGroup(99), "unknown"},
	}
	for _, tt := range tests {
		if tt.g.String() != tt.str {
			t.Errorf("expected %q, got %q", tt.str, tt.g.String())
		}
	}
}

func TestListenerPosition(t *testing.T) {
	initMixerForTest()
	m := MixerI()

	m.SetListenerPosition(100, 200)
	x, y := m.ListenerPosition()
	if x != 100 || y != 200 {
		t.Errorf("expected (100, 200), got (%f, %f)", x, y)
	}
}

func TestComputeSpatialPan(t *testing.T) {
	initMixerForTest()
	m := MixerI()
	m.SetListenerPosition(0, 0)

	tests := []struct {
		soundX float64
		want   float64
	}{
		{0, 0},       // same position -> centre
		{100, 0.5},   // 100px right -> 50% right
		{-100, -0.5}, // 100px left -> 50% left
		{300, 1.0},   // beyond range -> full right
		{-300, -1.0}, // beyond range -> full left
	}

	for _, tt := range tests {
		got := m.computeSpatialPan(tt.soundX)
		if got != tt.want {
			t.Errorf("computeSpatialPan(%f) = %f, want %f", tt.soundX, got, tt.want)
		}
	}
}

func TestDefaultPlayOpts(t *testing.T) {
	initMixerForTest()
	opts := defaultPlayOpts()
	if opts.Group != GroupSFX {
		t.Errorf("expected group SFX, got %v", opts.Group)
	}
	if opts.Volume != 1.0 {
		t.Errorf("expected volume 1.0, got %f", opts.Volume)
	}
	if opts.Pitch != 1.0 {
		t.Errorf("expected pitch 1.0, got %f", opts.Pitch)
	}
	if opts.Pan != 0 {
		t.Errorf("expected pan 0, got %f", opts.Pan)
	}
	if opts.Loop {
		t.Error("expected loop false")
	}
}

func TestPlay(t *testing.T) {
	initMixerForTest()
	m := MixerI()

	// Playing nil sound should return 0
	id := m.Play(nil, defaultPlayOpts())
	if id != 0 {
		t.Errorf("expected 0 for nil sound, got %d", id)
	}
}

func TestStopNonExistentChannel(t *testing.T) {
	initMixerForTest()
	m := MixerI()

	// Should not panic
	m.Stop(9999)
	m.Pause(9999)
	m.Resume(9999)
	m.SetChannelVolume(9999, 0.5)
}

func TestChannelInfoNil(t *testing.T) {
	initMixerForTest()
	m := MixerI()

	info := m.ChannelInfo(9999)
	if info != nil {
		t.Error("expected nil info for non-existent channel")
	}
}

func TestGC(t *testing.T) {
	initMixerForTest()
	m := MixerI()

	// GC on empty mixer should not panic
	m.GC()
}

func TestClear(t *testing.T) {
	initMixerForTest()
	m := MixerI()

	// Clear on empty mixer should not panic
	m.Clear()
	if m.Len() != 0 {
		t.Errorf("expected 0 after clear, got %d", m.Len())
	}
}

func TestActiveChannels(t *testing.T) {
	initMixerForTest()
	m := MixerI()

	channels := m.ActiveChannels()
	if len(channels) != 0 {
		t.Errorf("expected 0 active channels, got %d", len(channels))
	}
}

func TestConcurrentAccess(t *testing.T) {
	initMixerForTest()
	m := MixerI()

	// Run concurrent operations to check for races
	done := make(chan bool)
	go func() {
		m.SetGroupVolume(GroupMusic, 0.5)
		done <- true
	}()
	go func() {
		m.SetListenerPosition(100, 200)
		done <- true
	}()
	go func() {
		m.Len()
		done <- true
	}()
	go func() {
		m.ActiveChannels()
		done <- true
	}()
	go func() {
		m.computeSpatialPan(50)
		done <- true
	}()

	for i := 0; i < 5; i++ {
		<-done
	}
}
