package discovery

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/eslam-amin/didyoumean/types"
)

const sampleSwagger2YAML = `
swagger: "2.0"
info:
  title: Sample API
  version: "1.0.0"
basePath: /api/v1
paths:
  /users:
    get:
      summary: List all users
      description: Returns a paginated list of users
      operationId: listUsers
    post:
      summary: Create a user
  /users/{id}:
    get:
      summary: Get user by ID
`

const sampleOpenAPI3JSON = `{
  "openapi": "3.0.1",
  "info": {
    "title": "OpenAPI 3 Sample",
    "version": "1.0.0"
  },
  "servers": [
    { "url": "https://api.example.com/api/v2" }
  ],
  "paths": {
    "/products": {
      "get": {
        "summary": "List products"
      }
    },
    "/products/{id}": {
      "delete": {
        "summary": "Delete product"
      }
    }
  }
}`

func TestParseOpenAPISpec_Swagger2YAML(t *testing.T) {
	routes, err := ParseOpenAPISpec(sampleSwagger2YAML)
	if err != nil {
		t.Fatalf("failed to parse Swagger 2.0 YAML: %v", err)
	}

	if len(routes) != 3 {
		t.Fatalf("expected 3 route entries, got %d", len(routes))
	}

	foundListUsers := false
	for _, r := range routes {
		if r.Path == "/api/v1/users" && len(r.Methods) > 0 && r.Methods[0] == types.MethodGet {
			foundListUsers = true
			if r.Summary != "List all users" {
				t.Errorf("expected summary 'List all users', got %q", r.Summary)
			}
			if r.Description != "Returns a paginated list of users" {
				t.Errorf("expected description 'Returns a paginated list of users', got %q", r.Description)
			}
		}
	}

	if !foundListUsers {
		t.Errorf("expected to find GET /api/v1/users")
	}
}

func TestParseOpenAPISpec_OpenAPI3JSON(t *testing.T) {
	routes, err := ParseOpenAPISpec([]byte(sampleOpenAPI3JSON))
	if err != nil {
		t.Fatalf("failed to parse OpenAPI 3.0 JSON: %v", err)
	}

	if len(routes) != 2 {
		t.Fatalf("expected 2 route entries, got %d", len(routes))
	}

	foundGet := false
	for _, r := range routes {
		if r.Path == "/api/v2/products" && len(r.Methods) > 0 && r.Methods[0] == types.MethodGet {
			foundGet = true
			if r.Summary != "List products" {
				t.Errorf("expected summary 'List products', got %q", r.Summary)
			}
		}
	}
	if !foundGet {
		t.Errorf("expected to find GET /api/v2/products")
	}
}

func TestParseOpenAPISpec_FileAndURL(t *testing.T) {
	// Test file reading
	tmpDir := t.TempDir()
	specFile := filepath.Join(tmpDir, "spec.yaml")
	if err := os.WriteFile(specFile, []byte(sampleSwagger2YAML), 0644); err != nil {
		t.Fatalf("failed to write temp spec file: %v", err)
	}

	routesFromFile, err := ParseOpenAPISpec(specFile)
	if err != nil {
		t.Fatalf("failed to parse spec from file: %v", err)
	}
	if len(routesFromFile) != 3 {
		t.Fatalf("expected 3 routes from file, got %d", len(routesFromFile))
	}

	// Test HTTP remote URL fetching
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sampleOpenAPI3JSON))
	}))
	defer server.Close()

	routesFromURL, err := ParseOpenAPISpec(server.URL)
	if err != nil {
		t.Fatalf("failed to parse spec from remote URL: %v", err)
	}
	if len(routesFromURL) != 2 {
		t.Fatalf("expected 2 routes from remote URL, got %d", len(routesFromURL))
	}
}
