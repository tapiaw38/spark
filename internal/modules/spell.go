package modules

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// SpellSearch checks spelling and suggests corrections.
func SpellSearch(query string) []Result {
	word, ok := spellTerm(query)
	if !ok || len(word) < 2 {
		return nil
	}

	if result := runSpellChecker(word); result != nil {
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

func spellTerm(query string) (string, bool) {
	q := strings.TrimSpace(query)
	lower := strings.ToLower(q)
	for _, prefix := range []string{"spell ", "spelling "} {
		if strings.HasPrefix(lower, prefix) {
			return strings.TrimSpace(q[len(prefix):]), true
		}
	}
	return "", false
}

func runSpellChecker(word string) *Result {
	if _, err := exec.LookPath("aspell"); err == nil {
		return parseSpellOutput(word, "aspell")
	}
	if _, err := exec.LookPath("hunspell"); err == nil {
		return parseSpellOutput(word, "hunspell")
	}
	return nil
}

func parseSpellOutput(word, cmdName string) *Result {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, cmdName, "-a")
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
				Action: func() { exec.Command("wl-copy", word).Run() },
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
				Action: func() { exec.Command("wl-copy", copyText).Run() },
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
