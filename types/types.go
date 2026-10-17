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