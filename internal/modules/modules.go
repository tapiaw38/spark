package modules

type ResultType string

const (
	TypeApp              ResultType = "app"
	TypeBookmark         ResultType = "bookmark"
	TypeCalc             ResultType = "calc"
	TypeClipboard        ResultType = "clipboard"
	TypeContact          ResultType = "contact"
	TypeDevtool          ResultType = "devtool"
	TypeDictionary       ResultType = "dictionary"
	TypeDirectory        ResultType = "directory"
	TypeEmail            ResultType = "email"
	TypeEmoji            ResultType = "emoji"
	TypeFile             ResultType = "file"
	TypeFileAction       ResultType = "file-action"
	TypeFileBuffer       ResultType = "file-buffer"
	TypeFileBufferAction ResultType = "file-buffer-action"
	TypeFileOp           ResultType = "file-op"
	TypeHelp             ResultType = "help"
	TypeKill             ResultType = "kill"
	TypeLargeType        ResultType = "large-type"
	TypeMediaControl     ResultType = "media-control"
	TypeMusic            ResultType = "music"
	TypeNavigation       ResultType = "navigation"
	TypePass             ResultType = "pass"
	TypeRecent           ResultType = "recent"
	TypeScreenshot       ResultType = "screenshot"
	TypeShell            ResultType = "shell"
	TypeSnippet          ResultType = "snippet"
	TypeSpell            ResultType = "spell"
	TypeSpotify          ResultType = "spotify"
	TypeSSH              ResultType = "ssh"
	TypeStats            ResultType = "stats"
	TypeStatus           ResultType = "status"
	TypeSync             ResultType = "sync"
	TypeSystem           ResultType = "system"
	TypeTimer            ResultType = "timer"
	TypeUndo             ResultType = "undo"
	TypeWeather          ResultType = "weather"
	TypeWeb              ResultType = "web"
	TypeWindow           ResultType = "window"
	TypeWorkspace        ResultType = "workspace"
	TypeYouTube          ResultType = "youtube"
	TypeYouTubePlayer    ResultType = "youtube-player"
)

func AllResultTypes() []ResultType {
	return []ResultType{
		TypeApp, TypeBookmark, TypeCalc, TypeClipboard, TypeContact,
		TypeDevtool, TypeDictionary, TypeDirectory, TypeEmail, TypeEmoji,
		TypeFile, TypeFileAction, TypeFileBuffer, TypeFileBufferAction,
		TypeFileOp, TypeHelp, TypeKill, TypeLargeType, TypeMediaControl,
		TypeMusic, TypeNavigation, TypePass, TypeRecent, TypeScreenshot,
		TypeShell, TypeSnippet, TypeSpell, TypeSpotify, TypeSSH,
		TypeStats, TypeStatus, TypeSync, TypeSystem, TypeTimer, TypeUndo,
		TypeWeather, TypeWeb, TypeWindow, TypeWorkspace, TypeYouTube,
		TypeYouTubePlayer,
	}
}

type ActionKind string

const (
	ActionNone        ActionKind = ""
	ActionOpen        ActionKind = "open"
	ActionCopy        ActionKind = "copy"
	ActionStart       ActionKind = "start"
	ActionRun         ActionKind = "run"
	ActionTerminal    ActionKind = "terminal"
	ActionEmail       ActionKind = "email"
	ActionEmailWindow ActionKind = "email-window"
	ActionFile        ActionKind = "file"
	ActionClipboard   ActionKind = "clipboard"
	ActionScreenshot  ActionKind = "screenshot"
	ActionMusic       ActionKind = "music"
	ActionPaste       ActionKind = "paste"
	ActionWeather     ActionKind = "weather"
	ActionPlayer      ActionKind = "player"
	ActionState       ActionKind = "state"
	ActionSync        ActionKind = "sync"
	ActionSystem      ActionKind = "system"
	ActionApp         ActionKind = "app-launch"
)

func AllActionKinds() []ActionKind {
	return []ActionKind{
		ActionOpen, ActionCopy, ActionStart, ActionRun, ActionTerminal,
		ActionEmail, ActionEmailWindow, ActionFile, ActionClipboard,
		ActionScreenshot, ActionMusic, ActionPaste, ActionWeather,
		ActionPlayer, ActionState, ActionSync, ActionSystem, ActionApp,
	}
}

type ActionSpec struct {
	Kind          ActionKind
	Target        string
	Args          []string
	SuccessStatus string
}

func OpenAction(path string) ActionSpec {
	return ActionSpec{Kind: ActionOpen, Target: path}
}

func OpenPathsAction(paths ...string) ActionSpec {
	if len(paths) == 0 {
		return ActionSpec{}
	}
	return ActionSpec{Kind: ActionOpen, Target: paths[0], Args: paths[1:]}
}

func CopyAction(text string) ActionSpec {
	return ActionSpec{Kind: ActionCopy, Target: text}
}

func StartAction(name string, args ...string) ActionSpec {
	return ActionSpec{Kind: ActionStart, Target: name, Args: args}
}

func RunAction(name string, args ...string) ActionSpec {
	return ActionSpec{Kind: ActionRun, Target: name, Args: args}
}

func TerminalAction(command string) ActionSpec {
	return ActionSpec{Kind: ActionTerminal, Target: command}
}

func EmailAction(to, subject, body string, attachments ...string) ActionSpec {
	args := append([]string{subject, body}, attachments...)
	return ActionSpec{Kind: ActionEmail, Target: to, Args: args}
}

func EmailWindowAction(to, subject, body string) ActionSpec {
	return ActionSpec{Kind: ActionEmailWindow, Target: to, Args: []string{subject, body}}
}

func FileAction(op string, args ...string) ActionSpec {
	return ActionSpec{Kind: ActionFile, Target: op, Args: args}
}

func ClipboardHistoryAction(id string, paste bool) ActionSpec {
	args := []string{}
	if paste {
		args = []string{"paste"}
	}
	return ActionSpec{Kind: ActionClipboard, Target: id, Args: args}
}

func ScreenshotAction(mode string) ActionSpec {
	return ActionSpec{Kind: ActionScreenshot, Target: mode}
}

func MusicAction(mode string, args ...string) ActionSpec {
	return ActionSpec{Kind: ActionMusic, Target: mode, Args: args}
}

func PasteAction(text string) ActionSpec {
	return ActionSpec{Kind: ActionPaste, Target: text}
}

func WeatherAction(page, desc string, copyOnEnter bool) ActionSpec {
	args := []string{desc}
	if copyOnEnter {
		args = append(args, "copy")
	}
	return ActionSpec{Kind: ActionWeather, Target: page, Args: args}
}

func StateAction(op string, args ...string) ActionSpec {
	return ActionSpec{Kind: ActionState, Target: op, Args: args}
}

func SyncAction(op string, args ...string) ActionSpec {
	return ActionSpec{Kind: ActionSync, Target: op, Args: args}
}

func SystemAction(op string) ActionSpec {
	return ActionSpec{Kind: ActionSystem, Target: op}
}

func AppAction(name, exec, icon string) ActionSpec {
	return ActionSpec{Kind: ActionApp, Target: name, Args: []string{exec, icon}}
}

func (a ActionSpec) WithStatus(status string) ActionSpec {
	a.SuccessStatus = status
	return a
}

func (a ActionSpec) IsZero() bool {
	return a.Kind == ActionNone
}

type Result struct {
	Type            ResultType
	Title           string
	Desc            string
	Icon            string
	IconText        string
	Preview         string
	PreviewImage    string
	PreviewImageURL string
	Data            string
	KeepOpen        bool
	Confirm         bool
	NavigateQuery   string
	ActionSpec      ActionSpec
}
