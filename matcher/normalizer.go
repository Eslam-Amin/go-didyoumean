package matcher

import (
	"path"
	"regexp"
	"strings"
)

var (
	// regexMultiSlash matches consecutive slashes
	regexMultiSlash = regexp.MustCompile(`/+`)
	// regexCurlyParam matches OpenAPI/Chi curly-brace parameters like {id}
	regexCurlyParam = regexp.MustCompile(`\{([^}]+)\}`)
	// regexColonParam matches Gin/Echo colon parameters like :id
	regexColonParam = regexp.MustCompile(`:([a-zA-Z0-9_]+)`)
)

// NormalizePath normalizes a URL path by:
// 1. Stripping query strings (?) and hash fragments (#)
// 2. Collapsing consecutive duplicate slashes (e.g. //api///users -> /api/users)
// 3. Trimming trailing slashes (while preserving single root "/")
// 4. Resolving path directory dots via path.Clean
func NormalizePath(rawPath string) string {
	if rawPath == "" {
		return "/"
	}

	// 1. Strip query string
	if idx := strings.IndexByte(rawPath, '?'); idx != -1 {
		rawPath = rawPath[:idx]
	}
	// Strip fragment
	if idx := strings.IndexByte(rawPath, '#'); idx != -1 {
		rawPath = rawPath[:idx]
	}

	rawPath = strings.TrimSpace(rawPath)
	if rawPath == "" {
		return "/"
	}

	// 2. Collapse slashes
	rawPath = regexMultiSlash.ReplaceAllString(rawPath, "/")

	// 3. Clean path (handles . and ..)
	clean := path.Clean(rawPath)
	if !strings.HasPrefix(clean, "/") {
		clean = "/" + clean
	}

	// 4. Trim trailing slash except root
	if len(clean) > 1 && strings.HasSuffix(clean, "/") {
		clean = strings.TrimSuffix(clean, "/")
	}

	return clean
}