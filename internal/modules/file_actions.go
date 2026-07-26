package modules

import (
	"encoding/json"
	"fmt"
	"os"
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
			Type:       TypeFileAction,
			Title:      "Open",
			Desc:       path,
			Icon:       "document-open",
			ActionSpec: OpenAction(path),
		},
		{
			Type:       TypeFileAction,
			Title:      "Reveal in Files",
			Desc:       filepath.Dir(path),
			Icon:       "folder-open",
			ActionSpec: FileAction("reveal", path),
		},
		{
			Type:       TypeFileAction,
			Title:      "Copy Path",
			Desc:       path,
			Icon:       "edit-copy",
			ActionSpec: CopyAction(path),
		},
		{
			Type:       TypeFileAction,
			Title:      "Rename...",
			Desc:       "Edit name in file operation window",
			Icon:       "edit-rename",
			ActionSpec: FileAction("op-window", "rename", path, filepath.Base(path)),
		},
		{
			Type:       TypeFileAction,
			Title:      "Copy To...",
			Desc:       "Choose destination in file operation window",
			Icon:       "edit-copy",
			ActionSpec: FileAction("op-window", "copy", path, filepath.Dir(path)),
		},
		{
			Type:       TypeFileAction,
			Title:      "Move To...",
			Desc:       "Choose destination in file operation window",
			Icon:       "go-jump",
			ActionSpec: FileAction("op-window", "move", path, filepath.Dir(path)),
		},
		{
			Type:       TypeFileAction,
			Title:      "Add to Buffer",
			Desc:       bufferSummary(1),
			Icon:       "list-add",
			KeepOpen:   true,
			ActionSpec: StateAction("file-buffer-add", path),
		},
		{
			Type:       TypeFileAction,
			Title:      "Email File",
			Desc:       path,
			Icon:       "internet-mail",
			ActionSpec: EmailAction("", "", "", path),
		},
		{
			Type:       TypeFileAction,
			Title:      "Move to Trash",
			Desc:       path,
			Icon:       "user-trash",
			Confirm:    true,
			ActionSpec: FileAction("trash", path),
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
			Type:  TypeFileBuffer,
			Title: "File Buffer Empty",
			Desc:  "Select file result, press Tab, choose Add to Buffer",
			Icon:  "folder",
		}}
	}

	results := []Result{
		{
			Type:       TypeFileBufferAction,
			Title:      "Open Buffered Files",
			Desc:       bufferSummary(len(paths)),
			Icon:       "document-open",
			ActionSpec: MusicAction("open-files", paths...),
		},
		{
			Type:       TypeFileBufferAction,
			Title:      "Copy Buffered Paths",
			Desc:       bufferSummary(len(paths)),
			Icon:       "edit-copy",
			ActionSpec: CopyAction(strings.Join(paths, "\n")),
		},
		{
			Type:       TypeFileBufferAction,
			Title:      "Reveal First Buffered File",
			Desc:       filepath.Dir(paths[0]),
			Icon:       "folder-open",
			ActionSpec: FileAction("reveal", paths[0]),
		},
		{
			Type:       TypeFileBufferAction,
			Title:      "Email Buffered Files",
			Desc:       bufferSummary(len(paths)),
			Icon:       "internet-mail",
			ActionSpec: EmailAction("", "", "", paths...),
		},
		{
			Type:       TypeFileBufferAction,
			Title:      "Clear Buffer",
			Desc:       bufferSummary(len(paths)),
			Icon:       "edit-clear",
			KeepOpen:   true,
			ActionSpec: StateAction("file-buffer-clear"),
		},
	}

	for _, path := range paths {
		results = append(results, Result{
			Type:       TypeFile,
			Title:      filepath.Base(path),
			Desc:       shortenPath(filepath.Dir(path)),
			Icon:       getFileIcon(path),
			ActionSpec: OpenAction(path),
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

func bufferSummary(count int) string {
	if count == 1 {
		return "1 file"
	}
	return fmt.Sprintf("%d files", count)
}
