package matcher

import (
	"math"
	"sort"
	"strings"

	"github.com/eslam-amin/didyoumean/types"
)

// RankRoutes evaluates candidate routes against a requested HTTP method and URL path,
// returning the highest-scoring suggestions sorted by relevance and method mismatch priority.
func RankRoutes(requestMethod, requestPath string, routes []types.RouteEntry, opts types.Options) []types.RouteSuggestion {
	opts.ApplyDefaults()

	normReqPath := NormalizePath(requestPath)
	reqSegments := SplitSegments(normReqPath)
	reqMethod := types.HttpMethod(strings.ToUpper(strings.TrimSpace(requestMethod)))

	var allSuggestions []types.RouteSuggestion
	seen := make(map[string]int) // key -> index in allSuggestions

	for _, route := range routes {
		// Filter out excluded routes (e.g. /health, /metrics)
		if opts.IsRouteExcluded(route.Path) {
			continue
		}

		normRoutePath := NormalizePath(route.Path)
		// Skip root catch-all routes (e.g. /* or *) which represent fallback 404 handlers
		if normRoutePath == "/*" || normRoutePath == "*" || route.Path == "/*" || route.Path == "*" {
			continue
		}
		routeSegments := SplitSegments(normRoutePath)

		hasMethod := containsMethod(route.Methods, reqMethod)
		isMatch, isExact := IsStructuralMatch(reqSegments, routeSegments)

		if isMatch {
			if !hasMethod {
				// Method mismatch on exact/structural path!
				if opts.ShouldIncludeMethodMismatches() {
					for _, m := range route.Methods {
						sug := types.RouteSuggestion{
							Method:           m,
							Path:             route.Path,
							Score:            1.0,
							IsMethodMismatch: true,
							Summary:          route.Summary,
							Description:      route.Description,
						}
						addOrUpdateSuggestion(&allSuggestions, seen, sug)
					}
				}
			} else {
				// Matched method and path
				score := 1.0
				if !isExact {
					score = 0.99
				}
				sug := types.RouteSuggestion{
					Method:           reqMethod,
					Path:             route.Path,
					Score:            score,
					IsMethodMismatch: false,
					Summary:          route.Summary,
					Description:      route.Description,
				}
				addOrUpdateSuggestion(&allSuggestions, seen, sug)
			}
			continue
		}

		// Fuzzy matching
		baseScore := HybridScore(normReqPath, normRoutePath, reqSegments, routeSegments, false)

		if hasMethod {
			if baseScore >= opts.Threshold {
				sug := types.RouteSuggestion{
					Method:           reqMethod,
					Path:             route.Path,
					Score:            baseScore,
					IsMethodMismatch: false,
					Summary:          route.Summary,
					Description:      route.Description,
				}
				addOrUpdateSuggestion(&allSuggestions, seen, sug)
			}
		} else if opts.ShouldIncludeMethodMismatches() {
			// Path is fuzzy AND method is different; apply slight penalty
			adjustedScore := math.Round(baseScore*0.9*10000) / 10000
			if adjustedScore >= opts.Threshold {
				for _, m := range route.Methods {
					sug := types.RouteSuggestion{
						Method:           m,
						Path:             route.Path,
						Score:            adjustedScore,
						IsMethodMismatch: false,
						Summary:          route.Summary,
						Description:      route.Description,
					}
					addOrUpdateSuggestion(&allSuggestions, seen, sug)
				}
			}
		}
	}

	// Sort suggestions
	sort.SliceStable(allSuggestions, func(i, j int) bool {
		a, b := allSuggestions[i], allSuggestions[j]

		// 1. Method mismatch takes priority
		if a.IsMethodMismatch != b.IsMethodMismatch {
			return a.IsMethodMismatch
		}

		// 2. Higher score comes first
		if math.Abs(a.Score-b.Score) > 0.0001 {
			return a.Score > b.Score
		}

		// 3. Prefer path with closer length to request path
		diffA := int(math.Abs(float64(len(a.Path) - len(normReqPath))))
		diffB := int(math.Abs(float64(len(b.Path) - len(normReqPath))))
		if diffA != diffB {
			return diffA < diffB
		}

		// 4. Deterministic alphabetical sorting
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		return a.Method < b.Method
	})

	// Cut to maxSuggestions
	if len(allSuggestions) > opts.MaxSuggestions {
		allSuggestions = allSuggestions[:opts.MaxSuggestions]
	}

	return allSuggestions
}