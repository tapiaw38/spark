package modules

import (
	"bytes"
	"os/exec"
	"strings"
)

// ClipboardSearch searches clipboard history
// ponytail: relies on cliphist (common Wayland clipboard manager)
// Install: yay -S cliphist
// Setup: wl-paste --watch cliphist store (in autostart)
func ClipboardSearch(query string) []Result {
	lowerQuery := strings.ToLower(query)
	if !strings.HasPrefix(lowerQuery, "clip") && !strings.HasPrefix(lowerQuery, "cb") {
		return nil
	}

	// Get search term after prefix
	searchTerm := ""
	if strings.HasPrefix(lowerQuery, "clipboard ") {
		searchTerm = query[len("clipboard "):]
	} else if strings.HasPrefix(lowerQuery, "clip ") {
		searchTerm = query[len("clip "):]
	} else if strings.HasPrefix(lowerQuery, "cb ") {
		searchTerm = query[len("cb "):]
	}

	// Check if cliphist is available
	if _, err := exec.LookPath("cliphist"); err != nil {
		return []Result{{
			Type:   "clipboard",
			Title:  "Clipboard: cliphist not installed",
			Desc:   "Install with: yay -S cliphist",
			Icon:   "edit-paste",
			Action: func() {},
		}}
	}

	// Get clipboard history
	cmd := exec.Command("cliphist", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var results []Result

	for _, line := range lines {
		if line == "" {
			continue
		}

		// cliphist format: "id\tpreview"
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) < 2 {
			continue
		}

		id := parts[0]
		preview := parts[1]

		// Filter by search term
		if searchTerm != "" && !strings.Contains(strings.ToLower(preview), strings.ToLower(searchTerm)) {
			continue
		}

		// Truncate preview
		if len(preview) > 50 {
			preview = preview[:50] + "..."
		}

		clipID := id // capture
		icon, desc := clipboardDisplay(preview)
		results = append(results, Result{
			Type:  "clipboard",
			Title: preview,
			Desc:  desc,
			Icon:  icon,
			Action: func() {
				// Decode and copy to clipboard
				decode := exec.Command("cliphist", "decode")
				decode.Stdin = strings.NewReader(clipID)
				decoded, _ := decode.Output()

				copy := exec.Command("wl-copy")
				copy.Stdin = bytes.NewReader(decoded)
				copy.Run()
			},
		})

		if len(results) >= 50 {
			break
		}
	}

	if len(results) == 0 && (query == "clip" || query == "cb" || query == "clipboard") {
		return []Result{{
			Type:   "clipboard",
			Title:  "Clipboard History",
			Desc:   "Type 'clip <search>' to filter",
			Icon:   "edit-paste",
			Action: func() {},
		}}
	}

	return results
}

func clipboardDisplay(preview string) (string, string) {
	lower := strings.ToLower(preview)
	switch {
	case strings.Contains(lower, "image/") || strings.Contains(lower, ".png") || strings.Contains(lower, ".jpg") || strings.Contains(lower, ".jpeg") || strings.Contains(lower, ".webp"):
		return "image-x-generic", "Copy image from clipboard history"
	case strings.Contains(lower, "file://") || strings.HasPrefix(lower, "/") || strings.HasPrefix(lower, "~"):
		return "text-x-generic", "Copy file/path from clipboard history"
	case strings.HasPrefix(strings.TrimSpace(lower), "#") && len(strings.TrimSpace(lower)) == 7:
		return "applications-graphics", "Copy color from clipboard history"
	default:
		return "edit-paste", "Copy from clipboard history"
	}
}
