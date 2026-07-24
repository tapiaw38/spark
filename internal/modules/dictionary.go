package modules

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var (
	dictCache   = make(map[string]string)
	dictCacheMu sync.RWMutex
)

// DictionarySearch looks up word definitions
func DictionarySearch(query string) []Result {
	var word string
	if strings.HasPrefix(strings.ToLower(query), "define ") {
		word = strings.TrimSpace(strings.TrimPrefix(strings.ToLower(query), "define "))
	} else if strings.HasPrefix(strings.ToLower(query), "def ") {
		word = strings.TrimSpace(strings.TrimPrefix(strings.ToLower(query), "def "))
	} else {
		return nil
	}

	if len(word) < 2 {
		return nil
	}

	// Check cache first
	dictCacheMu.RLock()
	if cached, ok := dictCache[word]; ok {
		dictCacheMu.RUnlock()
		return []Result{{
			Type:  "dictionary",
			Title: word,
			Desc:  cached,
			Icon:  "accessories-dictionary",
			Action: func() {
				copyToClipboard(cached)
			},
		}}
	}
	dictCacheMu.RUnlock()

	// Try local dict (fast)
	if def := localDict(word); def != "" {
		dictCacheMu.Lock()
		dictCache[word] = def
		dictCacheMu.Unlock()

		return []Result{{
			Type:  "dictionary",
			Title: word,
			Desc:  def,
			Icon:  "accessories-dictionary",
			Action: func() {
				copyToClipboard(def)
			},
		}}
	}

	// Async online lookup - return placeholder, cache when done
	go func() {
		if def := onlineDict(word); def != "" {
			dictCacheMu.Lock()
			dictCache[word] = def
			dictCacheMu.Unlock()
		}
	}()

	return []Result{{
		Type:   "dictionary",
		Title:  word,
		Desc:   "Looking up...",
		Icon:   "accessories-dictionary",
		Action: func() {},
	}}
}

func localDict(word string) string {
	if _, err := exec.LookPath("dict"); err != nil {
		return ""
	}

	// Timeout 200ms
	cmd := exec.Command("dict", "-d", "wn", word)
	done := make(chan []byte, 1)
	go func() {
		out, _ := cmd.Output()
		done <- out
	}()

	var out []byte
	select {
	case out = <-done:
	case <-time.After(200 * time.Millisecond):
		cmd.Process.Kill()
		return ""
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "1 definition") || strings.HasPrefix(line, "From") || line == "" {
			continue
		}
		if len(line) > 10 {
			if len(line) > 80 {
				return line[:80] + "..."
			}
			return line
		}
	}
	return ""
}

func onlineDict(word string) string {
	client := http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get("https://api.dictionaryapi.dev/api/v2/entries/en/" + word)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return ""
	}

	var data []struct {
		Meanings []struct {
			PartOfSpeech string `json:"partOfSpeech"`
			Definitions  []struct {
				Definition string `json:"definition"`
			} `json:"definitions"`
		} `json:"meanings"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return ""
	}

	if len(data) > 0 && len(data[0].Meanings) > 0 && len(data[0].Meanings[0].Definitions) > 0 {
		def := data[0].Meanings[0].Definitions[0].Definition
		pos := data[0].Meanings[0].PartOfSpeech
		result := "(" + pos + ") " + def
		if len(result) > 80 {
			return result[:80] + "..."
		}
		return result
	}

	return ""
}
