package modules

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PassSearch searches the pass(1) password store and copies entries.
func PassSearch(query string) []Result {
	lower := strings.ToLower(strings.TrimSpace(query))
	if lower != "pass" && !strings.HasPrefix(lower, "pass ") {
		return nil
	}
	if _, err := exec.LookPath("pass"); err != nil {
		return nil
	}
	filter := strings.TrimSpace(query[len("pass"):])

	var out []Result
	for _, entry := range passEntries() {
		if filter != "" && !strings.Contains(strings.ToLower(entry), strings.ToLower(filter)) {
			continue
		}
		entry := entry
		out = append(out, Result{
			Type:  "pass",
			Title: "Pass: " + entry,
			Desc:  "Copy password to clipboard",
			Icon:  "dialog-password",
			Action: func() {
				if err := exec.Command("pass", "-c", entry).Run(); err != nil {
					SetStatus(false, "Pass failed: "+err.Error())
				} else {
					SetStatus(true, "Copied password: "+entry)
				}
			},
		})
		if len(out) >= 8 {
			break
		}
	}
	return out
}

func passEntries() []string {
	store := os.Getenv("PASSWORD_STORE_DIR")
	if store == "" {
		store = filepath.Join(os.Getenv("HOME"), ".password-store")
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
