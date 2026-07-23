package modules

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// MusicSearch searches local audio files under ~/Music.
func MusicSearch(query string) []Result {
	if !strings.HasPrefix(strings.ToLower(query), "m ") &&
		!strings.HasPrefix(strings.ToLower(query), "music ") {
		return nil
	}

	term := strings.TrimSpace(query[2:])
	if strings.HasPrefix(strings.ToLower(query), "music ") {
		term = strings.TrimSpace(query[6:])
	}
	if len(term) < 2 {
		return nil
	}

	musicDir := filepath.Join(os.Getenv("HOME"), "Music")
	if _, err := os.Stat(musicDir); err != nil {
		return []Result{{
			Type:   "music",
			Title:  "Music folder not found",
			Desc:   musicDir,
			Icon:   "folder-music",
			Action: func() {},
		}}
	}

	paths := findAudioFiles(musicDir, term)
	results := make([]Result, 0, len(paths))
	for _, path := range paths {
		p := path
		results = append(results, Result{
			Type:  "music",
			Title: strings.TrimSuffix(filepath.Base(p), filepath.Ext(p)),
			Desc:  shortenPath(filepath.Dir(p)),
			Icon:  "audio-x-generic",
			Action: func() {
				exec.Command("xdg-open", p).Start()
			},
		})
	}
	return results
}

func findAudioFiles(dir, term string) []string {
	var cmd *exec.Cmd
	if _, err := exec.LookPath("fd"); err == nil {
		cmd = exec.Command("fd", "--max-results", "50", "--type", "f", "--extension", "mp3", "--extension", "flac", "--extension", "ogg", "--extension", "wav", "--extension", "m4a", term, dir)
	} else {
		cmd = exec.Command("find", dir, "-maxdepth", "5", "-type", "f", "-iname", "*"+term+"*")
	}

	done := make(chan []byte, 1)
	go func() {
		out, _ := cmd.Output()
		done <- out
	}()

	var out []byte
	select {
	case out = <-done:
	case <-time.After(700 * time.Millisecond):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return nil
	}

	var paths []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" || !isAudioFile(line) {
			continue
		}
		paths = append(paths, line)
		if len(paths) >= 50 {
			break
		}
	}
	return paths
}

func isAudioFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp3", ".flac", ".ogg", ".wav", ".m4a", ".aac":
		return true
	default:
		return false
	}
}
