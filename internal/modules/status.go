package modules

import (
	"strings"
	"sync"
)

var (
	lastStatus   string
	lastStatusOK bool
	statusMu     sync.Mutex
)

func SetStatus(ok bool, message string) {
	statusMu.Lock()
	defer statusMu.Unlock()
	lastStatusOK = ok
	lastStatus = message
}

func StatusSearch(query string) []Result {
	q := strings.ToLower(strings.TrimSpace(query))
	if q != "status" && q != "last" {
		return nil
	}
	statusMu.Lock()
	defer statusMu.Unlock()
	if lastStatus == "" {
		return []Result{{
			Type:   "status",
			Title:  "No Status Yet",
			Desc:   "Actions report here",
			Icon:   "dialog-information",
			Action: func() {},
		}}
	}
	icon := "dialog-error"
	title := "Last Error"
	if lastStatusOK {
		icon = "dialog-information"
		title = "Last Action"
	}
	return []Result{{
		Type:   "status",
		Title:  title,
		Desc:   lastStatus,
		Icon:   icon,
		Action: func() {},
	}}
}
