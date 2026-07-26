package modules

import (
	"os"
	"strings"
)

func LargeTypeSearch(query string) []Result {
	text, ok := largeTypeText(query)
	if !ok {
		return nil
	}
	if text == "" {
		return []Result{{
			Type:  TypeLargeType,
			Title: "Large Type",
			Desc:  "Type: large text",
			Icon:  "preferences-desktop-font",
		}}
	}
	allMonitors := false
	if strings.HasPrefix(strings.ToLower(text), "all ") {
		allMonitors = true
		text = strings.TrimSpace(text[4:])
	}
	largeTypeAction := ActionSpec{}
	if exe, err := os.Executable(); err == nil {
		args := []string{"--large-type", text}
		if allMonitors {
			args = []string{"--large-type-all", text}
		}
		largeTypeAction = StartAction(exe, args...)
	}
	return []Result{
		{
			Type:       TypeLargeType,
			Title:      "Show Large Type",
			Desc:       text,
			Icon:       "preferences-desktop-font",
			ActionSpec: largeTypeAction,
		},
		{
			Type:       TypeLargeType,
			Title:      "Copy Text",
			Desc:       text,
			Icon:       "edit-copy",
			ActionSpec: CopyAction(text),
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
