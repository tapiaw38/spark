package modules

import (
	"os/exec"
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
		Type:  TypeShell,
		Title: "Run: " + cmd,
		Desc:  "Execute in terminal",
		Icon:  "utilities-terminal",
		Action: func() {
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
