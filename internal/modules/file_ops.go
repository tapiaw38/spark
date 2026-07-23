package modules

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var lastFileUndo *fileUndo

type fileUndo struct {
	Title  string `json:"title"`
	Kind   string `json:"kind"`
	Source string `json:"source"`
	Target string `json:"target"`
}

// FileOperationSearch runs file operations generated from file actions.
func FileOperationSearch(query string) []Result {
	if result := UndoSearch(query); result != nil {
		return result
	}

	op, source, target, ok := parseFileOperation(query)
	if !ok {
		return nil
	}
	if source == "" || target == "" {
		return []Result{{
			Type:   "file-op",
			Title:  operationTitle(op),
			Desc:   "Use: " + op + " source | target",
			Icon:   operationIcon(op),
			Action: func() {},
		}}
	}

	source = expandHome(source)
	target = expandHome(target)
	if op == "rename" && !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(source), target)
	}
	src := source
	dst := target

	return []Result{{
		Type:    "file-op",
		Title:   operationTitle(op),
		Desc:    shortenPath(src) + " -> " + shortenPath(dst),
		Icon:    operationIcon(op),
		Confirm: true,
		Action: func() {
			RunFileOperation(op, src, dst)
		},
	}}
}

func UndoSearch(query string) []Result {
	q := strings.ToLower(strings.TrimSpace(query))
	if q != "undo" {
		return nil
	}
	undo := currentFileUndo()
	if undo == nil {
		return []Result{{
			Type:   "undo",
			Title:  "Nothing to Undo",
			Desc:   "File operations set undo state",
			Icon:   "edit-undo",
			Action: func() {},
		}}
	}
	return []Result{{
		Type:    "undo",
		Title:   "Undo: " + undo.Title,
		Desc:    "Confirm undo",
		Icon:    "edit-undo",
		Confirm: true,
		Action: func() {
			runUndo(undo)
			clearUndo()
		},
	}}
}

func RunFileOperation(op, source, target string) {
	source = expandHome(source)
	target = expandHome(target)
	if op == "rename" && !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(source), target)
	}
	if info, err := os.Stat(target); err == nil && info.IsDir() && op != "rename" {
		target = filepath.Join(target, filepath.Base(source))
	}

	src := source
	dst := target
	switch op {
	case "rename", "move":
		if err := os.Rename(src, dst); err == nil {
			saveUndo(fileUndo{Title: operationTitle(op), Kind: "rename", Source: dst, Target: src})
			SetStatus(true, operationTitle(op)+": "+shortenPath(src)+" -> "+shortenPath(dst))
		} else {
			SetStatus(false, operationTitle(op)+" failed: "+err.Error())
		}
	case "copy":
		if err := copyPath(src, dst); err == nil {
			saveUndo(fileUndo{Title: "Copy File", Kind: "delete", Source: dst})
			SetStatus(true, "Copy File: "+shortenPath(src)+" -> "+shortenPath(dst))
		} else {
			SetStatus(false, "Copy File failed: "+err.Error())
		}
	}
}

func SetTrashUndo() {
	saveUndo(fileUndo{Title: "Move to Trash", Kind: "gio-trash-restore"})
}

func currentFileUndo() *fileUndo {
	if lastFileUndo != nil {
		return lastFileUndo
	}
	data, err := os.ReadFile(fileUndoPath())
	if err != nil {
		return nil
	}
	var undo fileUndo
	if json.Unmarshal(data, &undo) != nil || undo.Kind == "" {
		return nil
	}
	lastFileUndo = &undo
	return &undo
}

func saveUndo(undo fileUndo) {
	lastFileUndo = &undo
	os.MkdirAll(filepath.Dir(fileUndoPath()), 0755)
	data, _ := json.Marshal(undo)
	os.WriteFile(fileUndoPath(), data, 0644)
}

func clearUndo() {
	lastFileUndo = nil
	os.Remove(fileUndoPath())
}

func fileUndoPath() string {
	return filepath.Join(os.Getenv("HOME"), ".local", "share", "spark", "undo.json")
}

func runUndo(undo *fileUndo) {
	if undo == nil {
		return
	}
	switch undo.Kind {
	case "rename":
		if err := os.Rename(undo.Source, undo.Target); err != nil {
			SetStatus(false, "Undo failed: "+err.Error())
		} else {
			SetStatus(true, "Undo: "+undo.Title)
		}
	case "delete":
		if err := os.RemoveAll(undo.Source); err != nil {
			SetStatus(false, "Undo failed: "+err.Error())
		} else {
			SetStatus(true, "Undo: "+undo.Title)
		}
	case "gio-trash-restore":
		if _, err := exec.LookPath("gio"); err == nil {
			if err := exec.Command("gio", "trash", "--restore").Run(); err != nil {
				SetStatus(false, "Undo trash failed: "+err.Error())
			} else {
				SetStatus(true, "Undo: "+undo.Title)
			}
		}
	}
}

func parseFileOperation(query string) (string, string, string, bool) {
	q := strings.TrimSpace(query)
	lower := strings.ToLower(q)
	for _, op := range []string{"rename", "copy", "move"} {
		if lower == op {
			return op, "", "", true
		}
		prefix := op + " "
		if strings.HasPrefix(lower, prefix) {
			body := strings.TrimSpace(q[len(prefix):])
			parts := strings.SplitN(body, "|", 2)
			if len(parts) < 2 {
				return op, strings.TrimSpace(body), "", true
			}
			return op, strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), true
		}
	}
	return "", "", "", false
}

func operationTitle(op string) string {
	switch op {
	case "rename":
		return "Rename File"
	case "copy":
		return "Copy File"
	case "move":
		return "Move File"
	default:
		return "File Operation"
	}
}

func operationIcon(op string) string {
	switch op {
	case "rename":
		return "edit-rename"
	case "copy":
		return "edit-copy"
	case "move":
		return "go-jump"
	default:
		return "document"
	}
}

func copyPath(source, target string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyDir(source, target)
	}
	return copyFile(source, target)
}

func copyFile(source, target string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()

	if info, err := os.Stat(target); err == nil && info.IsDir() {
		target = filepath.Join(target, filepath.Base(source))
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	out, err := os.Create(target)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	info, _ := os.Stat(source)
	if info == nil {
		return nil
	}
	return os.Chmod(target, info.Mode())
}

func copyDir(source, target string) error {
	return filepath.WalkDir(source, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		dst := filepath.Join(target, rel)
		if d.IsDir() {
			return os.MkdirAll(dst, 0755)
		}
		return copyFile(path, dst)
	})
}
