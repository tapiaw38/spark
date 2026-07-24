package modules

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ScreenshotSearch captures the screen via grim/slurp (Wayland).
func ScreenshotSearch(query string) []Result {
	q := strings.ToLower(strings.TrimSpace(query))
	if q != "screenshot" && q != "ss" && !strings.HasPrefix(q, "screenshot ") && !strings.HasPrefix(q, "ss ") {
		return nil
	}
	if _, err := exec.LookPath("grim"); err != nil {
		return nil
	}

	return []Result{
		{
			Type:   "screenshot",
			Title:  "Screenshot: Full screen",
			Desc:   "Save to ~/Pictures",
			Icon:   "camera-photo",
			Action: func() { grimTo(shotPath(), false) },
		},
		{
			Type:   "screenshot",
			Title:  "Screenshot: Select area",
			Desc:   "grim + slurp, save to ~/Pictures",
			Icon:   "camera-photo",
			Action: func() { grimTo(shotPath(), true) },
		},
		{
			Type:   "screenshot",
			Title:  "Screenshot: Area → clipboard",
			Desc:   "Copy region to clipboard",
			Icon:   "camera-photo",
			Action: grimClip,
		},
	}
}

func shotPath() string {
	dir := filepath.Join(os.Getenv("HOME"), "Pictures")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "Screenshot-"+strconv.FormatInt(time.Now().Unix(), 10)+".png")
}

func grimTo(path string, area bool) {
	if area {
		exec.Command("sh", "-c", "grim -g \"$(slurp)\" "+shellQuote(path)).Start()
		return
	}
	exec.Command("grim", path).Start()
}

func grimClip() {
	exec.Command("sh", "-c", "grim -g \"$(slurp)\" - | wl-copy").Start()
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
