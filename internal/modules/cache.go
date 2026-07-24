package modules

import (
	"os"
	"path/filepath"
)

// cacheSubdir returns ~/.cache/spark/<name>, creating it. Empty string on failure.
func cacheSubdir(name string) string {
	dir := filepath.Join(os.Getenv("HOME"), ".cache", "spark", name)
	if os.MkdirAll(dir, 0755) != nil {
		return ""
	}
	return dir
}
