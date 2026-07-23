package modules

import (
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

type systemCommand struct {
	keywords []string
	name     string
	desc     string
	icon     string
	confirm  bool
	enabled  func() bool
	reason   string
	action   func()
}

// SystemSearch finds matching system commands.
func SystemSearch(query string) []Result {
	if len(query) < 2 {
		return nil
	}
	query = strings.ToLower(strings.TrimSpace(query))

	var results []Result
	for _, sc := range systemCommands() {
		for _, kw := range sc.keywords {
			if strings.Contains(kw, query) || strings.Contains(strings.ToLower(sc.name), query) {
				cmd := sc
				if cmd.enabled != nil && !cmd.enabled() {
					results = append(results, Result{
						Type:   "system",
						Title:  cmd.name + " unavailable",
						Desc:   cmd.reason,
						Icon:   "dialog-warning",
						Action: func() {},
					})
					break
				}
				results = append(results, Result{
					Type:    "system",
					Title:   cmd.name,
					Desc:    cmd.desc,
					Icon:    cmd.icon,
					Confirm: cmd.confirm,
					Action:  cmd.action,
				})
				break
			}
		}
	}
	return results
}

func systemCommands() []systemCommand {
	return []systemCommand{
		{[]string{"lock", "screensaver", "screen saver"}, "Lock Screen", "Lock session", "system-lock-screen", false, hasLocker, "Install swaylock, hyprlock, gtklock, or loginctl", lockScreen},
		{[]string{"sleep", "suspend"}, "Sleep", "Suspend system", "system-suspend", true, hasSystemctl, "systemctl not available", func() { exec.Command("systemctl", "suspend").Start() }},
		{[]string{"hibernate"}, "Hibernate", "Hibernate system", "system-suspend-hibernate", true, hasSystemctl, "systemctl not available", func() { exec.Command("systemctl", "hibernate").Start() }},
		{[]string{"restart", "reboot"}, "Restart", "Restart system", "system-reboot", true, hasSystemctl, "systemctl not available", func() { exec.Command("systemctl", "reboot").Start() }},
		{[]string{"shutdown", "poweroff", "power off"}, "Shutdown", "Power off system", "system-shutdown", true, hasSystemctl, "systemctl not available", func() { exec.Command("systemctl", "poweroff").Start() }},
		{[]string{"logout", "log out", "exit session"}, "Logout", "Terminate current user session", "system-log-out", true, hasLoginctl, "loginctl not available", logout},
		{[]string{"trash", "empty trash", "clear trash"}, "Empty Trash", "Delete files from user trash", "user-trash", true, hasTrashBackend, "Install gio or kioclient6", emptyTrash},
	}
}

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func hasSystemctl() bool { return hasCommand("systemctl") }
func hasLoginctl() bool  { return hasCommand("loginctl") }

func hasLocker() bool {
	return hasCommand("swaylock") || hasCommand("hyprlock") || hasCommand("gtklock") || hasCommand("loginctl")
}

func hasTrashBackend() bool {
	return hasCommand("gio") || hasCommand("kioclient6")
}

func lockScreen() {
	for _, cmd := range [][]string{
		{"swaylock", "-f", "-c", "000000"},
		{"hyprlock"},
		{"gtklock"},
		{"loginctl", "lock-session"},
	} {
		if _, err := exec.LookPath(cmd[0]); err == nil {
			exec.Command(cmd[0], cmd[1:]...).Start()
			return
		}
	}
}

func logout() {
	if u, err := user.Current(); err == nil && u.Username != "" {
		exec.Command("loginctl", "terminate-user", u.Username).Start()
		return
	}
	exec.Command("loginctl", "terminate-session", os.Getenv("XDG_SESSION_ID")).Start()
}

func emptyTrash() {
	for _, dir := range []string{
		filepath.Join(os.Getenv("HOME"), ".local/share/Trash/files"),
		filepath.Join(os.Getenv("HOME"), ".local/share/Trash/info"),
	} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			os.RemoveAll(filepath.Join(dir, entry.Name()))
		}
	}
}
