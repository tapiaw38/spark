package modules

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/tapiaw38/spark/internal/config"
)

// SpellSearch checks spelling and suggests corrections.
func SpellSearch(query string) []Result {
	word, lang, ok := spellTerm(query)
	if !ok || len(word) < 2 {
		return nil
	}
	if lang == "" {
		lang = config.Current.SpellLanguage
	}

	if result := runSpellChecker(word, lang); result != nil {
		return []Result{*result}
	}

	return []Result{{
		Type:   "spell",
		Title:  "Spell: " + word,
		Desc:   "Install aspell or hunspell for spelling suggestions",
		Icon:   "accessories-dictionary",
		Action: func() {},
	}}
}

func spellTerm(query string) (string, string, bool) {
	q := strings.TrimSpace(query)
	lower := strings.ToLower(q)
	for _, prefix := range []string{"spell ", "spelling "} {
		if strings.HasPrefix(lower, prefix) {
			body := strings.TrimSpace(q[len(prefix):])
			parts := strings.Fields(body)
			if len(parts) >= 2 && len(parts[0]) == 2 {
				return strings.TrimSpace(strings.TrimPrefix(body, parts[0])), parts[0], true
			}
			return body, "", true
		}
	}
	return "", "", false
}

func runSpellChecker(word, lang string) *Result {
	if _, err := exec.LookPath("aspell"); err == nil {
		return parseSpellOutput(word, "aspell", lang)
	}
	if _, err := exec.LookPath("hunspell"); err == nil {
		return parseSpellOutput(word, "hunspell", lang)
	}
	return nil
}

func parseSpellOutput(word, cmdName, lang string) *Result {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	args := []string{"-a"}
	if lang != "" {
		if cmdName == "aspell" {
			args = append([]string{"-l", lang}, args...)
		} else {
			args = append([]string{"-d", lang}, args...)
		}
	}
	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Stdin = strings.NewReader(word + "\n")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "@") {
			continue
		}
		if line == "*" {
			return &Result{
				Type:   "spell",
				Title:  word + " is spelled correctly",
				Desc:   "Copy word",
				Icon:   "accessories-dictionary",
				Action: func() { copyToClipboard(word) },
			}
		}
		if strings.HasPrefix(line, "&") || strings.HasPrefix(line, "#") {
			suggestions := spellSuggestions(line)
			title := "No suggestions for " + word
			copyText := word
			if len(suggestions) > 0 {
				title = word + " -> " + suggestions[0]
				copyText = suggestions[0]
			}
			desc := strings.Join(suggestions, ", ")
			return &Result{
				Type:   "spell",
				Title:  title,
				Desc:   desc,
				Icon:   "accessories-dictionary",
				Action: func() { copyToClipboard(copyText) },
			}
		}
	}
	return nil
}

func spellSuggestions(line string) []string {
	colon := strings.Index(line, ":")
	if colon < 0 || colon+1 >= len(line) {
		return nil
	}
	parts := strings.Split(line[colon+1:], ",")
	var out []string
	for _, part := range parts {
		s := strings.TrimSpace(part)
		if s != "" {
			out = append(out, s)
		}
		if len(out) >= 6 {
			break
		}
	}
	return out
}
