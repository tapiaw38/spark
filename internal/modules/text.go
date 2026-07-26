package modules

import "unicode/utf8"

const ellipsis = "..."

const (
	SnippetPreviewLen  = 40
	ClipboardLen       = 50
	FilePreviewLineLen = 50
	DefinitionLen      = 80
	PDFPreviewLen      = 200
	DocumentPreviewLen = 300
)

func Truncate(s string, maxRunes int) string {
	if maxRunes <= 0 || utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	count := 0
	for i := range s {
		if count == maxRunes {
			return s[:i] + ellipsis
		}
		count++
	}
	return s
}
