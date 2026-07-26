package main

import (
	"sync/atomic"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/tapiaw38/spark/internal/modules"
)

const (
	resultBatchSize   = 20
	resultsPageSize   = 60
	loadMoreThreshold = 480
)

func clearResultRows() {
	isClearing = true
	defer func() { isClearing = false }()

	for {
		row := listBox.RowAtIndex(0)
		if row == nil {
			return
		}
		listBox.Remove(row)
	}
}

func setResults(results []modules.Result) {
	clearResultRows()
	currentResults = results
	resultsRenderedN = 0
	resultsLoadingMore = false

	if len(results) == 0 {
		resultsScroll.Hide()
		return
	}

	listBox.Show()
	resultsScroll.Show()

	scheduleResultRendering(atomic.AddUint64(&resultsRenderVersion, 1))
}

func renderNextResultChunk() bool {
	if resultsRenderedN >= len(currentResults) {
		return false
	}
	start := resultsRenderedN
	end := min(start+resultBatchSize, len(currentResults))

	for _, r := range currentResults[start:end] {
		row := createResultRow(r)
		listBox.Add(row)
		row.ShowAll()
	}
	resultsRenderedN = end

	if start == 0 {
		selectFirstRow()
	}
	return true
}

func selectFirstRow() {
	if first := listBox.RowAtIndex(0); first != nil {
		listBox.SelectRow(first)
	}
}

func scheduleResultRendering(version uint64) {
	if atomic.LoadUint64(&resultsRenderVersion) != version {
		return
	}
	if !renderNextResultChunk() {
		return
	}
	if resultsRenderedN < len(currentResults) && resultsRenderedN < resultsPageSize {
		glib.IdleAdd(func() { scheduleResultRendering(version) })
	}
}

func maybeLoadMoreResults() {
	if resultsScroll == nil || resultsLoadingMore || resultsRenderedN >= len(currentResults) {
		return
	}
	adj := resultsScroll.VAdjustment()
	if adj == nil {
		return
	}
	if adj.Value()+adj.PageSize() < adj.Upper()-loadMoreThreshold {
		return
	}

	resultsLoadingMore = true
	glib.IdleAdd(func() {
		resultsLoadingMore = false
		renderNextResultChunk()
	})
}
