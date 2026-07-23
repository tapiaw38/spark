package history

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var historyPath = filepath.Join(os.Getenv("HOME"), ".local/share/spark/history.json")

// counts maps app name -> launch count
var counts map[string]int

func init() {
	counts = make(map[string]int)
	load()
}

func load() {
	data, err := os.ReadFile(historyPath)
	if err != nil {
		return
	}
	json.Unmarshal(data, &counts)
}

func save() {
	os.MkdirAll(filepath.Dir(historyPath), 0755)
	data, _ := json.Marshal(counts)
	os.WriteFile(historyPath, data, 0644)
}

// Record increments launch count for app
func Record(appName string) {
	counts[appName]++
	save()
}

// Score returns launch count for app (used as boost in search)
func Score(appName string) int {
	return counts[appName]
}
