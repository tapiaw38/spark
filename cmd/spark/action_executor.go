package main

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	cmdrunner "github.com/tapiaw38/spark/internal/platform/commands"

	"github.com/tapiaw38/spark/internal/apps"
	"github.com/tapiaw38/spark/internal/config"
	"github.com/tapiaw38/spark/internal/history"
	"github.com/tapiaw38/spark/internal/modules"
)

var terminalExecArgs = map[string][]string{
	"gnome-terminal": {"--"},
}

var terminals = []string{"ghostty", "alacritty", "kitty", "foot", "gnome-terminal"}

var clipboardWriters = [][]string{
	{"wl-copy"},
	{"xclip", "-selection", "clipboard"},
	{"xsel", "--clipboard", "--input"},
}

var actionHandlers = map[modules.ActionKind]func(modules.ActionSpec) bool{
	modules.ActionOpen: func(a modules.ActionSpec) bool {
		return openPaths(append([]string{a.Target}, a.Args...))
	},
	modules.ActionCopy:  func(a modules.ActionSpec) bool { return copyToClipboard(a.Target) },
	modules.ActionStart: func(a modules.ActionSpec) bool { return startCommand(a.Target, a.Args...) },
	modules.ActionRun:   func(a modules.ActionSpec) bool { return runCommand(a.Target, a.Args...) },

	modules.ActionTerminal:    func(a modules.ActionSpec) bool { return openTerminal(a.Target) },
	modules.ActionPaste:       func(a modules.ActionSpec) bool { return pasteText(a.Target) },
	modules.ActionEmail:       executeEmailAction,
	modules.ActionEmailWindow: executeEmailWindowAction,
	modules.ActionFile:        executeFileAction,
	modules.ActionClipboard:   executeClipboardHistoryAction,
	modules.ActionScreenshot:  executeScreenshotAction,
	modules.ActionMusic:       executeMusicAction,
	modules.ActionWeather:     executeWeatherAction,
	modules.ActionPlayer:      executePlayerAction,
	modules.ActionState:       executeStateAction,
	modules.ActionSync:        executeSyncAction,
	modules.ActionSystem:      executeSystemAction,
	modules.ActionApp:         executeAppAction,
}

func executeActionSpec(action modules.ActionSpec) bool {
	if action.IsZero() {
		return false
	}
	handler, known := actionHandlers[action.Kind]
	if !known {
		modules.SetStatus(false, "Unknown action: "+string(action.Kind))
		return false
	}
	ok := handler(action)
	if ok && action.SuccessStatus != "" {
		modules.SetStatus(true, action.SuccessStatus)
	}
	return ok
}

func openPaths(paths []string) bool {
	ok := true
	for _, path := range paths {
		if path == "" {
			continue
		}
		if !startCommand("xdg-open", path) {
			ok = false
		}
	}
	return ok
}

func startCommand(name string, args ...string) bool {
	cmd := cmdrunner.Command(name, args...)
	if err := cmd.Start(); err != nil {
		modules.SetStatus(false, "Failed to run "+name+": "+err.Error())
		return false
	}
	go func() {
		_ = cmd.Wait()
	}()
	return true
}

func runCommand(name string, args ...string) bool {
	if err := cmdrunner.Command(name, args...).Run(); err != nil {
		modules.SetStatus(false, "Failed to run "+name+": "+err.Error())
		return false
	}
	return true
}

func openTerminal(command string) bool {
	for _, term := range terminals {
		if _, err := cmdrunner.LookPath(term); err != nil {
			continue
		}
		execFlag := "-e"
		if flags, ok := terminalExecArgs[term]; ok {
			execFlag = flags[0]
		}
		return startCommand(term, execFlag, "sh", "-c", command)
	}
	modules.SetStatus(false, "No terminal found: install one of "+strings.Join(terminals, ", "))
	return false
}

func copyToClipboard(text string) bool {
	for _, writer := range clipboardWriters {
		if writeClipboard(writer[0], writer[1:], text) {
			modules.SetStatus(true, "Copied to clipboard")
			return true
		}
	}
	modules.SetStatus(false, "Clipboard copy failed: install wl-clipboard, xclip, or xsel")
	return false
}

func writeClipboard(name string, args []string, text string) bool {
	if _, err := cmdrunner.LookPath(name); err != nil {
		return false
	}
	cmd := cmdrunner.Command(name, args...)
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

func executeEmailAction(action modules.ActionSpec) bool {
	subject, body := "", ""
	if len(action.Args) > 0 {
		subject = action.Args[0]
	}
	if len(action.Args) > 1 {
		body = action.Args[1]
	}
	attachments := []string{}
	if len(action.Args) > 2 {
		attachments = action.Args[2:]
	}
	if len(attachments) == 0 && body != "" {
		link := "mailto:" + url.QueryEscape(action.Target) + "?subject=" + url.QueryEscape(subject) + "&body=" + url.QueryEscape(body)
		if startCommand("xdg-open", link) {
			modules.SetStatus(true, "Email compose opened")
			return true
		}
		return false
	}
	args := []string{}
	if subject != "" {
		args = append(args, "--subject", subject)
	}
	if body != "" {
		args = append(args, "--body", body)
	}
	for _, path := range attachments {
		if path != "" {
			args = append(args, "--attach", path)
		}
	}
	if action.Target != "" {
		args = append(args, action.Target)
	}
	if !startCommand("xdg-email", args...) {
		return false
	}
	if len(attachments) > 1 {
		modules.SetStatus(true, "Email compose opened with buffer")
	} else if len(attachments) == 1 {
		modules.SetStatus(true, "Email compose opened with attachment")
	} else {
		modules.SetStatus(true, "Email compose opened")
	}
	return true
}

func executeEmailWindowAction(action modules.ActionSpec) bool {
	subject, body := "", ""
	if len(action.Args) > 0 {
		subject = action.Args[0]
	}
	if len(action.Args) > 1 {
		body = action.Args[1]
	}
	exe, err := os.Executable()
	if err != nil {
		modules.SetStatus(false, "Email composer failed: "+err.Error())
		return false
	}
	return startCommand(exe, "--email-window", action.Target, subject, body)
}

func executeFileAction(action modules.ActionSpec) bool {
	switch action.Target {
	case "reveal":
		if len(action.Args) == 0 {
			return false
		}
		return revealFile(action.Args[0])
	case "trash":
		if len(action.Args) == 0 {
			return false
		}
		return moveToTrash(action.Args[0])
	case "op-window":
		if len(action.Args) < 3 {
			return false
		}
		exe, err := os.Executable()
		if err != nil {
			modules.SetStatus(false, "File operation window failed: "+err.Error())
			return false
		}
		return startCommand(exe, "--file-op-window", action.Args[0], action.Args[1], action.Args[2])
	case "op":
		if len(action.Args) < 3 {
			return false
		}
		modules.RunFileOperation(action.Args[0], action.Args[1], action.Args[2])
		return true
	}
	return false
}

func revealFile(path string) bool {
	if _, err := cmdrunner.LookPath("dbus-send"); err == nil {
		uri := (&url.URL{Scheme: "file", Path: filepath.ToSlash(path)}).String()
		return startCommand("dbus-send", "--session", "--dest=org.freedesktop.FileManager1", "--type=method_call", "/org/freedesktop/FileManager1", "org.freedesktop.FileManager1.ShowItems", "array:string:"+uri, "string:")
	}
	return startCommand("xdg-open", filepath.Dir(path))
}

func moveToTrash(path string) bool {
	if _, err := cmdrunner.LookPath("gio"); err == nil {
		if runCommand("gio", "trash", path) {
			modules.SetTrashUndo()
			return true
		}
		return false
	}
	return runCommand("kioclient6", "move", path, "trash:/")
}

func executeClipboardHistoryAction(action modules.ActionSpec) bool {
	decode := cmdrunner.Command("cliphist", "decode")
	decode.Stdin = strings.NewReader(action.Target)
	decoded, err := decode.Output()
	if err != nil {
		modules.SetStatus(false, "Clipboard decode failed: "+err.Error())
		return false
	}
	if !copyBytesToClipboard(decoded) {
		return false
	}
	if len(action.Args) > 0 && action.Args[0] == "paste" {
		pasteClipboard()
	}
	return true
}

func copyBytesToClipboard(data []byte) bool {
	cmd := cmdrunner.Command("wl-copy")
	cmd.Stdin = bytes.NewReader(data)
	if err := cmd.Run(); err != nil {
		modules.SetStatus(false, "Clipboard copy failed: "+err.Error())
		return false
	}
	return true
}

func pasteText(text string) bool {
	if !copyToClipboard(text) {
		return false
	}
	pasteClipboard()
	return true
}

func pasteClipboard() {
	if _, err := cmdrunner.LookPath("wtype"); err == nil {
		startCommand("wtype", "-M", "ctrl", "v", "-m", "ctrl")
		return
	}
	if _, err := cmdrunner.LookPath("ydotool"); err == nil {
		startCommand("ydotool", "key", "29:1", "47:1", "47:0", "29:0")
	}
}

func executeScreenshotAction(action modules.ActionSpec) bool {
	switch action.Target {
	case "full":
		return startCommand("grim", shotPath())
	case "area":
		return startCommand("sh", "-c", "grim -g \"$(slurp)\" "+shellQuote(shotPath()))
	case "clip":
		return startCommand("sh", "-c", "grim -g \"$(slurp)\" - | wl-copy")
	}
	return false
}

func shotPath() string {
	dir := config.HomeFile("Pictures")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "Screenshot-"+strconv.FormatInt(time.Now().Unix(), 10)+".png")
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func executeMusicAction(action modules.ActionSpec) bool {
	queue := modules.MusicQueue()
	if len(queue) == 0 {
		modules.SetStatus(false, "Music queue is empty")
		return false
	}
	switch action.Target {
	case "play":
		if _, err := cmdrunner.LookPath("mpv"); err == nil {
			return startCommand("mpv", queue...)
		}
		ok := true
		for _, path := range queue {
			if !startCommand("xdg-open", path) {
				ok = false
			}
		}
		return ok
	case "play-with":
		if len(action.Args) == 0 {
			return false
		}
		player := action.Args[0]
		if _, err := cmdrunner.LookPath(player); err != nil {
			modules.SetStatus(false, player+" not installed")
			return false
		}
		return startCommand(player, queue...)
	}
	return false
}

func executeWeatherAction(action modules.ActionSpec) bool {
	desc := ""
	if len(action.Args) > 0 {
		desc = action.Args[0]
	}
	if len(action.Args) > 1 && action.Args[1] == "copy" {
		if copyToClipboard(desc) {
			modules.SetStatus(true, "Weather copied: "+desc)
			return true
		}
		return false
	}
	text, err := currentWeather(action.Target)
	if err == nil && text != "" {
		if runCommand("notify-send", "Weather", text) {
			modules.SetStatus(true, text)
			return true
		}
		if copyToClipboard(text) {
			modules.SetStatus(true, "Weather copied: "+text)
			return true
		}
	}
	modules.SetStatus(false, "Weather failed; opening browser")
	return startCommand("xdg-open", action.Target)
}

func executePlayerAction(action modules.ActionSpec) bool {
	player := modules.MediaPlayer(modules.PlayerKind(action.Target))
	if player == "" {
		modules.SetStatus(false, action.Target+" player not detected")
		return false
	}
	args := append([]string{"--player=" + player}, action.Args...)
	if !runCommand("playerctl", args...) {
		modules.SetStatus(false, action.Target+" command failed: "+strings.Join(action.Args, " "))
		return false
	}
	modules.SetStatus(true, "")
	return true
}

func currentWeather(page string) (string, error) {
	client := http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(page + "?format=3")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(body)), nil
}

func executeStateAction(action modules.ActionSpec) bool {
	switch action.Target {
	case "file-buffer-add":
		if len(action.Args) == 0 {
			return false
		}
		modules.AddFileToBuffer(action.Args[0])
		return true
	case "file-buffer-clear":
		modules.ClearFileBuffer()
		return true
	case "music-queue-add":
		if len(action.Args) == 0 {
			return false
		}
		modules.AddMusicToQueue(action.Args[0])
		return true
	case "music-queue-clear":
		modules.ClearMusicQueue()
		return true
	case "file-undo":
		modules.RunLastFileUndo()
		return true
	}
	return false
}

func executeSyncAction(action modules.ActionSpec) bool {
	switch action.Target {
	case "import":
		if len(action.Args) < 2 {
			return false
		}
		if err := modules.ImportSyncZip(action.Args[0], action.Args[1]); err != nil {
			modules.SetStatus(false, "Sync import failed: "+err.Error())
			return false
		}
		modules.SetStatus(true, "Imported Spark settings from "+action.Args[0])
		return true
	case "export":
		if len(action.Args) < 3 {
			return false
		}
		if err := modules.ExportSyncZip(action.Args[0], action.Args[1:]); err != nil {
			modules.SetStatus(false, "Sync export failed: "+err.Error())
			return false
		}
		modules.SetStatus(true, "Exported Spark settings to "+action.Args[0])
		return true
	case "script":
		if len(action.Args) < 3 {
			return false
		}
		if err := modules.WriteSyncScript(action.Args[0], action.Args[1], action.Args[2]); err != nil {
			modules.SetStatus(false, "Git sync script failed: "+err.Error())
			return false
		}
		modules.SetStatus(true, "Git sync script created: "+action.Args[0])
		return true
	case "profile":
		if len(action.Args) < 3 {
			return false
		}
		if err := modules.WriteSyncProfile(action.Args[0], action.Args[1], action.Args[2]); err != nil {
			modules.SetStatus(false, "Sync profile failed: "+err.Error())
			return false
		}
		modules.SetStatus(true, "Sync profile created: "+action.Args[0])
		return true
	}
	return false
}

func executeSystemAction(action modules.ActionSpec) bool {
	switch action.Target {
	case "lock":
		for _, cmd := range [][]string{
			{"swaylock", "-f", "-c", "000000"},
			{"hyprlock"},
			{"gtklock"},
			{"loginctl", "lock-session"},
		} {
			if _, err := cmdrunner.LookPath(cmd[0]); err == nil {
				return startCommand(cmd[0], cmd[1:]...)
			}
		}
	case "logout":
		if u, err := user.Current(); err == nil && u.Username != "" {
			return startCommand("loginctl", "terminate-user", u.Username)
		}
		return startCommand("loginctl", "terminate-session", os.Getenv("XDG_SESSION_ID"))
	case "empty-trash":
		for _, dir := range []string{
			config.DataHomeFile("Trash", "files"),
			config.DataHomeFile("Trash", "info"),
		} {
			entries, err := os.ReadDir(dir)
			if err != nil {
				continue
			}
			for _, entry := range entries {
				os.RemoveAll(filepath.Join(dir, entry.Name()))
			}
		}
		return true
	}
	return false
}

func executeAppAction(action modules.ActionSpec) bool {
	if len(action.Args) == 0 {
		return false
	}
	app := apps.App{Name: action.Target, Exec: action.Args[0]}
	if len(action.Args) > 1 {
		app.Icon = action.Args[1]
	}
	history.Record(app.Name)
	if err := app.Launch(); err != nil {
		modules.SetStatus(false, "Failed to launch "+app.Name+": "+err.Error())
		return false
	}
	return true
}
