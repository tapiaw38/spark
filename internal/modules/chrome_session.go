package modules

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

var chromeYouTubeURLPattern = regexp.MustCompile(`https://(?:www\.)?(?:youtube\.com/watch\?v=|youtu\.be/|youtube\.com/shorts/|youtube\.com/embed/)[A-Za-z0-9_-]+[A-Za-z0-9_?=&%./:-]*`)

func chromeCurrentYouTubeVideo() firefoxVideo {
	files := chromeSessionFiles()
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		matches := chromeYouTubeURLPattern.FindAll(data, -1)
		for i := len(matches) - 1; i >= 0; i-- {
			url := cleanChromeSessionURL(string(matches[i]))
			if isYouTubeVideoURL(url) {
				return firefoxVideo{URL: url}
			}
		}
	}
	return firefoxVideo{}
}

func chromeSessionFiles() []string {
	home := os.Getenv("HOME")
	if home == "" {
		return nil
	}
	roots := []string{
		filepath.Join(home, ".config", "google-chrome"),
		filepath.Join(home, ".config", "chromium"),
		filepath.Join(home, ".config", "brave-browser"),
	}
	var files []string
	for _, root := range roots {
		for _, pattern := range []string{
			filepath.Join(root, "*", "Sessions", "Session_*"),
			filepath.Join(root, "*", "Sessions", "Tabs_*"),
			filepath.Join(root, "*", "Current Session"),
			filepath.Join(root, "*", "Current Tabs"),
		} {
			matches, _ := filepath.Glob(pattern)
			files = append(files, matches...)
		}
	}
	sort.Slice(files, func(i, j int) bool {
		a, errA := os.Stat(files[i])
		b, errB := os.Stat(files[j])
		if errA != nil || errB != nil {
			return files[i] < files[j]
		}
		return a.ModTime().After(b.ModTime())
	})
	return files
}

func cleanChromeSessionURL(raw string) string {
	url := raw
	for _, sep := range []byte{0, '"', '\'', '<', '>', ' ', '\n', '\r', '\t'} {
		if idx := bytes.IndexByte([]byte(url), sep); idx >= 0 {
			url = url[:idx]
		}
	}
	return url
}
