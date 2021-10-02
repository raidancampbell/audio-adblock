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

func main() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	var metadata1 = audioMetadata{}
	var data1 []byte
	go func() {
		data1, metadata1 = readFile("input/94.mp3")
		wg.Done()
	}()
	data2, metadata2 := readFile("input/96.mp3")
	wg.Wait()

	fmt.Println("files read in.  Finding longest subsequences...")
	length, aStop, bStop := LCSub(metadata1.fingerprints, metadata2.fingerprints)
	bStart := (bStop - length) + 1
	aStart := (aStop - length) + 1

	fmt.Printf("found subsequence of length %s\n", time.Duration((float64(metadata1.samplesPerFprint) / float64(metadata1.fprintSampleRate)) * float64(length)) * time.Second)

	of, err := os.Create("outputB.mp3")
	if err != nil {
		panic(err)
	}
	defer of.Close()
	enc := lame.NewEncoder(of)

	r := bytes.NewBuffer(data2[(bStart * metadata2.samplesPerFprint * 4 * 4):(bStop * metadata2.samplesPerFprint * 4 * 4)])
	r.WriteTo(enc)
	enc.Close()

	of, err = os.Create("outputA.mp3")
	if err != nil {
		panic(err)
	}
	defer of.Close()
	enc = lame.NewEncoder(of)
	defer enc.Close()

	r = bytes.NewBuffer(data1[(aStart * metadata1.samplesPerFprint * 4 * 4):(aStop * metadata1.samplesPerFprint * 4 * 4)])
	r.WriteTo(enc)
	enc.Close()

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
		fingerprints:     fprint,
		audioChannels:    channels,
		audioSampleRate:  sampleRate,
		samplesPerFprint: ctx.GetItemDurationSamples(),
		fprintChannels:   ctx.GetNumChannels(),
		fprintSampleRate: ctx.GetSampleRate(),
	}, err
}

type audioMetadata struct {
	fingerprints     []int32
	audioChannels    int
	audioSampleRate  int
	samplesPerFprint int
	fprintChannels   int
	fprintSampleRate int
}