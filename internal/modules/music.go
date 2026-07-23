package modules

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// MusicSearch searches local audio files under ~/Music.
func MusicSearch(query string) []Result {
	if results := MusicQueueSearch(query); results != nil {
		return results
	}

	if !strings.HasPrefix(strings.ToLower(query), "m ") &&
		!strings.HasPrefix(strings.ToLower(query), "music ") {
		return nil
	}

	term := strings.TrimSpace(query[2:])
	if strings.HasPrefix(strings.ToLower(query), "music ") {
		term = strings.TrimSpace(query[6:])
	}
	mode := "track"
	for _, prefix := range []string{"artist ", "album ", "genre "} {
		if strings.HasPrefix(strings.ToLower(term), prefix) {
			mode = strings.TrimSpace(prefix)
			term = strings.TrimSpace(term[len(prefix):])
			break
		}
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
	if mode != "track" {
		paths = findAudioFiles(musicDir, "")
		paths = filterMusicByTag(paths, mode, term)
	}
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
		results = append(results, Result{
			Type:     "music",
			Title:    "Queue: " + strings.TrimSuffix(filepath.Base(p), filepath.Ext(p)),
			Desc:     shortenPath(filepath.Dir(p)),
			Icon:     "list-add",
			KeepOpen: true,
			Action: func() {
				AddMusicToQueue(p)
			},
		})
	}
	return results
}

func MusicQueueSearch(query string) []Result {
	q := strings.ToLower(strings.TrimSpace(query))
	if q != "mq" && q != "music queue" && q != "queue music" {
		return nil
	}
	queue := MusicQueue()
	if len(queue) == 0 {
		return []Result{{
			Type:   "music",
			Title:  "Music Queue Empty",
			Desc:   "Search m song, choose Queue result",
			Icon:   "audio-x-generic",
			Action: func() {},
		}}
	}
	results := []Result{
		{
			Type:  "music",
			Title: "Play Queue",
			Desc:  stringInt(len(queue)) + " tracks",
			Icon:  "media-playback-start",
			Action: func() {
				playMusicQueue()
			},
		},
		{
			Type:     "music",
			Title:    "Clear Queue",
			Desc:     stringInt(len(queue)) + " tracks",
			Icon:     "edit-clear",
			KeepOpen: true,
			Action: func() {
				ClearMusicQueue()
			},
		},
	}
	for _, path := range queue {
		p := path
		results = append(results, Result{
			Type:  "music",
			Title: filepath.Base(p),
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
		args := []string{"--max-results", "200", "--type", "f", "--extension", "mp3", "--extension", "flac", "--extension", "ogg", "--extension", "wav", "--extension", "m4a"}
		if term != "" {
			args = append(args, term)
		}
		args = append(args, dir)
		cmd = exec.Command("fd", args...)
	} else {
		pattern := "*"
		if term != "" {
			pattern = "*" + term + "*"
		}
		cmd = exec.Command("find", dir, "-maxdepth", "5", "-type", "f", "-iname", pattern)
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

func filterMusicByTag(paths []string, tag, term string) []string {
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return filterMusicByPath(paths, term)
	}
	var out []string
	needle := strings.ToLower(term)
	for _, path := range paths {
		cmd := exec.Command("ffprobe", "-v", "quiet", "-show_entries", "format_tags="+tag, "-of", "default=noprint_wrappers=1:nokey=1", path)
		data, err := cmd.Output()
		if err != nil {
			continue
		}
		if strings.Contains(strings.ToLower(string(data)), needle) {
			out = append(out, path)
		}
		if len(out) >= 50 {
			break
		}
	}
	return out
}

func filterMusicByPath(paths []string, term string) []string {
	var out []string
	needle := strings.ToLower(term)
	for _, path := range paths {
		if strings.Contains(strings.ToLower(path), needle) {
			out = append(out, path)
		}
	}
	return out
}

func musicQueuePath() string {
	return filepath.Join(os.Getenv("HOME"), ".local", "share", "spark", "music-queue.json")
}

func MusicQueue() []string {
	data, err := os.ReadFile(musicQueuePath())
	if err != nil {
		return nil
	}
	var queue []string
	json.Unmarshal(data, &queue)
	return queue
}

func AddMusicToQueue(path string) {
	queue := MusicQueue()
	queue = append(queue, path)
	os.MkdirAll(filepath.Dir(musicQueuePath()), 0755)
	data, _ := json.Marshal(queue)
	os.WriteFile(musicQueuePath(), data, 0644)
	SetStatus(true, "Queued music: "+filepath.Base(path))
}

func ClearMusicQueue() {
	os.Remove(musicQueuePath())
	SetStatus(true, "Music queue cleared")
}

func playMusicQueue() {
	queue := MusicQueue()
	if len(queue) == 0 {
		return
	}
	if _, err := exec.LookPath("mpv"); err == nil {
		exec.Command("mpv", queue...).Start()
		return
	}
	for _, path := range queue {
		exec.Command("xdg-open", path).Start()
	}
}
