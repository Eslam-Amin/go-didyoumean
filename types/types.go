package types

import (
	"net/http"
	"strings"
)

// HttpMethod represents an HTTP method string.
type HttpMethod string

const (
	MethodGet     HttpMethod = http.MethodGet
	MethodPost    HttpMethod = http.MethodPost
	MethodPut     HttpMethod = http.MethodPut
	MethodDelete  HttpMethod = http.MethodDelete
	MethodPatch   HttpMethod = http.MethodPatch
	MethodHead    HttpMethod = http.MethodHead
	MethodOptions HttpMethod = http.MethodOptions
	MethodConnect HttpMethod = http.MethodConnect
	MethodTrace   HttpMethod = http.MethodTrace
)

// RouteEntry represents a registered API route discovered via OpenAPI or router inspection.
type RouteEntry struct {
	Path        string       `json:"path"`
	Methods     []HttpMethod `json:"methods"`
	Summary     string       `json:"summary,omitempty"`
	Description string       `json:"description,omitempty"`
	OperationID string       `json:"operationId,omitempty"`
	Source      string       `json:"source"` // "openapi", "gin", "chi", "echo", "http"
}

// RouteSuggestion represents a recommended matching route returned for a 404 request.
type RouteSuggestion struct {
	Method           HttpMethod `json:"method"`
	Path             string     `json:"path"`
	Score            float64    `json:"score"`
	IsMethodMismatch bool       `json:"isMethodMismatch"`
	Summary          string     `json:"summary,omitempty"`
	Description      string     `json:"description,omitempty"`
}

// Options configures the didyoumean middleware and ranking engine.
type Options struct {
	// OpenAPISpec can be a file path string, remote URL string, []byte, or io.Reader.
	OpenAPISpec any

	// MaxSuggestions is the maximum number of suggestions to return (default: 3).
	MaxSuggestions int

	// Threshold is the minimum similarity score required for a suggestion (default: 0.4).
	Threshold float64

	// Enabled explicitly forces the middleware to be enabled or disabled regardless of environment.
	Enabled *bool

	// EnabledEnvironments defines environment names where didyoumean is active.
	// If empty, defaults to enabling in all non-production and non-release environments.
	EnabledEnvironments []string

	// ExcludeRoutes specifies path patterns or exact paths to exclude from suggestions (e.g. "/health", "/metrics").
	ExcludeRoutes []string

	// IncludeMethodMismatches controls whether wrong-method exact/structural matches are prioritized.
	// Defaults to true.
	IncludeMethodMismatches *bool

	// FormatResponse allows providing a custom response object for JSON responses.
	FormatResponse func(suggestions []RouteSuggestion, r *http.Request) any

	// CustomHTMLRenderer allows providing a custom renderer for HTML developer cards.
	CustomHTMLRenderer func(w http.ResponseWriter, r *http.Request, suggestions []RouteSuggestion) error
}

// ApplyDefaults fills in default values for Options when not explicitly configured.
func (o *Options) ApplyDefaults() {
	if o.MaxSuggestions <= 0 {
		o.MaxSuggestions = 3
	}
	if o.Threshold <= 0 {
		o.Threshold = 0.4
	}
	if o.IncludeMethodMismatches == nil {
		t := true
		o.IncludeMethodMismatches = &t
	}
}

// ShouldIncludeMethodMismatches returns true if method mismatches should be included.
func (o *Options) ShouldIncludeMethodMismatches() bool {
	if o.IncludeMethodMismatches == nil {
		return true
	}
	return *o.IncludeMethodMismatches
}

// IsRouteExcluded checks if a path matches any pattern in ExcludeRoutes.
func (o *Options) IsRouteExcluded(path string) bool {
	cleanPath := strings.TrimSpace(path)
	for _, excluded := range o.ExcludeRoutes {
		ex := strings.TrimSpace(excluded)
		if ex == "" {
			continue
		}
		// Exact match
		if cleanPath == ex {
			return true
		}
		// Prefix match if excluded ends with wildcard /* or /
		if strings.HasSuffix(ex, "/*") {
			prefix := strings.TrimSuffix(ex, "/*")
			if strings.HasPrefix(cleanPath, prefix) {
				return true
			}
		} else if strings.HasSuffix(ex, "/") {
			if strings.HasPrefix(cleanPath, ex) {
				return true
			}
		}
	}
	return false
}