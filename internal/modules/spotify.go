package modules

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SpotifyInfo contains current playback info
type SpotifyInfo struct {
	Title   string
	Artist  string
	Album   string
	Status  string // Playing, Paused, Stopped
	ArtURL  string
	ArtPath string // Local cached path
}

// GetSpotifyInfo returns current playback information
func GetSpotifyInfo() *SpotifyInfo {
	title, _ := exec.Command("playerctl", "metadata", "title").Output()
	if len(strings.TrimSpace(string(title))) == 0 {
		return nil
	}

	artist, _ := exec.Command("playerctl", "metadata", "artist").Output()
	album, _ := exec.Command("playerctl", "metadata", "album").Output()
	status, _ := exec.Command("playerctl", "status").Output()
	artURL, _ := exec.Command("playerctl", "metadata", "mpris:artUrl").Output()

	info := &SpotifyInfo{
		Title:  strings.TrimSpace(string(title)),
		Artist: strings.TrimSpace(string(artist)),
		Album:  strings.TrimSpace(string(album)),
		Status: strings.TrimSpace(string(status)),
		ArtURL: strings.TrimSpace(string(artURL)),
	}

	// Cache album art locally
	if info.ArtURL != "" {
		info.ArtPath = cacheAlbumArt(info.ArtURL)
	}

	return info
}

// cacheAlbumArt downloads and caches album art, returns local path
func cacheAlbumArt(url string) string {
	cacheDir := filepath.Join(os.Getenv("HOME"), ".cache", "spark", "art")
	os.MkdirAll(cacheDir, 0755)

	// Use URL hash as filename
	hash := simpleHash(url)
	cachePath := filepath.Join(cacheDir, hash+".jpg")

	// Check if already cached
	if _, err := os.Stat(cachePath); err == nil {
		return cachePath
	}

	// Handle file:// URLs
	if strings.HasPrefix(url, "file://") {
		localPath := strings.TrimPrefix(url, "file://")
		return localPath
	}

	// Download from http/https
	if strings.HasPrefix(url, "http") {
		resp, err := http.Get(url)
		if err != nil {
			return ""
		}
		defer resp.Body.Close()

		f, err := os.Create(cachePath)
		if err != nil {
			return ""
		}
		defer f.Close()

		io.Copy(f, resp.Body)
		return cachePath
	}

	return ""
}

func simpleHash(s string) string {
	h := uint32(0)
	for i := 0; i < len(s); i++ {
		h = h*31 + uint32(s[i])
	}
	// Convert to hex string
	const hex = "0123456789abcdef"
	result := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		result[i] = hex[h&0xf]
		h >>= 4
	}
	return string(result)
}

// SpotifyControls returns control actions
func SpotifyControls() []Result {
	return []Result{
		{Type: "spotify-control", Title: "Play/Pause", Desc: "Toggle playback", Icon: "media-playback-start",
			Action: func() { exec.Command("playerctl", "play-pause").Run() }},
		{Type: "spotify-control", Title: "Next", Desc: "Next track", Icon: "media-skip-forward",
			Action: func() { exec.Command("playerctl", "next").Run() }},
		{Type: "spotify-control", Title: "Previous", Desc: "Previous track", Icon: "media-skip-backward",
			Action: func() { exec.Command("playerctl", "previous").Run() }},
		{Type: "spotify-control", Title: "Volume Up", Desc: "+10%", Icon: "audio-volume-high",
			Action: func() { exec.Command("playerctl", "volume", "0.1+").Run() }},
		{Type: "spotify-control", Title: "Volume Down", Desc: "-10%", Icon: "audio-volume-low",
			Action: func() { exec.Command("playerctl", "volume", "0.1-").Run() }},
	}
}

// IsSpotifyQuery returns true if query triggers spotify mode
func IsSpotifyQuery(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	return q == "sp" || q == "spotify" || strings.HasPrefix(q, "sp ")
}

// SpotifySearch returns music control results (legacy, for non-spotify-mode)
func SpotifySearch(query string) []Result {
	q := strings.ToLower(strings.TrimSpace(query))

	// Only trigger on specific control keywords outside spotify mode
	if q == "play" || q == "pause" || q == "next" || q == "prev" {
		switch q {
		case "play", "pause":
			return []Result{{Type: "spotify", Title: "Play/Pause", Icon: "media-playback-start",
				Action: func() { exec.Command("playerctl", "play-pause").Run() }}}
		case "next":
			return []Result{{Type: "spotify", Title: "Next Track", Icon: "media-skip-forward",
				Action: func() { exec.Command("playerctl", "next").Run() }}}
		case "prev":
			return []Result{{Type: "spotify", Title: "Previous Track", Icon: "media-skip-backward",
				Action: func() { exec.Command("playerctl", "previous").Run() }}}
		}
	}

	return nil
}
