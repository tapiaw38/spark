package modules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type firefoxVideo struct {
	URL   string
	Title string
}

type firefoxSession struct {
	Windows []struct {
		Selected int `json:"selected"`
		Tabs     []struct {
			Index   int `json:"index"`
			Entries []struct {
				URL   string `json:"url"`
				Title string `json:"title"`
			} `json:"entries"`
		} `json:"tabs"`
	} `json:"windows"`
}

func firefoxCurrentYouTubeVideo() firefoxVideo {
	files := firefoxSessionFiles()
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		jsonData, err := decodeMozLz4(data)
		if err != nil {
			continue
		}
		video := youtubeVideoFromFirefoxSession(jsonData)
		if video.URL != "" {
			return video
		}
	}
	return firefoxVideo{}
}

func firefoxSessionFiles() []string {
	home := os.Getenv("HOME")
	if home == "" {
		return nil
	}
	var files []string
	root := filepath.Join(home, ".mozilla", "firefox")
	patterns := []string{
		filepath.Join(root, "*", "sessionstore-backups", "recovery.jsonlz4"),
		filepath.Join(root, "*", "sessionstore.jsonlz4"),
	}
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		files = append(files, matches...)
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

func youtubeVideoFromFirefoxSession(data []byte) firefoxVideo {
	var session firefoxSession
	if json.Unmarshal(data, &session) != nil {
		return firefoxVideo{}
	}

	for _, window := range session.Windows {
		tabIdx := window.Selected - 1
		if tabIdx < 0 || tabIdx >= len(window.Tabs) {
			continue
		}
		if video := youtubeVideoFromFirefoxTab(window.Tabs[tabIdx].Index, window.Tabs[tabIdx].Entries); video.URL != "" {
			return video
		}
	}

	for _, window := range session.Windows {
		for _, tab := range window.Tabs {
			if video := youtubeVideoFromFirefoxTab(tab.Index, tab.Entries); video.URL != "" {
				return video
			}
		}
	}
	return firefoxVideo{}
}

func youtubeVideoFromFirefoxTab(index int, entries []struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}) firefoxVideo {
	if len(entries) == 0 {
		return firefoxVideo{}
	}
	entryIdx := index - 1
	if entryIdx < 0 || entryIdx >= len(entries) {
		entryIdx = len(entries) - 1
	}
	entry := entries[entryIdx]
	if isYouTubeVideoURL(entry.URL) {
		return firefoxVideo{URL: entry.URL, Title: cleanYouTubeTitle(entry.Title)}
	}
	return firefoxVideo{}
}

func cleanYouTubeTitle(title string) string {
	title = strings.TrimSpace(title)
	title = strings.TrimSuffix(title, " - YouTube")
	if title == "" {
		return "YouTube video"
	}
	return title
}

func decodeMozLz4(data []byte) ([]byte, error) {
	const header = "mozLz40\x00"
	if len(data) < len(header)+4 || string(data[:len(header)]) != header {
		return nil, os.ErrInvalid
	}
	return decodeLz4Block(data[len(header)+4:])
}

func decodeLz4Block(src []byte) ([]byte, error) {
	dst := make([]byte, 0, len(src)*3)
	for i := 0; i < len(src); {
		token := int(src[i])
		i++

		litLen := token >> 4
		if litLen == 15 {
			for {
				if i >= len(src) {
					return nil, os.ErrInvalid
				}
				n := int(src[i])
				i++
				litLen += n
				if n != 255 {
					break
				}
			}
		}
		if i+litLen > len(src) {
			return nil, os.ErrInvalid
		}
		dst = append(dst, src[i:i+litLen]...)
		i += litLen
		if i >= len(src) {
			break
		}
		if i+2 > len(src) {
			return nil, os.ErrInvalid
		}
		offset := int(src[i]) | int(src[i+1])<<8
		i += 2
		if offset <= 0 || offset > len(dst) {
			return nil, os.ErrInvalid
		}

		matchLen := token&0x0f + 4
		if token&0x0f == 15 {
			for {
				if i >= len(src) {
					return nil, os.ErrInvalid
				}
				n := int(src[i])
				i++
				matchLen += n
				if n != 255 {
					break
				}
			}
		}
		for n := 0; n < matchLen; n++ {
			dst = append(dst, dst[len(dst)-offset])
		}
	}
	return dst, nil
}
