package modules

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var snippetsPath = filepath.Join(os.Getenv("HOME"), ".config", "spark", "snippets.json")

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
		// Create default snippets file
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

// SnippetSearch finds snippets matching query
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

			snippet := s // capture
			preview := snippet.Content
			if len(preview) > 40 {
				preview = preview[:40] + "..."
			}
			preview = strings.ReplaceAll(preview, "\n", " ")

			results = append(results, Result{
				Type:  "snippet",
				Title: snippet.Name + " (" + snippet.Keyword + ")",
				Desc:  preview,
				Icon:  "edit-paste",
				Action: func() {
					content := expandSnippet(snippet.Content)
					copyToClipboard(content)
					// Simulate paste with wtype
					if _, err := exec.LookPath("wtype"); err == nil {
						exec.Command("wtype", "-M", "ctrl", "v", "-m", "ctrl").Run()
					}
				},
			})
		}

		if len(results) >= 5 {
			break
		}
	}

	return results
}

func expandSnippet(content string) string {
	// Replace dynamic placeholders
	content = strings.ReplaceAll(content, "{{DATE}}", currentDate())
	content = strings.ReplaceAll(content, "{{TIME}}", currentTime())
	return content
}

func currentDate() string {
	out, _ := exec.Command("date", "+%Y-%m-%d").Output()
	return strings.TrimSpace(string(out))
}

func currentTime() string {
	out, _ := exec.Command("date", "+%H:%M").Output()
	return strings.TrimSpace(string(out))
}
