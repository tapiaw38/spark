package modules

import (
	"net/url"
	"os/exec"
	"strings"
)

// WeatherSearch shows current weather from wttr.in.
func WeatherSearch(query string) []Result {
	lower := strings.ToLower(strings.TrimSpace(query))
	if lower != "weather" && !strings.HasPrefix(lower, "weather ") {
		return nil
	}
	city := strings.TrimSpace(query[len("weather"):])

	title := "Weather"
	if city != "" {
		title = "Weather: " + city
	}
	page := "https://wttr.in/" + url.QueryEscape(city)

	return []Result{{
		Type:  "weather",
		Title: title,
		Desc:  "Notify current conditions (needs curl), else open in browser",
		Icon:  "weather-clear",
		Action: func() {
			if _, err := exec.LookPath("curl"); err == nil {
				exec.Command("sh", "-c", "notify-send Weather \"$(curl -s '"+page+"?format=3')\"").Start()
				return
			}
			exec.Command("xdg-open", page).Start()
		},
	}}
}
