package modules

import "os/exec"

func copyToClipboard(text string) {
	if cmd := exec.Command("wl-copy"); runClipboard(cmd, text) {
		SetStatus(true, "Copied to clipboard")
		return
	}
	if cmd := exec.Command("xclip", "-selection", "clipboard"); runClipboard(cmd, text) {
		SetStatus(true, "Copied to clipboard")
		return
	}
	if cmd := exec.Command("xsel", "--clipboard", "--input"); runClipboard(cmd, text) {
		SetStatus(true, "Copied to clipboard")
		return
	}
	SetStatus(false, "Clipboard copy failed: install wl-clipboard, xclip, or xsel")
}

func runClipboard(cmd *exec.Cmd, text string) bool {
	if _, err := exec.LookPath(cmd.Path); err != nil {
		return false
	}
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
	for _, term := range []string{"ghostty", "alacritty", "kitty", "foot", "gnome-terminal"} {
		if _, err := exec.LookPath(term); err != nil {
			continue
		}
		if term == "gnome-terminal" {
			exec.Command(term, "--", "sh", "-c", command).Start()
		} else {
			exec.Command(term, "-e", "sh", "-c", command).Start()
		}
		return
	}
}
