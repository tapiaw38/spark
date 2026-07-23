package apps

import (
	"os"
	"strings"
	"testing"
)

func TestApplicationDirsUsesXDGDataDirs(t *testing.T) {
	t.Setenv("HOME", "/home/test")
	t.Setenv("XDG_DATA_DIRS", "/var/lib/flatpak/exports/share:/usr/share")

	dirs := applicationDirs()
	want := "/var/lib/flatpak/exports/share/applications"
	for _, dir := range dirs {
		if dir == want {
			return
		}
	}
	t.Fatalf("missing %s in %v", want, dirs)
}

func TestCleanDesktopExecRemovesDesktopFieldCodes(t *testing.T) {
	execCmd := "/usr/bin/flatpak run --branch=stable --arch=x86_64 --command=com.slack.Slack --file-forwarding com.slack.Slack @@u %U @@"
	got := cleanDesktopExec(execCmd)
	want := "/usr/bin/flatpak run --branch=stable --arch=x86_64 --command=com.slack.Slack --file-forwarding com.slack.Slack"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestLoadFindsFlatpakSlackWhenInstalled(t *testing.T) {
	if _, err := os.Stat("/var/lib/flatpak/exports/share/applications/com.slack.Slack.desktop"); err != nil {
		t.Skip("Slack flatpak desktop file not installed")
	}

	apps := Load()
	for _, app := range apps {
		if app.Name == "Slack" {
			if !strings.Contains(app.Exec, "com.slack.Slack") {
				t.Fatalf("Slack Exec not cleaned: %q", app.Exec)
			}
			return
		}
	}
	t.Fatalf("Slack not loaded from XDG application dirs")
}
