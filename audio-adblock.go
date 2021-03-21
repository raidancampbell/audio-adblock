package main

import (
	"bytes"
	"fmt"
	"github.com/go-fingerprint/fingerprint"
	"github.com/go-fingerprint/gochroma"
	"github.com/hajimehoshi/oto"
	"github.com/raidancampbell/gochroma/chromaprint"
	"github.com/tosone/minimp3"
	"github.com/viert/go-lame"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"time"
)

func main() {
	// sample length 31,207 or 34,494
	// 94-95 had 14,938 at 95% similarity. 7213 at 99.5% similarity.  4581 at 99.9%. 2771 at 99.99%
	// 95-96 had 16,660

	_, one, _ := readFile("input/95.mp3")
	data2, two, samples2 := readFile("input/96.mp3")
	//fmt.Println(dur2.String())
	//three := readFile("input/96.mp3")

	fmt.Println("files read in.  Finding longest subsequences...")
	//fmt.Println(longestCommonSubsequence(one, two))
	fprintStart, fprintStop := getLongestSubsequence(one, two)

	// test data has 123ms per fingerprint item
	fmt.Printf("found subsequence of length %s\n", time.Millisecond * time.Duration(fprintStop - fprintStart) * 123)
	fmt.Printf("location by millisecond offset: start %s, end: %s\n", time.Duration(fprintStart) * 123 * time.Millisecond,  time.Duration(fprintStop) * 123 * time.Millisecond)
	fmt.Printf("location by bytes offset: start %f, end: %f\n", float64(fprintStart * samples2 * 2 * 2)/float64(len(data2)),  float64(fprintStop * samples2 * 2 * 2)/float64(len(data2)))

	var context *oto.Context
	var err error
	if context, err = oto.NewContext(44100, 2, 2, 1024); err != nil {
		panic(err)
	}
	defer context.Close()

	var player = context.NewPlayer()
	// item starting point * samples per item * channels * bit depth
	fmt.Printf("expected bytes: %d\n", (fprintStop * samples2 * 2 * 2) - (fprintStart * samples2 * 2 * 2))
	bytesWritten, err := player.Write(data2[(fprintStart * samples2 * 2 * 2) : (fprintStop * samples2 * 2 * 2)])
	fmt.Printf("wrote %d bytes to audio\n", bytesWritten)
	if err != nil {
		panic(err)
	}
	fmt.Println("write complete, waiting 1 seconds to close player")

	<-time.After(1 * time.Second)

	if err := player.Close(); err != nil {
		log.Fatal(err)
	}


	of, err := os.Create("output.mp3")
	if err != nil {
		panic(err)
	}
	defer of.Close()
	enc := lame.NewEncoder(of)
	defer enc.Close()

	r := bytes.NewBuffer(data2[(fprintStart * samples2 * 2 * 2) : (fprintStop * samples2 * 2 * 2)])
	r.WriteTo(enc)

}

func longestCommonSubsequence(A, B []int32) int {
	tmp := float64(math.MaxInt32) * 0.0001
	tolerance := int32(tmp)

	pre := make([]int, len(B)+1)
	cur := make([]int, len(B)+1)

	for j := 0; j < len(A); j++ {
		for i := 0; i < len(B); i++ {
			temp := max(pre[i+1], cur[i])
			if int32Abs(int32Abs(B[i])-int32Abs(A[j])) < tolerance {
				cur[i+1] = max(temp, pre[i]+1)
			} else {
				cur[i+1] = temp
			}
		}
		cur, pre = pre, cur
	}
	return pre[len(B)]
}

func getLongestSubsequence(A, B []int32) (int, int) {
	tmp := float64(math.MaxInt32) * 0.4
	tolerance := int32(tmp)

	maxStart := 0
	maxLen := 0

	startIdx := 0
	stopIdx := -1

	//TODO: the logic on this is wrong
	// it finds the longest subsequence in B that looks like a single item in A
	// what I really need is a suffix array
	// https://github.com/vmarkovtsev/go-lcss
	// and https://en.wikipedia.org/wiki/Suffix_array
	for j := 0; j < len(A); j++ {
		for i := 0; i < len(B); i++ {
			if int32Abs(int32Abs(B[i])-int32Abs(A[j])) < tolerance {
				if stopIdx != i-1 {
					startIdx = i
				}
				stopIdx = i
			}
			if stopIdx-startIdx > maxLen {
				maxStart = startIdx
				maxLen = stopIdx - startIdx
			}
		}
	}
	return maxStart, maxStart+maxLen
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

	fpcalc := gochroma.New(gochroma.AlgorithmDefault)
	defer fpcalc.Close()
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

	if i.MaxSeconds < 120 {
		i.MaxSeconds = 120
	}
	rate, channels := i.Rate, i.Channels
	if err := ctx.Start(int(rate), int(channels)); err != nil {
		return nil, 0, err
	}
	numbytes := 2 * 10 * rate * channels
	buf := make([]byte, numbytes)
	for total := uint(0); total <= i.MaxSeconds; total += 10 {
		read, err := i.Src.Read(buf)
		if err != nil && err != io.EOF {
			return nil, 0, err
		}
		if read == 0 {
			break
		}
		if err := ctx.Feed(buf[:read]); err != nil {
			return nil, 0, err
		}
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
