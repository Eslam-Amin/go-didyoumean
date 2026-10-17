package matcher

import (
	"testing"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		a, b     string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "b", 1},
		{"users", "users", 0},
		{"usr", "user", 1},
		{"kitten", "sitting", 3},
		{"/api/v1/users", "/api/v1/user", 1},
		{"café", "cafe", 1},
	}

	for _, tt := range tests {
		dist := LevenshteinDistance(tt.a, tt.b)
		if dist != tt.expected {
			t.Errorf("LevenshteinDistance(%q, %q) = %d, expected %d", tt.a, tt.b, dist, tt.expected)
		}
	}
}

func TestLevenshteinSimilarity(t *testing.T) {
	if sim := LevenshteinSimilarity("users", "users"); sim != 1.0 {
		t.Errorf("expected 1.0, got %f", sim)
	}
	if sim := LevenshteinSimilarity("", ""); sim != 1.0 {
		t.Errorf("expected 1.0, got %f", sim)
	}
	sim := LevenshteinSimilarity("users", "user")
	if sim < 0.79 || sim > 0.81 {
		t.Errorf("expected ~0.80, got %f", sim)
	}
}

func TestJaroWinklerSimilarity(t *testing.T) {
	// Identical
	if sim := JaroWinklerSimilarity("hello", "hello"); sim != 1.0 {
		t.Errorf("expected 1.0, got %f", sim)
	}
	// Common prefix boost: /api/v1/users vs /api/v1/user
	sim1 := JaroWinklerSimilarity("/api/v1/users", "/api/v1/user")
	if sim1 < 0.9 {
		t.Errorf("expected high similarity for prefix match, got %f", sim1)
	}
	// Completely disjoint
	sim2 := JaroWinklerSimilarity("abc", "xyz")
	if sim2 != 0.0 {
		t.Errorf("expected 0.0, got %f", sim2)
	}
}

func TestSorensenDiceSimilarity(t *testing.T) {
	if sim := SorensenDiceSimilarity("night", "nacht"); sim <= 0.0 || sim >= 1.0 {
		t.Errorf("expected valid dice similarity, got %f", sim)
	}
	if sim := SorensenDiceSimilarity("same", "same"); sim != 1.0 {
		t.Errorf("expected 1.0, got %f", sim)
	}
}

func TestSegmentSimilarity(t *testing.T) {
	// Exact
	s1 := []string{"api", "v1", "users"}
	s2 := []string{"api", "v1", "users"}
	if sim := SegmentSimilarity(s1, s2); sim != 1.0 {
		t.Errorf("expected 1.0, got %f", sim)
	}

	// Minor typo in last segment
	s3 := []string{"api", "v1", "usr"}
	sim := SegmentSimilarity(s1, s3)
	if sim < 0.8 {
		t.Errorf("expected > 0.8 for segment typo, got %f", sim)
	}

	// Dynamic parameter placeholder
	sRoute := []string{"api", "v1", "users", ":id"}
	sReq := []string{"api", "v1", "users", "123"}
	simParam := SegmentSimilarity(sReq, sRoute)
	if simParam < 0.95 {
		t.Errorf("expected >= 0.95 for param segment match, got %f", simParam)
	}
}
