package modules

import (
	"slices"
	"strings"
)

func MatchesAny(query string, names ...string) bool {
	return slices.Contains(names, normalizeQuery(query))
}

func MatchCommand(query string, names ...string) (string, bool) {
	trimmed := strings.TrimSpace(query)
	q := strings.ToLower(trimmed)
	for _, name := range names {
		if q == name {
			return "", true
		}
		if strings.HasPrefix(q, name+" ") {
			return strings.TrimSpace(trimmed[len(name):]), true
		}
	}
	return "", false
}

func normalizeQuery(query string) string {
	return strings.ToLower(strings.TrimSpace(query))
}
