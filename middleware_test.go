package didyoumean

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/eslam-amin/go-didyoumean/types"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	"github.com/labstack/echo/v4"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGinMiddleware_FuzzyAndMethodMismatch(t *testing.T) {
	engine := gin.New()
	opts := Options{
		MaxSuggestions: 3,
		Threshold:      0.4,
	}
	engine.Use(GinMiddleware(engine, opts))

	// Register routes
	engine.GET("/api/v1/users", func(c *gin.Context) {
		c.String(http.StatusOK, "users list")
	})
	engine.GET("/api/v1/orders", func(c *gin.Context) {
		c.String(http.StatusOK, "orders list")
	})

	// 1. Test Typo Match: GET /api/v1/usr
	{
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/usr", nil)
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}

		var res ErrorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to decode JSON response: %v, body: %s", err, rec.Body.String())
		}

		if len(res.Suggestions) == 0 {
			t.Fatalf("expected suggestions for /api/v1/usr, got none")
		}

		top := res.Suggestions[0]
		if top.Path != "/api/v1/users" {
			t.Errorf("expected top suggestion /api/v1/users, got %s", top.Path)
		}
		if top.Method != MethodGet {
			t.Errorf("expected GET method, got %s", top.Method)
		}
		if top.IsMethodMismatch {
			t.Errorf("expected isMethodMismatch to be false for GET")
		}
	}

	// 2. Test Method Mismatch: POST /api/v1/users (only GET is registered)
	{
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/users", nil)
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}

		var res ErrorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to decode JSON: %v", err)
		}

		if len(res.Suggestions) == 0 {
			t.Fatalf("expected suggestions for POST /api/v1/users, got none")
		}

		top := res.Suggestions[0]
		if top.Path != "/api/v1/users" {
			t.Errorf("expected /api/v1/users, got %s", top.Path)
		}
		if top.Method != MethodGet {
			t.Errorf("expected GET, got %s", top.Method)
		}
		if !top.IsMethodMismatch {
			t.Errorf("expected isMethodMismatch to be true for POST on GET route")
		}
		if top.Score != 1.0 {
			t.Errorf("expected score 1.0, got %f", top.Score)
		}
	}
}

func TestGinMiddleware_ProductionGuardrail(t *testing.T) {
	// Simulate production environment
	os.Setenv("ENV", "production")
	defer os.Unsetenv("ENV")

	engine := gin.New()
	opts := Options{}
	engine.Use(GinMiddleware(engine, opts))

	engine.GET("/api/v1/users", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/usr", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	// Should NOT return didyoumean JSON suggestions in production
	if strings.Contains(rec.Body.String(), "suggestions") {
		t.Errorf("production mode should bypass didyoumean suggestions, body: %s", rec.Body.String())
	}
}

func TestGinMiddleware_ContentNegotiation(t *testing.T) {
	engine := gin.New()
	opts := Options{}
	engine.Use(GinMiddleware(engine, opts))

	engine.GET("/api/v1/users", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// HTML Accept header
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/usr", nil)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("expected text/html content type, got %s", contentType)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Errorf("expected HTML document, got %s", body)
	}
	if !strings.Contains(body, "/api/v1/users") {
		t.Errorf("expected suggested path in HTML body")
	}
}

func TestGinMiddleware_CustomFormatResponse(t *testing.T) {
	engine := gin.New()
	opts := Options{
		FormatResponse: func(suggestions []RouteSuggestion, r *http.Request) any {
			return map[string]any{
				"custom_error": "Route Not Found",
				"hint_count":   len(suggestions),
			}
		},
	}
	engine.Use(GinMiddleware(engine, opts))

	engine.GET("/api/v1/users", func(c *gin.Context) {})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/usr", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	var customRes map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &customRes); err != nil {
		t.Fatalf("failed to parse custom response: %v", err)
	}

	if customRes["custom_error"] != "Route Not Found" {
		t.Errorf("expected custom_error, got %v", customRes["custom_error"])
	}
	if customRes["hint_count"].(float64) < 1 {
		t.Errorf("expected hint_count >= 1")
	}
}

func TestChiNotFoundHandler(t *testing.T) {
	r := chi.NewRouter()
	opts := Options{
		OpenAPISpec: `
swagger: "2.0"
paths:
  /articles:
    get:
      summary: Get articles
`,
	}
	r.NotFound(ChiNotFoundHandler(r, opts))

	r.Get("/items", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Request /item -> should suggest /items from router
	req, _ := http.NewRequest(http.MethodGet, "/item", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	var res ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if len(res.Suggestions) == 0 {
		t.Fatalf("expected suggestions on Chi 404, got none")
	}
	if res.Suggestions[0].Path != "/items" {
		t.Errorf("expected /items suggestion, got %s", res.Suggestions[0].Path)
	}

	// Request /article -> should suggest /articles from OpenAPI
	req2, _ := http.NewRequest(http.MethodGet, "/article", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	var res2 ErrorResponse
	_ = json.Unmarshal(rec2.Body.Bytes(), &res2)
	if len(res2.Suggestions) == 0 || res2.Suggestions[0].Path != "/articles" {
		t.Errorf("expected /articles suggestion from OpenAPI on Chi")
	}
}

func TestEchoNotFoundHandler(t *testing.T) {
	e := echo.New()
	opts := Options{}
	e.RouteNotFound("/*", EchoNotFoundHandler(e, opts))

	e.GET("/customers", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/customer", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	var res ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse echo response: %v", err)
	}

	if len(res.Suggestions) == 0 {
		t.Fatalf("expected suggestions from Echo")
	}
	if res.Suggestions[0].Path != "/customers" {
		t.Errorf("expected /customers suggestion, got %s", res.Suggestions[0].Path)
	}
}

func TestHTTPNotFoundHandler_StandardNetHTTP(t *testing.T) {
	knownRoutes := []types.RouteEntry{
		{
			Path:    "/api/v1/reports",
			Methods: []types.HttpMethod{MethodGet},
		},
	}

	opts := Options{}
	handler := HTTPNotFoundHandler(knownRoutes, opts)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/report", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	var res ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(res.Suggestions) == 0 {
		t.Fatalf("expected suggestions from standard net/http handler")
	}
	if res.Suggestions[0].Path != "/api/v1/reports" {
		t.Errorf("expected /api/v1/reports, got %s", res.Suggestions[0].Path)
	}
}

func TestGinNoRouteHandler(t *testing.T) {
	engine := gin.New()
	opts := Options{}
	engine.NoRoute(GinNoRouteHandler(engine, opts))

	engine.GET("/api/v1/books", func(c *gin.Context) {})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/bok", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	var res ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(res.Suggestions) == 0 || res.Suggestions[0].Path != "/api/v1/books" {
		t.Errorf("expected /api/v1/books suggestion from GinNoRouteHandler")
	}
}

func TestIsEnabled_Variations(t *testing.T) {
	// 1. Explicit Enabled = true
	tTrue := true
	if !IsEnabled(Options{Enabled: &tTrue}) {
		t.Errorf("expected true when Enabled is explicitly true")
	}

	// 2. Explicit Enabled = false
	tFalse := false
	if IsEnabled(Options{Enabled: &tFalse}) {
		t.Errorf("expected false when Enabled is explicitly false")
	}

	// 3. EnabledEnvironments matching
	os.Setenv("ENV", "staging")
	defer os.Unsetenv("ENV")
	if !IsEnabled(Options{EnabledEnvironments: []string{"development", "staging"}}) {
		t.Errorf("expected true for staging in EnabledEnvironments")
	}
	if IsEnabled(Options{EnabledEnvironments: []string{"development"}}) {
		t.Errorf("expected false for staging when not in EnabledEnvironments")
	}
}

func TestCustomHTMLRenderer(t *testing.T) {
	calledCustom := false
	opts := Options{
		CustomHTMLRenderer: func(w http.ResponseWriter, r *http.Request, suggestions []RouteSuggestion) error {
			calledCustom = true
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("<h1>Custom 404</h1>"))
			return nil
		},
	}

	routes := []types.RouteEntry{{Path: "/dashboard", Methods: []types.HttpMethod{MethodGet}}}
	handler := HTTPNotFoundHandler(routes, opts)

	req := httptest.NewRequest(http.MethodGet, "/dashbord", nil)
	req.Header.Set("Accept", "text/html")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !calledCustom {
		t.Errorf("expected custom HTML renderer to be invoked")
	}
	if rec.Body.String() != "<h1>Custom 404</h1>" {
		t.Errorf("expected custom body, got %s", rec.Body.String())
	}
}
