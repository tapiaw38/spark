package modules

import (
	"net/url"
	"strings"

	"github.com/tapiaw38/spark/internal/config"
)

func WebSearch(query string) []Result {
	parts := strings.SplitN(query, " ", 2)
	if len(parts) < 2 {
		return nil
	}

	prefix := strings.ToLower(parts[0])
	searchQuery := parts[1]

	shortcut, ok := config.Current.WebShortcuts[prefix]
	if !ok {
		return nil
	}

	searchURL := strings.Replace(shortcut.URL, "%s", url.QueryEscape(searchQuery), 1)

	return []Result{{
		Type:       TypeWeb,
		Title:      shortcut.Name + ": " + searchQuery,
		Desc:       "Search on " + shortcut.Name,
		Icon:       shortcut.Icon,
		ActionSpec: OpenAction(searchURL),
	}}
}

func FallbackWebSearch(query string) []Result {
	if query == "" {
		return nil
	}

	searchURL := "https://www.google.com/search?q=" + url.QueryEscape(query)

	return []Result{{
		Type:       TypeWeb,
		Title:      "Search Google: " + query,
		Desc:       "No results found, search on Google",
		Icon:       "web-browser",
		ActionSpec: OpenAction(searchURL),
	}}
}
