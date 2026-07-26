package main

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/tapiaw38/spark/internal/modules"
)

func activeList() *gtk.ListBox {
	if inSpotifyMode && spotifyList != nil {
		return spotifyList
	}
	return listBox
}

func selectNext() {
	list := activeList()
	if list == nil {
		return
	}
	selected := list.SelectedRow()
	if selected == nil {
		selectRow(list, list.RowAtIndex(0))
		return
	}

	idx := selected.Index()
	next := list.RowAtIndex(idx + 1)
	if next == nil && list == listBox && resultsRenderedN < len(currentResults) {
		renderNextResultChunk()
		next = list.RowAtIndex(idx + 1)
	}
	selectRow(list, next)
}

func selectPrev() {
	list := activeList()
	if list == nil {
		return
	}
	selected := list.SelectedRow()
	if selected == nil || selected.Index() <= 0 {
		return
	}
	selectRow(list, list.RowAtIndex(selected.Index()-1))
}

func selectRow(list *gtk.ListBox, row *gtk.ListBoxRow) {
	if row == nil {
		return
	}
	list.SelectRow(row)
	scrollToRow(row)
}

func scrollToRow(row *gtk.ListBoxRow) {
	if row == nil || resultsScroll == nil {
		return
	}
	adj := resultsScroll.VAdjustment()
	if adj == nil {
		return
	}

	alloc := row.Allocation()
	rowTop := float64(alloc.Y())
	rowBottom := rowTop + float64(alloc.Height())
	viewTop := adj.Value()
	viewBottom := viewTop + adj.PageSize()

	switch {
	case rowBottom > viewBottom:
		adj.SetValue(rowBottom - adj.PageSize())
	case rowTop < viewTop:
		adj.SetValue(rowTop)
	}
}

func executeSelected() {
	if inSpotifyMode && spotifyList != nil {
		executePlayerControl()
		return
	}

	selected := listBox.SelectedRow()
	r, ok := selectedResult(selected)
	if !ok {
		return
	}

	if r.NavigateQuery != "" {
		searchEntry.SetText(r.NavigateQuery)
		searchEntry.SetPosition(-1)
		return
	}
	if r.Confirm && !confirmAction(r) {
		return
	}
	executeActionSpec(r.ActionSpec)
	if !r.KeepOpen {
		gtk.MainQuit()
		return
	}
	if inActionMode {
		inActionMode = false
		updateResults(searchEntry.Text())
	}
}

func executePlayerControl() {
	selected := spotifyList.SelectedRow()
	if selected == nil || selected.Index() < 0 {
		return
	}
	controls := modules.PlayerControls(playerMode)
	if selected.Index() >= len(controls) {
		return
	}
	runPlayerResultAction(controls[selected.Index()])
}

func confirmAction(r modules.Result) bool {
	dialog := gtk.NewMessageDialog(nil, gtk.DialogModal, gtk.MessageWarning, gtk.ButtonsOKCancel)
	dialog.SetMarkup("<b>" + r.Title + "</b>\n" + r.Desc)
	dialog.ShowAll()
	response := dialog.Run()
	dialog.Destroy()
	return gtk.ResponseType(response) == gtk.ResponseOK
}
