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

const quickSearchResultLimit = 6

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

type scoredApp struct {
	app   App
	score int
}

func sortByScoreDescending(results []scoredApp) {
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].score > results[i].score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

func toApps(results []scoredApp) []App {
	out := make([]App, len(results))
	for i, r := range results {
		out[i] = r.app
	}
	return out
}

func QuickSearch(apps []App, query string) []App {
	if query == "" {
		return nil
	}
	query = strings.ToLower(query)

	var results []scoredApp
	for _, app := range apps {
		name := strings.ToLower(app.Name)
		switch {
		case strings.HasPrefix(name, query):
			results = append(results, scoredApp{app, 100 + history.Score(app.Name)*3})
		case strings.Contains(name, " "+query):
			results = append(results, scoredApp{app, 50 + history.Score(app.Name)*3})
		}
	}

	sortByScoreDescending(results)
	if len(results) > quickSearchResultLimit {
		results = results[:quickSearchResultLimit]
	}
	return toApps(results)
}

func Search(apps []App, query string) []App {
	if query == "" {
		return apps
	}
	query = strings.ToLower(query)

	var results []scoredApp
	for _, app := range apps {
		if score := fuzzyScore(strings.ToLower(app.Name), query); score > 0 {
			score += history.Score(app.Name) * 3
			results = append(results, scoredApp{app, score})
		}
	}

	sortByScoreDescending(results)
	return toApps(results)
}

const (
	fuzzyScoreCharMatch       = 1
	fuzzyScoreStartOfString   = 5
	fuzzyScoreAfterSeparator  = 3
	fuzzyScoreConsecutiveStep = 2
)

func isWordSeparator(b byte) bool {
	return b == ' ' || b == '-' || b == '_'
}

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
		if name[ni] != query[qi] {
			continue
		}
		score += fuzzyScoreCharMatch
		if ni == 0 {
			score += fuzzyScoreStartOfString
		}
		if ni > 0 && isWordSeparator(name[ni-1]) {
			score += fuzzyScoreAfterSeparator
		}
		if ni == prevMatch+1 {
			consecutive++
			score += consecutive * fuzzyScoreConsecutiveStep
		} else {
			consecutive = 0
		}
		prevMatch = ni
		qi++
	}

	if qi < len(query) {
		return 0
	}
	return score
}

func Launch(app App) error {
	parts := strings.Fields(app.Exec)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}
