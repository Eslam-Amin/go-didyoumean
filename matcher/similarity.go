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

// JaroWinklerSimilarity computes the Jaro-Winkler similarity between two strings in range [0.0, 1.0].
// Gives extra weight to common prefixes, which is ideal for matching REST API URL paths.
func JaroWinklerSimilarity(s1, s2 string) float64 {
	jaro := JaroSimilarity(s1, s2)
	if jaro <= 0.0 || s1 == s2 {
		return jaro
	}

	r1 := []rune(s1)
	r2 := []rune(s2)
	maxPrefix := min(4, min(len(r1), len(r2)))
	prefixLen := 0
	for i := 0; i < maxPrefix; i++ {
		if r1[i] == r2[i] {
			prefixLen++
		} else {
			break
		}
	}

	const scalingFactor = 0.1
	jw := jaro + float64(prefixLen)*scalingFactor*(1.0-jaro)
	if jw > 1.0 {
		return 1.0
	}
	return jw
}

// SorensenDiceSimilarity computes the Dice similarity coefficient based on character bigrams.
func SorensenDiceSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}
	r1 := []rune(s1)
	r2 := []rune(s2)

	if len(r1) < 2 || len(r2) < 2 {
		if len(r1) == 0 && len(r2) == 0 {
			return 1.0
		}
		return LevenshteinSimilarity(s1, s2)
	}

	bg1 := make(map[string]int, len(r1)-1)
	for i := 0; i < len(r1)-1; i++ {
		bg1[string(r1[i:i+2])]++
	}

	intersection := 0
	for i := 0; i < len(r2)-1; i++ {
		bg := string(r2[i : i+2])
		if count, ok := bg1[bg]; ok && count > 0 {
			bg1[bg]--
			intersection++
		}
	}

	total := (len(r1) - 1) + (len(r2) - 1)
	return (2.0 * float64(intersection)) / float64(total)
}

// SegmentSimilarity computes similarity between two URL path segment slices.
// Incorporates dynamic parameter placeholders (like :id or {id}) and positional alignment.
func SegmentSimilarity(aSegments, bSegments []string) float64 {
	lenA := len(aSegments)
	lenB := len(bSegments)

	if lenA == 0 && lenB == 0 {
		return 1.0
	}
	if lenA == 0 || lenB == 0 {
		return 0.0
	}

	maxLen := max(lenA, lenB)
	minLen := min(lenA, lenB)

	// Positional comparison
	totalPositionalScore := 0.0
	for i := 0; i < minLen; i++ {
		segA := strings.ToLower(aSegments[i])
		segB := strings.ToLower(bSegments[i])

		if isParameterSegment(segA) || isParameterSegment(segB) {
			// Parameter match
			totalPositionalScore += 0.95
		} else if segA == segB {
			totalPositionalScore += 1.0
		} else {
			lev := LevenshteinSimilarity(segA, segB)
			jw := JaroWinklerSimilarity(segA, segB)
			totalPositionalScore += 0.6*jw + 0.4*lev
		}
	}
	positionalScore := totalPositionalScore / float64(maxLen)

	// Token set overlap (for shifted segments like /api/users vs /api/v1/users)
	matchedB := make([]bool, lenB)
	tokenMatchScore := 0.0
	for _, sA := range aSegments {
		lowerA := strings.ToLower(sA)
		bestSegScore := 0.0
		bestIdx := -1

		for j, sB := range bSegments {
			if matchedB[j] {
				continue
			}
			lowerB := strings.ToLower(sB)
			var curScore float64
			if isParameterSegment(lowerB) {
				curScore = 0.9
			} else if lowerA == lowerB {
				curScore = 1.0
			} else {
				curScore = 0.5*JaroWinklerSimilarity(lowerA, lowerB) + 0.5*LevenshteinSimilarity(lowerA, lowerB)
			}
			if curScore > bestSegScore {
				bestSegScore = curScore
				bestIdx = j
			}
		}

		if bestIdx >= 0 && bestSegScore >= 0.5 {
			matchedB[bestIdx] = true
			tokenMatchScore += bestSegScore
		}
	}
	tokenScore := tokenMatchScore / float64(maxLen)

	// Blend positional and token set matching (prioritizing positional)
	finalScore := 0.7*positionalScore + 0.3*tokenScore
	return math.Min(1.0, math.Max(0.0, finalScore))
}

// isParameterSegment checks if a segment represents a route parameter.
func isParameterSegment(seg string) bool {
	if strings.HasPrefix(seg, ":") || strings.HasPrefix(seg, "*") {
		return true
	}
	if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
		return true
	}
	return false
}

// HybridScore calculates the combined multi-metric score as defined in the system architecture:
// Score = 0.5 * SegmentSim + 0.3 * JaroWinkler + 0.2 * Levenshtein
func HybridScore(reqNormalizedPath, candidateNormalizedPath string, reqSegments, candidateSegments []string, isStructuralMatch bool) float64 {
	if isStructuralMatch {
		// Exact structural match (static segments match and dynamic params have values)
		return 0.99
	}

	normReq := strings.ToLower(reqNormalizedPath)
	normCand := strings.ToLower(candidateNormalizedPath)

	if normReq == normCand {
		return 1.0
	}

	segSim := SegmentSimilarity(reqSegments, candidateSegments)
	jwSim := JaroWinklerSimilarity(normReq, normCand)
	levSim := LevenshteinSimilarity(normReq, normCand)

	score := (0.5 * segSim) + (0.3 * jwSim) + (0.2 * levSim)

	// Round to 4 decimal places for clean representation
	rounded := math.Round(score*10000) / 10000
	return math.Min(1.0, math.Max(0.0, rounded))
}