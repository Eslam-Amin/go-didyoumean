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

// JaroSimilarity computes the Jaro distance between two strings in range [0.0, 1.0].
func JaroSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	r1 := []rune(s1)
	r2 := []rune(s2)
	l1 := len(r1)
	l2 := len(r2)

	if l1 == 0 || l2 == 0 {
		return 0.0
	}

	matchWindow := max(l1, l2)/2 - 1
	if matchWindow < 0 {
		matchWindow = 0
	}

	s1Matches := make([]bool, l1)
	s2Matches := make([]bool, l2)
	matches := 0

	for i := 0; i < l1; i++ {
		start := max(0, i-matchWindow)
		end := min(i+matchWindow+1, l2)

		for j := start; j < end; j++ {
			if s2Matches[j] {
				continue
			}
			if r1[i] == r2[j] {
				s1Matches[i] = true
				s2Matches[j] = true
				matches++
				break
			}
		}
	}

	if matches == 0 {
		return 0.0
	}

	transpositions := 0
	k := 0
	for i := 0; i < l1; i++ {
		if !s1Matches[i] {
			continue
		}
		for !s2Matches[k] {
			k++
		}
		if r1[i] != r2[k] {
			transpositions++
		}
		k++
	}

	m := float64(matches)
	t := float64(transpositions) / 2.0
	return (m/float64(l1) + m/float64(l2) + (m-t)/m) / 3.0
}