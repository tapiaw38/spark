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

type PlayerKind string

const (
	PlayerSpotify PlayerKind = "spotify"
	PlayerYouTube PlayerKind = "youtube"
)

func GetPlayerInfo(kind PlayerKind) *SpotifyInfo {
	player := mediaPlayer(kind)
	if player == "" {
		return nil
	}

	title, _ := playerctlMedia(player, "metadata", "title").Output()
	if len(strings.TrimSpace(string(title))) == 0 {
		return nil
	}

	artist, _ := playerctlMedia(player, "metadata", "artist").Output()
	album, _ := playerctlMedia(player, "metadata", "album").Output()
	status, _ := playerctlMedia(player, "status").Output()
	artURL, _ := playerctlMedia(player, "metadata", "mpris:artUrl").Output()

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

func spotifyPlayer() string {
	return mediaPlayer(PlayerSpotify)
}

func mediaPlayer(kind PlayerKind) string {
	out, err := exec.Command("playerctl", "-l").Output()
	if err != nil {
		return ""
	}
	var fallback string
	for _, line := range strings.Split(string(out), "\n") {
		player := strings.TrimSpace(line)
		if player == "" {
			continue
		}
		lower := strings.ToLower(player)
		switch kind {
		case PlayerSpotify:
			if strings.Contains(lower, "spotify") {
				return player
			}
		case PlayerYouTube:
			if strings.Contains(lower, "youtube") {
				return player
			}
			if fallback == "" && (strings.Contains(lower, "firefox") || strings.Contains(lower, "chrome") || strings.Contains(lower, "chromium") || strings.Contains(lower, "brave") || strings.Contains(lower, "vivaldi")) {
				fallback = player
			}
		}
	}
	return fallback
}

func playerctlMedia(player string, args ...string) *exec.Cmd {
	all := append([]string{"--player=" + player}, args...)
	return exec.Command("playerctl", all...)
}

func playerctlSpotify(player string, args ...string) *exec.Cmd {
	return playerctlMedia(player, args...)
}

func playerAction(kind PlayerKind, args ...string) func() {
	return func() {
		player := mediaPlayer(kind)
		if player == "" {
			SetStatus(false, string(kind)+" player not detected")
			return
		}
		playerctlMedia(player, args...).Run()
	}
}

func PlayerControls(kind PlayerKind) []Result {
	label := "Spotify"
	if kind == PlayerYouTube {
		label = "YouTube"
	}
	return []Result{
		{Type: "media-control", Title: "Play/Pause", Desc: label, Icon: "media-playback-start", Action: playerAction(kind, "play-pause")},
		{Type: "media-control", Title: "Next", Desc: label, Icon: "media-skip-forward", Action: playerAction(kind, "next")},
		{Type: "media-control", Title: "Previous", Desc: label, Icon: "media-skip-backward", Action: playerAction(kind, "previous")},
		{Type: "media-control", Title: "Volume Up", Desc: "+10%", Icon: "audio-volume-high", Action: playerAction(kind, "volume", "0.1+")},
		{Type: "media-control", Title: "Volume Down", Desc: "-10%", Icon: "audio-volume-low", Action: playerAction(kind, "volume", "0.1-")},
	}
}

// cacheAlbumArt downloads and caches album art, returns local path
func cacheAlbumArt(url string) string {
	cacheDir := cacheSubdir("art")
	if cacheDir == "" {
		return ""
	}

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
	return PlayerControls(PlayerSpotify)
}

// IsSpotifyQuery returns true if query triggers spotify mode
func IsSpotifyQuery(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	return q == "sp" || q == "spotify" || strings.HasPrefix(q, "sp ")
}

func IsYouTubePlayerQuery(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	return q == "yp" || q == "youtube player" || strings.HasPrefix(q, "yp ") || strings.HasPrefix(q, "youtube player ")
}

// SpotifySearch returns music control results (legacy, for non-spotify-mode)
func SpotifySearch(query string) []Result {
	q := strings.ToLower(strings.TrimSpace(query))

	// Only trigger on specific control keywords outside spotify mode
	if q == "play" || q == "pause" || q == "next" || q == "prev" {
		player := spotifyPlayer()
		run := func(args ...string) func() {
			return func() {
				if player == "" {
					SetStatus(false, "Spotify player not detected")
					return
				}
				playerctlSpotify(player, args...).Run()
			}
		}
		switch q {
		case "play", "pause":
			return []Result{{Type: "spotify", Title: "Play/Pause", Icon: "media-playback-start",
				Action: run("play-pause")}}
		case "next":
			return []Result{{Type: "spotify", Title: "Next Track", Icon: "media-skip-forward",
				Action: run("next")}}
		case "prev":
			return []Result{{Type: "spotify", Title: "Previous Track", Icon: "media-skip-backward",
				Action: run("previous")}}
		}
	}

	return nil
}
