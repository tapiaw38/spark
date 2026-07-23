package modules

import (
	"encoding/xml"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RecentSearch lists recently used local documents from GTK recent files.
func RecentSearch(query string) []Result {
	q := strings.TrimSpace(query)
	lower := strings.ToLower(q)
	if lower != "recent" && !strings.HasPrefix(lower, "recent ") {
		return nil
	}

	filter := ""
	appFilter := ""
	if strings.HasPrefix(lower, "recent ") {
		filter = strings.TrimSpace(q[len("recent "):])
	}
	if strings.HasPrefix(strings.ToLower(filter), "app ") {
		appFilter = strings.TrimSpace(filter[len("app "):])
		filter = ""
	}

	paths := recentFiles(filter, appFilter)
	if len(paths) == 0 {
		return []Result{{
			Type:   "recent",
			Title:  "No Recent Documents",
			Desc:   "~/.local/share/recently-used.xbel",
			Icon:   "document-open-recent",
			Action: func() {},
		}}
	}

	results := make([]Result, 0, len(paths))
	for _, path := range paths {
		p := path
		results = append(results, Result{
			Type:  "file",
			Title: filepath.Base(p),
			Desc:  shortenPath(filepath.Dir(p)),
			Icon:  getFileIcon(p),
			Action: func() {
				exec.Command("xdg-open", p).Start()
			},
		})
	}
	return results
}

func recentFiles(filter, appFilter string) []string {
	path := filepath.Join(os.Getenv("HOME"), ".local/share/recently-used.xbel")
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	decoder := xml.NewDecoder(f)
	var results []string
	seen := make(map[string]bool)
	lowerFilter := strings.ToLower(filter)

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}
		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "bookmark" {
			continue
		}
		for _, attr := range start.Attr {
			if attr.Name.Local != "href" || !strings.HasPrefix(attr.Value, "file://") {
				continue
			}
			u, err := url.Parse(attr.Value)
			if err != nil {
				continue
			}
			localPath := u.Path
			if localPath == "" || seen[localPath] {
				continue
			}
			if lowerFilter != "" && !strings.Contains(strings.ToLower(filepath.Base(localPath)), lowerFilter) && !strings.Contains(strings.ToLower(localPath), lowerFilter) {
				continue
			}
			if appFilter != "" && !recentBookmarkHasApp(decoder, strings.ToLower(appFilter)) {
				continue
			}
			if _, err := os.Stat(localPath); err != nil {
				continue
			}
			seen[localPath] = true
			results = append(results, localPath)
			if len(results) >= 50 {
				return results
			}
		}
	}
	return results
}

func recentBookmarkHasApp(decoder *xml.Decoder, appFilter string) bool {
	for {
		token, err := decoder.Token()
		if err != nil {
			return false
		}
		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "application" {
				for _, attr := range t.Attr {
					if attr.Name.Local == "name" && strings.Contains(strings.ToLower(attr.Value), appFilter) {
						return true
					}
				}
			}
		case xml.EndElement:
			if t.Name.Local == "bookmark" {
				return false
			}
		}
	}
}
