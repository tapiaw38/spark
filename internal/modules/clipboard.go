package modules

import (
	"bytes"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// copyToClipboard writes text to the Wayland clipboard.
// ponytail: single point — add xclip/X11 fallback here if needed
func copyToClipboard(text string) {
	exec.Command("wl-copy", text).Run()
}

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
			Data:         clipID,
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

func GetClipboardPreviewImage(r Result) string {
	if r.Type != "clipboard" {
		return ""
	}
	if r.PreviewImage != "" {
		return expandHome(r.PreviewImage)
	}
	if r.Data == "" || !clipboardLooksImage(r.Title) {
		return ""
	}
	return cacheClipboardImage(r.Data)
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

func clipboardLooksImage(preview string) bool {
	lower := strings.ToLower(preview)
	return strings.Contains(lower, "image/") || strings.Contains(lower, "png") || strings.Contains(lower, "jpeg") || strings.Contains(lower, "jpg") || strings.Contains(lower, "webp")
}

func cacheClipboardImage(id string) string {
	decode := exec.Command("cliphist", "decode")
	decode.Stdin = strings.NewReader(id)
	data, err := decode.Output()
	if err != nil || len(data) < 12 {
		return ""
	}
	ext := ""
	switch {
	case bytes.HasPrefix(data, []byte{0x89, 'P', 'N', 'G'}):
		ext = ".png"
	case bytes.HasPrefix(data, []byte{0xff, 0xd8, 0xff}):
		ext = ".jpg"
	case bytes.HasPrefix(data, []byte("RIFF")) && len(data) > 12 && string(data[8:12]) == "WEBP":
		ext = ".webp"
	default:
		return ""
	}
	dir := cacheSubdir("clipboard")
	if dir == "" {
		return ""
	}
	path := filepath.Join(dir, simpleHash(id)+ext)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	if os.WriteFile(path, data, 0600) != nil {
		return ""
	}
	return path
}

func colorSwatch(color string) string {
	raw, err := hex.DecodeString(strings.TrimPrefix(color, "#"))
	if err != nil || len(raw) != 3 {
		return ""
	}
	dir := cacheSubdir("swatches")
	if dir == "" {
		return ""
	}
	path := filepath.Join(dir, strings.TrimPrefix(color, "#")+".ppm")
	if _, err := os.Stat(path); err == nil {
		return path
	}
	pixel := strconv.Itoa(int(raw[0])) + " " + strconv.Itoa(int(raw[1])) + " " + strconv.Itoa(int(raw[2])) + "\n"
	ppm := "P3\n64 64\n255\n" + strings.Repeat(pixel, 64*64)
	if os.WriteFile(path, []byte(ppm), 0644) != nil {
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
