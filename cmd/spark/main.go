package main

import (
	"os"

	"github.com/diamondburned/gotk4-layer-shell/pkg/gtklayershell"
	"github.com/diamondburned/gotk4/pkg/gdk/v3"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/tapiaw38/spark/internal/apps"
	"github.com/tapiaw38/spark/internal/config"
)

func main() {
	if runSubcommand() {
		return
	}

	gtk.Init()
	config.Load()
	allApps = apps.Load()
	go preloadIcons()

	window := gtk.NewWindow(gtk.WindowToplevel)
	window.SetTitle("Spark")
	window.SetDefaultSize(config.Current.Width, -1)
	window.SetSizeRequest(config.Current.Width, -1)
	window.SetDecorated(false)

	// Layer shell setup
	gtklayershell.InitForWindow(window)
	gtklayershell.SetLayer(window, gtklayershell.LayerShellLayerTop)
	gtklayershell.SetKeyboardMode(window, gtklayershell.LayerShellKeyboardModeExclusive)
	gtklayershell.SetAnchor(window, gtklayershell.LayerShellEdgeTop, true)
	gtklayershell.SetMargin(window, gtklayershell.LayerShellEdgeTop, config.Current.MarginTop)

	// Main container
	mainBox = gtk.NewBox(gtk.OrientationVertical, 8)
	mainBox.SetMarginStart(12)
	mainBox.SetMarginEnd(12)
	mainBox.SetMarginTop(12)
	mainBox.SetMarginBottom(12)

	// Create spotify view (hidden initially)
	createSpotifyView()

	// Search entry
	entry := gtk.NewEntry()
	searchEntry = entry
	entry.SetPlaceholderText("Search apps, ;snippet, define word, > shell...")
	entry.SetName("spark-entry")

	// Results list
	listBox = gtk.NewListBox()
	listBox.SetName("spark-results")
	listBox.SetSelectionMode(gtk.SelectionSingle)
	listBox.SetCanFocus(false)

	// Scroll container for results - fixed height for 6 rows
	resultsScroll = gtk.NewScrolledWindow(nil, nil)
	resultsScroll.SetCanFocus(false)
	resultsScroll.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	resultsScroll.SetSizeRequest(-1, 288) // 6 rows * 48px
	resultsScroll.SetNoShowAll(true)
	resultsScroll.Add(listBox)

	buildPreviewPane()

	listBox.Connect("row-selected", func(_ *gtk.ListBox, row *gtk.ListBoxRow) {
		if quickLookActive {
			updatePreview(row)
		} else {
			hidePreview()
		}
	})

	entry.Connect("changed", onSearchChanged)

	window.Connect("key-press-event", func(_ *gtk.Window, event *gdk.Event) bool {
		return onKeyPress(event)
	})

	window.Connect("destroy", gtk.MainQuit)

	mainBox.PackStart(entry, false, false, 0)
	mainBox.PackStart(resultsScroll, false, false, 0)
	mainBox.PackStart(spotifyView, false, false, 0)
	mainBox.PackStart(previewBox, false, false, 0)
	window.Add(mainBox)

	loadCSS()

	window.ShowAll()
	spotifyView.Hide()   // Hide spotify view initially
	resultsScroll.Hide() // Hide results list initially
	entry.GrabFocus()

	gtk.Main()
	os.Exit(0)
}
