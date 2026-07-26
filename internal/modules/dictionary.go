package modules

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/tapiaw38/spark/internal/platform/commands"
)

var (
	dictCache   = make(map[string]string)
	dictCacheMu sync.RWMutex
)

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

	dictCacheMu.RLock()
	if cached, ok := dictCache[word]; ok {
		dictCacheMu.RUnlock()
		return []Result{{
			Type:       TypeDictionary,
			Title:      word,
			Desc:       cached,
			Icon:       "accessories-dictionary",
			ActionSpec: CopyAction(cached),
		}}
	}
	dictCacheMu.RUnlock()

	if def := localDict(word); def != "" {
		dictCacheMu.Lock()
		dictCache[word] = def
		dictCacheMu.Unlock()

		return []Result{{
			Type:       TypeDictionary,
			Title:      word,
			Desc:       def,
			Icon:       "accessories-dictionary",
			ActionSpec: CopyAction(def),
		}}
	}

	go func() {
		if def := onlineDict(word); def != "" {
			dictCacheMu.Lock()
			dictCache[word] = def
			dictCacheMu.Unlock()
		}
	}()

	return []Result{{
		Type:  TypeDictionary,
		Title: word,
		Desc:  "Looking up...",
		Icon:  "accessories-dictionary",
	}}
}

func localDict(word string) string {
	if _, err := commands.LookPath("dict"); err != nil {
		return ""
	}

	cmd := commands.Command("dict", "-d", "wn", word)
	done := make(chan []byte, 1)
	go func() {
		out, _ := cmd.Output()
		done <- out
	}()

	var out []byte
	select {
	case out = <-done:
	case <-time.After(DictionaryLookupTimeout):
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
			return Truncate(line, DefinitionLen)
		}
	}
	return ""
}

func onlineDict(word string) string {
	client := http.Client{Timeout: DictionaryHTTPTimeout}
	resp, err := client.Get("https://api.dictionaryapi.dev/api/v2/entries/en/" + word)
	if err != nil || resp.StatusCode != 200 {
		return ""
	}
	defer resp.Body.Close()

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
		return Truncate("("+pos+") "+def, DefinitionLen)
	}

	return ""
}
