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