package modules

import (
	"strings"
)

func ShellSearch(query string) []Result {
	if !strings.HasPrefix(query, ">") {
		return nil
	}

	cmd := strings.TrimSpace(strings.TrimPrefix(query, ">"))
	if cmd == "" {
		return nil
	}

	return []Result{{
		Type:       TypeShell,
		Title:      "Run: " + cmd,
		Desc:       "Execute in terminal",
		Icon:       "utilities-terminal",
		ActionSpec: TerminalAction(cmd),
	}}
}
