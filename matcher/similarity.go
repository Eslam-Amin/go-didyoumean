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

// LevenshteinSimilarity computes normalized Levenshtein similarity in range [0.0, 1.0].
// 1.0 indicates identical strings, 0.0 indicates completely disjoint strings.
func LevenshteinSimilarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	ar := []rune(a)
	br := []rune(b)
	maxLen := max(len(ar), len(br))
	if maxLen == 0 {
		return 1.0
	}
	dist := LevenshteinDistance(a, b)
	sim := 1.0 - float64(dist)/float64(maxLen)
	if sim < 0.0 {
		return 0.0
	}
	return sim
}