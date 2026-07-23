package modules

import (
	"os/exec"
	"strings"
)

// ShellSearch handles "> command" prefix for running shell commands
func ShellSearch(query string) []Result {
	if !strings.HasPrefix(query, ">") {
		return nil
	}

	cmd := strings.TrimSpace(strings.TrimPrefix(query, ">"))
	if cmd == "" {
		return nil
	}

	return []Result{{
		Type:  "shell",
		Title: "Run: " + cmd,
		Desc:  "Execute in terminal",
		Icon:  "utilities-terminal",
		Action: func() {
			// ponytail: ghostty from user's config, fallback to common terminals
			terminals := []string{"ghostty", "alacritty", "kitty", "foot", "gnome-terminal"}
			for _, term := range terminals {
				if _, err := exec.LookPath(term); err == nil {
					var c *exec.Cmd
					switch term {
					case "gnome-terminal":
						c = exec.Command(term, "--", "sh", "-c", cmd)
					default:
						c = exec.Command(term, "-e", "sh", "-c", cmd)
					}
					c.Start()
					return
				}
			}
		},
	}}
}
