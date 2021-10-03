package main

import (
	"github.com/raidancampbell/gochroma"
	"github.com/raidancampbell/gochroma/chromaprint"
	"github.com/tosone/minimp3"
	"os"
)

// readFile reads the mp3 from the given path and returns its raw data, alongside its metadata/fingerprint
func readFile(filename string) ([]byte, audioMetadata, error) {
	var err error

	var file []byte
	if file, err = os.ReadFile(filename); err != nil {
		return nil, audioMetadata{}, err
	}

	dec, data, err := minimp3.DecodeFull(file)
	if err != nil {
		return nil, audioMetadata{}, err
	}
	defer dec.Close()

	metadata, err := rawFingerprint(data, dec.Channels, dec.SampleRate)
	if err != nil {
		return nil, audioMetadata{}, err
	}

	return data, metadata, nil
}

// rawFingerprint makes CGO callouts to the chromaprint bindings to fingerprint the given data, and returns
// the fingerprint/metadata of the given input data
func rawFingerprint(data []byte, channels, sampleRate int) (audioMetadata, error) {
	ctx := chromaprint.NewChromaprint(gochroma.AlgorithmDefault)
	defer ctx.Free()

	if err := ctx.Start(sampleRate, channels); err != nil {
		return audioMetadata{}, err
	}

	if err := ctx.Feed(data); err != nil {
		return audioMetadata{}, err
	}

	if err := ctx.Finish(); err != nil {
		return audioMetadata{}, err
	}
	fprint, err := ctx.GetRawFingerprint()

	metadata := audioMetadata{
		fingerprints:     fprint,                                // len 31207
		audioChannels:    float64(channels),                     // 2
		audioSampleRate:  float64(sampleRate),                   // 44100
		samplesPerFprint: float64(ctx.GetItemDurationSamples()), // 1365
		fprintChannels:   float64(ctx.GetNumChannels()),         // 1
		fprintSampleRate: float64(ctx.GetSampleRate()),          //11025
	}
	metadata.audioDurationSec = float64(len(data)) / metadata.audioSampleRate / metadata.audioChannels / byteDepth
	metadata.calcSamplesPerFprint = metadata.audioDurationSec * metadata.fprintSampleRate / float64(len(metadata.fingerprints))

	return metadata, err
}

type audioMetadata struct {
	fingerprints     []int32
	audioChannels    float64
	audioSampleRate  float64
	samplesPerFprint float64
	fprintChannels   float64
	fprintSampleRate float64

	audioDurationSec     float64
	calcSamplesPerFprint float64
}

