package gosound

import (
	"fmt"
	"os"
	"strings"

	"github.com/gordonklaus/portaudio"
	"github.com/thomasluce/util"
)

// SampleRate is the sample rate for playback of our files in Hz
const SampleRate = 44100

type soundFormat struct {
	name, extension string
	decoder         func(*os.File) (*Sound, error)
}

// Some package-global vars that we use
var soundFormats map[string]soundFormat
var ioStream *portaudio.Stream
var sounds []Sound

func init() {
	soundFormats = make(map[string]soundFormat)
}

// TODO: make a loader for the wav format, and start effing around with this ish.

// RegisterAudioFormat is called by any audio file format handlers to register
// the fact that they know how to load a certain thing.
func RegisterAudioFormat(name, fileExtension string, decodeFunction func(*os.File) (*Sound, error)) {
	soundFormats[fileExtension] = soundFormat{
		name:      name,
		extension: fileExtension,
		decoder:   decodeFunction,
	}
}

// Init is called by the engine at the start of the engine. You should not call
// it yourself.
func Init() error {
	err := portaudio.Initialize()
	if err != nil {
		return fmt.Errorf("Could not initialize audio sub-system: %v", err)
	}

	util.Infof("Using portaudio: %s", portaudio.VersionText())

	return nil
}

// Terminate is called by the engine at the engine termination. You should not
// call this yourself.
func Terminate() {
	for i := range sounds {
		sounds[i].Terminate()
	}
	// We ignore the errors here because we are closing the app anyway, so fuck
	// it.
	portaudio.Terminate()
}

// Play plays a given file by filename.
func Play(filename string) error {
	util.Infof("Playing sound, %s", filename)
	parts := strings.Split(filename, ".")
	extension := parts[len(parts)-1]
	format, ok := soundFormats[extension]
	if !ok {
		return fmt.Errorf("Unknown file format, %s, for file, %s", extension, filename)
	}

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("Error laoding sound file: %v", err)
	}
	defer file.Close()

	sound, err := format.decoder(file)
	if err != nil {
		return fmt.Errorf("Error decoding file: %v", err)
	}
	err = sound.Init()
	if err != nil {
		return fmt.Errorf("Error initializing sound: %v", err)
	}
	sound.Play()

	sounds = append(sounds, *sound)

	return nil
}

func Playing() bool {
	// TODO: I need a new way to determine this now that I've changed everything around.
	return true
}
