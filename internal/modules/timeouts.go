package modules

import "time"

const (
	DictionaryLookupTimeout = 200 * time.Millisecond
	SpellCheckTimeout       = 300 * time.Millisecond
	MusicProbeTimeout       = 700 * time.Millisecond
	FileSearchTimeout       = 900 * time.Millisecond
	DictionaryHTTPTimeout   = 1 * time.Second
	ThumbnailHTTPTimeout    = 4 * time.Second
	YouTubeSearchTimeout    = 5 * time.Second
)

const (
	FileSearchDebounce = 250 * time.Millisecond
	FileCacheTTL       = 5 * time.Second
	StatusClearDelay   = 3 * time.Second
)
