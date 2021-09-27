package main

import (
	"fmt"
	"math"
)

// taken from wikipedia's pseudocode: https://en.wikipedia.org/wiki/Longest_common_substring_problem
// function LCSubstr(S[1..r], T[1..n])
//    L := array(1..r, 1..n)
//    z := 0
//    ret := {}
//
//    for i := 1..r
//        for j := 1..n
//            if S[i] = T[j]
//                if i = 1 or j = 1
//                    L[i, j] := 1
//                else
//                    L[i, j] := L[i − 1, j − 1] + 1
//                if L[i, j] > z
//                    z := L[i, j]
//                    ret := {S[i − z + 1..i]}
//                else if L[i, j] = z
//                    ret := ret ∪ {S[i − z + 1..i]}
//            else
//                L[i, j] := 0
//    return ret

func LCSub(a, b []int32) (length, stopA, stopB int) {
	tmp := float64(math.MaxInt32) * 0.050
	tolerance := int32(tmp)
	r, n := len(a), len(b)
	z := 0
	l := make([][]int32, r)
	for i := range l {
		l[i] = make([]int32, n)
	}
	var ret []int32
	for i := 0; i < r; i++ {
		for j := 0; j < n; j++ {
			if closeEnough(a[i], b[j], tolerance) {
				if i == 0 || j == 0 {
					l[i][j] = 1
				} else {
					l[i][j] = l[i-1][j-1] + 1
				}
				if int(l[i][j]) > z {
					stopA = i
					stopB = j
					z = int(l[i][j])
					ret = a[i-z + 1:i]
				} else if int(l[i][j]) == z {
					// ret := ret ∪ {S[i − z + 1..i]}
				}
			} else {
				l[i][j] = 0
			}
		}
	}
	_ = fmt.Sprintf("%d", ret) // just in case I want it later
	return z, stopA, stopB
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