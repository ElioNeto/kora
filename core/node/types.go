// Package node implements the core node system for Kora Engine
package node

// AudioPlayer represents a 2D audio player node
type AudioPlayer struct {
	Node2D
	volume  float32
	pitch   float32
	loop    bool
	autoPlay bool
}

// NewAudioPlayer creates a new AudioPlayer node
func NewAudioPlayer(name string, id uint64) *AudioPlayer {
	return &AudioPlayer{
		Node2D: *NewNode2D(name, id),
		volume:   1.0,
		pitch:    1.0,
		loop:     false,
		autoPlay: false,
	}
}
