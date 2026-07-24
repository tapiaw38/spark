package main

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/tapiaw38/spark/internal/apps"
	"github.com/tapiaw38/spark/internal/history"
	"github.com/tapiaw38/spark/internal/modules"
)

func onSearchChanged() {
	query := searchEntry.Text()
	inActionMode = false
	if debounceTimer != nil {
		debounceTimer.Stop()
	}
	if len(query) <= 1 {
		updateResults(query)
		return
	}
	debounceTimer = time.AfterFunc(50*time.Millisecond, func() {
		glib.IdleAdd(func() { updateResults(query) })
	})
}

func updateResults(query string) {
	version := atomic.AddUint64(&searchVersion, 1)
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

	// Check for Spotify mode
	if modules.IsSpotifyQuery(query) {
		showPlayerView(modules.PlayerSpotify)
		return
	}
	if modules.IsYouTubePlayerQuery(query) {
		showPlayerView(modules.PlayerYouTube)
		return
	}
	hideSpotifyView()

	if modules.IsYouTubeQuery(query) {
		setResults(modules.YouTubeLoading(query))
		go func(q string, v uint64) {
			results := modules.YouTubeSearch(q)
			glib.IdleAdd(func() {
				if atomic.LoadUint64(&searchVersion) != v {
					return
				}
				setResults(results)
			})
		}(query, version)
		return
	}

	if modules.IsFileQuery(query) {
		setResults(modules.FileLoading(query))
		fileSearchMu.Lock()
		if fileSearchStop != nil {
			fileSearchStop()
			fileSearchStop = nil
		}
		ctx, cancel := context.WithCancel(context.Background())
		fileSearchStop = cancel
		fileSearchMu.Unlock()
		if !modules.IsFileQueryReady(query) {
			return
		}
		go func(q string, v uint64) {
			time.Sleep(250 * time.Millisecond)
			if atomic.LoadUint64(&searchVersion) != v {
				cancel()
				return
			}
			results := modules.FileSearchContext(ctx, q)
			glib.IdleAdd(func() {
				if atomic.LoadUint64(&searchVersion) != v {
					return
				}
				fileSearchMu.Lock()
				fileSearchStop = nil
				fileSearchMu.Unlock()
				setResults(results)
			})
		}(query, version)
		return
	}
	fileSearchMu.Lock()
	if fileSearchStop != nil {
		fileSearchStop()
		fileSearchStop = nil
	}
	fileSearchMu.Unlock()

	results, terminal := modules.SearchAll(query)
	if terminal {
		setResults(results)
		return
	}

	results = append(results, appResults(query)...)
	if len(results) == 0 {
		results = modules.FallbackWebSearch(query)
	}
	setResults(results)
}

func appResults(query string) []modules.Result {
	var matches []apps.App
	if len(query) <= 2 {
		matches = apps.QuickSearch(allApps, query)
	} else {
		matches = apps.Search(allApps, query)
	}
	out := make([]modules.Result, 0, len(matches))
	for _, app := range matches {
		a := app
		out = append(out, modules.Result{
			Type:  "app",
			Title: a.Name,
			Icon:  a.Icon,
			Action: func() {
				history.Record(a.Name)
				apps.Launch(a)
			},
		})
	}
	return out
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
	path := modules.GetFilePath(currentResults[idx])
	if path == "" {
		return
	}
	inActionMode = true
	setResults(modules.FileActions(path))
}

func clearResultRows() {
	isClearing = true
	for {
		row := listBox.RowAtIndex(0)
		if row == nil {
			break
		}
		listBox.Remove(row)
	}
	isClearing = false
}

func setResults(results []modules.Result) {
	clearResultRows()
	currentResults = results

	maxScrollResults := 50
	if len(currentResults) > maxScrollResults {
		currentResults = currentResults[:maxScrollResults]
	}

	// Create rows
	for _, r := range currentResults {
		row := createResultRow(r)
		listBox.Add(row)
		row.ShowAll()
	}

	if len(currentResults) == 0 {
		resultsScroll.Hide()
		return
	}

	listBox.Show()
	resultsScroll.Show()

	if first := listBox.RowAtIndex(0); first != nil {
		listBox.SelectRow(first)
	}
}

func selectNext() {
	list := listBox
	if inSpotifyMode && spotifyList != nil {
		list = spotifyList
	}
	if list == nil {
		return
	}
	selected := list.SelectedRow()
	if selected == nil {
		// Select first if none selected
		if first := list.RowAtIndex(0); first != nil {
			list.SelectRow(first)
			scrollToRow(first)
		}
		return
	}
	idx := selected.Index()
	if next := list.RowAtIndex(idx + 1); next != nil {
		list.SelectRow(next)
		scrollToRow(next)
	}
}

func selectPrev() {
	list := listBox
	if inSpotifyMode && spotifyList != nil {
		list = spotifyList
	}
	if list == nil {
		return
	}
	selected := list.SelectedRow()
	if selected == nil {
		return
	}
	idx := selected.Index()
	if idx > 0 {
		if prev := list.RowAtIndex(idx - 1); prev != nil {
			list.SelectRow(prev)
			scrollToRow(prev)
		}
	}
}

func scrollToRow(row *gtk.ListBoxRow) {
	if row == nil || resultsScroll == nil {
		return
	}
	adj := resultsScroll.VAdjustment()
	if adj == nil {
		return
	}

	// Get row allocation
	alloc := row.Allocation()
	rowY := float64(alloc.Y())
	rowH := float64(alloc.Height())

	// Get scroll viewport
	scrollY := adj.Value()
	scrollH := adj.PageSize()

	// Scroll down if row below viewport
	if rowY+rowH > scrollY+scrollH {
		adj.SetValue(rowY + rowH - scrollH)
	}
	// Scroll up if row above viewport
	if rowY < scrollY {
		adj.SetValue(rowY)
	}
}

func executeSelected() {
	if inSpotifyMode && spotifyList != nil {
		selected := spotifyList.SelectedRow()
		if selected == nil {
			return
		}
		idx := selected.Index()
		if idx < 0 {
			return
		}
		ctrls := modules.SpotifyControls()
		if idx < len(ctrls) {
			ctrls[idx].Action()
			glib.TimeoutAdd(300, func() bool { refreshSpotifyInfo(); return false })
		}
		return // Don't quit, stay in spotify mode
	}

	selected := listBox.SelectedRow()
	if selected == nil {
		return
	}

	idx := selected.Index()
	if idx < 0 || idx >= len(currentResults) {
		return
	}

	r := currentResults[idx]
	if r.NavigateQuery != "" {
		searchEntry.SetText(r.NavigateQuery)
		searchEntry.SetPosition(-1)
		return
	}
	if r.Confirm {
		if !confirmAction(r) {
			return
		}
	}
	if r.Action != nil {
		r.Action()
	}
	if r.KeepOpen {
		if inActionMode {
			inActionMode = false
			updateResults(searchEntry.Text())
		}
		return
	}
	gtk.MainQuit()
}

func confirmAction(r modules.Result) bool {
	dialog := gtk.NewMessageDialog(nil, gtk.DialogModal, gtk.MessageWarning, gtk.ButtonsOKCancel)
	dialog.SetMarkup("<b>" + r.Title + "</b>\n" + r.Desc)
	dialog.ShowAll()
	response := dialog.Run()
	dialog.Destroy()
	return gtk.ResponseType(response) == gtk.ResponseOK
}
