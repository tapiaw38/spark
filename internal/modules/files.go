package modules

import (
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

// FileSearch finds files matching query using fd or find
// ponytail: requires "f:" prefix to avoid slow search on every keystroke
func FileSearch(query string) []Result {
	// Require explicit prefix to trigger file search
	if !strings.HasPrefix(strings.ToLower(query), "f:") &&
		!strings.HasPrefix(strings.ToLower(query), "file:") {
		return nil
	}

	// Extract search term
	term := query
	if strings.HasPrefix(strings.ToLower(query), "file:") {
		term = strings.TrimPrefix(query, "file:")
		term = strings.TrimPrefix(term, "File:")
	} else {
		term = strings.TrimPrefix(query, "f:")
		term = strings.TrimPrefix(term, "F:")
	}
	term = strings.TrimSpace(term)

	if len(term) < 2 {
		return nil
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
	results := doFileSearch(term)

	// Update cache
	fileCacheMu.Lock()
	fileCache = results
	fileCacheTerm = term
	fileCacheTime = time.Now()
	fileCacheMu.Unlock()

	return results
}

func doFileSearch(term string) []Result {
	var cmd *exec.Cmd
	if _, err := exec.LookPath("fd"); err == nil {
		cmd = exec.Command("fd", "--max-results", "50", "--type", "f", term, os.Getenv("HOME"))
	} else {
		cmd = exec.Command("find", os.Getenv("HOME"), "-maxdepth", "4", "-type", "f", "-iname", "*"+term+"*")
	}

	// Timeout after 500ms
	done := make(chan []byte, 1)
	go func() {
		out, _ := cmd.Output()
		done <- out
	}()

	var output []byte
	select {
	case output = <-done:
	case <-time.After(500 * time.Millisecond):
		cmd.Process.Kill()
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
