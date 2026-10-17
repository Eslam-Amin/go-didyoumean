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

// GinMiddleware creates a Gin middleware that intercepts 404 responses and appends route suggestions.
func GinMiddleware(engine *gin.Engine, opts Options) gin.HandlerFunc {
	opts.ApplyDefaults()

	var cachedRoutes []types.RouteEntry
	var harvestOnce sync.Once

	return func(c *gin.Context) {
		// Production Guardrail: immediate transparent no-op with zero allocations
		if !IsEnabled(opts) {
			c.Next()
			return
		}

		c.Next()

		if c.Writer.Status() == http.StatusNotFound {
			harvestOnce.Do(func() {
				cachedRoutes = harvestAllRoutes(engine, opts)
			})

			suggestions := matcher.RankRoutes(c.Request.Method, c.Request.URL.Path, cachedRoutes, opts)
			if len(suggestions) > 0 {
				if PrefersHTML(c.Request) {
					if opts.CustomHTMLRenderer != nil {
						if err := opts.CustomHTMLRenderer(c.Writer, c.Request, suggestions); err == nil {
							return
						}
					}
					_ = RenderHTMLCard(c.Writer, c.Request, suggestions)
					return
				}

				if opts.FormatResponse != nil {
					c.JSON(http.StatusNotFound, opts.FormatResponse(suggestions, c.Request))
				} else {
					c.JSON(http.StatusNotFound, DefaultFormatResponse(c.Request.Method, c.Request.URL.Path, suggestions))
				}
			}
		}
	}
}

// GinNoRouteHandler creates a gin.HandlerFunc to use with engine.NoRoute(...).
func GinNoRouteHandler(engine *gin.Engine, opts Options) gin.HandlerFunc {
	opts.ApplyDefaults()

	var cachedRoutes []types.RouteEntry
	var harvestOnce sync.Once

	return func(c *gin.Context) {
		if !IsEnabled(opts) {
			c.String(http.StatusNotFound, "404 page not found")
			return
		}

		harvestOnce.Do(func() {
			cachedRoutes = harvestAllRoutes(engine, opts)
		})

		suggestions := matcher.RankRoutes(c.Request.Method, c.Request.URL.Path, cachedRoutes, opts)
		if len(suggestions) > 0 {
			if PrefersHTML(c.Request) {
				if opts.CustomHTMLRenderer != nil {
					if err := opts.CustomHTMLRenderer(c.Writer, c.Request, suggestions); err == nil {
						return
					}
				}
				_ = RenderHTMLCard(c.Writer, c.Request, suggestions)
				return
			}

			if opts.FormatResponse != nil {
				c.JSON(http.StatusNotFound, opts.FormatResponse(suggestions, c.Request))
			} else {
				c.JSON(http.StatusNotFound, DefaultFormatResponse(c.Request.Method, c.Request.URL.Path, suggestions))
			}
			return
		}

		c.String(http.StatusNotFound, "404 page not found")
	}
}

// ChiNotFoundHandler creates an http.HandlerFunc to use with router.NotFound(...).
func ChiNotFoundHandler(r chi.Routes, opts Options) http.HandlerFunc {
	opts.ApplyDefaults()

	var cachedRoutes []types.RouteEntry
	var harvestOnce sync.Once

	return func(w http.ResponseWriter, req *http.Request) {
		if !IsEnabled(opts) {
			http.NotFound(w, req)
			return
		}

		harvestOnce.Do(func() {
			var openapiRoutes []types.RouteEntry
			if opts.OpenAPISpec != nil {
				if parsed, err := discovery.ParseOpenAPISpec(opts.OpenAPISpec); err == nil {
					openapiRoutes = parsed
				}
			}
			chiRoutes := discovery.HarvestChiRoutes(r)
			cachedRoutes = discovery.MergeRoutes(openapiRoutes, chiRoutes)
		})

		suggestions := matcher.RankRoutes(req.Method, req.URL.Path, cachedRoutes, opts)
		if len(suggestions) > 0 {
			WriteResponse(w, req, suggestions, opts)
			return
		}

		http.NotFound(w, req)
	}
}