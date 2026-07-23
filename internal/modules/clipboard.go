package modules

import (
	"os/exec"
	"strings"
)

// ClipboardSearch searches clipboard history
// ponytail: relies on cliphist (common Wayland clipboard manager)
// Install: yay -S cliphist
// Setup: wl-paste --watch cliphist store (in autostart)
func ClipboardSearch(query string) []Result {
	if !strings.HasPrefix(query, "clip") && !strings.HasPrefix(query, "cb") {
		return nil
	}

	// Get search term after prefix
	searchTerm := ""
	if strings.HasPrefix(query, "clipboard ") {
		searchTerm = strings.TrimPrefix(query, "clipboard ")
	} else if strings.HasPrefix(query, "clip ") {
		searchTerm = strings.TrimPrefix(query, "clip ")
	} else if strings.HasPrefix(query, "cb ") {
		searchTerm = strings.TrimPrefix(query, "cb ")
	}

	// Check if cliphist is available
	if _, err := exec.LookPath("cliphist"); err != nil {
		return []Result{{
			Type:  "clipboard",
			Title: "Clipboard: cliphist not installed",
			Desc:  "Install with: yay -S cliphist",
			Icon:  "edit-paste",
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
		results = append(results, Result{
			Type:  "clipboard",
			Title: preview,
			Desc:  "Paste from clipboard history",
			Icon:  "edit-paste",
			Action: func() {
				// Decode and copy to clipboard
				decode := exec.Command("cliphist", "decode")
				decode.Stdin = strings.NewReader(clipID)
				decoded, _ := decode.Output()

				copy := exec.Command("wl-copy")
				copy.Stdin = strings.NewReader(string(decoded))
				copy.Run()
			},
		})

		if len(results) >= 5 {
			break
		}
	}

	if len(results) == 0 && (query == "clip" || query == "cb" || query == "clipboard") {
		return []Result{{
			Type:  "clipboard",
			Title: "Clipboard History",
			Desc:  "Type 'clip <search>' to filter",
			Icon:  "edit-paste",
			Action: func() {},
		}}
	}

	return results
}
