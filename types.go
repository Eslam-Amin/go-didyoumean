package didyoumean

import (
	"net/http"

	"github.com/eslam-amin/didyoumean/types"
)

// HttpMethod represents an HTTP method string.
type HttpMethod = types.HttpMethod

const (
	MethodGet     = types.MethodGet
	MethodPost    = types.MethodPost
	MethodPut     = types.MethodPut
	MethodDelete  = types.MethodDelete
	MethodPatch   = types.MethodPatch
	MethodHead    = types.MethodHead
	MethodOptions = types.MethodOptions
	MethodConnect = types.MethodConnect
	MethodTrace   = types.MethodTrace
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
func DefaultFormatResponse(method, path string, suggestions []RouteSuggestion) ErrorResponse {
	return ErrorResponse{
		StatusCode:  http.StatusNotFound,
		Error:       "Not Found",
		Message:     "Cannot " + method + " " + path,
		Suggestions: suggestions,
	}
}
