package modules

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	fileCache     []Result
	fileCacheTerm string
	fileCacheMu   sync.Mutex
	fileCacheTime time.Time
)

// FileSearch finds files matching query using fd or find.
// Requires "f " prefix to avoid slow search on every keystroke.
func FileSearch(query string) []Result {
	return FileSearchContext(context.Background(), query)
}

func FileSearchContext(ctx context.Context, query string) []Result {
	term, ok := fileSearchTerm(query)
	if !ok {
		return nil
	}

	if len(term) < 3 {
		return FileLoading(query)
	}

	// Check cache (valid for 5 seconds)
	fileCacheMu.Lock()
	if term == fileCacheTerm && time.Since(fileCacheTime) < 5*time.Second {
		result := fileCache
		fileCacheMu.Unlock()
		return result
	}
	fileCacheMu.Unlock()

	// Run search
	results := doFileSearch(ctx, term)

	// Update cache
	fileCacheMu.Lock()
	fileCache = results
	fileCacheTerm = term
	fileCacheTime = time.Now()
	fileCacheMu.Unlock()

	return results
}

func IsFileQueryReady(query string) bool {
	term, ok := fileSearchTerm(query)
	return ok && len(term) >= 3
}

func IsFileQuery(query string) bool {
	_, ok := fileSearchTerm(query)
	return ok
}

func FileLoading(query string) []Result {
	term, ok := fileSearchTerm(query)
	if !ok {
		return nil
	}
	if len(term) < 3 {
		return []Result{{
			Type:   "file",
			Title:  "Find Files",
			Desc:   "Type f <name>, for example f pdf",
			Icon:   "system-search",
			Action: func() {},
		}}
	}
	return []Result{{
		Type:   "file",
		Title:  "Searching files...",
		Desc:   term,
		Icon:   "system-search",
		Action: func() {},
	}}
}

func fileSearchTerm(query string) (string, bool) {
	q := strings.TrimSpace(query)
	lower := strings.ToLower(q)
	switch {
	case lower == "f" || lower == "file":
		return "", true
	case strings.HasPrefix(lower, "f "):
		return strings.TrimSpace(q[2:]), true
	case strings.HasPrefix(lower, "file "):
		return strings.TrimSpace(q[len("file "):]), true
	default:
		return "", false
	}
}

func doFileSearch(ctx context.Context, term string) []Result {
	ctx, cancel := context.WithTimeout(ctx, 900*time.Millisecond)
	defer cancel()

	var cmd *exec.Cmd
	if _, err := exec.LookPath("fd"); err == nil {
		cmd = exec.CommandContext(ctx, "fd", "--max-results", "50", "--type", "f", term, os.Getenv("HOME"))
	} else {
		cmd = exec.CommandContext(ctx, "find", os.Getenv("HOME"), "-maxdepth", "4", "-type", "f", "-iname", "*"+term+"*")
	}

	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var results []Result

	for _, path := range lines {
		if path == "" {
			continue
		}
		p := path
		name := filepath.Base(p)
		dir := filepath.Dir(p)
		icon := getFileIcon(name)

		results = append(results, Result{
			Type:  "file",
			Title: name,
			Desc:  shortenPath(dir),
			Icon:  icon,
			Action: func() {
				exec.Command("xdg-open", p).Start()
			},
		})

		if len(results) >= 50 {
			break
		}
	}

	return results
}

func getFileIcon(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".pdf":
		return "application-pdf"
	case ".doc", ".docx", ".odt":
		return "x-office-document"
	case ".xls", ".xlsx", ".ods":
		return "x-office-spreadsheet"
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
		return "image-x-generic"
	case ".mp3", ".wav", ".flac", ".ogg":
		return "audio-x-generic"
	case ".mp4", ".mkv", ".avi", ".webm":
		return "video-x-generic"
	case ".go", ".py", ".js", ".ts", ".rs", ".c", ".cpp":
		return "text-x-script"
	default:
		return "text-x-generic"
	}
}

func shortenPath(path string) string {
	home := os.Getenv("HOME")
	if strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}
