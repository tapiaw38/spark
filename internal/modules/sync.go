package modules

import (
	"archive/zip"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// SyncSearch exposes settings paths for external sync tools.
func SyncSearch(query string) []Result {
	q := strings.ToLower(strings.TrimSpace(query))
	if q != "sync" && q != "settings sync" && !strings.HasPrefix(q, "sync import ") {
		return nil
	}

	configDir := filepath.Join(os.Getenv("HOME"), ".config", "spark")
	dataDir := filepath.Join(os.Getenv("HOME"), ".local", "share", "spark")
	paths := configDir + "\n" + dataDir
	exportPath := filepath.Join(os.Getenv("HOME"), "spark-settings.zip")

	if strings.HasPrefix(q, "sync import ") {
		zipPath := strings.TrimSpace(query[len("sync import "):])
		return []Result{{
			Type:    "sync",
			Title:   "Import Spark Settings",
			Desc:    zipPath,
			Icon:    "document-open",
			Confirm: true,
			Action: func() {
				if err := importSyncZip(expandHome(zipPath), os.Getenv("HOME")); err != nil {
					SetStatus(false, "Sync import failed: "+err.Error())
				} else {
					SetStatus(true, "Imported Spark settings from "+zipPath)
				}
			},
		}}
	}

	return []Result{
		{
			Type:  "sync",
			Title: "Open Settings Folder",
			Desc:  configDir,
			Icon:  "folder-open",
			Action: func() {
				exec.Command("xdg-open", configDir).Start()
			},
		},
		{
			Type:  "sync",
			Title: "Copy Sync Paths",
			Desc:  "~/.config/spark + ~/.local/share/spark",
			Icon:  "edit-copy",
			Action: func() {
				copyText(paths)
			},
		},
		{
			Type:  "sync",
			Title: "Export Settings Zip",
			Desc:  exportPath,
			Icon:  "document-save",
			Action: func() {
				if err := exportSyncZip(exportPath, []string{configDir, dataDir}); err != nil {
					SetStatus(false, "Sync export failed: "+err.Error())
				} else {
					SetStatus(true, "Exported Spark settings to "+exportPath)
					revealFile(exportPath)
				}
			},
		},
		{
			Type:  "sync",
			Title: "Sync with Git/Syncthing",
			Desc:  "Track copied paths with external sync",
			Icon:  "emblem-synchronizing",
			Action: func() {
				copyText("Sync these Spark paths:\n" + paths)
			},
		},
	}
}

func exportSyncZip(target string, paths []string) error {
	os.MkdirAll(filepath.Dir(target), 0755)
	out, err := os.Create(target)
	if err != nil {
		return err
	}
	defer out.Close()

	writer := zip.NewWriter(out)
	defer writer.Close()

	for _, root := range paths {
		filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(os.Getenv("HOME"), path)
			if err != nil {
				return nil
			}
			header, err := zip.FileInfoHeader(fileInfo(path))
			if err != nil {
				return nil
			}
			header.Name = rel
			header.Method = zip.Deflate
			w, err := writer.CreateHeader(header)
			if err != nil {
				return nil
			}
			in, err := os.Open(path)
			if err != nil {
				return nil
			}
			defer in.Close()
			io.Copy(w, in)
			return nil
		})
	}
	return nil
}

func importSyncZip(source, home string) error {
	reader, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		target := filepath.Clean(filepath.Join(home, file.Name))
		if !strings.HasPrefix(target, home) {
			continue
		}
		if file.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}
		os.MkdirAll(filepath.Dir(target), 0755)
		rc, err := file.Open()
		if err != nil {
			continue
		}
		out, err := os.Create(target)
		if err != nil {
			rc.Close()
			continue
		}
		io.Copy(out, rc)
		out.Close()
		rc.Close()
		os.Chmod(target, file.FileInfo().Mode())
	}
	return nil
}

func fileInfo(path string) os.FileInfo {
	info, err := os.Stat(path)
	if err != nil {
		return fakeFileInfo{name: filepath.Base(path)}
	}
	return info
}

type fakeFileInfo struct{ name string }

func (f fakeFileInfo) Name() string       { return f.name }
func (f fakeFileInfo) Size() int64        { return 0 }
func (f fakeFileInfo) Mode() os.FileMode  { return 0644 }
func (f fakeFileInfo) ModTime() time.Time { return time.Now() }
func (f fakeFileInfo) IsDir() bool        { return false }
func (f fakeFileInfo) Sys() any           { return nil }
