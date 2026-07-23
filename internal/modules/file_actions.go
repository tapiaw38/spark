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
)

var (
	fileBuffer   []string
	bufferLoaded bool
	fileBufferMu sync.Mutex
)

// FileActions returns actions for a selected file result.
func FileActions(path string) []Result {
	if path == "" {
		return nil
	}

	return []Result{
		{
			Type:  "file-action",
			Title: "Open",
			Desc:  path,
			Icon:  "document-open",
			Action: func() {
				exec.Command("xdg-open", path).Start()
			},
		},
		{
			Type:  "file-action",
			Title: "Reveal in Files",
			Desc:  filepath.Dir(path),
			Icon:  "folder-open",
			Action: func() {
				revealFile(path)
			},
		},
		{
			Type:  "file-action",
			Title: "Copy Path",
			Desc:  path,
			Icon:  "edit-copy",
			Action: func() {
				copyText(path)
			},
		},
		{
			Type:          "file-action",
			Title:         "Rename...",
			Desc:          "rename " + path + " | " + filepath.Base(path),
			Icon:          "edit-rename",
			KeepOpen:      true,
			NavigateQuery: "rename " + path + " | " + filepath.Base(path),
			Action:        func() {},
		},
		{
			Type:          "file-action",
			Title:         "Copy To...",
			Desc:          "pick copy " + path + " | " + filepath.Dir(path),
			Icon:          "edit-copy",
			KeepOpen:      true,
			NavigateQuery: "pick copy " + path + " | " + filepath.Dir(path),
			Action:        func() {},
		},
		{
			Type:          "file-action",
			Title:         "Move To...",
			Desc:          "pick move " + path + " | " + filepath.Dir(path),
			Icon:          "go-jump",
			KeepOpen:      true,
			NavigateQuery: "pick move " + path + " | " + filepath.Dir(path),
			Action:        func() {},
		},
		{
			Type:     "file-action",
			Title:    "Add to Buffer",
			Desc:     bufferSummary(1),
			Icon:     "list-add",
			KeepOpen: true,
			Action: func() {
				AddFileToBuffer(path)
			},
		},
		{
			Type:  "file-action",
			Title: "Email File",
			Desc:  path,
			Icon:  "internet-mail",
			Action: func() {
				EmailFile(path)
			},
		},
		{
			Type:    "file-action",
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

// FileBufferSearch lists buffered files and buffer actions.
func FileBufferSearch(query string) []Result {
	q := strings.ToLower(strings.TrimSpace(query))
	if q != "buffer" && q != "buf" && !strings.HasPrefix(q, "buffer ") && !strings.HasPrefix(q, "buf ") {
		return nil
	}

	paths := FileBuffer()
	if len(paths) == 0 {
		return []Result{{
			Type:   "file-buffer",
			Title:  "File Buffer Empty",
			Desc:   "Select file result, press Tab, choose Add to Buffer",
			Icon:   "folder",
			Action: func() {},
		}}
	}

	results := []Result{
		{
			Type:  "file-buffer-action",
			Title: "Open Buffered Files",
			Desc:  bufferSummary(len(paths)),
			Icon:  "document-open",
			Action: func() {
				for _, p := range FileBuffer() {
					exec.Command("xdg-open", p).Start()
				}
			},
		},
		{
			Type:  "file-buffer-action",
			Title: "Copy Buffered Paths",
			Desc:  bufferSummary(len(paths)),
			Icon:  "edit-copy",
			Action: func() {
				copyText(strings.Join(FileBuffer(), "\n"))
			},
		},
		{
			Type:  "file-buffer-action",
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
			Type:  "file-buffer-action",
			Title: "Email Buffered Files",
			Desc:  bufferSummary(len(paths)),
			Icon:  "internet-mail",
			Action: func() {
				EmailFiles(FileBuffer())
			},
		},
		{
			Type:     "file-buffer-action",
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
			Type:  "file",
			Title: filepath.Base(p),
			Desc:  shortenPath(filepath.Dir(p)),
			Icon:  getFileIcon(p),
			Action: func() {
				exec.Command("xdg-open", p).Start()
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
	return filepath.Join(os.Getenv("HOME"), ".local", "share", "spark", "file-buffer.json")
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
		exec.Command("dbus-send", "--session", "--dest=org.freedesktop.FileManager1", "--type=method_call", "/org/freedesktop/FileManager1", "org.freedesktop.FileManager1.ShowItems", "array:string:"+uri, "string:").Start()
		return
	}
	exec.Command("xdg-open", filepath.Dir(path)).Start()
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
	exec.Command("kioclient6", "move", path, "trash:/").Run()
}

func bufferSummary(count int) string {
	if count == 1 {
		return "1 file"
	}
	return fmt.Sprintf("%d files", count)
}
