package modules

import (
	"github.com/tapiaw38/spark/internal/platform/commands"
	"os"
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
	spec     ActionSpec
}

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
						Type:  TypeSystem,
						Title: cmd.name + " unavailable",
						Desc:  cmd.reason,
						Icon:  "dialog-warning",
					})
					break
				}
				results = append(results, Result{
					Type:       TypeSystem,
					Title:      cmd.name,
					Desc:       cmd.desc,
					Icon:       cmd.icon,
					Confirm:    cmd.confirm,
					ActionSpec: cmd.spec,
				})
				break
			}
		}
	}
	return results
}

func systemCommands() []systemCommand {
	return []systemCommand{
		{keywords: []string{"lock", "screensaver", "screen saver"}, name: "Lock Screen", desc: "Lock session", icon: "system-lock-screen", enabled: hasLocker, reason: lockerReason(), spec: SystemAction("lock")},
		{keywords: []string{"sleep", "suspend"}, name: "Sleep", desc: "Suspend system", icon: "system-suspend", confirm: true, enabled: hasSystemctl, reason: systemctlReason(), spec: StartAction("systemctl", "suspend")},
		{keywords: []string{"hibernate"}, name: "Hibernate", desc: "Hibernate system", icon: "system-suspend-hibernate", confirm: true, enabled: hasSystemctl, reason: systemctlReason(), spec: StartAction("systemctl", "hibernate")},
		{keywords: []string{"restart", "reboot"}, name: "Restart", desc: "Restart system", icon: "system-reboot", confirm: true, enabled: hasSystemctl, reason: systemctlReason(), spec: StartAction("systemctl", "reboot")},
		{keywords: []string{"shutdown", "poweroff", "power off"}, name: "Shutdown", desc: "Power off system", icon: "system-shutdown", confirm: true, enabled: hasSystemctl, reason: systemctlReason(), spec: StartAction("systemctl", "poweroff")},
		{keywords: []string{"logout", "log out", "exit session"}, name: "Logout", desc: "Terminate current user session", icon: "system-log-out", confirm: true, enabled: hasLoginctl, reason: loginctlReason(), spec: SystemAction("logout")},
		{keywords: []string{"trash", "empty trash", "clear trash"}, name: "Empty Trash", desc: "Delete files from user trash", icon: "user-trash", confirm: true, enabled: hasTrashBackend, reason: "Install gio or kioclient6; session " + sessionSummary(), spec: SystemAction("empty-trash")},
	}
}

func hasCommand(name string) bool {
	_, err := commands.LookPath(name)
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

func sessionSummary() string {
	parts := []string{}
	for _, key := range []string{"XDG_SESSION_TYPE", "XDG_CURRENT_DESKTOP", "DESKTOP_SESSION", "WAYLAND_DISPLAY", "DISPLAY"} {
		if value := os.Getenv(key); value != "" {
			parts = append(parts, key+"="+value)
		}
	}
	if len(parts) == 0 {
		return "unknown"
	}
	return strings.Join(parts, " ")
}

func lockerReason() string {
	return "Install swaylock, hyprlock, gtklock, or loginctl; session " + sessionSummary()
}

func systemctlReason() string {
	return "systemctl not available; session " + sessionSummary()
}

func loginctlReason() string {
	return "loginctl not available; XDG_SESSION_ID=" + os.Getenv("XDG_SESSION_ID")
}
