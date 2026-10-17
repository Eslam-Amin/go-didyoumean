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

// MergeRoutes merges two route lists (e.g. OpenAPI spec and Router harvested routes),
// deduplicating by normalized path and HTTP method while preserving rich metadata (summary, description).
func MergeRoutes(primary, secondary []types.RouteEntry) []types.RouteEntry {
	type key struct {
		normPath string
		method   string
	}

	routeMap := make(map[key]*types.RouteEntry)
	var orderedKeys []key

	// Add routes helper
	addRoute := func(entry types.RouteEntry) {
		normPath := matcher.NormalizePath(entry.Path)
		for _, m := range entry.Methods {
			mUpper := strings.ToUpper(string(m))
			k := key{normPath: normPath, method: mUpper}

			if existing, found := routeMap[k]; found {
				// Merge metadata if empty
				if existing.Summary == "" && entry.Summary != "" {
					existing.Summary = entry.Summary
				}
				if existing.Description == "" && entry.Description != "" {
					existing.Description = entry.Description
				}
				if existing.OperationID == "" && entry.OperationID != "" {
					existing.OperationID = entry.OperationID
				}
			} else {
				newEntry := &types.RouteEntry{
					Path:        entry.Path,
					Methods:     []types.HttpMethod{types.HttpMethod(mUpper)},
					Summary:     entry.Summary,
					Description: entry.Description,
					OperationID: entry.OperationID,
					Source:      entry.Source,
				}
				routeMap[k] = newEntry
				orderedKeys = append(orderedKeys, k)
			}
		}
	}

	for _, r := range primary {
		addRoute(r)
	}
	for _, r := range secondary {
		addRoute(r)
	}

	result := make([]types.RouteEntry, 0, len(orderedKeys))
	for _, k := range orderedKeys {
		result = append(result, *routeMap[k])
	}

	return result
}