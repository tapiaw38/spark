package modules

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func WeatherSearch(query string) []Result {
	if !IsWeatherQuery(query) {
		return nil
	}
	return weatherResult(query, "", false)
}

func IsWeatherQuery(query string) bool {
	lower := strings.ToLower(strings.TrimSpace(query))
	return lower == "weather" || strings.HasPrefix(lower, "weather ") || lower == "wt" || strings.HasPrefix(lower, "wt ")
}

func WeatherLoading(query string) []Result {
	if !IsWeatherQuery(query) {
		return nil
	}
	return weatherResult(query, "Loading weather...", false)
}

func WeatherSearchAsync(query string) []Result {
	if !IsWeatherQuery(query) {
		return nil
	}
	page := weatherPage(query)
	text, err := currentWeather(page)
	if err != nil {
		return weatherResult(query, "Weather unavailable; press Enter to open wttr.in", false)
	}
	return weatherResult(query, text, true)
}

func weatherResult(query, desc string, copyOnEnter bool) []Result {
	city := weatherCity(query)
	displayCity := titleCity(city)

	title := "Weather"
	if city != "" {
		title = "Weather: " + displayCity
	}
	page := weatherPage(query)
	if desc == "" {
		desc = "Show current conditions"
	}

	return []Result{{
		Type:       TypeWeather,
		Title:      title,
		Desc:       desc,
		Icon:       "weather-clear",
		ActionSpec: WeatherAction(page, desc, copyOnEnter),
	}}
}

func weatherPage(query string) string {
	city := weatherCity(query)
	return "https://wttr.in/" + url.PathEscape(city)
}

func weatherCity(query string) string {
	q := strings.TrimSpace(query)
	lower := strings.ToLower(q)
	switch {
	case lower == "wt":
		return ""
	case strings.HasPrefix(lower, "wt "):
		return strings.TrimSpace(q[2:])
	case lower == "weather":
		return ""
	case strings.HasPrefix(lower, "weather "):
		return strings.TrimSpace(q[len("weather"):])
	default:
		return ""
	}
}

func titleCity(city string) string {
	parts := strings.Fields(strings.ToLower(strings.TrimSpace(city)))
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func currentWeather(page string) (string, error) {
	client := http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(page + "?format=3")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("weather status %d", resp.StatusCode)
	}
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	text := strings.TrimSpace(string(out))
	if text == "" {
		return "", fmt.Errorf("empty weather response")
	}
	return text, nil
}
