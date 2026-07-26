package commands_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOnlyThisPackageImportsOSExec(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	allowed := filepath.Join("internal", "platform", "commands")

	var offenders []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "vendor" || d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		if strings.HasPrefix(rel, allowed) {
			return nil
		}
		src, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if strings.Contains(string(src), `"os/exec"`) {
			offenders = append(offenders, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}

	for _, file := range offenders {
		t.Errorf("%s imports os/exec directly; route subprocesses through internal/platform/commands", file)
	}
}
