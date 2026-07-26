package main

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/tapiaw38/spark/internal/apps"
	"github.com/tapiaw38/spark/internal/modules"
)

const quickAppSearchQueryLen = 2

var (
	searchRegistry *modules.Registry
	asyncSearchers []modules.AsyncSearcher
)

func initSearch() {
	searchRegistry = modules.DefaultRegistry().
		Append(modules.SearchFunc(appSearch)).
		WithFallback(modules.SearchFunc(modules.FallbackWebSearch))
	asyncSearchers = modules.DefaultAsyncSearchers()
}

func appSearch(query string) []modules.Result {
	var matches []apps.App
	if len(query) <= quickAppSearchQueryLen {
		matches = apps.QuickSearch(allApps, query)
	} else {
		matches = apps.Search(allApps, query)
	}

	out := make([]modules.Result, 0, len(matches))
	for _, app := range matches {
		out = append(out, modules.Result{
			Type:       modules.TypeApp,
			Title:      app.Name,
			Icon:       app.Icon,
			ActionSpec: modules.AppAction(app.Name, app.Exec, app.Icon),
		})
	}
	return out
}

func updateResults(query string) {
	version := atomic.AddUint64(&searchVersion, 1)
	cancelPendingSearch()
	clearResultRows()
	if quickLookActive {
		cancelPreviewLoad()
	} else {
		hidePreview()
	}

	if query == "" {
		currentResults = nil
		resultsScroll.Hide()
		hideSpotifyView()
		return
	}

	if handlePlayerQuery(query) {
		return
	}
	hideSpotifyView()

	for _, s := range asyncSearchers {
		if s.Match(query) {
			runAsyncSearch(s, query, version)
			return
		}
	}

	if results := modules.FileOperationSearch(query); results != nil {
		setResults(results)
		return
	}

	currentResults = searchRegistry.Search(query)
	setResults(currentResults)
}

func handlePlayerQuery(query string) bool {
	if modules.IsSpotifyQuery(query) {
		showPlayerView(modules.PlayerSpotify)
		return true
	}
	if results := modules.YouTubePlayerStatus(query); results != nil {
		setResults(results)
		return true
	}
	if modules.IsYouTubePlayerQuery(query) {
		showPlayerView(modules.PlayerYouTube)
		return true
	}
	return false
}

func runAsyncSearch(s modules.AsyncSearcher, query string, version uint64) {
	setResults(s.Loading(query))
	if s.Ready != nil && !s.Ready(query) {
		return
	}

	ctx := newSearchContext()
	go func() {
		if s.Debounce > 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(s.Debounce):
			}
		}
		if atomic.LoadUint64(&searchVersion) != version {
			return
		}
		results := s.Search(ctx, query)
		glib.IdleAdd(func() {
			if atomic.LoadUint64(&searchVersion) != version {
				return
			}
			setResults(results)
		})
	}()
}

func newSearchContext() context.Context {
	searchCancelMu.Lock()
	defer searchCancelMu.Unlock()

	if searchCancel != nil {
		searchCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	searchCancel = cancel
	return ctx
}

func cancelPendingSearch() {
	searchCancelMu.Lock()
	defer searchCancelMu.Unlock()

	if searchCancel != nil {
		searchCancel()
		searchCancel = nil
	}
}

func showSelectedFileActions() {
	if inSpotifyMode || listBox == nil {
		return
	}
	selected := listBox.SelectedRow()
	if selected == nil {
		return
	}
	idx := selected.Index()
	if idx < 0 || idx >= len(currentResults) {
		return
	}
	path := currentResults[idx].FilePath()
	if path == "" {
		return
	}
	inActionMode = true
	setResults(modules.FileActions(path))
}
