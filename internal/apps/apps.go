package apps

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tapiaw38/spark/internal/history"
)

type App struct {
	Name string
	Exec string
	Icon string
}

// Load reads all .desktop files from standard locations
func Load() []App {
	var apps []App
	dirs := applicationDirs()

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), ".desktop") {
				continue
			}
			if app, ok := parseDesktop(filepath.Join(dir, e.Name())); ok {
				apps = append(apps, app)
			}
		}
	}
	return apps
}

func applicationDirs() []string {
	seen := make(map[string]bool)
	var dirs []string

	add := func(dir string) {
		if dir == "" || seen[dir] {
			return
		}
		seen[dir] = true
		dirs = append(dirs, dir)
	}

	add(filepath.Join(os.Getenv("HOME"), ".local/share/applications"))
	for _, dataDir := range strings.Split(os.Getenv("XDG_DATA_DIRS"), ":") {
		if dataDir == "" {
			continue
		}
		add(filepath.Join(dataDir, "applications"))
	}
	add("/usr/local/share/applications")
	add("/usr/share/applications")

	return dirs
}

func parseDesktop(path string) (App, bool) {
	f, err := os.Open(path)
	if err != nil {
		return App{}, false
	}
	defer f.Close()

	var app App
	inDesktopEntry := false
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "[Desktop Entry]" {
			inDesktopEntry = true
			continue
		}
		if strings.HasPrefix(line, "[") {
			inDesktopEntry = false
			continue
		}
		if !inDesktopEntry {
			continue
		}

		if strings.HasPrefix(line, "Name=") && app.Name == "" {
			app.Name = strings.TrimPrefix(line, "Name=")
		} else if strings.HasPrefix(line, "Exec=") {
			app.Exec = cleanDesktopExec(strings.TrimPrefix(line, "Exec="))
		} else if strings.HasPrefix(line, "Icon=") {
			app.Icon = strings.TrimPrefix(line, "Icon=")
		} else if strings.HasPrefix(line, "NoDisplay=true") {
			return App{}, false
		}
	}

	if app.Name == "" || app.Exec == "" {
		return App{}, false
	}
	return app, true
}

func cleanDesktopExec(execCmd string) string {
	fields := strings.Fields(execCmd)
	clean := fields[:0]
	for _, field := range fields {
		if strings.HasPrefix(field, "%") || strings.HasPrefix(field, "@@") {
			continue
		}
		clean = append(clean, field)
	}
	return strings.Join(clean, " ")
}

// QuickSearch does fast prefix match for short queries (1-2 chars)
// ponytail: skip fuzzy scoring, just prefix match + history sort
func QuickSearch(apps []App, query string) []App {
	if query == "" {
		return nil
	}
	query = strings.ToLower(query)

	type scored struct {
		app   App
		score int
	}
	var results []scored

	for _, app := range apps {
		name := strings.ToLower(app.Name)
		// Simple prefix match
		if strings.HasPrefix(name, query) {
			score := 100 + history.Score(app.Name)*3
			results = append(results, scored{app, score})
		} else if strings.Contains(name, " "+query) {
			// Word start match
			score := 50 + history.Score(app.Name)*3
			results = append(results, scored{app, score})
		}
	}

	// Sort by score descending (simple bubble, small N)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].score > results[i].score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Limit to 6
	if len(results) > 6 {
		results = results[:6]
	}

	out := make([]App, len(results))
	for i, r := range results {
		out[i] = r.app
	}
	return out
}

// Search filters apps by fuzzy match, sorted by score
func Search(apps []App, query string) []App {
	if query == "" {
		return apps
	}
	query = strings.ToLower(query)

	type scored struct {
		app   App
		score int
	}
	var results []scored

	for _, app := range apps {
		if score := fuzzyScore(strings.ToLower(app.Name), query); score > 0 {
			// ponytail: history boost - each launch adds weight
			score += history.Score(app.Name) * 3
			results = append(results, scored{app, score})
		}
	}

	// Sort by score descending
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].score > results[i].score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	out := make([]App, len(results))
	for i, r := range results {
		out[i] = r.app
	}
	return out
}

// fuzzyScore returns 0 if no match, higher score = better match
// ponytail: simple subsequence match with bonuses for consecutive/start matches
func fuzzyScore(name, query string) int {
	if len(query) == 0 {
		return 1
	}
	if len(name) == 0 {
		return 0
	}

	score := 0
	qi := 0
	consecutive := 0
	prevMatch := -2

	for ni := 0; ni < len(name) && qi < len(query); ni++ {
		if name[ni] == query[qi] {
			score += 1
			// Bonus: start of string
			if ni == 0 {
				score += 5
			}
			// Bonus: after separator (space, dash, underscore)
			if ni > 0 && (name[ni-1] == ' ' || name[ni-1] == '-' || name[ni-1] == '_') {
				score += 3
			}
			// Bonus: consecutive
			if ni == prevMatch+1 {
				consecutive++
				score += consecutive * 2
			} else {
				consecutive = 0
			}
			prevMatch = ni
			qi++
		}
	}

	// All query chars matched?
	if qi < len(query) {
		return 0
	}
	return score
}

// Launch starts the app
func Launch(app App) error {
	parts := strings.Fields(app.Exec)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}
