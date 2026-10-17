package types

import "testing"

func TestOptions_ApplyDefaults(t *testing.T) {
	opts := Options{}
	opts.ApplyDefaults()

	if opts.MaxSuggestions != 3 {
		t.Errorf("expected default MaxSuggestions 3, got %d", opts.MaxSuggestions)
	}
	if opts.Threshold != 0.4 {
		t.Errorf("expected default Threshold 0.4, got %f", opts.Threshold)
	}
	if !opts.ShouldIncludeMethodMismatches() {
		t.Errorf("expected IncludeMethodMismatches default true")
	}

	f := false
	opts.IncludeMethodMismatches = &f
	if opts.ShouldIncludeMethodMismatches() {
		t.Errorf("expected IncludeMethodMismatches to be false")
	}
}

func TestOptions_IsRouteExcluded(t *testing.T) {
	opts := Options{
		ExcludeRoutes: []string{
			"/health",
			"/metrics/*",
			"/internal/",
		},
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{"/health", true},
		{"/health/check", false},
		{"/metrics/prometheus", true},
		{"/metrics", true},
		{"/internal/debug", true},
		{"/api/v1/users", false},
	}

	for _, tt := range tests {
		if got := opts.IsRouteExcluded(tt.path); got != tt.expected {
			t.Errorf("IsRouteExcluded(%q) = %v, expected %v", tt.path, got, tt.expected)
		}
	}
}
