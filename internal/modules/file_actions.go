package modules

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/tapiaw38/spark/internal/config"
)

var (
	fileBuffer   []string
	bufferLoaded bool
	fileBufferMu sync.Mutex
)

func FileActions(path string) []Result {
	if path == "" {
		return nil
	}

	return []Result{
		{
			Type:  TypeFileAction,
			Title: "Open",
			Desc:  path,
			Icon:  "document-open",
			Action: func() {
				Open(path)
			},
		},
		{
			Type:  TypeFileAction,
			Title: "Reveal in Files",
			Desc:  filepath.Dir(path),
			Icon:  "folder-open",
			Action: func() {
				revealFile(path)
			},
		},
		{
			Type:  TypeFileAction,
			Title: "Copy Path",
			Desc:  path,
			Icon:  "edit-copy",
			Action: func() {
				copyText(path)
			},
		},
		{
			Type:  TypeFileAction,
			Title: "Rename...",
			Desc:  "Edit name in file operation window",
			Icon:  "edit-rename",
			Action: func() {
				openFileOpWindow("rename", path, filepath.Base(path))
			},
		},
		{
			Type:  TypeFileAction,
			Title: "Copy To...",
			Desc:  "Choose destination in file operation window",
			Icon:  "edit-copy",
			Action: func() {
				openFileOpWindow("copy", path, filepath.Dir(path))
			},
		},
		{
			Type:  TypeFileAction,
			Title: "Move To...",
			Desc:  "Choose destination in file operation window",
			Icon:  "go-jump",
			Action: func() {
				openFileOpWindow("move", path, filepath.Dir(path))
			},
		},
		{
			Type:     TypeFileAction,
			Title:    "Add to Buffer",
			Desc:     bufferSummary(1),
			Icon:     "list-add",
			KeepOpen: true,
			Action: func() {
				AddFileToBuffer(path)
			},
		},
		{
			Type:  TypeFileAction,
			Title: "Email File",
			Desc:  path,
			Icon:  "internet-mail",
			Action: func() {
				EmailFile(path)
			},
		},
		{
			Type:    TypeFileAction,
			Title:   "Move to Trash",
			Desc:    path,
			Icon:    "user-trash",
			Confirm: true,
			Action: func() {
				moveToTrash(path)
			},
		},
	}
}

func FileBufferSearch(query string) []Result {
	if _, ok := MatchCommand(query, "buffer", "buf"); !ok {
		return nil
	}

	paths := FileBuffer()
	if len(paths) == 0 {
		return []Result{{
			Type:   TypeFileBuffer,
			Title:  "File Buffer Empty",
			Desc:   "Select file result, press Tab, choose Add to Buffer",
			Icon:   "folder",
			Action: func() {},
		}}
	}

	results := []Result{
		{
			Type:  TypeFileBufferAction,
			Title: "Open Buffered Files",
			Desc:  bufferSummary(len(paths)),
			Icon:  "document-open",
			Action: func() {
				for _, p := range FileBuffer() {
					Open(p)
				}
			},
		},
		{
			Type:  TypeFileBufferAction,
			Title: "Copy Buffered Paths",
			Desc:  bufferSummary(len(paths)),
			Icon:  "edit-copy",
			Action: func() {
				copyText(strings.Join(FileBuffer(), "\n"))
			},
		},
		{
			Type:  TypeFileBufferAction,
			Title: "Reveal First Buffered File",
			Desc:  filepath.Dir(paths[0]),
			Icon:  "folder-open",
			Action: func() {
				current := FileBuffer()
				if len(current) > 0 {
					revealFile(current[0])
				}
			},
		},
		{
			Type:  TypeFileBufferAction,
			Title: "Email Buffered Files",
			Desc:  bufferSummary(len(paths)),
			Icon:  "internet-mail",
			Action: func() {
				EmailFiles(FileBuffer())
			},
		},
		{
			Type:     TypeFileBufferAction,
			Title:    "Clear Buffer",
			Desc:     bufferSummary(len(paths)),
			Icon:     "edit-clear",
			KeepOpen: true,
			Action: func() {
				ClearFileBuffer()
			},
		},
	}

	for _, path := range paths {
		p := path
		results = append(results, Result{
			Type:  TypeFile,
			Title: filepath.Base(p),
			Desc:  shortenPath(filepath.Dir(p)),
			Icon:  getFileIcon(p),
			Action: func() {
				Open(p)
			},
		})
	}
	return results
}

func AddFileToBuffer(path string) {
	fileBufferMu.Lock()
	defer fileBufferMu.Unlock()
	loadFileBufferLocked()
	for _, existing := range fileBuffer {
		if existing == path {
			return
		}
	}
	fileBuffer = append(fileBuffer, path)
	saveFileBufferLocked()
}

func FileBuffer() []string {
	fileBufferMu.Lock()
	defer fileBufferMu.Unlock()
	loadFileBufferLocked()
	out := make([]string, len(fileBuffer))
	copy(out, fileBuffer)
	return out
}

func ClearFileBuffer() {
	fileBufferMu.Lock()
	defer fileBufferMu.Unlock()
	loadFileBufferLocked()
	fileBuffer = nil
	saveFileBufferLocked()
}

func fileBufferPath() string {
	return config.DataFile("file-buffer.json")
}

func loadFileBufferLocked() {
	if bufferLoaded {
		return
	}
	bufferLoaded = true
	data, err := os.ReadFile(fileBufferPath())
	if err != nil {
		return
	}
	json.Unmarshal(data, &fileBuffer)
}

func saveFileBufferLocked() {
	os.MkdirAll(filepath.Dir(fileBufferPath()), 0755)
	data, _ := json.Marshal(fileBuffer)
	os.WriteFile(fileBufferPath(), data, 0644)
}

func revealFile(path string) {
	if _, err := exec.LookPath("dbus-send"); err == nil {
		uri := (&url.URL{Scheme: "file", Path: filepath.ToSlash(path)}).String()
		Start("dbus-send", "--session", "--dest=org.freedesktop.FileManager1", "--type=method_call", "/org/freedesktop/FileManager1", "org.freedesktop.FileManager1.ShowItems", "array:string:"+uri, "string:")
		return
	}
	Start("xdg-open", filepath.Dir(path))
}

func copyText(text string) {
	cmd := exec.Command("wl-copy")
	cmd.Stdin = strings.NewReader(text)
	cmd.Run()
}

func moveToTrash(path string) {
	if _, err := exec.LookPath("gio"); err == nil {
		if exec.Command("gio", "trash", path).Run() == nil {
			SetTrashUndo()
		}
		return
	}
	Run("kioclient6", "move", path, "trash:/")
}

func openFileOpWindow(op, source, target string) {
	if exe, err := os.Executable(); err == nil {
		Start(exe, "--file-op-window", op, source, target)
	}
}

func bufferSummary(count int) string {
	if count == 1 {
		return "1 file"
	}
	return fmt.Sprintf("%d files", count)
}
