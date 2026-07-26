package modules

import (
	"archive/zip"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tapiaw38/spark/internal/platform/commands"

	"github.com/tapiaw38/spark/internal/config"
)

func GetPreview(r Result) string {
	if r.Preview != "" {
		return r.Preview
	}

	switch r.Type {
	case TypeFile:
		return previewFile(r.Title, r.Desc)
	case TypeSnippet:
		return ""
	case TypeClipboard:
		return r.Title
	case TypeDictionary:
		return r.Desc
	case TypeCalc:
		return r.Title
	default:
		return ""
	}
}

func GetPreviewImage(r Result) string {
	return GetPreviewImageAt(r, 1, 360)
}

func GetPreviewImageAt(r Result, page, scale int) string {
	if r.PreviewImage != "" {
		return expandHome(r.PreviewImage)
	}

	if !isImageFile(r.Title) {
		ext := strings.ToLower(filepath.Ext(r.Title))
		if ext == ".pdf" {
			return previewPDFImageAt(GetFilePath(r), page, scale)
		}
		if ext == ".docx" || ext == ".odt" {
			if pdf := previewOfficePDF(GetFilePath(r)); pdf != "" {
				return previewPDFImageAt(pdf, page, scale)
			}
		}
		return ""
	}

	return GetFilePath(r)
}

func GetFilePath(r Result) string {
	if r.Type != TypeFile && r.Type != TypeDirectory {
		return ""
	}
	if r.Type == TypeDirectory && strings.HasPrefix(r.NavigateQuery, "nav ") {
		return expandHome(strings.TrimSpace(strings.TrimPrefix(r.NavigateQuery, "nav ")))
	}
	return filepath.Join(expandHome(r.Desc), r.Title)
}

func previewFile(name, dir string) string {
	path := filepath.Join(expandHome(dir), name)

	ext := strings.ToLower(filepath.Ext(name))

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
		text := Truncate(stripXMLTags(string(data[:n])), DocumentPreviewLen)
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

func IsImageFile(name string) bool {
	return isImageFile(name)
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

	var preview []string
	chars := 0
	for i, line := range lines {
		if i >= 5 || chars > 200 {
			break
		}
		line = Truncate(line, FilePreviewLineLen)
		preview = append(preview, line)
		chars += len(line)
	}

	return strings.Join(preview, "\n")
}

func previewPDF(path string) string {
	if _, err := commands.LookPath("pdftotext"); err != nil {
		return "[PDF file]"
	}

	out, err := commands.Command("pdftotext", "-f", "1", "-l", "1", "-layout", path, "-").Output()
	if err != nil {
		return "[PDF file]"
	}

	return Truncate(string(out), PDFPreviewLen)
}

func previewPDFImage(path string) string {
	return previewPDFImageAt(path, 1, 360)
}

func previewPDFImageAt(path string, page, scale int) string {
	if _, err := commands.LookPath("pdftoppm"); err != nil {
		return ""
	}
	if page < 1 {
		page = 1
	}
	if scale < 120 {
		scale = 360
	}
	cacheDir := config.CacheSubdir("pdf-preview")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return ""
	}
	base := filepath.Join(cacheDir, simpleHash(path)+"-p"+stringInt(page)+"-s"+stringInt(scale))
	png := base + ".png"
	if _, err := os.Stat(png); err == nil {
		return png
	}
	commands.Command("pdftoppm", "-png", "-singlefile", "-f", stringInt(page), "-l", stringInt(page), "-scale-to", stringInt(scale), path, base).Run()
	if _, err := os.Stat(png); err == nil {
		return png
	}
	return ""
}

func previewOfficePDF(path string) string {
	cacheDir := config.CacheSubdir("office-preview")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return ""
	}
	target := filepath.Join(cacheDir, simpleHash(path)+".pdf")
	if _, err := os.Stat(target); err == nil {
		return target
	}
	cmdName := "libreoffice"
	if _, err := commands.LookPath(cmdName); err != nil {
		cmdName = "soffice"
	}
	tmpDir := filepath.Join(cacheDir, simpleHash(path)+"-work")
	os.RemoveAll(tmpDir)
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)
	if !convertOfficeToPDF(path, tmpDir) {
		return ""
	}
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".pdf") {
			continue
		}
		src := filepath.Join(tmpDir, entry.Name())
		data, err := os.ReadFile(src)
		if err != nil {
			return ""
		}
		if os.WriteFile(target, data, 0644) == nil {
			return target
		}
	}
	return ""
}

func convertOfficeToPDF(path, outDir string) bool {
	for _, attempt := range officeConvertCommands(path, outDir) {
		if _, err := commands.LookPath(attempt[0]); err != nil {
			continue
		}
		if commands.Command(attempt[0], attempt[1:]...).Run() == nil {
			if hasPDF(outDir) {
				return true
			}
		}
	}
	return false
}

func officeConvertCommands(path, outDir string) [][]string {
	return [][]string{
		{"libreoffice", "--headless", "--convert-to", "pdf", "--outdir", outDir, path},
		{"soffice", "--headless", "--convert-to", "pdf", "--outdir", outDir, path},
		{"onlyoffice-desktopeditors", "--convert-to", "pdf", "--outdir", outDir, path},
		{"onlyoffice-desktopeditors", "--convert-to", "pdf", "--output-dir", outDir, path},
		{"onlyoffice-desktopeditors", "--headless", "--convert-to", "pdf", "--outdir", outDir, path},
		{"desktopeditors", "--convert-to", "pdf", "--outdir", outDir, path},
		{"desktopeditors", "--convert-to", "pdf", "--output-dir", outDir, path},
		{"desktopeditors", "--headless", "--convert-to", "pdf", "--outdir", outDir, path},
	}
}

func hasPDF(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".pdf") {
			return true
		}
	}
	return false
}

func previewAudio(path string) string {
	if _, err := commands.LookPath("ffprobe"); err != nil {
		return "[Audio file]"
	}

	out, err := commands.Command("ffprobe", "-v", "quiet", "-show_entries",
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
	whole := int(size)
	frac := int((size - float64(whole)) * 10)
	return string(rune('0'+whole/10)) + string(rune('0'+whole%10)) + "." + string(rune('0'+frac)) + " " + units[exp]
}
