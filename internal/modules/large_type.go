package modules

import (
	"os"
	"os/exec"
	"strings"
)

// LargeTypeSearch opens text in a large overlay.
func LargeTypeSearch(query string) []Result {
	text, ok := largeTypeText(query)
	if !ok {
		return nil
	}
	if text == "" {
		return []Result{{
			Type:   "large-type",
			Title:  "Large Type",
			Desc:   "Type: large text",
			Icon:   "preferences-desktop-font",
			Action: func() {},
		}}
	}
	allMonitors := false
	if strings.HasPrefix(strings.ToLower(text), "all ") {
		allMonitors = true
		text = strings.TrimSpace(text[4:])
	}
	return []Result{
		{
			Type:  "large-type",
			Title: "Show Large Type",
			Desc:  text,
			Icon:  "preferences-desktop-font",
			Action: func() {
				if exe, err := os.Executable(); err == nil {
					args := []string{"--large-type", text}
					if allMonitors {
						args = []string{"--large-type-all", text}
					}
					exec.Command(exe, args...).Start()
				}
			},
		},
		{
			Type:  "large-type",
			Title: "Copy Text",
			Desc:  text,
			Icon:  "edit-copy",
			Action: func() {
				copyText(text)
			},
		},
	}
}

func largeTypeText(query string) (string, bool) {
	q := strings.TrimSpace(query)
	lower := strings.ToLower(q)
	for _, prefix := range []string{"large ", "type ", "lt "} {
		if strings.HasPrefix(lower, prefix) {
			return strings.TrimSpace(q[len(prefix):]), true
		}
	}
	return "", lower == "large" || lower == "type" || lower == "lt"
}
