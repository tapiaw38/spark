package modules

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func IsNavQuery(query string) bool {
	lower := strings.ToLower(strings.TrimSpace(query))
	return lower == "nav" || lower == "browse" || strings.HasPrefix(lower, "nav ") || strings.HasPrefix(lower, "browse ")
}

func NavigationLoading(query string) []Result {
	return []Result{{
		Type:   TypeNavigation,
		Title:  "Loading...",
		Icon:   "folder",
		Action: func() {},
	}}
}

func NavigationSearch(query string) []Result {
	q := strings.TrimSpace(query)
	lower := strings.ToLower(q)
	if !IsNavQuery(query) {
		return nil
	}

	path := os.Getenv("HOME")
	if strings.HasPrefix(lower, "nav ") {
		path = strings.TrimSpace(q[len("nav "):])
	} else if strings.HasPrefix(lower, "browse ") {
		path = strings.TrimSpace(q[len("browse "):])
	}
	path = expandHome(path)
	if path == "" {
		path = os.Getenv("HOME")
	}

	path, filter := splitPathFilter(path)
	return directoryResults("nav", path, filter, "", "")
}

func IsPickQuery(query string) bool {
	lower := strings.ToLower(strings.TrimSpace(query))
	return strings.HasPrefix(lower, "pick copy ") || strings.HasPrefix(lower, "pick move ")
}

func DestinationPickerSearch(query string) []Result {
	q := strings.TrimSpace(query)
	lower := strings.ToLower(q)
	if !IsPickQuery(query) {
		return nil
	}

	op := "copy"
	body := strings.TrimSpace(q[len("pick copy "):])
	if strings.HasPrefix(lower, "pick move ") {
		op = "move"
		body = strings.TrimSpace(q[len("pick move "):])
	}
	parts := strings.SplitN(body, "|", 2)
	if len(parts) < 2 {
		return nil
	}
	source := strings.TrimSpace(parts[0])
	path := expandHome(strings.TrimSpace(parts[1]))
	if path == "" {
		path = os.Getenv("HOME")
	}
	path, filter := splitPathFilter(path)
	return directoryResults("pick "+op, path, filter, op, source)
}

func splitPathFilter(input string) (string, string) {
	input = expandHome(strings.TrimSpace(input))
	if input == "" {
		return os.Getenv("HOME"), ""
	}
	if info, err := os.Stat(input); err == nil && info.IsDir() {
		return input, ""
	}
	path := input
	filter := ""
	for {
		dir := filepath.Dir(path)
		base := filepath.Base(path)
		if dir == path || base == "." {
			return input, ""
		}
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			filter = base
			return dir, filter
		}
		path = dir
	}
}

func directoryResults(prefix, path, filter, op, source string) []Result {
	info, err := os.Stat(path)
	if err != nil {
		return []Result{{
			Type:   TypeNavigation,
			Title:  "Folder Not Found",
			Desc:   path,
			Icon:   "dialog-error",
			Action: func() {},
		}}
	}
	if !info.IsDir() {
		dir := filepath.Dir(path)
		return []Result{{
			Type:  TypeFile,
			Title: filepath.Base(path),
			Desc:  shortenPath(dir),
			Icon:  getFileIcon(path),
			Action: func() {
				Open(path)
			},
		}}
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
	})

	results := []Result{}
	if op == "copy" || op == "move" {
		dst := path
		results = append(results, Result{
			Type:    TypeFileOp,
			Title:   operationTitle(op) + " Here",
			Desc:    shortenPath(source) + " -> " + shortenPath(dst),
			Icon:    operationIcon(op),
			Confirm: true,
			Action: func() {
				RunFileOperation(op, source, dst)
			},
		})
	}

	results = append(results, Result{
		Type:          TypeDirectory,
		Title:         "..",
		Desc:          shortenPath(filepath.Dir(path)),
		Icon:          "go-up",
		NavigateQuery: navQuery(prefix, source, filepath.Dir(path)),
		KeepOpen:      true,
	})
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if filter != "" && !strings.Contains(strings.ToLower(entry.Name()), strings.ToLower(filter)) {
			continue
		}
		p := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			results = append(results, Result{
				Type:          TypeDirectory,
				Title:         entry.Name(),
				Desc:          shortenPath(p),
				Icon:          "folder",
				NavigateQuery: navQuery(prefix, source, p),
				KeepOpen:      true,
			})
		} else {
			filePath := p
			results = append(results, Result{
				Type:  TypeFile,
				Title: entry.Name(),
				Desc:  shortenPath(path),
				Icon:  getFileIcon(entry.Name()),
				Action: func() {
					Open(filePath)
				},
			})
		}
	}
	return results
}

func navQuery(prefix, source, path string) string {
	if strings.HasPrefix(prefix, "pick ") {
		return prefix + " " + source + " | " + path
	}
	return prefix + " " + path
}
