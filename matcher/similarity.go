package matcher

import (
	"math"
	"strings"
)

// LevenshteinDistance computes the character edit distance between two strings.
// Handles UTF-8 runes accurately with an O(min(N, M)) space dynamic programming approach.
func LevenshteinDistance(a, b string) int {
	if a == b {
		return 0
	}
	ar := []rune(a)
	br := []rune(b)

	if len(ar) == 0 {
		return len(br)
	}
	if len(br) == 0 {
		return len(ar)
	}

	if len(ar) < len(br) {
		ar, br = br, ar
	}

	bLen := len(br)
	prevRow := make([]int, bLen+1)
	currRow := make([]int, bLen+1)

	for j := 0; j <= bLen; j++ {
		prevRow[j] = j
	}

	for i := 0; i < len(ar); i++ {
		currRow[0] = i + 1
		for j := 0; j < bLen; j++ {
			cost := 1
			if ar[i] == br[j] {
				cost = 0
			}
			currRow[j+1] = min(currRow[j]+1, min(prevRow[j+1]+1, prevRow[j]+cost))
		}
		copy(prevRow, currRow)
	}

	return prevRow[bLen]
}