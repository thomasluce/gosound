package gosound

import (
	"fmt"

	"github.com/gordonklaus/portaudio"
	"github.com/thomasluce/util"
)

// Sample is a simple type that we use to represent a "frame" of audio, or a
// single sample. Each element in the slice is a different channel.
type Sample []float32

// Sound is a struct that holds all the information that we need for a sound to
// be played.
type Sound struct {
	// Samples are the actual sound frames themselves
	Samples []Sample

	// SampleRate is the sample rate of the sound.
	SampleRate float64

	// Position of the sound in samples
	Position uint

	paused bool
	stream portaudio.Stream
}

// Init initializes the sound
func (s *Sound) Init() error {
	stream, err := portaudio.OpenDefaultStream(
		0,                                    // no input channel
		2,                                    // 2 output channels (stereo sound)
		s.SampleRate,                         // Sample rate of the sound
		portaudio.FramesPerBufferUnspecified, // Let portaudio pick samples per frame for us
		s.processAudio,                       // The sound itself
	)
	if err != nil {
		return fmt.Errorf("Error starting sound stream: %v", err)
	}
	s.stream = *stream

	return nil
}

func (s *Sound) processAudio(out [][]float32) {
	if s.paused {
		return
	}

	// We assume that there are equal samples between left and right channels.
	for i := range out[0] {
		out[0][i] = float32(s.Samples[s.Position][0])
		out[1][i] = float32(s.Samples[s.Position][1])
		s.Position += 1
	}

	if int(s.Position) >= len(s.Samples) {
		s.stream.Stop()
	}
}

func (s *Sound) Terminate() {
	s.stream.Stop()
	s.stream.Close()
}

// Pause the sound
func (s *Sound) Pause() {
	s.paused = true
}

// Play the sound.
func (s *Sound) Play() {
	util.Debug("Playing")
	s.paused = false
	s.stream.Start()
}
