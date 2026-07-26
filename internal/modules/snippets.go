package modules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tapiaw38/spark/internal/config"
)

var snippetsPath = config.ConfigFile("snippets.json")

type Snippet struct {
	Keyword string `json:"keyword"`
	Content string `json:"content"`
	Name    string `json:"name"`
}

var snippets []Snippet

func init() {
	loadSnippets()
}

func loadSnippets() {
	data, err := os.ReadFile(snippetsPath)
	if err != nil {
		snippets = []Snippet{
			{Keyword: ";email", Content: "tu@email.com", Name: "Email"},
			{Keyword: ";tel", Content: "+54 9 XXX XXX XXXX", Name: "Teléfono"},
			{Keyword: ";firma", Content: "Saludos,\nWalter Tapia\nDeveloper", Name: "Firma"},
			{Keyword: ";date", Content: "{{DATE}}", Name: "Fecha actual"},
			{Keyword: ";shrug", Content: "¯\\_(ツ)_/¯", Name: "Shrug"},
		}
		saveSnippets()
		return
	}
	json.Unmarshal(data, &snippets)
}

func saveSnippets() {
	os.MkdirAll(filepath.Dir(snippetsPath), 0755)
	data, _ := json.MarshalIndent(snippets, "", "  ")
	os.WriteFile(snippetsPath, data, 0644)
}

func SnippetSearch(query string) []Result {
	if !strings.HasPrefix(query, ";") && !strings.HasPrefix(strings.ToLower(query), "snip") {
		return nil
	}

	searchTerm := query
	if strings.HasPrefix(strings.ToLower(query), "snip ") {
		searchTerm = strings.TrimPrefix(strings.ToLower(query), "snip ")
	}

	var results []Result
	for _, s := range snippets {
		if strings.Contains(strings.ToLower(s.Keyword), strings.ToLower(searchTerm)) ||
			strings.Contains(strings.ToLower(s.Name), strings.ToLower(searchTerm)) {

			snippet := s
			preview := Truncate(snippet.Content, SnippetPreviewLen)
			preview = strings.ReplaceAll(preview, "\n", " ")

			results = append(results, Result{
				Type:       TypeSnippet,
				Title:      snippet.Name + " (" + snippet.Keyword + ")",
				Desc:       preview,
				Icon:       "edit-paste",
				ActionSpec: PasteAction(expandSnippet(snippet.Content)),
			})
		}

		if len(results) >= MaxSnippetResults {
			break
		}
	}

	return results
}

func expandSnippet(content string) string {
	content = strings.ReplaceAll(content, "{{DATE}}", currentDate())
	content = strings.ReplaceAll(content, "{{TIME}}", currentTime())
	return content
}

func currentDate() string {
	return time.Now().Format("2006-01-02")
}

func currentTime() string {
	return time.Now().Format("15:04")
}
