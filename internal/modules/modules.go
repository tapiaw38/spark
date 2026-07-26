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
	Action          func()
}
