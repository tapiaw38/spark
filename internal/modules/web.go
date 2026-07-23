package modules

import (
	"net/url"
	"os/exec"
	"strings"

	"github.com/tapiaw38/spark/internal/config"
)

// WebSearch checks for web shortcuts like "g query"
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
		Type:  "web",
		Title: shortcut.Name + ": " + searchQuery,
		Desc:  "Search on " + shortcut.Name,
		Icon:  shortcut.Icon,
		Action: func() {
			cmd := exec.Command("xdg-open", searchURL)
			cmd.Start()
		},
	}}
}

// FallbackWebSearch provides web search when no results found
func FallbackWebSearch(query string) []Result {
	if query == "" {
		return nil
	}

	searchURL := "https://www.google.com/search?q=" + url.QueryEscape(query)

	return []Result{{
		Type:  "web",
		Title: "Search Google: " + query,
		Desc:  "No results found, search on Google",
		Icon:  "web-browser",
		Action: func() {
			cmd := exec.Command("xdg-open", searchURL)
			cmd.Start()
		},
	}}
}
