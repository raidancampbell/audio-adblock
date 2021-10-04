package main

import (
	"bytes"
	"fmt"
	"github.com/viert/go-lame"
	"os"
	"sort"
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
	aStarts, aStops := make([]float64, len(matches)), make([]float64, len(matches))
	bStarts, bStops := make([]float64, len(matches)), make([]float64, len(matches))
	for matchNum, match := range matches {
		fmt.Printf("found subsequence %d of length %s\n", matchNum, time.Duration((metadata1.calcSamplesPerFprint/metadata1.fprintSampleRate)*match.length*float64(time.Second)))

		aStarts[matchNum] = (match.AStop - match.length) + 1
		aStops[matchNum] = match.AStop
		bStarts[matchNum] = (match.BStop - match.length) + 1
		bStops[matchNum] = match.BStop
	}

	wg.Add(1)
	go func() {
		err = writeNonmatchesToFile("A.mp3", metadata1, aStarts, aStops, data1)
		if err != nil {
			panic(err)
		}
		wg.Done()
	}()
	err = writeNonmatchesToFile("B.mp3", metadata2, bStarts, bStops, data2)
	if err != nil {
		panic(err)
	}
	wg.Wait()
}

// TODO: one fingerprint's worth of static during patchwork
func writeNonmatchesToFile(filename string, metadata audioMetadata, fprintStarts, fprintStops []float64, data []byte) error {
	if len(fprintStops) != len(fprintStarts) {
		panic("input fingerprint index arrays are not of equal length")
	}

	// in the source data, which indies should be omitted
	filteredIndices := []int{}
	for i := range fprintStarts {
		// for each matching subsequence, convert the fingerprint index to the data index
		sampleRateAdjustment := metadata.audioSampleRate / metadata.fprintSampleRate // 4
		channelAdjustment := metadata.audioChannels / metadata.fprintChannels        // 2
		startIndex := int(fprintStarts[i] * metadata.calcSamplesPerFprint * sampleRateAdjustment * channelAdjustment * byteDepth)
		stopIndex := int(fprintStops[i] * metadata.calcSamplesPerFprint * sampleRateAdjustment * channelAdjustment * byteDepth)

		// then align the data index with an actual sample to ensure well-formed output
		for ; startIndex%(int(byteDepth)*int(metadata.audioChannels)) != 0; startIndex++ {
		}
		for ; stopIndex%(int(byteDepth)*int(metadata.audioChannels)) != 0; stopIndex-- {
		}
		// create a slice of all data indices held within this subsequence
		slice := make([]int, stopIndex - startIndex)
		_i := 0
		for i := startIndex; i < stopIndex; i++ {
			slice[_i] = i
			_i++
		}
		// and add the data indices for this subsequence to the omission set
		filteredIndices = append(filteredIndices, slice...)
	}
	// sort the filter slice to make it easier to walk through
	sort.Slice(filteredIndices, func(i, j int) bool {
		return filteredIndices[i] < filteredIndices[j]
	})

	redactedData := make([]byte, (len(data) - len(filteredIndices)) + 1)

	// index that we're writing data to
	redactedIdx := 0
	// index of the next filtering value, which contains the data index to filter
	filterIdx := 0
	// for each byte in the source data
	for i := range data {
		// if we're not already past the last filter section, check if the current data index should be omitted
		if (filterIdx < len(filteredIndices) && i != filteredIndices[filterIdx]) || filterIdx > len(filteredIndices) {
			if redactedIdx > cap(redactedData) {
				break
			}
			// if it should be included, then write it and increment the writing index
			redactedData[redactedIdx] = data[i]
			redactedIdx++
		} else {
			// if it should not be included, then increment the filter index
			filterIdx++
		}
	}

	r := bytes.NewBuffer(redactedData)

	of, err := os.Create(filename)
	if err != nil {
		return err
	}

	enc := lame.NewEncoder(of)
	_, err = r.WriteTo(enc)
	if err != nil {
		return err
	}
	enc.Close()
	return of.Close()
}

// writeMatchesToFile is a function useful for debugging: given a data/metadata stream and the start/stop indices of the
// fingerprint matches, the equivalent data is written to a file.  This is useful for diagnosing "is this actually identical"
// and whether the matcher is working as expected
func writeMatchesToFile(filename string, metadata audioMetadata, fprintStart, fPrintStop float64, data []byte) error{
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

	of, err := os.Create(filename)
	if err != nil {
		return err
	}

	enc := lame.NewEncoder(of)
	_, err = r.WriteTo(enc)
	if err != nil {
		return err
	}
	enc.Close()
	return of.Close()
}