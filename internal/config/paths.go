package config

import (
	"os"
	"path/filepath"
)

const AppName = "spark"

func homeDir() string {
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return home
	}
	return os.Getenv("HOME")
}

func xdgDir(envVar, fallback string) string {
	if dir := os.Getenv(envVar); dir != "" {
		return dir
	}
	return filepath.Join(homeDir(), fallback)
}

func ConfigHome() string {
	return xdgDir("XDG_CONFIG_HOME", ".config")
}

func DataHome() string {
	return xdgDir("XDG_DATA_HOME", filepath.Join(".local", "share"))
}

func CacheHome() string {
	return xdgDir("XDG_CACHE_HOME", ".cache")
}

func ConfigDir() string {
	return filepath.Join(ConfigHome(), AppName)
}

func DataDir() string {
	return filepath.Join(DataHome(), AppName)
}

func CacheDir() string {
	return filepath.Join(CacheHome(), AppName)
}

func ConfigFile(name string) string {
	return filepath.Join(ConfigDir(), name)
}

func DataFile(name string) string {
	return filepath.Join(DataDir(), name)
}

func CacheSubdir(name string) string {
	return filepath.Join(CacheDir(), name)
}

func HomeFile(parts ...string) string {
	return filepath.Join(append([]string{homeDir()}, parts...)...)
}

func DataHomeFile(parts ...string) string {
	return filepath.Join(append([]string{DataHome()}, parts...)...)
}
