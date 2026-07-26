package modules

import (
	"github.com/tapiaw38/spark/internal/platform/commands"
)

func ScreenshotSearch(query string) []Result {
	if _, ok := MatchCommand(query, "screenshot", "ss"); !ok {
		return nil
	}
	if _, err := commands.LookPath("grim"); err != nil {
		return nil
	}

	return []Result{
		{
			Type:       TypeScreenshot,
			Title:      "Screenshot: Full screen",
			Desc:       "Save to ~/Pictures",
			Icon:       "camera-photo",
			ActionSpec: ScreenshotAction("full"),
		},
		{
			Type:       TypeScreenshot,
			Title:      "Screenshot: Select area",
			Desc:       "grim + slurp, save to ~/Pictures",
			Icon:       "camera-photo",
			ActionSpec: ScreenshotAction("area"),
		},
		{
			Type:       TypeScreenshot,
			Title:      "Screenshot: Area → clipboard",
			Desc:       "Copy region to clipboard",
			Icon:       "camera-photo",
			ActionSpec: ScreenshotAction("clip"),
		},
	}
}
