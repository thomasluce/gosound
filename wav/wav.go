package wav

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/thomasluce/gosound"
	"github.com/thomasluce/util"
)

func init() {
	gosound.RegisterAudioFormat("wav", "wav", decodeWavFile)
}

type sampleReader interface {
	ReadSample(io.Reader, *wavFile) (float32, error)
}

type soundConfig struct {
	id          [4]byte
	size        uint32
	format      uint16
	numChannels uint16
	sampleRate  uint32
	// byteRate == sampleRate * numChannels * bitsPerSample/8
	byteRate uint32
	// blockAlign == numChannels * bitsPerSample/8
	blockAlign    uint16
	bitsPerSample uint16

	reader sampleReader
}

type dataChunk struct {
	id [4]byte
	// size == numSamples * numChannels * bitsPerSample/8
	size int32
	// TODO: do I need the soundData to be some other format (a slice of some
	// other format)
	soundData []byte
}

type factChunk struct {
	id           [4]byte
	size         int32
	sampleLength int32
}

type wavFile struct {
	riff      [4]byte
	chunkSize uint32
	format    [4]byte

	config soundConfig
	data   []dataChunk
}

// see http://soundfile.sapp.org/doc/WaveFormat/ for more information
func decodeWavFile(file *os.File) (*gosound.Sound, error) {
	wav := wavFile{}

	err := binary.Read(file, binary.BigEndian, &wav.riff)
	if err != nil {
		return nil, fmt.Errorf("Error reading RIFF of wav: %v", err)
	}
	if string(wav.riff[:]) != "RIFF" {
		return nil, fmt.Errorf("Not the expected magic number, RIFF: %v", wav.riff)
	}

	err = binary.Read(file, binary.LittleEndian, &wav.chunkSize)
	if err != nil {
		return nil, fmt.Errorf("Could not read chunksize: %v", err)
	}

	err = binary.Read(file, binary.BigEndian, &wav.format)
	if err != nil {
		return nil, fmt.Errorf("Could not read file format: %v", err)
	}
	if string(wav.format[:]) != "WAVE" {
		return nil, fmt.Errorf("RIFF format not wave: %v", wav.format)
	}

	util.Debugf("RIFF WAVE Chunk size: %d", wav.chunkSize)

	wav.config, err = readWavConfig(file)
	if err != nil {
		return nil, err
	}

	// If we are a PCM file, skip 2 bytes, for the extraparam's that aren't
	// there.
	//if wav.config.format == 1 {
	file.Seek(2, 1)
	//}

	// Non-PCM formats have to have a fact chunk
	if wav.config.format != 1 {
		// Get the fact chunk.
		fchunk := factChunk{}
		err = binary.Read(file, binary.BigEndian, &fchunk.id)
		if err != nil {
			return nil, fmt.Errorf("Error reading fact chunk header: %v", err)
		}
		if string(fchunk.id[:]) != "fact" {
			return nil, fmt.Errorf("Expecting fact chunk, got %s (%v)", string(fchunk.id[:]), fchunk.id)
		}

		// We don't actually care about the the rest of the chunk, we just need to
		// know it is here so that we can skip it correctly.
		err = binary.Read(file, binary.LittleEndian, &fchunk.size)
		if err != nil {
			return nil, err
		}
		util.Debugf("Fact chunk size: %d", fchunk.size)
		file.Seek(int64(fchunk.size), 1)
	}

	wav.data = []dataChunk{}
	chunk := dataChunk{}
	for {
		err = binary.Read(file, binary.BigEndian, &chunk.id)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Error reading data chunk header: %v", err)
		}

		if string(chunk.id[:]) == "afsp" {
			// This is a special chunk that contains things like copywrite
			// information. We have only ever seen this at the end of a file, so we
			// just special-case and skip it.
			break
		}

		if string(chunk.id[:]) != "data" {
			return nil, fmt.Errorf("Expecting a data chunk, got %s (%v)", string(chunk.id[:]), chunk.id)
		}

		util.Debugf("Data chunk id %v", chunk.id)

		err = binary.Read(file, binary.LittleEndian, &chunk.size)
		if err != nil {
			return nil, fmt.Errorf("Error reading data chunk size: %v", err)
		}

		util.Debugf("Data chunk size: %d", chunk.size)

		chunk.soundData = make([]byte, chunk.size)
		err = binary.Read(file, binary.LittleEndian, &chunk.soundData)
		if err != nil {
			return nil, fmt.Errorf("Error reading data chunk data: %v. Expected %d bytes", err, chunk.size)
		}

		if chunk.size%2 != 0 {
			util.Debug("Odd chunk size, skipping pad byte")
			file.Seek(1, 1)
		}

		wav.data = append(wav.data, chunk)
	}

	sound := gosound.Sound{
		Samples:    wav.ToSamples(),
		SampleRate: float64(wav.config.sampleRate),
	}
	return &sound, nil
}

// ToSamples turns a fully read wav file into gosound.Samples that we can use for
// playback.
func (w *wavFile) ToSamples() []gosound.Sample {
	samples := []gosound.Sample{}

	bytesPerSample := int(w.config.bitsPerSample) / 8 * int(w.config.numChannels)
	var err error
	for _, data := range w.data {
		reader := bytes.NewReader(data.soundData)
		util.Debugf("%d-byte data (%d samples)", len(data.soundData), len(data.soundData)/bytesPerSample)
		for i := 0; i < len(data.soundData); i += int(w.config.blockAlign) {
			sample := make([]float32, w.config.numChannels)

			for j := 0; j < int(w.config.numChannels); j++ {
				sample[j], err = w.config.reader.ReadSample(reader, w)
				if err != nil {
					// TODO: do something with err
				}
			}

			samples = append(samples, sample)
		}
	}
	return samples
}

func readWavConfig(file *os.File) (soundConfig, error) {
	config := soundConfig{}

	err := binary.Read(file, binary.BigEndian, &config.id)
	if err != nil {
		return config, fmt.Errorf("Could not get wave fmt chunk id: %v", err)
	}
	if string(config.id[:]) != "fmt " {
		return config, fmt.Errorf("Expexted fmt sub-chunk, got %v", config.id)
	}

	util.Debugf("Sub-chunk: %v", config.id)

	err = binary.Read(file, binary.LittleEndian, &config.size)
	if err != nil {
		return config, fmt.Errorf("Could not get fmt config subchunk size: %v", err)
	}

	util.Debugf("Sub-chunk size: %d", config.size)

	err = binary.Read(file, binary.LittleEndian, &config.format)
	if err != nil {
		return config, fmt.Errorf("Could not get fmt config subchunk format: %v", err)
	}

	switch config.format {
	case 1:
		config.reader = PCMSampler
	case 3:
		config.reader = IEEEFloatSampler
	case 6:
		// https://github.com/deftio/companders/blob/master/companders.c
		// see function DIO_s16 DIO_ALawToLinear(DIO_s8 aLawByte)
		// config.reader = AlphaLawSampler
		config.reader = AlawSampler
	case 7:
		// https://github.com/haha01haha01/NAudio/blob/master/Codecs/MuLawDecoder.cs
		config.reader = UlawSampler
	default:
		// TODO; do the rest of them.
		return config, fmt.Errorf("Un-supported sample format (we don't do compressed wav's): %d", config.format)
	}

	util.Debugf("Sub-chunk format: %v", config.format)

	err = binary.Read(file, binary.LittleEndian, &config.numChannels)
	if err != nil {
		return config, fmt.Errorf("Could not read number of channels: %v", err)
	}

	util.Debugf("%d channels", config.numChannels)

	err = binary.Read(file, binary.LittleEndian, &config.sampleRate)
	if err != nil {
		return config, fmt.Errorf("Could not read sample rate: %v", err)
	}

	util.Debugf("Sample rate: %d", config.sampleRate)

	err = binary.Read(file, binary.LittleEndian, &config.byteRate)
	if err != nil {
		return config, fmt.Errorf("Could not read byteRate: %v", err)
	}

	util.Debugf("Byte rate: %d", config.byteRate)

	err = binary.Read(file, binary.LittleEndian, &config.blockAlign)
	if err != nil {
		return config, fmt.Errorf("Could not read block alignment: %v", err)
	}

	util.Debugf("Block alignment %v", config.blockAlign)

	err = binary.Read(file, binary.LittleEndian, &config.bitsPerSample)
	if err != nil {
		return config, fmt.Errorf("Could not read bits per sample: %v", err)
	}

	util.Debugf("%d bits per sample", config.bitsPerSample)

	return config, nil
}
