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