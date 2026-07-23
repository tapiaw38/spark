package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	// Appearance
	Width           int     `yaml:"width"`
	MaxResults      int     `yaml:"max_results"`
	BackgroundColor string  `yaml:"background_color"`
	BackgroundAlpha float64 `yaml:"background_alpha"`
	BorderRadius    int     `yaml:"border_radius"`
	FontSize        int     `yaml:"font_size"`
	TextColor       string  `yaml:"text_color"`
	SelectionColor  string  `yaml:"selection_color"`

	// Web shortcuts
	WebShortcuts map[string]WebShortcut `yaml:"web_shortcuts"`

	// Behavior
	ShowIcons     bool   `yaml:"show_icons"`
	IconSize      int    `yaml:"icon_size"`
	MarginTop     int    `yaml:"margin_top"`
	HistoryBoost  int    `yaml:"history_boost"`
	SpellLanguage string `yaml:"spell_language"`

	// Hotkey (for mango WM: SUPER,s or SUPER+SHIFT,space etc.)
	Hotkey string `yaml:"hotkey"`
}

type WebShortcut struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
	Icon string `yaml:"icon"`
}

var defaultConfig = Config{
	Width:           600,
	MaxResults:      6,
	BackgroundColor: "30, 30, 40",
	BackgroundAlpha: 0.95,
	BorderRadius:    12,
	FontSize:        18,
	TextColor:       "white",
	SelectionColor:  "100, 150, 255",
	ShowIcons:       true,
	IconSize:        24,
	MarginTop:       100,
	HistoryBoost:    3,
	SpellLanguage:   "en",
	Hotkey:          "Alt,space",
	WebShortcuts: map[string]WebShortcut{
		"g":    {Name: "Google", URL: "https://www.google.com/search?q=%s", Icon: "web-browser"},
		"yt":   {Name: "YouTube", URL: "https://www.youtube.com/results?search_query=%s", Icon: "youtube"},
		"gh":   {Name: "GitHub", URL: "https://github.com/search?q=%s", Icon: "github"},
		"wiki": {Name: "Wikipedia", URL: "https://en.wikipedia.org/wiki/Special:Search?search=%s", Icon: "wikipedia"},
		"ddg":  {Name: "DuckDuckGo", URL: "https://duckduckgo.com/?q=%s", Icon: "web-browser"},
		"r":    {Name: "Reddit", URL: "https://www.reddit.com/search/?q=%s", Icon: "reddit"},
		"so":   {Name: "Stack Overflow", URL: "https://stackoverflow.com/search?q=%s", Icon: "web-browser"},
	},
}

var Current = defaultConfig

func Load() error {
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "spark", "config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		// No config file, use defaults and create one
		return Save()
	}

	if err := yaml.Unmarshal(data, &Current); err != nil {
		return err
	}

	return nil
}

func Save() error {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "spark")
	os.MkdirAll(configDir, 0755)

	configPath := filepath.Join(configDir, "config.yaml")

	data, err := yaml.Marshal(&Current)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// SetupHotkey updates mango WM bind.conf with configured hotkey
func SetupHotkey(sparkPath string) error {
	bindPath := filepath.Join(os.Getenv("HOME"), ".config", "mango", "bind.conf")
	bindLine := fmt.Sprintf("bind=%s,spawn,%s", Current.Hotkey, sparkPath)

	// Read existing config
	data, err := os.ReadFile(bindPath)
	if err != nil {
		// Create new file with just the bind
		os.MkdirAll(filepath.Dir(bindPath), 0755)
		return os.WriteFile(bindPath, []byte(bindLine+"\n"), 0644)
	}

	content := string(data)
	var newLines []string
	found := false

	// Remove old spark bindings, add new one
	for _, line := range splitLines(content) {
		if containsSparkBind(line) {
			if !found {
				newLines = append(newLines, bindLine)
				found = true
			}
			// Skip old spark bind
		} else {
			newLines = append(newLines, line)
		}
	}

	if !found {
		newLines = append(newLines, bindLine)
	}

	return os.WriteFile(bindPath, []byte(joinLines(newLines)), 0644)
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		result += line
		if i < len(lines)-1 {
			result += "\n"
		}
	}
	if len(result) > 0 && result[len(result)-1] != '\n' {
		result += "\n"
	}
	return result
}

func containsSparkBind(line string) bool {
	// Check if line is a spark binding
	for i := 0; i < len(line); i++ {
		if i+5 <= len(line) && line[i:i+5] == "spark" {
			return true
		}
	}
	return false
}

// GetCSS generates CSS from config
func GetCSS() string {
	c := Current
	return fmt.Sprintf(`
		window {
			background: rgba(%s, %.2f);
			border-radius: %dpx;
		}
		#spark-entry {
			font-size: %dpx;
			padding: 12px;
			background: rgba(255, 255, 255, 0.1);
			border: none;
			border-radius: 8px;
			color: %s;
		}
		#spark-results {
			background: transparent;
		}
		#spark-row {
			background: transparent;
			color: %s;
			border-radius: 6px;
			padding: 4px 8px;
			outline: none;
			box-shadow: none;
		}
		#spark-row:hover,
		#spark-row:focus,
		#spark-row:active {
			background: transparent;
			outline: none;
			box-shadow: none;
		}
		#spark-row:selected,
		#spark-row:selected:hover,
		#spark-row:selected:focus {
			background: rgba(%s, 0.3);
			outline: none;
			box-shadow: none;
		}
		#spark-title {
			color: %s;
			font-size: 14px;
		}
		#spark-desc {
			color: rgba(255, 255, 255, 0.6);
			font-size: 11px;
		}
	`, c.BackgroundColor, c.BackgroundAlpha, c.BorderRadius,
		c.FontSize, c.TextColor, c.TextColor, c.SelectionColor, c.TextColor)
}
