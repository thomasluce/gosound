package wav

import (
	"encoding/binary"
	"errors"
	"io"
)

// The PCM sampler reads wav files encoded using PCM integer sampling.

type pcmSampleReader bool

var PCMSampler pcmSampleReader

func (r pcmSampleReader) ReadSample(reader io.Reader, wav *wavFile) (float32, error) {
	if wav.config.bitsPerSample == 8 {
		// We read it as an unsigned 8-bit integer, range 0 - 255
		var sample uint8
		err := binary.Read(reader, binary.LittleEndian, &sample)
		if err != nil {
			return 0, err
		}
		return (float32(sample) / 255.0) - 0.5, nil
	} else if wav.config.bitsPerSample == 16 {
		// We read it as a signed 16-bit integer, range -32768 - 32767
		var sample int16
		err := binary.Read(reader, binary.LittleEndian, &sample)
		if err != nil {
			return 0, err
		}
		return (float32(sample) / 32767.0), nil
	}
	return 0, errors.New("Unknown sample size")
}
