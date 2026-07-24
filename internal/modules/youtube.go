package modules

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const maxYouTubeResults = 12

type youtubeVideo struct {
	ID         string             `json:"id"`
	Title      string             `json:"title"`
	Uploader   string             `json:"uploader"`
	Channel    string             `json:"channel"`
	Duration   int                `json:"duration"`
	ViewCount  int64              `json:"view_count"`
	WebpageURL string             `json:"webpage_url"`
	Thumbnail  string             `json:"thumbnail"`
	Thumbnails []youtubeThumbnail `json:"thumbnails"`
}

type youtubeThumbnail struct {
	URL string `json:"url"`
}

// IsYouTubeQuery returns true if query triggers YouTube result mode.
func IsYouTubeQuery(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	return q == "yt" || q == "youtube" || strings.HasPrefix(q, "yt ") || strings.HasPrefix(q, "youtube ")
}

// YouTubeSearch returns video results with cached thumbnails.
func YouTubeSearch(query string) []Result {
	searchQuery := strings.TrimSpace(query)
	for _, prefix := range []string{"youtube ", "yt "} {
		if strings.HasPrefix(strings.ToLower(searchQuery), prefix) {
			searchQuery = strings.TrimSpace(searchQuery[len(prefix):])
			break
		}
	}
	if searchQuery == "" {
		return []Result{youtubeFallback("YouTube", "Type: yt search terms", "https://www.youtube.com")}
	}

	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return []Result{youtubeFallback("YouTube: "+searchQuery, "Install yt-dlp for video previews", youtubeSearchURL(searchQuery))}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "yt-dlp", "--ignore-config", "--no-warnings", "--dump-json", "--flat-playlist", fmt.Sprintf("ytsearch%d:%s", maxYouTubeResults, searchQuery))
	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		return []Result{youtubeFallback("YouTube: "+searchQuery, "Open search in browser", youtubeSearchURL(searchQuery))}
	}

	videos := parseYouTubeVideos(out)
	if len(videos) == 0 {
		return []Result{youtubeFallback("YouTube: "+searchQuery, "Open search in browser", youtubeSearchURL(searchQuery))}
	}

	results := make([]Result, 0, len(videos))
	for _, video := range videos {
		v := video
		videoURL := v.WebpageURL
		if videoURL == "" && v.ID != "" {
			videoURL = "https://www.youtube.com/watch?v=" + v.ID
		}

		desc := v.Channel
		if desc == "" {
			desc = v.Uploader
		}
		if duration := formatDuration(v.Duration); duration != "" {
			if desc != "" {
				desc += " · "
			}
			desc += duration
		}

		results = append(results, Result{
			Type:            "youtube",
			Title:           v.Title,
			Desc:            desc,
			Icon:            "youtube",
			Preview:         videoURL,
			PreviewImageURL: youtubeThumbnailURL(v),
			Action: func() {
				exec.Command("xdg-open", videoURL).Start()
			},
		})
	}

	return results
}

func YouTubeLoading(query string) []Result {
	searchQuery := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(query, "yt "), "youtube "))
	return []Result{{
		Type:  "youtube",
		Title: "YouTube: " + searchQuery,
		Desc:  "Searching videos...",
		Icon:  "youtube",
	}}
}

func parseYouTubeVideos(out []byte) []youtubeVideo {
	var single struct {
		Entries []youtubeVideo `json:"entries"`
	}
	if err := json.Unmarshal(out, &single); err == nil && len(single.Entries) > 0 {
		return cleanYouTubeVideos(single.Entries)
	}

	var videos []youtubeVideo
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var video youtubeVideo
		if err := json.Unmarshal([]byte(line), &video); err == nil {
			videos = append(videos, video)
		}
	}
	return cleanYouTubeVideos(videos)
}

func cleanYouTubeVideos(videos []youtubeVideo) []youtubeVideo {
	clean := make([]youtubeVideo, 0, len(videos))
	for _, video := range videos {
		if video.Title == "" {
			continue
		}
		clean = append(clean, video)
	}
	return clean
}

func youtubeThumbnailURL(video youtubeVideo) string {
	thumbnailURL := video.Thumbnail
	if thumbnailURL == "" && len(video.Thumbnails) > 0 {
		thumbnailURL = video.Thumbnails[len(video.Thumbnails)-1].URL
	}
	if thumbnailURL == "" && video.ID != "" {
		thumbnailURL = "https://i.ytimg.com/vi/" + video.ID + "/hqdefault.jpg"
	}
	return thumbnailURL
}

func CacheYouTubeThumbnail(thumbnailURL string) string {
	if thumbnailURL == "" {
		return ""
	}

	cacheDir := cacheSubdir("youtube")
	if cacheDir == "" {
		return ""
	}

	name := simpleHash(thumbnailURL)
	cachePath := filepath.Join(cacheDir, name+".jpg")
	if _, err := os.Stat(cachePath); err == nil {
		return cachePath
	}

	client := http.Client{Timeout: 4 * time.Second}
	resp, err := client.Get(thumbnailURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ""
	}

	f, err := os.Create(cachePath)
	if err != nil {
		return ""
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return ""
	}
	return cachePath
}

func youtubeFallback(title, desc, link string) Result {
	return Result{
		Type:  "web",
		Title: title,
		Desc:  desc,
		Icon:  "youtube",
		Action: func() {
			exec.Command("xdg-open", link).Start()
		},
	}
}

func youtubeSearchURL(query string) string {
	return "https://www.youtube.com/results?search_query=" + url.QueryEscape(query)
}

func formatDuration(seconds int) string {
	if seconds <= 0 {
		return ""
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}
