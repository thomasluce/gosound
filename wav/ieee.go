package wav

import (
	"encoding/binary"
	"fmt"
	"io"
)

// The ieee sampler reads wav files encoded using ieee floating-point encoding
// in either 32 or 64 bit.

type ieeeFloatReader bool

func (r ieeeFloatReader) ReadSample(reader io.Reader, wav *wavFile) (float32, error) {
	if wav.config.bitsPerSample == 32 {
		var s float32
		err := binary.Read(reader, binary.LittleEndian, &s)
		if err != nil {
			return 0, err
		}

		return s, nil
	} else if wav.config.bitsPerSample == 64 {
		var s float64
		err := binary.Read(reader, binary.LittleEndian, &s)
		if err != nil {
			return 0, err
		}

		return float32(s), nil
	}

	return 0, fmt.Errorf("Unknown bit-width for ieee floating-point samples, %d", wav.config.bitsPerSample)
}

var IEEEFloatSampler ieeeFloatReader
