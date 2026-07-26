package modules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tapiaw38/spark/internal/platform/commands"

	"github.com/tapiaw38/spark/internal/config"
)

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
	if results := musicBrowseSearch(term); results != nil {
		return results
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
		return []Result{{
			Type:  TypeMusic,
			Title: "Browse Music",
			Desc:  "m artists / m albums / m genres / m <song>",
			Icon:  "folder-music",
		}}
	}

	musicDir := config.HomeFile("Music")
	if _, err := os.Stat(musicDir); err != nil {
		return []Result{{
			Type:  TypeMusic,
			Title: "Music folder not found",
			Desc:  musicDir,
			Icon:  "folder-music",
		}}
	}

	paths := findAudioFiles(musicDir, term)
	if mode != "track" {
		paths = findAudioFiles(musicDir, "")
		paths = filterMusicByTag(paths, mode, term)
	}
	results := make([]Result, 0, len(paths))
	for _, path := range paths {
		results = append(results, Result{
			Type:       TypeMusic,
			Title:      strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
			Desc:       shortenPath(filepath.Dir(path)),
			Icon:       "audio-x-generic",
			ActionSpec: OpenAction(path),
		})
		results = append(results, Result{
			Type:       TypeMusic,
			Title:      "Queue: " + strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
			Desc:       shortenPath(filepath.Dir(path)),
			Icon:       "list-add",
			KeepOpen:   true,
			ActionSpec: StateAction("music-queue-add", path),
		})
	}
	return results
}

func MusicQueueSearch(query string) []Result {
	if !MatchesAny(query, "mq", "music queue", "queue music") {
		return nil
	}
	queue := MusicQueue()
	if len(queue) == 0 {
		return []Result{{
			Type:  TypeMusic,
			Title: "Music Queue Empty",
			Desc:  "Search m song, choose Queue result",
			Icon:  "audio-x-generic",
		}}
	}
	results := []Result{
		{
			Type:       TypeMusic,
			Title:      "Play Queue",
			Desc:       stringInt(len(queue)) + " tracks",
			Icon:       "media-playback-start",
			ActionSpec: MusicAction("play"),
		},
		{
			Type:       TypeMusic,
			Title:      "Play Queue with mpv",
			Desc:       stringInt(len(queue)) + " tracks",
			Icon:       "media-playback-start",
			ActionSpec: MusicAction("play-with", "mpv"),
		},
		{
			Type:       TypeMusic,
			Title:      "Clear Queue",
			Desc:       stringInt(len(queue)) + " tracks",
			Icon:       "edit-clear",
			KeepOpen:   true,
			ActionSpec: StateAction("music-queue-clear"),
		},
	}
	for _, path := range queue {
		results = append(results, Result{
			Type:       TypeMusic,
			Title:      filepath.Base(path),
			Desc:       shortenPath(filepath.Dir(path)),
			Icon:       "audio-x-generic",
			ActionSpec: OpenAction(path),
		})
	}
	return results
}

func musicBrowseSearch(term string) []Result {
	q := strings.ToLower(strings.TrimSpace(term))
	mode := ""
	switch q {
	case "browse", "artists", "artist":
		mode = "artist"
	case "albums", "album":
		mode = "album"
	case "genres", "genre":
		mode = "genre"
	default:
		return nil
	}
	values := musicTagValues(mode)
	if len(values) == 0 {
		return []Result{{
			Type:  TypeMusic,
			Title: "No " + mode + " tags found",
			Desc:  "Requires ffprobe tags or filenames",
			Icon:  "audio-x-generic",
		}}
	}
	results := make([]Result, 0, len(values))
	for _, value := range values {
		v := value
		results = append(results, Result{
			Type:          TypeMusic,
			Title:         strings.Title(mode) + ": " + v,
			Desc:          "Browse " + mode,
			Icon:          "audio-x-generic",
			NavigateQuery: "m " + mode + " " + v,
			KeepOpen:      true,
		})
	}
	return results
}

func musicTagValues(tag string) []string {
	musicDir := config.HomeFile("Music")
	paths := findAudioFiles(musicDir, "")
	seen := make(map[string]bool)
	var values []string
	if _, err := commands.LookPath("ffprobe"); err == nil {
		for _, path := range paths {
			cmd := commands.Command("ffprobe", "-v", "quiet", "-show_entries", "format_tags="+tag, "-of", "default=noprint_wrappers=1:nokey=1", path)
			data, err := cmd.Output()
			if err != nil {
				continue
			}
			for _, line := range strings.Split(string(data), "\n") {
				value := strings.TrimSpace(line)
				key := strings.ToLower(value)
				if value == "" || seen[key] {
					continue
				}
				seen[key] = true
				values = append(values, value)
				if len(values) >= MaxBrowsingResults {
					return values
				}
			}
		}
	}
	if len(values) == 0 {
		for _, path := range paths {
			value := filepath.Base(filepath.Dir(path))
			key := strings.ToLower(value)
			if value == "" || seen[key] {
				continue
			}
			seen[key] = true
			values = append(values, value)
			if len(values) >= MaxBrowsingResults {
				break
			}
		}
	}
	return values
}

func findAudioFiles(dir, term string) []string {
	var cmd *commands.Cmd
	if _, err := commands.LookPath("fd"); err == nil {
		args := []string{"--max-results", "200", "--type", "f", "--extension", "mp3", "--extension", "flac", "--extension", "ogg", "--extension", "wav", "--extension", "m4a"}
		if term != "" {
			args = append(args, term)
		}
		args = append(args, dir)
		cmd = commands.Command("fd", args...)
	} else {
		pattern := "*"
		if term != "" {
			pattern = "*" + term + "*"
		}
		cmd = commands.Command("find", dir, "-maxdepth", "5", "-type", "f", "-iname", pattern)
	}

	done := make(chan []byte, 1)
	go func() {
		out, _ := cmd.Output()
		done <- out
	}()

	var out []byte
	select {
	case out = <-done:
	case <-time.After(MusicProbeTimeout):
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
		if len(paths) >= MaxBrowsingResults {
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
	if _, err := commands.LookPath("ffprobe"); err != nil {
		return filterMusicByPath(paths, term)
	}
	var out []string
	needle := strings.ToLower(term)
	for _, path := range paths {
		cmd := commands.Command("ffprobe", "-v", "quiet", "-show_entries", "format_tags="+tag, "-of", "default=noprint_wrappers=1:nokey=1", path)
		data, err := cmd.Output()
		if err != nil {
			continue
		}
		if strings.Contains(strings.ToLower(string(data)), needle) {
			out = append(out, path)
		}
		if len(out) >= MaxBrowsingResults {
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
	return config.DataFile("music-queue.json")
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
