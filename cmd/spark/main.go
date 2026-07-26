package main

import (
	"os"

	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/tapiaw38/spark/internal/config"
)

func main() {
	if cmd, args, ok := lookupCommand(os.Args[1:]); ok {
		cmd.run(args)
		return
	}
	runLauncher()
}

type command struct {
	minArgs int
	run     func(args []string)
}

var commands = map[string]command{
	"--large-type":     {minArgs: 1, run: func(a []string) { runWindow(func() { showLargeType(a[0], -1) }) }},
	"--large-type-all": {minArgs: 1, run: func(a []string) { runWindow(func() { showLargeTypeAll(a[0]) }) }},
	"--stats-window":   {minArgs: 0, run: func([]string) { runWindow(showStatsWindow) }},
	"--file-op-window": {minArgs: 3, run: func(a []string) { runWindow(func() { showFileOpWindow(a[0], a[1], a[2]) }) }},
	"--email-window": {minArgs: 0, run: func(a []string) {
		runWindow(func() { showEmailWindow(argAt(a, 0), argAt(a, 1), argAt(a, 2)) })
	}},
	"--setup": {minArgs: 0, run: func([]string) { runSetup() }},
}

func lookupCommand(argv []string) (command, []string, bool) {
	if len(argv) == 0 {
		return command{}, nil, false
	}
	cmd, ok := commands[argv[0]]
	if !ok {
		return command{}, nil, false
	}
	args := argv[1:]
	if len(args) < cmd.minArgs {
		return command{}, nil, false
	}
	return cmd, args, true
}

func argAt(args []string, i int) string {
	if i >= len(args) {
		return ""
	}
	return args[i]
}

func runWindow(show func()) {
	gtk.Init()
	show()
	gtk.Main()
	os.Exit(0)
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
