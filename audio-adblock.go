package main

import (
	"bytes"
	"fmt"
	"github.com/raidancampbell/gochroma"
	"github.com/raidancampbell/gochroma/chromaprint"
	"github.com/tosone/minimp3"
	"github.com/viert/go-lame"
	"os"
	"sync"
	"time"
)

const byteDepth = 2.0

func main() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	var metadata1 = audioMetadata{}
	var data1 []byte
	go func() {
		data1, metadata1 = readFile("input/96.mp3")
		wg.Done()
	}()
	data2, metadata2 := readFile("input/96.mp3")
	wg.Wait()

	fmt.Println("files read in.  Finding longest subsequences...")
	length, aStop, bStop := LCSub(metadata1.fingerprints, metadata2.fingerprints)
	bStart := (bStop - length) + 1
	aStart := (aStop - length) + 1

	fmt.Printf("found subsequence of length %s\n", time.Duration((metadata1.calcSamplesPerFprint/metadata1.fprintSampleRate)*length)*time.Second)

	of, err := os.Create("outputB.mp3")
	if err != nil {
		panic(err)
	}
	enc := lame.NewEncoder(of)

	// data bytes, starting at the beginning of the matched fingerprint
	// then multiplied by samples included in each fingerprint
	// then adjusted by the internal fingerprinting sample rate vs source's sample rate
	// then adjusted by the internal channels vs source's channels
	// then adjusted by the bit depth of the source (hardcoded to two bytes)
	sampleRateAdjustment := metadata1.audioSampleRate / metadata1.fprintSampleRate // 4
	channelAdjustment := metadata1.audioChannels / metadata1.fprintChannels        // 2
	startIndex := int(bStart * metadata2.calcSamplesPerFprint * sampleRateAdjustment * channelAdjustment * byteDepth)
	stopIndex := int(bStop * metadata2.calcSamplesPerFprint * sampleRateAdjustment * channelAdjustment * byteDepth)
	// latch the start index to the beginning of a sample
	for ; startIndex%(byteDepth*int(metadata2.audioChannels)) != 0; startIndex++ {
	}
	r := bytes.NewBuffer(data2[startIndex:stopIndex])
	_, err = r.WriteTo(enc)
	if err != nil {
		panic(err)
	}
	enc.Close()
	err = of.Close()
	if err != nil {
		panic(err)
	}

	of, err = os.Create("outputA.mp3")
	if err != nil {
		panic(err)
	}
	enc = lame.NewEncoder(of)

	startIndex = int(aStart * metadata1.calcSamplesPerFprint * sampleRateAdjustment * channelAdjustment * byteDepth)
	stopIndex = int(aStop * metadata1.calcSamplesPerFprint * sampleRateAdjustment * channelAdjustment * byteDepth)
	for ; startIndex%(byteDepth*int(metadata1.audioChannels)) != 0; startIndex++ {
	}
	r = bytes.NewBuffer(data1[startIndex:stopIndex])
	r.WriteTo(enc)
	enc.Close()
	of.Close()
}

func readFile(filename string) ([]byte, audioMetadata) {
	var err error

	var file []byte
	if file, err = os.ReadFile(filename); err != nil {
		panic(err)
	}

	dec, data, err := minimp3.DecodeFull(file)
	if err != nil {
		panic(err)
	}

	metadata, err := rawFingerprint(data, dec.Channels, dec.SampleRate)
	if err != nil {
		panic(err)
	}

	metadata.audioDurationSec = float64(len(data)) / metadata.audioSampleRate / metadata.audioChannels / byteDepth
	metadata.calcSamplesPerFprint = metadata.audioDurationSec * metadata.fprintSampleRate / float64(len(metadata.fingerprints))
	return data, metadata
}

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

	return audioMetadata{
		fingerprints:     fprint,                                // len 31207
		audioChannels:    float64(channels),                     // 2
		audioSampleRate:  float64(sampleRate),                   // 44100
		samplesPerFprint: float64(ctx.GetItemDurationSamples()), // 1365
		fprintChannels:   float64(ctx.GetNumChannels()),         // 1
		fprintSampleRate: float64(ctx.GetSampleRate()),          //11025
	}, err
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
