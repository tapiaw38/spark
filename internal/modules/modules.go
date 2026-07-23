package modules

// Result represents a search result from any module
type Result struct {
	Type            string // "app", "calc", "web", "system", "file", "shell", "spotify"
	Title           string
	Desc            string
	Icon            string
	Preview         string // Optional preview text
	PreviewImage    string // Optional preview image path
	PreviewImageURL string // Optional remote preview image URL
	Data            string // Optional module-specific opaque data
	KeepOpen        bool   // Keep launcher open after action
	Confirm         bool   // Require confirmation before action
	NavigateQuery   string // Optional query to set instead of quitting
	Action          func() // Execute when selected
}
