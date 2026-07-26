package modules

import (
	"context"
	"time"
)

type Searcher interface {
	Search(query string) []Result
}

type SearchFunc func(query string) []Result

func (f SearchFunc) Search(query string) []Result { return f(query) }

type Registry struct {
	searchers []Searcher
	fallback  Searcher
}

func NewRegistry(searchers ...Searcher) *Registry {
	return &Registry{searchers: searchers}
}

func (r *Registry) Append(searchers ...Searcher) *Registry {
	r.searchers = append(r.searchers, searchers...)
	return r
}

func (r *Registry) WithFallback(s Searcher) *Registry {
	r.fallback = s
	return r
}

func (r *Registry) Search(query string) []Result {
	var out []Result
	for _, s := range r.searchers {
		out = append(out, s.Search(query)...)
	}
	if len(out) == 0 && r.fallback != nil {
		out = append(out, r.fallback.Search(query)...)
	}
	return out
}

func DefaultRegistry() *Registry {
	return NewRegistry(
		SearchFunc(ShellSearch),
		SearchFunc(HelpSearch),
		SearchFunc(LargeTypeSearch),
		SearchFunc(RecentSearch),
		SearchFunc(ContactsSearch),
		SearchFunc(EmailSearch),
		SearchFunc(StatsSearch),
		SearchFunc(SyncSearch),
		SearchFunc(StatusSearch),
		SearchFunc(SnippetSearch),
		SearchFunc(DictionarySearch),
		SearchFunc(SpellSearch),
		SearchFunc(CalcSearch),
		SearchFunc(ClipboardSearch),
		SearchFunc(WebSearch),
		SearchFunc(SystemSearch),
		SearchFunc(SpotifySearch),
		SearchFunc(MusicSearch),
		SearchFunc(FileBufferSearch),
		SearchFunc(FileSearch),
		SearchFunc(EmojiSearch),
		SearchFunc(DevToolsSearch),
		SearchFunc(UnitSearch),
		SearchFunc(SSHSearch),
		SearchFunc(KillSearch),
		SearchFunc(ScreenshotSearch),
		SearchFunc(WindowSearch),
		SearchFunc(TimerSearch),
		SearchFunc(WeatherSearch),
		SearchFunc(PassSearch),
		SearchFunc(BookmarksSearch),
	)
}

type AsyncSearcher struct {
	Name     string
	Match    func(query string) bool
	Loading  func(query string) []Result
	Search   func(ctx context.Context, query string) []Result
	Ready    func(query string) bool
	Debounce time.Duration
}

func DefaultAsyncSearchers() []AsyncSearcher {
	return []AsyncSearcher{
		{
			Name:    "youtube",
			Match:   IsYouTubeQuery,
			Loading: YouTubeLoading,
			Search:  func(_ context.Context, q string) []Result { return YouTubeSearch(q) },
		},
		{
			Name:     "file",
			Match:    IsFileQuery,
			Loading:  FileLoading,
			Search:   FileSearchContext,
			Ready:    IsFileQueryReady,
			Debounce: 250 * time.Millisecond,
		},
		{
			Name:    "weather",
			Match:   IsWeatherQuery,
			Loading: WeatherLoading,
			Search:  func(_ context.Context, q string) []Result { return WeatherSearchAsync(q) },
		},
		{
			Name:    "navigation",
			Match:   IsNavQuery,
			Loading: NavigationLoading,
			Search:  func(_ context.Context, q string) []Result { return NavigationSearch(q) },
		},
		{
			Name:    "destination-picker",
			Match:   IsPickQuery,
			Loading: NavigationLoading,
			Search:  func(_ context.Context, q string) []Result { return DestinationPickerSearch(q) },
		},
	}
}
