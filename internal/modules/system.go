package modules

import (
	"os/exec"
	"strings"
)

var systemCommands = []struct {
	keywords []string
	name     string
	desc     string
	icon     string
	cmd      string
	args     []string
}{
	{[]string{"lock"}, "Lock Screen", "Lock the screen", "system-lock-screen", "swaylock", []string{"-f", "-c", "000000"}},
	{[]string{"sleep", "suspend"}, "Sleep", "Suspend the system", "system-suspend", "systemctl", []string{"suspend"}},
	{[]string{"restart", "reboot"}, "Restart", "Restart the system", "system-reboot", "systemctl", []string{"reboot"}},
	{[]string{"shutdown", "poweroff"}, "Shutdown", "Power off the system", "system-shutdown", "systemctl", []string{"poweroff"}},
	{[]string{"logout", "exit"}, "Logout", "Log out of session", "system-log-out", "loginctl", []string{"terminate-user", ""}},
	{[]string{"trash", "empty trash"}, "Empty Trash", "Empty the trash", "user-trash", "rm", []string{"-rf", "~/.local/share/Trash/files/*"}},
}

// SystemSearch finds matching system commands
func SystemSearch(query string) []Result {
	if len(query) < 2 {
		return nil
	}
	query = strings.ToLower(query)

	var results []Result
	for _, sc := range systemCommands {
		for _, kw := range sc.keywords {
			if strings.Contains(kw, query) {
				cmdName := sc.cmd
				cmdArgs := sc.args
				results = append(results, Result{
					Type:  "system",
					Title: sc.name,
					Desc:  sc.desc,
					Icon:  sc.icon,
					Action: func() {
						cmd := exec.Command(cmdName, cmdArgs...)
						cmd.Start()
					},
				})
				break
			}
		}
	}
	return results
}
