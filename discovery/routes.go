package discovery

import (
	"net/http"
	"strings"
	"sync"

	"github.com/eslam-amin/didyoumean/matcher"
	"github.com/eslam-amin/didyoumean/types"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	"github.com/labstack/echo/v4"
)

// HarvestGinRoutes inspects registered routes on a Gin engine.
func HarvestGinRoutes(engine *gin.Engine) []types.RouteEntry {
	if engine == nil {
		return nil
	}

	routesInfo := engine.Routes()
	entries := make([]types.RouteEntry, 0, len(routesInfo))

	for _, r := range routesInfo {
		if isRootCatchAll(r.Path) {
			continue
		}
		entries = append(entries, types.RouteEntry{
			Path:    r.Path,
			Methods: []types.HttpMethod{types.HttpMethod(strings.ToUpper(r.Method))},
			Source:  "gin",
		})
	}

	return entries
}

// HarvestChiRoutes inspects registered routes on a Chi router using chi.Walk.
func HarvestChiRoutes(r chi.Routes) []types.RouteEntry {
	if r == nil {
		return nil
	}

	var entries []types.RouteEntry

	_ = chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if isRootCatchAll(route) {
			return nil
		}
		entries = append(entries, types.RouteEntry{
			Path:    route,
			Methods: []types.HttpMethod{types.HttpMethod(strings.ToUpper(method))},
			Source:  "chi",
		})
		return nil
	})

	return entries
}

// HarvestEchoRoutes inspects registered routes on an Echo instance.
func HarvestEchoRoutes(e *echo.Echo) []types.RouteEntry {
	if e == nil {
		return nil
	}

	routes := e.Routes()
	entries := make([]types.RouteEntry, 0, len(routes))

	for _, r := range routes {
		if isRootCatchAll(r.Path) {
			continue
		}
		entries = append(entries, types.RouteEntry{
			Path:    r.Path,
			Methods: []types.HttpMethod{types.HttpMethod(strings.ToUpper(r.Method))},
			Source:  "echo",
		})
	}

	return entries
}

func isRootCatchAll(p string) bool {
	p = strings.TrimSpace(p)
	return p == "/*" || p == "*" || p == "/*any" || p == "/*filepath"
}