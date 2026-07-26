package modules

import (
	"os/exec"
	"strings"
)

var clipboardWriters = [][]string{
	{"wl-copy"},
	{"xclip", "-selection", "clipboard"},
	{"xsel", "--clipboard", "--input"},
}

var terminalExecArgs = map[string][]string{
	"gnome-terminal": {"--"},
}

var terminals = []string{"ghostty", "alacritty", "kitty", "foot", "gnome-terminal"}

func copyToClipboard(text string) {
	for _, writer := range clipboardWriters {
		if writeClipboard(writer[0], writer[1:], text) {
			SetStatus(true, "Copied to clipboard")
			return
		}
	}
	SetStatus(false, "Clipboard copy failed: install wl-clipboard, xclip, or xsel")
}

func writeClipboard(name string, args []string, text string) bool {
	if _, err := exec.LookPath(name); err != nil {
		return false
	}
	cmd := exec.Command(name, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return false
	}
	if err := cmd.Start(); err != nil {
		return false
	}
	_, _ = stdin.Write([]byte(text))
	_ = stdin.Close()
	return cmd.Wait() == nil
}

func openTerminal(command string) {
	for _, term := range terminals {
		if _, err := exec.LookPath(term); err != nil {
			continue
		}
		execFlag := "-e"
		if flags, ok := terminalExecArgs[term]; ok {
			execFlag = flags[0]
		}
		Start(term, execFlag, "sh", "-c", command)
		return
	}
	SetStatus(false, "No terminal found: install one of "+strings.Join(terminals, ", "))
}
