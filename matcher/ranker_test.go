package matcher

import (
	"testing"

	"github.com/eslam-amin/go-didyoumean/types"
)

func TestRankRoutes_MethodMismatchPrioritization(t *testing.T) {
	routes := []types.RouteEntry{
		{
			Path:    "/api/v1/users",
			Methods: []types.HttpMethod{types.MethodGet},
		},
		{
			Path:    "/api/v1/user-profiles",
			Methods: []types.HttpMethod{types.MethodPost},
		},
	}

	opts := types.Options{}
	opts.ApplyDefaults()

	// Client calls POST /api/v1/users
	// Expected: method mismatch on /api/v1/users with score 1.0, isMethodMismatch = true, sorted to top
	suggestions := RankRoutes("POST", "/api/v1/users", routes, opts)

	if len(suggestions) == 0 {
		t.Fatalf("expected suggestions, got none")
	}

	top := suggestions[0]
	if top.Path != "/api/v1/users" {
		t.Errorf("expected top path /api/v1/users, got %s", top.Path)
	}
	if top.Method != types.MethodGet {
		t.Errorf("expected top method GET, got %s", top.Method)
	}
	if top.Score != 1.0 {
		t.Errorf("expected score 1.0, got %f", top.Score)
	}
	if !top.IsMethodMismatch {
		t.Errorf("expected isMethodMismatch to be true")
	}
}

func TestRankRoutes_TypoMatching(t *testing.T) {
	routes := []types.RouteEntry{
		{
			Path:    "/api/v1/users",
			Methods: []types.HttpMethod{types.MethodGet},
		},
		{
			Path:    "/api/v1/orders",
			Methods: []types.HttpMethod{types.MethodGet},
		},
	}

	opts := types.Options{Threshold: 0.5, MaxSuggestions: 3}
	suggestions := RankRoutes("GET", "/api/v1/usr", routes, opts)

	if len(suggestions) == 0 {
		t.Fatalf("expected suggestions for /api/v1/usr, got none")
	}

	top := suggestions[0]
	if top.Path != "/api/v1/users" {
		t.Errorf("expected /api/v1/users, got %s", top.Path)
	}
	if top.IsMethodMismatch {
		t.Errorf("expected isMethodMismatch false for same method")
	}
	if top.Score < 0.7 {
		t.Errorf("expected score > 0.7 for typo, got %f", top.Score)
	}
}

func TestRankRoutes_ExcludeRoutes(t *testing.T) {
	routes := []types.RouteEntry{
		{
			Path:    "/health",
			Methods: []types.HttpMethod{types.MethodGet},
		},
		{
			Path:    "/api/v1/users",
			Methods: []types.HttpMethod{types.MethodGet},
		},
	}

	opts := types.Options{
		ExcludeRoutes: []string{"/health"},
	}

	suggestions := RankRoutes("GET", "/healt", routes, opts)
	for _, s := range suggestions {
		if s.Path == "/health" {
			t.Errorf("route /health should have been excluded")
		}
	}
}
