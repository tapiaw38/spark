package modules

import (
	"sort"
	"strings"

	"github.com/tapiaw38/spark/internal/history"
)

// StatsSearch shows app usage stats.
func StatsSearch(query string) []Result {
	q := strings.ToLower(strings.TrimSpace(query))
	if q != "stats" && q != "usage" {
		return nil
	}

	counts := history.Snapshot()
	if len(counts) == 0 {
		return []Result{{
			Type:   "stats",
			Title:  "No Usage Stats Yet",
			Desc:   "Launch apps to build stats",
			Icon:   "utilities-system-monitor",
			Action: func() {},
		}}
	}

	type stat struct {
		name  string
		count int
	}
	stats := make([]stat, 0, len(counts))
	total := 0
	for name, count := range counts {
		stats = append(stats, stat{name: name, count: count})
		total += count
	}
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].count > stats[j].count
	})

	results := []Result{{
		Type:   "stats",
		Title:  "Total App Launches",
		Desc:   stringInt(total),
		Icon:   "utilities-system-monitor",
		Action: func() {},
	}}
	for _, stat := range stats {
		results = append(results, Result{
			Type:   "stats",
			Title:  stat.name,
			Desc:   statBar(stat.count, stats[0].count) + " " + stringInt(stat.count) + " launches",
			Icon:   "utilities-system-monitor",
			Action: func() {},
		})
		if len(results) >= 12 {
			break
		}
	}
	return results
}

func statBar(count, max int) string {
	if max <= 0 {
		return ""
	}
	width := count * 12 / max
	if width < 1 {
		width = 1
	}
	return strings.Repeat("#", width)
}

func stringInt(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
