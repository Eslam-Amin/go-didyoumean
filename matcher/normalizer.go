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

// SplitSegments splits a normalized URL path into non-empty path segment strings.
// For example, "/api/v1/users" -> ["api", "v1", "users"]
func SplitSegments(p string) []string {
	normalized := NormalizePath(p)
	trimmed := strings.Trim(normalized, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

// IsParamSegment checks if a segment represents a route parameter or wildcard.
// Recognizes ":param" (Gin/Echo), "{param}" (OpenAPI/Chi), and "*param" (catch-all).
func IsParamSegment(seg string) bool {
	if seg == "" {
		return false
	}
	if strings.HasPrefix(seg, ":") || strings.HasPrefix(seg, "*") {
		return true
	}
	if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") && len(seg) > 2 {
		return true
	}
	return false
}

// IsCatchAllSegment checks if a segment is a wildcard catch-all (e.g. *filepath, *).
func IsCatchAllSegment(seg string) bool {
	return strings.HasPrefix(seg, "*")
}

// StandardizeWildcards converts wildcard representations ({param}, :param) into a standard format ":param".
func StandardizeWildcards(p string) string {
	normalized := NormalizePath(p)
	// Replace {id} with :id
	replaced := regexCurlyParam.ReplaceAllString(normalized, ":$1")
	return replaced
}