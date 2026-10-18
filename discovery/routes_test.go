package discovery

import (
	"net/http"
	"sync"
	"testing"

	"github.com/eslam-amin/go-didyoumean/types"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	"github.com/labstack/echo/v4"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHarvestGinRoutes(t *testing.T) {
	engine := gin.New()
	engine.GET("/api/v1/users", func(c *gin.Context) {})
	engine.POST("/api/v1/users", func(c *gin.Context) {})
	engine.DELETE("/api/v1/users/:id", func(c *gin.Context) {})

	routes := HarvestGinRoutes(engine)
	if len(routes) < 3 {
		t.Fatalf("expected at least 3 routes, got %d", len(routes))
	}

	foundGet := false
	for _, r := range routes {
		if r.Path == "/api/v1/users" && len(r.Methods) > 0 && r.Methods[0] == types.MethodGet {
			foundGet = true
		}
	}
	if !foundGet {
		t.Errorf("expected to harvest GET /api/v1/users from Gin")
	}
}

func TestHarvestChiRoutes(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/articles", func(w http.ResponseWriter, r *http.Request) {})
	r.Post("/articles", func(w http.ResponseWriter, r *http.Request) {})
	r.Get("/articles/{id}", func(w http.ResponseWriter, r *http.Request) {})

	routes := HarvestChiRoutes(r)
	if len(routes) < 3 {
		t.Fatalf("expected at least 3 routes, got %d", len(routes))
	}

	foundGet := false
	for _, rt := range routes {
		if rt.Path == "/articles" && len(rt.Methods) > 0 && rt.Methods[0] == types.MethodGet {
			foundGet = true
		}
	}
	if !foundGet {
		t.Errorf("expected to harvest GET /articles from Chi")
	}
}

func TestHarvestEchoRoutes(t *testing.T) {
	e := echo.New()
	e.GET("/items", func(c echo.Context) error { return nil })
	e.POST("/items", func(c echo.Context) error { return nil })

	routes := HarvestEchoRoutes(e)
	if len(routes) < 2 {
		t.Fatalf("expected at least 2 routes, got %d", len(routes))
	}

	foundItem := false
	for _, rt := range routes {
		if rt.Path == "/items" {
			foundItem = true
		}
	}
	if !foundItem {
		t.Errorf("expected to harvest /items from Echo")
	}
}

func TestMergeRoutes(t *testing.T) {
	openapiRoutes := []types.RouteEntry{
		{
			Path:        "/api/v1/users",
			Methods:     []types.HttpMethod{types.MethodGet},
			Summary:     "List users",
			Description: "Returns all active users",
			Source:      "openapi",
		},
	}

	routerRoutes := []types.RouteEntry{
		{
			Path:    "/api/v1/users",
			Methods: []types.HttpMethod{types.MethodGet},
			Source:  "gin",
		},
		{
			Path:    "/api/v1/users",
			Methods: []types.HttpMethod{types.MethodPost},
			Source:  "gin",
		},
		{
			Path:    "/api/v1/health",
			Methods: []types.HttpMethod{types.MethodGet},
			Source:  "gin",
		},
	}

	merged := MergeRoutes(openapiRoutes, routerRoutes)

	// Expected:
	// GET /api/v1/users (with summary "List users")
	// POST /api/v1/users
	// GET /api/v1/health
	if len(merged) != 3 {
		t.Fatalf("expected 3 merged routes, got %d", len(merged))
	}

	for _, m := range merged {
		if m.Path == "/api/v1/users" && len(m.Methods) > 0 && m.Methods[0] == types.MethodGet {
			if m.Summary != "List users" {
				t.Errorf("expected metadata to be preserved from OpenAPI, got %q", m.Summary)
			}
		}
	}
}

func TestRouteRegistry_ThreadSafety(t *testing.T) {
	reg := NewRouteRegistry()
	var wg sync.WaitGroup

	harvestFn := func() []types.RouteEntry {
		return []types.RouteEntry{
			{Path: "/test", Methods: []types.HttpMethod{types.MethodGet}},
		}
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			routes := reg.EnsureHarvested(harvestFn)
			if len(routes) != 1 {
				t.Errorf("expected 1 route, got %d", len(routes))
			}
		}()
	}

	wg.Wait()
	if !reg.IsCached() {
		t.Errorf("expected registry to be marked as cached")
	}
}
