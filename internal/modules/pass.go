package modules

import (
	"github.com/tapiaw38/spark/internal/platform/commands"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/tapiaw38/spark/internal/config"
)

func PassSearch(query string) []Result {
	lower := strings.ToLower(strings.TrimSpace(query))
	if lower != "pass" && !strings.HasPrefix(lower, "pass ") {
		return nil
	}
	if _, err := commands.LookPath("pass"); err != nil {
		return nil
	}
	filter := strings.TrimSpace(query[len("pass"):])

	var out []Result
	for _, entry := range passEntries() {
		if filter != "" && !strings.Contains(strings.ToLower(entry), strings.ToLower(filter)) {
			continue
		}
		out = append(out, Result{
			Type:       TypePass,
			Title:      "Pass: " + entry,
			Desc:       "Copy password to clipboard",
			Icon:       "dialog-password",
			ActionSpec: RunAction("pass", "-c", entry).WithStatus("Copied password: " + entry),
		})
		if len(out) >= MaxCompactResults {
			break
		}
	}
	return out
}

func passEntries() []string {
	store := os.Getenv("PASSWORD_STORE_DIR")
	if store == "" {
		store = config.HomeFile(".password-store")
	}
	var entries []string
	filepath.WalkDir(store, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".gpg") {
			return nil
		}
		rel, err := filepath.Rel(store, path)
		if err != nil {
			return nil
		}
		entries = append(entries, strings.TrimSuffix(rel, ".gpg"))
		return nil
	})
	return entries
}
