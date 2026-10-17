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