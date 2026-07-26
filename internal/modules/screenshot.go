package modules

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tapiaw38/spark/internal/config"
)

func ScreenshotSearch(query string) []Result {
	if _, ok := MatchCommand(query, "screenshot", "ss"); !ok {
		return nil
	}
	if _, err := exec.LookPath("grim"); err != nil {
		return nil
	}

	return []Result{
		{
			Type:   TypeScreenshot,
			Title:  "Screenshot: Full screen",
			Desc:   "Save to ~/Pictures",
			Icon:   "camera-photo",
			Action: func() { grimTo(shotPath(), false) },
		},
		{
			Type:   TypeScreenshot,
			Title:  "Screenshot: Select area",
			Desc:   "grim + slurp, save to ~/Pictures",
			Icon:   "camera-photo",
			Action: func() { grimTo(shotPath(), true) },
		},
		{
			Type:   TypeScreenshot,
			Title:  "Screenshot: Area → clipboard",
			Desc:   "Copy region to clipboard",
			Icon:   "camera-photo",
			Action: grimClip,
		},
	}
}

func shotPath() string {
	dir := config.HomeFile("Pictures")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "Screenshot-"+strconv.FormatInt(time.Now().Unix(), 10)+".png")
}

func grimTo(path string, area bool) {
	if area {
		Start("sh", "-c", "grim -g \"$(slurp)\" "+shellQuote(path))
		return
	}
	Start("grim", path)
}

func grimClip() {
	Start("sh", "-c", "grim -g \"$(slurp)\" - | wl-copy")
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
