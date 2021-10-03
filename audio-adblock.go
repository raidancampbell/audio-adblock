package main

import (
	"bytes"
	"fmt"
	"github.com/viert/go-lame"
	"os"
	"sync"
	"time"
)

// 16-bit audio == 2 bytes
const byteDepth = 2.0
// what's the smallest subsequence in seconds that should be considered a duplicate
const minDurationSec = 5.0

func main() {
	fmt.Println("reading in files...")
	wg := sync.WaitGroup{}
	wg.Add(1)
	var metadata1 = audioMetadata{}
	var data1 []byte
	go func() {
		var err error
		data1, metadata1, err = readFile("input/94.mp3")
		if err != nil {
			panic(err)
		}
		wg.Done()
	}()
	data2, metadata2, err := readFile("input/96.mp3")
	if err != nil {
		panic(err)
	}
	wg.Wait()

	fmt.Println("files read in.  Finding longest subsequences...")
	gt := minDurationSec/(metadata1.calcSamplesPerFprint/metadata1.fprintSampleRate)

	matches := LCSubs(metadata1.fingerprints, metadata2.fingerprints, int(gt))
	for matchNum, match := range matches {
		fmt.Printf("found subsequence %d of length %s\n", matchNum, time.Duration((metadata1.calcSamplesPerFprint/metadata1.fprintSampleRate)*match.length*float64(time.Second)))

		aStart := (match.AStop - match.length) + 1
		bStart := (match.BStop - match.length) + 1

		fmt.Printf("writing match %d to files...\n", matchNum)
		wg.Add(1)
		go func() {
			writeMatchesToFile(fmt.Sprintf("outputA-%d.mp3", matchNum), metadata1, aStart, match.AStop, data1)
			wg.Done()
		}()
		writeMatchesToFile(fmt.Sprintf("outputB-%d.mp3", matchNum), metadata2, bStart, match.BStop, data2)
		wg.Wait()
	}

}

// writeMatchesToFile is a function useful for debugging: given a data/metadata stream and the start/stop indices of the
// fingerprint matches, the equivalent data is written to a file.  This is useful for diagnosing "is this actually identical"
// and whether the matcher is working as expected
func writeMatchesToFile(filename string, metadata audioMetadata, fprintStart, fPrintStop float64, data []byte) error{
	of, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	enc := lame.NewEncoder(of)

	// data bytes, starting at the beginning of the matched fingerprint
	// then multiplied by samples included in each fingerprint
	// then adjusted by the internal fingerprinting sample rate vs source's sample rate
	// then adjusted by the internal channels vs source's channels
	// then adjusted by the bit depth of the source (hardcoded to two bytes)
	sampleRateAdjustment := metadata.audioSampleRate / metadata.fprintSampleRate // 4
	channelAdjustment := metadata.audioChannels / metadata.fprintChannels        // 2
	startIndex := int(fprintStart * metadata.calcSamplesPerFprint * sampleRateAdjustment * channelAdjustment * byteDepth)
	stopIndex := int(fPrintStop * metadata.calcSamplesPerFprint * sampleRateAdjustment * channelAdjustment * byteDepth)
	// latch the indices to the beginning of a sample
	for ; startIndex%int(byteDepth*metadata.audioChannels) != 0; startIndex++ {
	}
	for ; stopIndex%int(byteDepth*metadata.audioChannels) != 0; stopIndex-- {
	}
	r := bytes.NewBuffer(data[startIndex:stopIndex])
	_, err = r.WriteTo(enc)
	if err != nil {
		return err
	}
	enc.Close()
	return of.Close()
}