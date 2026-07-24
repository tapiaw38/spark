package main

import (
	"os"

	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/tapiaw38/spark/internal/config"
)

func runSubcommand() bool {
	if len(os.Args) < 2 {
		return false
	}
	switch os.Args[1] {
	case "--large-type":
		return uiCommand(3, func() { showLargeType(os.Args[2], -1) })
	case "--large-type-all":
		return uiCommand(3, func() { showLargeTypeAll(os.Args[2]) })
	case "--stats-window":
		return uiCommand(2, showStatsWindow)
	case "--email-window":
		return uiCommand(2, func() { showEmailWindow(arg(2), arg(3), arg(4)) })
	case "--file-op-window":
		return uiCommand(5, func() { showFileOpWindow(os.Args[2], os.Args[3], os.Args[4]) })
	case "--setup":
		runSetup()
		return true
	}
	return false
}

func uiCommand(minArgs int, show func()) bool {
	if len(os.Args) < minArgs {
		return false
	}
	gtk.Init()
	show()
	gtk.Main()
	os.Exit(0)
	return true
}

func arg(i int) string {
	if len(os.Args) > i {
		return os.Args[i]
	}
	return ""
}

func runSetup() {
	config.Load()
	sparkPath, _ := os.Executable()
	if err := config.SetupHotkey(sparkPath); err != nil {
		os.Stderr.WriteString("Failed to setup hotkey: " + err.Error() + "\n")
		os.Exit(1)
	}
	os.Stdout.WriteString("Hotkey configured: " + config.Current.Hotkey + "\nRestart mango to apply.\n")
	os.Exit(0)
}
