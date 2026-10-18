// Package didyoumean provides a zero-overhead, developer-friendly 404 route suggester
// for Go REST applications (net/http, Gin, Chi, and Echo).
//
// When an incoming HTTP request does not match any registered route, the middleware
// catches the 404, inspects all registered endpoints (via OpenAPI/Swagger specifications
// and live router tables), computes fuzzy string and structural similarities, and returns
// the top suggestions via JSON or an interactive developer HTML error card.
package didyoumean

import (
	"net/http"

	"github.com/eslam-amin/didyoumean/types"
)

// HttpMethod represents an HTTP method string.
type HttpMethod = types.HttpMethod

const (
	// MethodGet represents the HTTP GET method.
	MethodGet = types.MethodGet
	// MethodPost represents the HTTP POST method.
	MethodPost = types.MethodPost
	// MethodPut represents the HTTP PUT method.
	MethodPut = types.MethodPut
	// MethodDelete represents the HTTP DELETE method.
	MethodDelete = types.MethodDelete
	// MethodPatch represents the HTTP PATCH method.
	MethodPatch = types.MethodPatch
	// MethodHead represents the HTTP HEAD method.
	MethodHead = types.MethodHead
	// MethodOptions represents the HTTP OPTIONS method.
	MethodOptions = types.MethodOptions
	// MethodConnect represents the HTTP CONNECT method.
	MethodConnect = types.MethodConnect
	// MethodTrace represents the HTTP TRACE method.
	MethodTrace = types.MethodTrace
)

// RouteEntry represents a registered API route discovered via OpenAPI or router inspection.
type RouteEntry = types.RouteEntry

// RouteSuggestion represents a recommended matching route returned for a 404 request.
type RouteSuggestion = types.RouteSuggestion

// Options configures the didyoumean middleware and ranking engine.
type Options = types.Options

// ErrorResponse represents the default JSON 404 response payload.
type ErrorResponse = types.ErrorResponse

// DefaultFormatResponse creates the standard ErrorResponse struct.
// It formats the error as "Cannot <METHOD> <PATH>" alongside the slice of closest route suggestions.
func DefaultFormatResponse(method, path string, suggestions []RouteSuggestion) ErrorResponse {
	return ErrorResponse{
		StatusCode:  http.StatusNotFound,
		Error:       "Not Found",
		Message:     "Cannot " + method + " " + path,
		Suggestions: suggestions,
	}
}
