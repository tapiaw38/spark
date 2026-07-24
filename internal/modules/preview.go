package modules

import (
	"archive/zip"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// GetPreview returns preview content for a result
func GetPreview(r Result) string {
	// Use explicit preview if set
	if r.Preview != "" {
		return r.Preview
	}

	switch r.Type {
	case "file":
		return previewFile(r.Title, r.Desc)
	case "snippet":
		// Already shown in Desc
		return ""
	case "clipboard":
		return r.Title // Full clipboard content
	case "dictionary":
		return r.Desc
	case "calc":
		return r.Title
	default:
		return ""
	}
}

// GetPreviewImage returns a local image path for visual previews.
func GetPreviewImage(r Result) string {
	return GetPreviewImageAt(r, 1, 360)
}

func GetPreviewImageAt(r Result, page, scale int) string {
	if r.PreviewImage != "" {
		return expandHome(r.PreviewImage)
	}

	if !isImageFile(r.Title) {
		if strings.EqualFold(filepath.Ext(r.Title), ".pdf") {
			return previewPDFImageAt(GetFilePath(r), page, scale)
		}
		return ""
	}

	return GetFilePath(r)
}

// GetFilePath returns the local path represented by a file result.
func GetFilePath(r Result) string {
	if r.Type != "file" && r.Type != "directory" {
		return ""
	}
	if r.Type == "directory" && strings.HasPrefix(r.NavigateQuery, "nav ") {
		return expandHome(strings.TrimSpace(strings.TrimPrefix(r.NavigateQuery, "nav ")))
	}
	return filepath.Join(expandHome(r.Desc), r.Title)
}

func previewFile(name, dir string) string {
	path := filepath.Join(expandHome(dir), name)

	ext := strings.ToLower(filepath.Ext(name))

	// Text files - show first lines
	switch ext {
	case ".txt", ".md", ".go", ".py", ".js", ".ts", ".json", ".yaml", ".yml",
		".toml", ".sh", ".bash", ".zsh", ".html", ".css", ".rs", ".c", ".cpp",
		".h", ".java", ".rb", ".lua", ".sql", ".xml", ".env", ".conf", ".ini":
		return previewTextFile(path)
	case ".pdf":
		return previewPDF(path)
	case ".docx", ".odt":
		return previewOfficeText(path)
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
		return "[Image: " + name + "]"
	case ".mp3", ".wav", ".flac", ".ogg":
		return previewAudio(path)
	case ".mp4", ".mkv", ".avi", ".webm":
		return "[Video: " + name + "]"
	default:
		info, err := os.Stat(path)
		if err != nil {
			return ""
		}
		return formatSize(info.Size())
	}
}

func previewOfficeText(path string) string {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return "[Document file]"
	}
	defer reader.Close()

	target := "word/document.xml"
	if strings.EqualFold(filepath.Ext(path), ".odt") {
		target = "content.xml"
	}
	for _, file := range reader.File {
		if file.Name != target {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return "[Document file]"
		}
		defer rc.Close()
		data := make([]byte, 8192)
		n, _ := rc.Read(data)
		text := stripXMLTags(string(data[:n]))
		if len(text) > 300 {
			text = text[:300] + "..."
		}
		if strings.TrimSpace(text) == "" {
			return "[Document file]"
		}
		return text
	}
	return "[Document file]"
}

func stripXMLTags(s string) string {
	s = strings.ReplaceAll(s, "</w:p>", "\n")
	s = strings.ReplaceAll(s, "</text:p>", "\n")
	re := regexp.MustCompile(`<[^>]+>`)
	s = re.ReplaceAllString(s, " ")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	return strings.Join(strings.Fields(s), " ")
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~") {
		return strings.Replace(path, "~", os.Getenv("HOME"), 1)
	}
	return path
}

func isImageFile(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
		return true
	default:
		return false
	}
}

func previewTextFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	// First 5 lines, max 200 chars
	var preview []string
	chars := 0
	for i, line := range lines {
		if i >= 5 || chars > 200 {
			break
		}
		if len(line) > 50 {
			line = line[:50] + "..."
		}
		preview = append(preview, line)
		chars += len(line)
	}

	return strings.Join(preview, "\n")
}

func previewPDF(path string) string {
	// Use pdftotext if available
	if _, err := exec.LookPath("pdftotext"); err != nil {
		return "[PDF file]"
	}

	out, err := exec.Command("pdftotext", "-f", "1", "-l", "1", "-layout", path, "-").Output()
	if err != nil {
		return "[PDF file]"
	}

	content := string(out)
	if len(content) > 200 {
		return content[:200] + "..."
	}
	return content
}

func previewPDFImage(path string) string {
	return previewPDFImageAt(path, 1, 360)
}

func previewPDFImageAt(path string, page, scale int) string {
	if _, err := exec.LookPath("pdftoppm"); err != nil {
		return ""
	}
	if page < 1 {
		page = 1
	}
	if scale < 120 {
		scale = 360
	}
	cacheDir := filepath.Join(os.Getenv("HOME"), ".cache", "spark", "pdf-preview")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return ""
	}
	base := filepath.Join(cacheDir, simpleHash(path)+"-p"+stringInt(page)+"-s"+stringInt(scale))
	png := base + "-1.png"
	if _, err := os.Stat(png); err == nil {
		return png
	}
	exec.Command("pdftoppm", "-png", "-singlefile", "-f", stringInt(page), "-l", stringInt(page), "-scale-to", stringInt(scale), path, base).Run()
	if _, err := os.Stat(png); err == nil {
		return png
	}
	return ""
}

func previewAudio(path string) string {
	// Use ffprobe if available
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return "[Audio file]"
	}

	out, err := exec.Command("ffprobe", "-v", "quiet", "-show_entries",
		"format=duration", "-of", "default=noprint_wrappers=1:nokey=1", path).Output()
	if err != nil {
		return "[Audio file]"
	}

	duration := strings.TrimSpace(string(out))
	return "Duration: " + duration + "s"
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return string(rune('0'+bytes/100)) + string(rune('0'+(bytes/10)%10)) + string(rune('0'+bytes%10)) + " B"
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	size := float64(bytes) / float64(div)
	// Simple format
	whole := int(size)
	frac := int((size - float64(whole)) * 10)
	return string(rune('0'+whole/10)) + string(rune('0'+whole%10)) + "." + string(rune('0'+frac)) + " " + units[exp]
}
