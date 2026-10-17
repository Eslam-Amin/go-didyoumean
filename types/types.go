// Package types defines the core domain models, configuration options,
// and response structures utilized throughout the didyoumean package.
package types

import (
	"net/http"
	"strings"
)

// HttpMethod represents an HTTP method string (e.g., "GET", "POST").
type HttpMethod string

const (
	// MethodGet represents the HTTP GET method.
	MethodGet HttpMethod = http.MethodGet
	// MethodPost represents the HTTP POST method.
	MethodPost HttpMethod = http.MethodPost
	// MethodPut represents the HTTP PUT method.
	MethodPut HttpMethod = http.MethodPut
	// MethodDelete represents the HTTP DELETE method.
	MethodDelete HttpMethod = http.MethodDelete
	// MethodPatch represents the HTTP PATCH method.
	MethodPatch HttpMethod = http.MethodPatch
	// MethodHead represents the HTTP HEAD method.
	MethodHead HttpMethod = http.MethodHead
	// MethodOptions represents the HTTP OPTIONS method.
	MethodOptions HttpMethod = http.MethodOptions
	// MethodConnect represents the HTTP CONNECT method.
	MethodConnect HttpMethod = http.MethodConnect
	// MethodTrace represents the HTTP TRACE method.
	MethodTrace HttpMethod = http.MethodTrace
)

// RouteEntry represents a registered API route discovered via OpenAPI or router inspection.
type RouteEntry struct {
	// Path is the URL pattern of the registered endpoint.
	Path string `json:"path"`
	// Methods is the list of supported HTTP methods for this path.
	Methods []HttpMethod `json:"methods"`
	// Summary is an optional human-readable brief summary from OpenAPI specs.
	Summary string `json:"summary,omitempty"`
	// Description is an optional detailed endpoint description from OpenAPI specs.
	Description string `json:"description,omitempty"`
	// OperationID is the unique OpenAPI operation identifier if present.
	OperationID string `json:"operationId,omitempty"`
	// Source indicates where the route was harvested from ("openapi", "gin", "chi", "echo", "http").
	Source string `json:"source"`
}

// RouteSuggestion represents a recommended matching route returned for a 404 request.
type RouteSuggestion struct {
	// Method is the recommended HTTP method for this endpoint.
	Method HttpMethod `json:"method"`
	// Path is the recommended URL path.
	Path string `json:"path"`
	// Score is the computed hybrid similarity score ranging from 0.0 to 1.0.
	Score float64 `json:"score"`
	// IsMethodMismatch is true if the requested path matches structurally but with the wrong HTTP method.
	IsMethodMismatch bool `json:"isMethodMismatch"`
	// Summary is an optional description or summary of the suggested route.
	Summary string `json:"summary,omitempty"`
	// Description provides optional expanded details about the suggested route.
	Description string `json:"description,omitempty"`
}

// Options configures the didyoumean middleware and ranking engine.
type Options struct {
	// OpenAPISpec can be a file path string, remote URL string, []byte, or io.Reader
	// representing an OpenAPI 2.0 (Swagger) or 3.x specification.
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
// Sets MaxSuggestions to 3, Threshold to 0.4, and IncludeMethodMismatches to true.
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
// Defaults to true if IncludeMethodMismatches is unset.
func (o *Options) ShouldIncludeMethodMismatches() bool {
	if o.IncludeMethodMismatches == nil {
		return true
	}
	return *o.IncludeMethodMismatches
}

// IsRouteExcluded checks if a path matches any pattern in ExcludeRoutes.
// Supports exact matching, trailing wildcard matching (e.g. "/*"), and folder prefix matching ("/").
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

// ErrorResponse represents the default JSON 404 response payload.
type ErrorResponse struct {
	// StatusCode is the HTTP response code (typically 404).
	StatusCode int `json:"statusCode"`
	// Error describes the general error category (e.g., "Not Found").
	Error string `json:"error"`
	// Message contains a descriptive message including the requested method and path.
	Message string `json:"message"`
	// Suggestions contains the top closest matching routes.
	Suggestions []RouteSuggestion `json:"suggestions"`
}