package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestXDGEnvVarsWin(t *testing.T) {
	t.Setenv("HOME", "/home/tester")
	t.Setenv("XDG_CONFIG_HOME", "/custom/cfg")
	t.Setenv("XDG_DATA_HOME", "/custom/data")
	t.Setenv("XDG_CACHE_HOME", "/custom/cache")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ConfigDir", ConfigDir(), "/custom/cfg/spark"},
		{"DataDir", DataDir(), "/custom/data/spark"},
		{"CacheDir", CacheDir(), "/custom/cache/spark"},
		{"ConfigFile", ConfigFile("config.yaml"), "/custom/cfg/spark/config.yaml"},
		{"DataFile", DataFile("history.json"), "/custom/data/spark/history.json"},
		{"CacheSubdir", CacheSubdir("art"), "/custom/cache/spark/art"},
		{"DataHomeFile", DataHomeFile("applications"), "/custom/data/applications"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}

func TestFallsBackToHomeWhenXDGUnset(t *testing.T) {
	t.Setenv("HOME", "/home/tester")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ConfigDir", ConfigDir(), "/home/tester/.config/spark"},
		{"DataDir", DataDir(), "/home/tester/.local/share/spark"},
		{"CacheDir", CacheDir(), "/home/tester/.cache/spark"},
		{"DataHomeFile", DataHomeFile("Trash", "files"), "/home/tester/.local/share/Trash/files"},
		{"HomeFile", HomeFile(".ssh", "config"), "/home/tester/.ssh/config"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}

func TestSparkDirsAreAllUnderTheAppName(t *testing.T) {
	t.Setenv("HOME", "/home/tester")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")

	for _, dir := range []string{ConfigDir(), DataDir(), CacheDir()} {
		if filepath.Base(dir) != AppName {
			t.Errorf("%q does not end in %q", dir, AppName)
		}
	}
}

func TestPathsAreAbsolute(t *testing.T) {
	t.Setenv("HOME", "/home/tester")
	for _, p := range []string{ConfigDir(), DataDir(), CacheDir(), HomeFile("Music")} {
		if !strings.HasPrefix(p, "/") {
			t.Errorf("%q is not absolute", p)
		}
	}
}
