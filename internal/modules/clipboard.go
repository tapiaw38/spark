package modules

import (
	"bytes"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ClipboardSearch searches clipboard history
// ponytail: relies on cliphist (common Wayland clipboard manager)
// Install: yay -S cliphist
// Setup: wl-paste --watch cliphist store (in autostart)
func ClipboardSearch(query string) []Result {
	lowerQuery := strings.ToLower(query)
	directPaste := strings.HasPrefix(lowerQuery, "paste")
	if !strings.HasPrefix(lowerQuery, "clip") && !strings.HasPrefix(lowerQuery, "cb") && !directPaste {
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
	} else if strings.HasPrefix(lowerQuery, "paste ") {
		searchTerm = query[len("paste "):]
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
		icon, desc := clipboardDisplay(preview, directPaste)
		previewImage := clipboardPreviewImage(preview)
		results = append(results, Result{
			Type:         "clipboard",
			Title:        preview,
			Desc:         desc,
			Icon:         icon,
			PreviewImage: previewImage,
			Action: func() {
				// Decode and copy to clipboard
				decode := exec.Command("cliphist", "decode")
				decode.Stdin = strings.NewReader(clipID)
				decoded, _ := decode.Output()

				copy := exec.Command("wl-copy")
				copy.Stdin = bytes.NewReader(decoded)
				copy.Run()
				if directPaste {
					pasteClipboard()
				}
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

func clipboardPreviewImage(preview string) string {
	if path := clipboardImagePath(preview); path != "" {
		return path
	}
	color := strings.TrimSpace(preview)
	if len(color) == 7 && strings.HasPrefix(color, "#") {
		if path := colorSwatch(color); path != "" {
			return path
		}
	}
	return ""
}

func clipboardImagePath(preview string) string {
	for _, field := range strings.Fields(preview) {
		field = strings.TrimPrefix(field, "file://")
		field = strings.Trim(field, "'\"")
		lower := strings.ToLower(field)
		if !(strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") || strings.HasSuffix(lower, ".webp")) {
			continue
		}
		if strings.HasPrefix(field, "~") {
			field = expandHome(field)
		}
		if _, err := os.Stat(field); err == nil {
			return field
		}
	}
	return ""
}

func colorSwatch(color string) string {
	raw, err := hex.DecodeString(strings.TrimPrefix(color, "#"))
	if err != nil || len(raw) != 3 {
		return ""
	}
	dir := filepath.Join(os.Getenv("HOME"), ".cache", "spark", "swatches")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ""
	}
	path := filepath.Join(dir, strings.TrimPrefix(color, "#")+".ppm")
	if _, err := os.Stat(path); err == nil {
		return path
	}
	var b strings.Builder
	b.WriteString("P3\n64 64\n255\n")
	for i := 0; i < 64*64; i++ {
		b.WriteString(stringInt(int(raw[0])) + " " + stringInt(int(raw[1])) + " " + stringInt(int(raw[2])) + "\n")
	}
	if os.WriteFile(path, []byte(b.String()), 0644) != nil {
		return ""
	}
	return path
}

func clipboardDisplay(preview string, directPaste bool) (string, string) {
	action := "Copy"
	if directPaste {
		action = "Paste"
	}
	lower := strings.ToLower(preview)
	switch {
	case strings.Contains(lower, "image/") || strings.Contains(lower, ".png") || strings.Contains(lower, ".jpg") || strings.Contains(lower, ".jpeg") || strings.Contains(lower, ".webp"):
		return "image-x-generic", action + " image from clipboard history"
	case strings.Contains(lower, "file://") || strings.HasPrefix(lower, "/") || strings.HasPrefix(lower, "~"):
		return "text-x-generic", action + " file/path from clipboard history"
	case strings.HasPrefix(strings.TrimSpace(lower), "#") && len(strings.TrimSpace(lower)) == 7:
		return "applications-graphics", action + " color from clipboard history"
	default:
		return "edit-paste", action + " from clipboard history"
	}
}

func pasteClipboard() {
	if _, err := exec.LookPath("wtype"); err == nil {
		exec.Command("wtype", "-M", "ctrl", "v", "-m", "ctrl").Start()
		return
	}
	if _, err := exec.LookPath("ydotool"); err == nil {
		exec.Command("ydotool", "key", "29:1", "47:1", "47:0", "29:0").Start()
	}
}
