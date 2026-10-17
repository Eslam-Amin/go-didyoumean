package discovery

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/eslam-amin/didyoumean/types"
	"gopkg.in/yaml.v3"
)

// OpenAPISpec represents the parsed high-level structure of OpenAPI 2.0 or 3.x documents.
type rawSpec struct {
	Swagger  string                         `yaml:"swagger"`
	OpenAPI  string                         `yaml:"openapi"`
	BasePath string                         `yaml:"basePath"`
	Servers  []rawServer                    `yaml:"servers"`
	Paths    map[string]map[string]rawOp    `yaml:"paths"`
}

type rawServer struct {
	URL string `yaml:"url"`
}

type rawOp struct {
	Summary     string `yaml:"summary"`
	Description string `yaml:"description"`
	OperationID string `yaml:"operationId"`
	Deprecated  bool   `yaml:"deprecated"`
}

var supportedHTTPMethods = map[string]types.HttpMethod{
	"get":     types.MethodGet,
	"post":    types.MethodPost,
	"put":     types.MethodPut,
	"delete":  types.MethodDelete,
	"patch":   types.MethodPatch,
	"head":    types.MethodHead,
	"options": types.MethodOptions,
	"trace":   types.MethodTrace,
}

// ParseOpenAPISpec loads, parses, and converts an OpenAPI 2.0 or 3.x specification
// from a file path, remote URL, byte slice, raw string, or io.Reader.
func ParseOpenAPISpec(specInput any) ([]types.RouteEntry, error) {
	if specInput == nil {
		return nil, nil
	}

	rawBytes, err := readSpecInput(specInput)
	if err != nil {
		return nil, fmt.Errorf("failed to read OpenAPI spec input: %w", err)
	}

	if len(bytes.TrimSpace(rawBytes)) == 0 {
		return nil, nil
	}

	var spec rawSpec
	if err := yaml.Unmarshal(rawBytes, &spec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OpenAPI spec (JSON/YAML): %w", err)
	}

	// Determine base path prefix
	basePrefix := ""
	if spec.BasePath != "" {
		basePrefix = spec.BasePath
	} else if len(spec.Servers) > 0 {
		serverURL := strings.TrimSpace(spec.Servers[0].URL)
		if strings.HasPrefix(serverURL, "http://") || strings.HasPrefix(serverURL, "https://") {
			if u, err := url.Parse(serverURL); err == nil && u.Path != "" && u.Path != "/" {
				basePrefix = u.Path
			}
		} else if strings.HasPrefix(serverURL, "/") && serverURL != "/" {
			basePrefix = serverURL
		}
	}

	basePrefix = strings.TrimSuffix(basePrefix, "/")

	var routes []types.RouteEntry

	for routePath, methodsMap := range spec.Paths {
		fullPath := routePath
		if basePrefix != "" {
			fullPath = path.Join(basePrefix, routePath)
		}
		if !strings.HasPrefix(fullPath, "/") {
			fullPath = "/" + fullPath
		}

		for methodKey, op := range methodsMap {
			httpMethod, isMethod := supportedHTTPMethods[strings.ToLower(methodKey)]
			if !isMethod {
				continue
			}

			routes = append(routes, types.RouteEntry{
				Path:        fullPath,
				Methods:     []types.HttpMethod{httpMethod},
				Summary:     strings.TrimSpace(op.Summary),
				Description: strings.TrimSpace(op.Description),
				OperationID: strings.TrimSpace(op.OperationID),
				Source:      "openapi",
			})
		}
	}

	return routes, nil
}