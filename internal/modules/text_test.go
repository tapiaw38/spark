package modules

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncateLeavesShortStringsAlone(t *testing.T) {
	for _, s := range []string{"", "abc", "café", "🎵🎵🎵"} {
		if got := Truncate(s, 10); got != s {
			t.Errorf("Truncate(%q, 10) = %q, want unchanged", s, got)
		}
	}
}

func TestTruncateCountsRunesNotBytes(t *testing.T) {
	if got, want := Truncate("ñññññ", 3), "ííí"; got == want {
		t.Fatal("test fixture is wrong")
	}
	got := Truncate("ñññññ", 3)
	if want := "ñññ" + ellipsis; got != want {
		t.Errorf("Truncate = %q, want %q", got, want)
	}
}

func TestTruncateNeverProducesInvalidUTF8(t *testing.T) {
	inputs := []string{
		strings.Repeat("a", 79) + "éxito con acentos",
		strings.Repeat("a", 48) + "🎵 musica",
		strings.Repeat("ñ", 200),
		strings.Repeat("🎵", 100),
		"日本語のテキストはここにあります",
	}
	for _, in := range inputs {
		for _, limit := range []int{1, 5, 40, 50, 80, 200, 300} {
			got := Truncate(in, limit)
			if !utf8.ValidString(got) {
				t.Errorf("Truncate(%.12q..., %d) produced invalid UTF-8: %q", in, limit, got)
			}
		}
	}
}

func TestTruncateAppendsEllipsisOnlyWhenCut(t *testing.T) {
	if got := Truncate("abcdef", 3); got != "abc"+ellipsis {
		t.Errorf("Truncate(abcdef, 3) = %q, want abc%s", got, ellipsis)
	}
	if got := Truncate("abc", 3); got != "abc" {
		t.Errorf("Truncate(abc, 3) = %q, want abc with no ellipsis", got)
	}
}

func TestTruncateKeepsExactlyMaxRunes(t *testing.T) {
	got := Truncate(strings.Repeat("🎵", 10), 4)
	body := strings.TrimSuffix(got, ellipsis)
	if n := utf8.RuneCountInString(body); n != 4 {
		t.Errorf("kept %d runes, want 4", n)
	}
}

func TestModulePreviewsSurviveMultibyteContent(t *testing.T) {
	cases := []struct {
		name  string
		input string
		limit int
	}{
		{"snippet", strings.Repeat("ñ", 100), SnippetPreviewLen},
		{"clipboard", strings.Repeat("🎵", 100), ClipboardLen},
		{"file line", strings.Repeat("日", 100), FilePreviewLineLen},
		{"definition", strings.Repeat("é", 100), DefinitionLen},
		{"pdf", strings.Repeat("→", 300), PDFPreviewLen},
		{"document", strings.Repeat("Ω", 500), DocumentPreviewLen},
	}
	for _, tc := range cases {
		got := Truncate(tc.input, tc.limit)
		if !utf8.ValidString(got) {
			t.Errorf("%s: invalid UTF-8 after truncation", tc.name)
		}
		body := strings.TrimSuffix(got, ellipsis)
		if n := utf8.RuneCountInString(body); n != tc.limit {
			t.Errorf("%s: kept %d runes, want %d", tc.name, n, tc.limit)
		}
	}
}
