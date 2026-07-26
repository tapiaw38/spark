package modules

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type PlaybackStatus string

const (
	StatusPlaying PlaybackStatus = "Playing"
	StatusPaused  PlaybackStatus = "Paused"
	StatusStopped PlaybackStatus = "Stopped"
)

func ParsePlaybackStatus(raw string) PlaybackStatus {
	trimmed := strings.TrimSpace(raw)
	switch strings.ToLower(trimmed) {
	case "playing":
		return StatusPlaying
	case "paused":
		return StatusPaused
	case "stopped":
		return StatusStopped
	default:
		return PlaybackStatus(trimmed)
	}
}

type PlayerInfo struct {
	Title        string
	Artist       string
	Album        string
	Status       PlaybackStatus
	ArtURL       string
	ArtCachePath string
}

type playerMeta struct {
	name   string
	title  string
	artist string
	url    string
	status PlaybackStatus
}

type PlayerKind string

const (
	PlayerSpotify PlayerKind = "spotify"
	PlayerYouTube PlayerKind = "youtube"
)

func GetPlayerInfo(kind PlayerKind) *PlayerInfo {
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

	info := &PlayerInfo{
		Title:  strings.TrimSpace(string(title)),
		Artist: strings.TrimSpace(string(artist)),
		Album:  strings.TrimSpace(string(album)),
		Status: ParsePlaybackStatus(string(status)),
		ArtURL: strings.TrimSpace(string(artURL)),
	}
	if kind == PlayerYouTube {
		mediaURL := playerMetadata(player, "xesam:url")
		if !isYouTubeVideoURL(mediaURL) {
			if video := browserCurrentYouTubeVideo(player); video.URL != "" {
				info.Title = video.Title
				if info.Title == "" {
					info.Title = strings.TrimSpace(string(title))
				}
				if info.Title == "" {
					info.Title = "YouTube video"
				}
				info.Artist = "Firefox"
				if isChromeLikePlayer(player) {
					info.Artist = "Chrome"
				}
				info.Album = video.URL
				info.ArtURL = youtubePlayerThumbnailURL(video.URL)
			} else {
				info.Title = "YouTube player"
				info.Artist = player
				info.Album = ""
				info.ArtURL = ""
			}
		}
	}

	if info.ArtURL != "" {
		info.ArtCachePath = cacheAlbumArt(info.ArtURL)
	}

	return info
}

func mediaPlayer(kind PlayerKind) string {
	players := playerMetas()
	var fallback string
	for _, meta := range players {
		lower := strings.ToLower(meta.name)
		switch kind {
		case PlayerSpotify:
			if strings.Contains(lower, "spotify") {
				return meta.name
			}
		case PlayerYouTube:
			if isYouTubeMeta(meta) {
				return meta.name
			}
			if strings.Contains(lower, "youtube") {
				return meta.name
			}
			if isBrowserPlayer(meta.name) {
				if video := browserCurrentYouTubeVideo(meta.name); video.URL != "" {
					return meta.name
				}
			}
			if fallback == "" && meta.status == StatusPlaying &&
				isBrowserPlayer(meta.name) {
				fallback = meta.name
			}
		}
	}
	return fallback
}

func playerMetas() []playerMeta {
	out, err := exec.Command("playerctl", "-l").Output()
	if err != nil {
		return nil
	}
	var players []playerMeta
	for _, line := range strings.Split(string(out), "\n") {
		player := strings.TrimSpace(line)
		if player == "" {
			continue
		}
		players = append(players, playerMeta{
			name:   player,
			title:  playerMetadata(player, "title"),
			artist: playerMetadata(player, "artist"),
			url:    playerMetadata(player, "xesam:url"),
			status: playerStatus(player),
		})
	}
	return players
}

func playerMetadata(player, key string) string {
	out, err := playerctlMedia(player, "metadata", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func playerStatus(player string) PlaybackStatus {
	out, err := playerctlMedia(player, "status").Output()
	if err != nil {
		return ""
	}
	return ParsePlaybackStatus(string(out))
}

func isYouTubeMeta(meta playerMeta) bool {
	url := strings.ToLower(meta.url)
	haystack := strings.ToLower(meta.name + " " + meta.title + " " + meta.artist)
	if isYouTubeVideoURL(url) {
		return true
	}
	if meta.status == StatusPlaying && strings.Contains(url, "youtube.com") {
		return true
	}
	return strings.Contains(haystack, "youtube") && meta.status != StatusStopped
}

func isYouTubeVideoURL(raw string) bool {
	url := strings.ToLower(raw)
	return strings.Contains(url, "youtu.be") ||
		strings.Contains(url, "youtube.com/watch") ||
		strings.Contains(url, "youtube.com/shorts") ||
		strings.Contains(url, "youtube.com/embed")
}

func youtubePlayerThumbnailURL(raw string) string {
	id := youtubeVideoID(raw)
	if id == "" {
		return ""
	}
	return "https://img.youtube.com/vi/" + id + "/hqdefault.jpg"
}

func youtubeVideoID(raw string) string {
	for _, marker := range []string{"v=", "youtu.be/", "youtube.com/shorts/", "youtube.com/embed/"} {
		if idx := strings.Index(raw, marker); idx >= 0 {
			id := raw[idx+len(marker):]
			for cut, r := range id {
				if !(r == '-' || r == '_' || r >= '0' && r <= '9' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z') {
					return id[:cut]
				}
			}
			return id
		}
	}
	return ""
}

func browserCurrentYouTubeVideo(player string) firefoxVideo {
	if isFirefoxPlayer(player) {
		return firefoxCurrentYouTubeVideo()
	}
	if isChromeLikePlayer(player) {
		return chromeCurrentYouTubeVideo()
	}
	if video := firefoxCurrentYouTubeVideo(); video.URL != "" {
		return video
	}
	return chromeCurrentYouTubeVideo()
}

func isFirefoxPlayer(player string) bool {
	return strings.Contains(strings.ToLower(player), "firefox")
}

func isChromeLikePlayer(player string) bool {
	lower := strings.ToLower(player)
	return strings.Contains(lower, "chrome") ||
		strings.Contains(lower, "chromium") ||
		strings.Contains(lower, "brave") ||
		strings.Contains(lower, "vivaldi")
}

func isBrowserPlayer(player string) bool {
	return isFirefoxPlayer(player) || isChromeLikePlayer(player)
}

func playerctlMedia(player string, args ...string) *exec.Cmd {
	all := append([]string{"--player=" + player}, args...)
	return exec.Command("playerctl", all...)
}

func playerAction(kind PlayerKind, args ...string) func() {
	return func() {
		player := mediaPlayer(kind)
		if player == "" {
			SetStatus(false, string(kind)+" player not detected")
			return
		}
		if err := playerctlMedia(player, args...).Run(); err != nil {
			SetStatus(false, string(kind)+" command failed: "+strings.Join(args, " "))
			time.AfterFunc(3*time.Second, func() { SetStatus(true, "") })
			return
		}
		SetStatus(true, "")
	}
}

func PlayerControls(kind PlayerKind) []Result {
	label := "Spotify"
	if kind == PlayerYouTube {
		label = "YouTube"
	}
	return []Result{
		{Type: TypeMediaControl, Title: "Play/Pause", Desc: label, Icon: "media-playback-start", Action: playerAction(kind, "play-pause")},
		{Type: TypeMediaControl, Title: "Next", Desc: label, Icon: "media-skip-forward", Action: playerAction(kind, "next")},
		{Type: TypeMediaControl, Title: "Previous", Desc: label, Icon: "media-skip-backward", Action: playerAction(kind, "previous")},
		{Type: TypeMediaControl, Title: "Volume Up", Desc: "+10%", Icon: "audio-volume-high", Action: playerAction(kind, "volume", "0.1+")},
		{Type: TypeMediaControl, Title: "Volume Down", Desc: "-10%", Icon: "audio-volume-low", Action: playerAction(kind, "volume", "0.1-")},
	}
}

func playerForQuickControls(q string) (PlayerKind, string, bool) {
	switch {
	case strings.HasPrefix(q, "yp "):
		return PlayerYouTube, strings.TrimSpace(strings.TrimPrefix(q, "yp ")), true
	case strings.HasPrefix(q, "youtube player "):
		return PlayerYouTube, strings.TrimSpace(strings.TrimPrefix(q, "youtube player ")), true
	default:
		return PlayerSpotify, q, false
	}
}

func YouTubePlayerControls(query string) []Result {
	q := strings.ToLower(strings.TrimSpace(query))
	if q != "yp play" && q != "yp pause" && q != "yp next" && q != "yp prev" && q != "yp previous" {
		return nil
	}
	_, action, _ := playerForQuickControls(q)
	switch action {
	case "play", "pause":
		return []Result{{Type: TypeYouTubePlayer, Title: "YouTube Play/Pause", Icon: "media-playback-start", Action: playerAction(PlayerYouTube, "play-pause")}}
	case "next":
		return []Result{{Type: TypeYouTubePlayer, Title: "YouTube Next", Icon: "media-skip-forward", Action: playerAction(PlayerYouTube, "next")}}
	case "prev", "previous":
		return []Result{{Type: TypeYouTubePlayer, Title: "YouTube Previous", Icon: "media-skip-backward", Action: playerAction(PlayerYouTube, "previous")}}
	default:
		return nil
	}
}

func YouTubePlayerStatus(query string) []Result {
	if !MatchesAny(query, "yp status", "youtube player status") {
		return nil
	}
	metas := playerMetas()
	if len(metas) == 0 {
		return []Result{{
			Type:   TypeYouTubePlayer,
			Title:  "No MPRIS players",
			Desc:   "Open YouTube and run playerctl -l",
			Icon:   "dialog-warning",
			Action: func() {},
		}}
	}
	var results []Result
	selected := mediaPlayer(PlayerYouTube)
	for _, meta := range metas {
		m := meta
		title := m.name
		if m.name == selected {
			title = "Selected: " + title
		}
		metaTitle := m.title
		if isYouTubeMeta(m) && !isYouTubeVideoURL(m.url) {
			if video := browserCurrentYouTubeVideo(m.name); video.URL != "" {
				title := video.Title
				if title == "" {
					title = m.title
				}
				if title == "" {
					title = "YouTube video"
				}
				metaTitle = title + " | " + video.URL
			} else {
				metaTitle = "YouTube tab detected; browser did not expose current video URL"
			}
		}
		desc := strings.TrimSpace(string(m.status) + " | " + metaTitle)
		if m.url != "" {
			desc += " | " + m.url
		}
		if desc == "" {
			desc = "No title/url metadata"
		}
		results = append(results, Result{
			Type:  TypeYouTubePlayer,
			Title: title,
			Desc:  desc,
			Icon:  "media-playback-start",
			Action: func() {
				copyText("player=" + m.name + "\ntitle=" + m.title + "\nartist=" + m.artist + "\nurl=" + m.url)
			},
		})
	}
	return results
}

func cacheAlbumArt(url string) string {
	cacheDir := filepath.Join(os.Getenv("HOME"), ".cache", "spark", "art")
	os.MkdirAll(cacheDir, 0755)

	hash := simpleHash(url)
	cachePath := filepath.Join(cacheDir, hash+".jpg")

	if _, err := os.Stat(cachePath); err == nil {
		return cachePath
	}

	if strings.HasPrefix(url, "file://") {
		localPath := strings.TrimPrefix(url, "file://")
		return localPath
	}

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
	const hex = "0123456789abcdef"
	result := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		result[i] = hex[h&0xf]
		h >>= 4
	}
	return string(result)
}

func IsSpotifyQuery(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	return q == "sp" || q == "spotify" || strings.HasPrefix(q, "sp ")
}

func IsYouTubePlayerQuery(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	return q == "yp" || q == "youtube player" || strings.HasPrefix(q, "yp ") || strings.HasPrefix(q, "youtube player ")
}

var bareMediaCommands = map[string]struct {
	title string
	icon  string
	arg   string
}{
	"play":  {"Play/Pause", "media-playback-start", "play-pause"},
	"pause": {"Play/Pause", "media-playback-start", "play-pause"},
	"next":  {"Next Track", "media-skip-forward", "next"},
	"prev":  {"Previous Track", "media-skip-backward", "previous"},
}

func SpotifySearch(query string) []Result {
	cmd, ok := bareMediaCommands[normalizeQuery(query)]
	if !ok {
		return nil
	}
	return []Result{{
		Type:   TypeSpotify,
		Title:  cmd.title,
		Icon:   cmd.icon,
		Action: playerAction(PlayerSpotify, cmd.arg),
	}}
}
