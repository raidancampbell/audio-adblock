package main

import (
	"math"
)

// TOLERANCE_FACTOR holds "how close, across the range of all int32 values,
// should two values be in order for them to be considered equal"
const TOLERANCE_FACTOR = 0.05

// LCSubs finds the longest common subsequence of the given int32 slices
// the length and stop locations of each input are returned
// implementation is from wikipedia's pseudocode: https://en.wikipedia.org/wiki/Longest_common_substring_problem
func LCSubs(a, b []int32, gt int) (matches []Match) {
	tmp := float64(math.MaxInt32) * TOLERANCE_FACTOR
	tolerance := int32(tmp)
	r, n := len(a), len(b)
	l := make([][]int32, r)
	for i := range l {
		l[i] = make([]int32, n)
	}
	for i := 0; i < r; i++ {
		for j := 0; j < n; j++ {
			if closeEnough(a[i], b[j], tolerance) {
				if i == 0 || j == 0 {
					l[i][j] = 1
				} else {
					l[i][j] = l[i-1][j-1] + 1
				}
			}
		}
	}

	for i := 1; i < r; i++ {
		for j := 1; j < n; j++ {
			if int(l[i][j]) > gt {
				if i == r-1 || j == n-1 || int(l[i+1][j+1]) != int(l[i][j])+1 {
					matches = append(matches, Match{AStop: float64(i), BStop: float64(j), length: float64(l[i][j])})
				}
			}
		}
	}
	return matches
}

type Match struct {
	AStop, BStop, length float64
}

func closeEnough(A, B, tolerance int32) bool {
	if A >= 0 && B >= 0 {
		return int32Abs(A-B) < tolerance
	}
	if A >= 0 && B < 0 {
		return int32Abs(A - int32Abs(B)) < tolerance
	}
	if A < 0 && B >= 0 {
		return int32Abs(int32Abs(A) - B) < tolerance
	}
	return int32Abs(int32Abs(A) - int32Abs(B)) < tolerance
}

func int32Abs(i int32) int32 {
	mask := i >> 31
	return (mask + i) ^ mask
}
