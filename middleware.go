package didyoumean

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/eslam-amin/didyoumean/discovery"
	"github.com/eslam-amin/didyoumean/matcher"
	"github.com/eslam-amin/didyoumean/types"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	"github.com/labstack/echo/v4"
)

// RankRoutes evaluates candidate routes against a requested HTTP method and URL path,
// returning the closest matching route suggestions.
func RankRoutes(method, path string, routes []types.RouteEntry, opts types.Options) []types.RouteSuggestion {
	return matcher.RankRoutes(method, path, routes, opts)
}

// IsEnabled returns true if the didyoumean middleware is active in the current environment.
// Production safety: returns false if production is detected, executing zero route walking.
func IsEnabled(opts types.Options) bool {
	if opts.Enabled != nil {
		return *opts.Enabled
	}

	env := strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
	if env == "" {
		env = strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	}
	if env == "" {
		env = strings.ToLower(strings.TrimSpace(os.Getenv("GIN_MODE")))
	}
	if env == "" {
		env = strings.ToLower(strings.TrimSpace(os.Getenv("NODE_ENV")))
	}
	if env == "" {
		env = strings.ToLower(strings.TrimSpace(os.Getenv("GO_ENV")))
	}

	// If explicit allowed environments configured
	if len(opts.EnabledEnvironments) > 0 {
		for _, e := range opts.EnabledEnvironments {
			if strings.EqualFold(e, env) {
				return true
			}
		}
		return false
	}

	// Default guardrail: disabled in production / release
	if env == "production" || env == "prod" || env == "release" || gin.Mode() == gin.ReleaseMode {
		return false
	}

	return true
}

// PrefersHTML returns true if the client Accept header indicates preference for HTML.
func PrefersHTML(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "text/html") {
		if strings.Contains(accept, "application/json") {
			return strings.Index(accept, "text/html") < strings.Index(accept, "application/json")
		}
		return true
	}
	return false
}

// WriteResponse outputs suggestions to http.ResponseWriter respecting Content-Type negotiation.
func WriteResponse(w http.ResponseWriter, r *http.Request, suggestions []types.RouteSuggestion, opts types.Options) {
	if PrefersHTML(r) {
		if opts.CustomHTMLRenderer != nil {
			if err := opts.CustomHTMLRenderer(w, r, suggestions); err == nil {
				return
			}
		}
		_ = RenderHTMLCard(w, r, suggestions)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)

	var payload any
	if opts.FormatResponse != nil {
		payload = opts.FormatResponse(suggestions, r)
	} else {
		payload = DefaultFormatResponse(r.Method, r.URL.Path, suggestions)
	}

	_ = json.NewEncoder(w).Encode(payload)
}