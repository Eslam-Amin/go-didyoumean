package matcher

import (
	"reflect"
	"testing"
)

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "/"},
		{"/", "/"},
		{"//api///users//", "/api/users"},
		{"/api/v1/users?page=1&limit=10", "/api/v1/users"},
		{"/api/v1/users#section", "/api/v1/users"},
		{"/api/v1/users?foo=bar#baz", "/api/v1/users"},
		{"users", "/users"},
		{"/a/b/../c", "/a/c"},
	}

	for _, tt := range tests {
		got := NormalizePath(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizePath(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}

func TestSplitSegments(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"/", nil},
		{"", nil},
		{"/api/v1/users", []string{"api", "v1", "users"}},
		{"//api///users//", []string{"api", "users"}},
	}

	for _, tt := range tests {
		got := SplitSegments(tt.input)
		if !reflect.DeepEqual(got, tt.expected) {
			t.Errorf("SplitSegments(%q) = %v, expected %v", tt.input, got, tt.expected)
		}
	}
}

func TestIsStructuralMatch(t *testing.T) {
	tests := []struct {
		name          string
		req           []string
		route         []string
		expectMatches bool
		expectExact   bool
	}{
		{
			name:          "exact match",
			req:           []string{"api", "v1", "users"},
			route:         []string{"api", "v1", "users"},
			expectMatches: true,
			expectExact:   true,
		},
		{
			name:          "dynamic parameter :id",
			req:           []string{"api", "v1", "users", "123"},
			route:         []string{"api", "v1", "users", ":id"},
			expectMatches: true,
			expectExact:   false,
		},
		{
			name:          "dynamic parameter {id}",
			req:           []string{"api", "v1", "users", "456"},
			route:         []string{"api", "v1", "users", "{id}"},
			expectMatches: true,
			expectExact:   false,
		},
		{
			name:          "catch all wildcard *filepath",
			req:           []string{"static", "css", "main.css"},
			route:         []string{"static", "*filepath"},
			expectMatches: true,
			expectExact:   false,
		},
		{
			name:          "segment count mismatch",
			req:           []string{"api", "users"},
			route:         []string{"api", "v1", "users"},
			expectMatches: false,
			expectExact:   false,
		},
		{
			name:          "static segment mismatch",
			req:           []string{"api", "v2", "users"},
			route:         []string{"api", "v1", "users"},
			expectMatches: false,
			expectExact:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches, exact := IsStructuralMatch(tt.req, tt.route)
			if matches != tt.expectMatches || exact != tt.expectExact {
				t.Errorf("IsStructuralMatch(%v, %v) = (%v, %v), expected (%v, %v)",
					tt.req, tt.route, matches, exact, tt.expectMatches, tt.expectExact)
			}
		})
	}
}

func TestStandardizeWildcards(t *testing.T) {
	got := StandardizeWildcards("/users/{id}/posts/{slug}")
	expected := "/users/:id/posts/:slug"
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}
