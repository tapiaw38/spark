package modules

import (
	"strconv"
	"strings"
	"time"
)

func TimerSearch(query string) []Result {
	lower := strings.ToLower(strings.TrimSpace(query))
	if !strings.HasPrefix(lower, "timer ") {
		return nil
	}
	arg := strings.TrimSpace(lower[len("timer "):])
	d, err := time.ParseDuration(arg)
	if err != nil || d <= 0 {
		return nil
	}
	secs := strconv.Itoa(int(d.Seconds()))

	return []Result{{
		Type:  TypeTimer,
		Title: "Timer: " + arg,
		Desc:  "Notify when elapsed",
		Icon:  "alarm-symbolic",
		Action: func() {
			Start("sh", "-c", "sleep "+secs+" && notify-send 'Spark Timer' '"+arg+" elapsed'")
			SetStatus(true, "Timer set: "+arg)
		},
	}}
}
