package main

import (
	"bytes"
	"fmt"
	"github.com/go-fingerprint/fingerprint"
	"github.com/go-fingerprint/gochroma"
	"github.com/raidancampbell/gochroma/chromaprint"
	"github.com/tosone/minimp3"
	"github.com/viert/go-lame"
	"io"
	"io/ioutil"
	"math"
	"os"
	"sync"
	"time"
)

func main() {
	// sample length 31,207 or 34,494
	// 94-95 had 14,938 at 95% similarity. 7213 at 99.5% similarity.  4581 at 99.9%. 2771 at 99.99%
	// 95-96 had 16,660

	wg := sync.WaitGroup{}
	wg.Add(1)
	one := []int32{}
	data1 := []byte{}
	var samples1 int
	go func() {
		data1, one, samples1 = readFile("input/94.mp3")
		wg.Done()
	}()
	data2, two, samples2 := readFile("input/96.mp3")
	wg.Wait()

	fmt.Println("files read in.  Finding longest subsequences...")
	length, aStop, bStop := LCSub(one, two)
	bStart := (bStop - length) + 1
	aStart := (aStop - length) + 1

	// test data has 123ms per fingerprint item
	fmt.Printf("found subsequence of length %s\n", time.Millisecond*time.Duration(length)*123/4)
	fmt.Printf("location by millisecond offset in A: start %s, end: %s\n", time.Duration(aStart)*(123/4)*time.Millisecond, time.Duration(aStop)*(123/4)*time.Millisecond)
	fmt.Printf("location by millisecond offset in B: start %s, end: %s\n", time.Duration(bStart)*(123/4)*time.Millisecond, time.Duration(bStop)*(123/4)*time.Millisecond)

	of, err := os.Create("outputB.mp3")
	if err != nil {
		panic(err)
	}
	defer of.Close()
	enc := lame.NewEncoder(of)

	r := bytes.NewBuffer(data2[(bStart * samples2 * 4*4):(bStop * samples2 * 4*4)])
	r.WriteTo(enc)
	enc.Close()

	of, err = os.Create("outputA.mp3")
	if err != nil {
		panic(err)
	}
	defer of.Close()
	enc = lame.NewEncoder(of)
	defer enc.Close()

	r = bytes.NewBuffer(data1[(aStart * samples1 * 4*4):(aStop * samples1 * 4*4)])
	r.WriteTo(enc)
	enc.Close()

}

func int32Abs(i int32) int32 {
	mask := i >> 31
	return (mask + i) ^ mask
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func readFile(filename string) ([]byte, []int32, int) {
	var err error

	var file []byte
	if file, err = ioutil.ReadFile(filename); err != nil {
		panic(err)
	}

	dec, data, err := minimp3.DecodeFull(file)
	if err != nil {
		panic(err)
	}

	fprint, samples, err := rawFingerprint(
		fingerprint.RawInfo{
			Src:        bytes.NewReader(data),
			Channels:   uint(dec.Channels),
			Rate:       uint(dec.SampleRate),
			MaxSeconds: math.MaxUint64,
		})
	if err != nil {
		panic(err)
	}
	return data, fprint, samples
}
func rawFingerprint(i fingerprint.RawInfo) ([]int32, int, error) {
	ctx := chromaprint.NewChromaprint(gochroma.AlgorithmDefault)
	defer ctx.Free()

	rate, channels := i.Rate, i.Channels
	if err := ctx.Start(int(rate), int(channels)); err != nil {
		return nil, 0, err
	}

	read, err := io.ReadAll(i.Src)
	if err != nil && err != io.EOF {
		return nil, 0, err
	}
	if err := ctx.Feed(read); err != nil {
		return nil, 0, err
	}

	if err := ctx.Finish(); err != nil {
		return nil, 0, err
	}
	fprint, err := ctx.GetRawFingerprint()
	samples := ctx.GetItemDurationSamples()
	dur := ctx.GetItemDuration()
	fmt.Printf("duration per item %s\n", dur.String()) // 123
	//1000 / 123 = 8.130081301
	//TODO: these numbers do not agree for a sample rate of  44100
	fmt.Printf("samples per item %d\n", samples) // 1365
	// 1365 * 4 = 5460; 44100/5460 = 8.076923077
	// close enough?
	//TODO: where does the 4 come from with the samples?  Need to read the source
	// TODO: print timestamps of start and stop.  Also potentially multiply both by 4?
	return fprint, samples, err
}
