package modules

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// BookmarksSearch searches Chromium-family browser bookmarks.
func BookmarksSearch(query string) []Result {
	lower := strings.ToLower(strings.TrimSpace(query))
	if lower != "bm" && !strings.HasPrefix(lower, "bm ") {
		return nil
	}
	filter := strings.ToLower(strings.TrimSpace(query[len("bm"):]))

	var out []Result
	for _, bm := range chromiumBookmarks() {
		if filter != "" && !strings.Contains(strings.ToLower(bm.name), filter) && !strings.Contains(strings.ToLower(bm.url), filter) {
			continue
		}
		bm := bm
		title := bm.name
		if title == "" {
			title = bm.url
		}
		out = append(out, Result{
			Type:   "bookmark",
			Title:  title,
			Desc:   bm.url,
			Icon:   "user-bookmarks",
			Action: func() { exec.Command("xdg-open", bm.url).Start() },
		})
		if len(out) >= 8 {
			break
		}
	}
	return out
}

type bookmark struct{ name, url string }

func chromiumBookmarks() []bookmark {
	home := os.Getenv("HOME")
	profiles := []string{
		filepath.Join(home, ".config", "google-chrome", "Default", "Bookmarks"),
		filepath.Join(home, ".config", "chromium", "Default", "Bookmarks"),
		filepath.Join(home, ".config", "BraveSoftware", "Brave-Browser", "Default", "Bookmarks"),
	}
	for _, path := range profiles {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var root struct {
			Roots map[string]json.RawMessage `json:"roots"`
		}
		if json.Unmarshal(data, &root) != nil {
			continue
		}
		var all []bookmark
		for _, node := range root.Roots {
			all = append(all, walkBookmarks(node)...)
		}
		if len(all) > 0 {
			return all
		}
	}
	return nil
}

func walkBookmarks(raw json.RawMessage) []bookmark {
	var node struct {
		Type     string            `json:"type"`
		Name     string            `json:"name"`
		URL      string            `json:"url"`
		Children []json.RawMessage `json:"children"`
	}
	if json.Unmarshal(raw, &node) != nil {
		return nil
	}
	if node.Type == "url" {
		return []bookmark{{node.Name, node.URL}}
	}
	var out []bookmark
	for _, child := range node.Children {
		out = append(out, walkBookmarks(child)...)
	}
	return out
}
