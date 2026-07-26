package modules

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tapiaw38/spark/internal/config"
)

func SyncSearch(query string) []Result {
	q := strings.ToLower(strings.TrimSpace(query))
	if q != "sync" && q != "settings sync" && !strings.HasPrefix(q, "sync import ") {
		return nil
	}

	configDir := config.ConfigDir()
	dataDir := config.DataDir()
	paths := configDir + "\n" + dataDir
	exportPath := config.HomeFile("spark-settings.zip")

	if strings.HasPrefix(q, "sync import ") {
		zipPath := strings.TrimSpace(query[len("sync import "):])
		return []Result{{
			Type:    TypeSync,
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
			Type:  TypeSync,
			Title: "Open Settings Folder",
			Desc:  configDir,
			Icon:  "folder-open",
			Action: func() {
				Open(configDir)
			},
		},
		{
			Type:  TypeSync,
			Title: "Copy Sync Paths",
			Desc:  "~/.config/spark + ~/.local/share/spark",
			Icon:  "edit-copy",
			Action: func() {
				copyText(paths)
			},
		},
		{
			Type:  TypeSync,
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
			Type:  TypeSync,
			Title: "Sync with Git/Syncthing",
			Desc:  "Track copied paths with external sync",
			Icon:  "emblem-synchronizing",
			Action: func() {
				copyText("Sync these Spark paths:\n" + paths)
			},
		},
		{
			Type:  TypeSync,
			Title: "Copy Git Bootstrap",
			Desc:  "~/.config/spark",
			Icon:  "git",
			Action: func() {
				copyText("cd ~/.config/spark\ngit init\ngit add .\ngit commit -m 'sync spark settings'\n")
			},
		},
		{
			Type:  TypeSync,
			Title: "Create Git Bootstrap Script",
			Desc:  filepath.Join(configDir, "spark-sync-git.sh"),
			Icon:  "document-save",
			Action: func() {
				path := filepath.Join(configDir, "spark-sync-git.sh")
				if err := writeSyncScript(path, configDir, dataDir); err != nil {
					SetStatus(false, "Git sync script failed: "+err.Error())
				} else {
					SetStatus(true, "Git sync script created: "+path)
				}
			},
		},
		{
			Type:  TypeSync,
			Title: "Open Syncthing",
			Desc:  "http://127.0.0.1:8384",
			Icon:  "emblem-synchronizing",
			Action: func() {
				Open("http://127.0.0.1:8384")
			},
		},
		{
			Type:  TypeSync,
			Title: "Create Sync Profile",
			Desc:  filepath.Join(configDir, "sync-profile.txt"),
			Icon:  "document-save",
			Action: func() {
				path := filepath.Join(configDir, "sync-profile.txt")
				os.MkdirAll(configDir, 0755)
				content := "Spark sync paths:\n" + paths + "\n\nSyncthing:\n- Add both folders as send/receive folders.\n- Keep ignore patterns for cache files.\n\nGit:\ncd " + configDir + "\ngit init\ngit add .\ngit commit -m 'sync spark config'\n\nData:\ncd " + dataDir + "\ngit init\ngit add .\ngit commit -m 'sync spark data'\n"
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					SetStatus(false, "Sync profile failed: "+err.Error())
				} else {
					SetStatus(true, "Sync profile created: "+path)
				}
			},
		},
	}
}

func writeSyncScript(path, configDir, dataDir string) error {
	os.MkdirAll(filepath.Dir(path), 0755)
	content := "#!/bin/sh\nset -eu\nfor dir in '" + configDir + "' '" + dataDir + "'; do\n  mkdir -p \"$dir\"\n  cd \"$dir\"\n  if [ ! -d .git ]; then git init; fi\n  git add .\n  git commit -m 'sync spark settings' || true\ndone\n"
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		return err
	}
	return nil
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
