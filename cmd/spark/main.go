package main

import (
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/diamondburned/gotk4-layer-shell/pkg/gtklayershell"
	"github.com/diamondburned/gotk4/pkg/gdk/v3"
	"github.com/diamondburned/gotk4/pkg/gdkpixbuf/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/diamondburned/gotk4/pkg/pango"
	"github.com/tapiaw38/spark/internal/apps"
	"github.com/tapiaw38/spark/internal/config"
	"github.com/tapiaw38/spark/internal/history"
	"github.com/tapiaw38/spark/internal/modules"
)

var (
	allApps        []apps.App
	currentResults []modules.Result
	listBox        *gtk.ListBox
	resultsScroll  *gtk.ScrolledWindow
	previewBox     *gtk.Box
	previewImage   *gtk.Image
	previewLabel   *gtk.Label
	searchMu       sync.Mutex
	searchVersion  uint64
	previewVersion uint64
	debounceTimer  *time.Timer
	iconCache      = make(map[string]*gdkpixbuf.Pixbuf)
	iconCacheMu    sync.RWMutex

	// Spotify view widgets
	spotifyView     *gtk.Box
	spotifyArtBig   *gtk.Image
	spotifyArtSmall *gtk.Image
	spotifyTitle    *gtk.Label
	spotifyArtist   *gtk.Label
	spotifyAlbum    *gtk.Label
	spotifyStatus   *gtk.Label
	spotifyList     *gtk.ListBox
	inSpotifyMode   bool
	mainBox         *gtk.Box
	isClearing      bool
)

const (
	dragInfoURI uint = iota + 1
	dragInfoText
)

func main() {
	// Handle --setup flag
	if len(os.Args) > 1 && os.Args[1] == "--setup" {
		config.Load()
		sparkPath, _ := os.Executable()
		if err := config.SetupHotkey(sparkPath); err != nil {
			os.Stderr.WriteString("Failed to setup hotkey: " + err.Error() + "\n")
			os.Exit(1)
		}
		os.Stdout.WriteString("Hotkey configured: " + config.Current.Hotkey + "\nRestart mango to apply.\n")
		os.Exit(0)
	}

	gtk.Init()

	// Load config
	config.Load()

	// Load apps
	allApps = apps.Load()

	// Preload icons for first 20 apps (background, before first keystroke)
	go func() {
		theme := gtk.IconThemeGetDefault()
		for i, app := range allApps {
			if i >= 20 {
				break
			}
			if app.Icon != "" && app.Icon[0] != '/' {
				if pb, err := theme.LoadIcon(app.Icon, 24, gtk.IconLookupForceSize); err == nil {
					iconCacheMu.Lock()
					iconCache[app.Icon] = pb
					iconCacheMu.Unlock()
				}
			}
		}
	}()

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
	entry.SetPlaceholderText("Search apps, ;snippet, define word, > shell...")
	entry.SetName("spark-entry")

	// Results list
	listBox = gtk.NewListBox()
	listBox.SetName("spark-results")
	listBox.SetSelectionMode(gtk.SelectionSingle)

	// Scroll container for results - fixed height for 6 rows
	resultsScroll = gtk.NewScrolledWindow(nil, nil)
	resultsScroll.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	resultsScroll.SetSizeRequest(-1, 288) // 6 rows * 48px
	resultsScroll.SetNoShowAll(true)
	resultsScroll.Add(listBox)

	// Preview pane - fixed size to prevent layout jumps
	previewBox = gtk.NewBox(gtk.OrientationVertical, 8)
	previewBox.SetName("spark-preview")
	previewBox.SetSizeRequest(-1, 200)
	previewBox.SetNoShowAll(true)

	previewImage = gtk.NewImage()
	previewImage.SetName("spark-preview-image")
	previewImage.SetSizeRequest(220, 180)
	previewImage.SetNoShowAll(true)

	previewLabel = gtk.NewLabel("")
	previewLabel.SetName("spark-preview-label")
	previewLabel.SetXAlign(0)
	previewLabel.SetLineWrap(true)
	previewLabel.SetLineWrapMode(pango.WrapWordChar)
	previewLabel.SetMaxWidthChars(40)
	previewLabel.SetSizeRequest(300, -1) // Fixed width
	previewLabel.SetNoShowAll(true)
	previewBox.PackStart(previewImage, false, false, 0)
	previewBox.PackStart(previewLabel, false, false, 0)

	// Update preview on selection change
	listBox.Connect("row-selected", func(_ *gtk.ListBox, row *gtk.ListBoxRow) {
		updatePreview(row)
	})

	// Search on typing with debounce
	entry.Connect("changed", func() {
		query := entry.Text()

		// Cancel previous timer
		if debounceTimer != nil {
			debounceTimer.Stop()
		}

		// Immediate update for empty or very short
		if len(query) <= 1 {
			updateResults(query)
			return
		}

		// Debounce 50ms for longer queries
		debounceTimer = time.AfterFunc(50*time.Millisecond, func() {
			glib.IdleAdd(func() {
				updateResults(query)
			})
		})
	})

	// Keyboard navigation
	window.Connect("key-press-event", func(_ *gtk.Window, event *gdk.Event) bool {
		keyEvent := event.AsKey()
		switch keyEvent.Keyval() {
		case gdk.KEY_Escape:
			gtk.MainQuit()
			return true
		case gdk.KEY_Down:
			selectNext()
			return true
		case gdk.KEY_Up:
			selectPrev()
			return true
		case gdk.KEY_Return:
			executeSelected()
			return true
		}
		return false
	})

	window.Connect("destroy", func() {
		gtk.MainQuit()
	})

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

func updateResults(query string) {
	version := atomic.AddUint64(&searchVersion, 1)
	clearResultRows()
	hidePreview()

	if query == "" {
		currentResults = nil
		resultsScroll.Hide()
		hideSpotifyView()
		return
	}

	// Check for Spotify mode
	if modules.IsSpotifyQuery(query) {
		showSpotifyView()
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

	// Collect results from all modules (priority order)
	currentResults = nil

	// 1. Shell commands (> prefix)
	currentResults = append(currentResults, modules.ShellSearch(query)...)

	// 2. Snippets (; prefix or "snip")
	currentResults = append(currentResults, modules.SnippetSearch(query)...)

	// 3. Dictionary (define/def prefix)
	currentResults = append(currentResults, modules.DictionarySearch(query)...)

	// 4. Calculator
	currentResults = append(currentResults, modules.CalcSearch(query)...)

	// 5. Clipboard history (clip/cb prefix)
	currentResults = append(currentResults, modules.ClipboardSearch(query)...)

	// 6. Web shortcuts (g, gh, etc.)
	currentResults = append(currentResults, modules.WebSearch(query)...)

	// 8. System commands
	currentResults = append(currentResults, modules.SystemSearch(query)...)

	// 9. Spotify/music control (sp prefix)
	currentResults = append(currentResults, modules.SpotifySearch(query)...)

	// 10. File search (explicit f: prefix)
	currentResults = append(currentResults, modules.FileSearch(query)...)

	// 11. Apps (limit search for short queries)
	var appResults []apps.App
	if len(query) <= 2 {
		appResults = apps.QuickSearch(allApps, query)
	} else {
		appResults = apps.Search(allApps, query)
	}

	for _, app := range appResults {
		a := app // capture
		currentResults = append(currentResults, modules.Result{
			Type:  "app",
			Title: a.Name,
			Icon:  a.Icon,
			Action: func() {
				history.Record(a.Name)
				apps.Launch(a)
			},
		})
	}

	// 10. Fallback web search if no results
	if len(currentResults) == 0 {
		currentResults = append(currentResults, modules.FallbackWebSearch(query)...)
	}

	setResults(currentResults)
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

func updatePreview(row *gtk.ListBoxRow) {
	if isClearing || inSpotifyMode || row == nil {
		hidePreview()
		return
	}

	idx := row.Index()
	if idx < 0 || idx >= len(currentResults) {
		hidePreview()
		return
	}

	r := currentResults[idx]

	imagePath := modules.GetPreviewImage(r)
	if imagePath != "" {
		if pb, err := gdkpixbuf.NewPixbufFromFileAtScale(imagePath, 220, 180, true); err == nil {
			previewLabel.Hide()
			previewLabel.SetText("")
			previewImage.SetFromPixbuf(pb)
			previewImage.Show()
			previewBox.Show()
			return
		}
	}

	if r.PreviewImageURL != "" {
		version := atomic.AddUint64(&previewVersion, 1)
		previewImage.Hide()
		previewImage.Clear()
		previewLabel.SetText(r.Preview)
		previewLabel.Show()
		previewBox.Show()
		go func(imageURL string, v uint64) {
			path := modules.CacheYouTubeThumbnail(imageURL)
			if path == "" {
				return
			}
			glib.IdleAdd(func() {
				if atomic.LoadUint64(&previewVersion) != v {
					return
				}
				if pb, err := gdkpixbuf.NewPixbufFromFileAtScale(path, 220, 180, true); err == nil {
					previewLabel.Hide()
					previewLabel.SetText("")
					previewImage.SetFromPixbuf(pb)
					previewImage.Show()
					previewBox.Show()
				}
			})
		}(r.PreviewImageURL, version)
		return
	}

	preview := modules.GetPreview(r)
	if preview == "" {
		hidePreview()
		return
	}

	previewImage.Hide()
	previewImage.Clear()
	previewLabel.SetText(preview)
	previewLabel.Show()
	previewBox.Show()
}

func hidePreview() {
	atomic.AddUint64(&previewVersion, 1)
	previewLabel.Hide()
	previewLabel.SetText("")
	previewImage.Hide()
	previewImage.Clear()
	previewBox.Hide()
}

func createResultRow(r modules.Result) *gtk.ListBoxRow {
	row := gtk.NewListBoxRow()
	row.SetName("spark-row")

	hbox := gtk.NewBox(gtk.OrientationHorizontal, 8)
	hbox.SetMarginStart(8)
	hbox.SetMarginEnd(8)
	hbox.SetMarginTop(8)
	hbox.SetMarginBottom(8)

	// Icon
	if config.Current.ShowIcons {
		icon := loadIcon(r.Icon, config.Current.IconSize)
		if icon != nil {
			hbox.PackStart(icon, false, false, 0)
		}
	}

	// Title and description
	vbox := gtk.NewBox(gtk.OrientationVertical, 2)

	title := gtk.NewLabel(r.Title)
	title.SetXAlign(0)
	title.SetName("spark-title")
	title.SetEllipsize(pango.EllipsizeEnd)
	title.SetMaxWidthChars(50)
	vbox.PackStart(title, false, false, 0)

	if r.Desc != "" {
		desc := gtk.NewLabel(r.Desc)
		desc.SetXAlign(0)
		desc.SetName("spark-desc")
		desc.SetEllipsize(pango.EllipsizeEnd)
		desc.SetMaxWidthChars(60)
		vbox.PackStart(desc, false, false, 0)
	}

	hbox.PackStart(vbox, true, true, 0)

	eventBox := gtk.NewEventBox()
	eventBox.SetVisibleWindow(false)
	eventBox.Add(hbox)
	row.Add(eventBox)
	setupFileDragSource(eventBox, r)
	return row
}

type dragSourceWidget interface {
	DragSourceSet(gdk.ModifierType, []gtk.TargetEntry, gdk.DragAction)
	DragSourceSetIconPixbuf(*gdkpixbuf.Pixbuf)
	ConnectDragDataGet(func(context *gdk.DragContext, data *gtk.SelectionData, info, time uint)) glib.SignalHandle
}

func setupFileDragSource(widget dragSourceWidget, r modules.Result) {
	path := modules.GetFilePath(r)
	if path == "" {
		return
	}

	targets := []gtk.TargetEntry{
		*gtk.NewTargetEntry("text/uri-list", 0, dragInfoURI),
		*gtk.NewTargetEntry("text/plain", 0, dragInfoText),
		*gtk.NewTargetEntry("UTF8_STRING", 0, dragInfoText),
		*gtk.NewTargetEntry("STRING", 0, dragInfoText),
	}
	widget.DragSourceSet(gdk.Button1Mask, targets, gdk.ActionCopy)

	if imagePath := modules.GetPreviewImage(r); imagePath != "" {
		if pb, err := gdkpixbuf.NewPixbufFromFileAtScale(imagePath, 96, 96, true); err == nil {
			widget.DragSourceSetIconPixbuf(pb)
		}
	}

	widget.ConnectDragDataGet(func(_ *gdk.DragContext, data *gtk.SelectionData, info, _ uint) {
		absPath := absolutePath(path)
		switch info {
		case dragInfoURI:
			data.SetURIs([]string{fileURI(absPath)})
		default:
			data.SetText(absPath)
		}
	})
}

func absolutePath(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return path
}

func fileURI(path string) string {
	return (&url.URL{Scheme: "file", Path: filepath.ToSlash(path)}).String()
}

func loadIcon(iconName string, size int) *gtk.Image {
	if iconName == "" {
		return nil
	}

	cacheKey := iconName

	// Check cache
	iconCacheMu.RLock()
	if pixbuf, ok := iconCache[cacheKey]; ok {
		iconCacheMu.RUnlock()
		if pixbuf != nil {
			return gtk.NewImageFromPixbuf(pixbuf)
		}
		return nil
	}
	iconCacheMu.RUnlock()

	// Load icon
	var pixbuf *gdkpixbuf.Pixbuf
	if iconName[0] == '/' {
		pb, err := gdkpixbuf.NewPixbufFromFileAtSize(iconName, size, size)
		if err == nil {
			pixbuf = pb
		}
	} else {
		theme := gtk.IconThemeGetDefault()
		pb, err := theme.LoadIcon(iconName, size, gtk.IconLookupForceSize)
		if err == nil {
			pixbuf = pb
		}
	}

	// Cache result (even nil)
	iconCacheMu.Lock()
	iconCache[cacheKey] = pixbuf
	iconCacheMu.Unlock()

	if pixbuf != nil {
		return gtk.NewImageFromPixbuf(pixbuf)
	}
	return nil
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
	if r.Action != nil {
		r.Action()
	}
	gtk.MainQuit()
}

func loadCSS() {
	css := gtk.NewCSSProvider()
	css.LoadFromData(config.GetCSS() + `
		#spark-preview {
			background: rgba(0, 0, 0, 0.3);
			padding: 8px;
			border-radius: 6px;
		}
		#spark-preview-label {
			color: rgba(255, 255, 255, 0.8);
			font-family: monospace;
			font-size: 11px;
		}
		#spark-preview-image {
			background: rgba(255, 255, 255, 0.08);
			border: 1px solid rgba(255, 255, 255, 0.14);
			padding: 6px;
		}
		#spotify-view {
			background: rgba(0, 0, 0, 0.2);
			border-radius: 8px;
			padding: 12px;
		}
		#spotify-header {
			background: rgba(0, 0, 0, 0.3);
			border-radius: 8px;
			padding: 12px;
		}
		#spotify-title {
			color: white;
			font-size: 16px;
			font-weight: bold;
		}
		#spotify-artist {
			color: rgba(255, 255, 255, 0.7);
			font-size: 13px;
		}
		#spotify-album {
			color: rgba(255, 255, 255, 0.5);
			font-size: 12px;
		}
		#spotify-status {
			color: #1DB954;
			font-size: 11px;
		}
		#spotify-control {
			background: rgba(255, 255, 255, 0.1);
			border-radius: 50%;
			padding: 8px;
			min-width: 36px;
			min-height: 36px;
		}
		#spotify-control:hover {
			background: rgba(255, 255, 255, 0.2);
		}
		#spotify-list {
			background: transparent;
		}
		#spotify-list row {
			background: transparent;
			border-radius: 6px;
			padding: 6px;
		}
		#spotify-list row:selected {
			background: rgba(100, 150, 255, 0.3);
		}
	`)
	screen := gdk.ScreenGetDefault()
	gtk.StyleContextAddProviderForScreen(screen, css, uint(gtk.STYLE_PROVIDER_PRIORITY_APPLICATION))
}

func createSpotifyView() {
	spotifyView = gtk.NewBox(gtk.OrientationVertical, 8)
	spotifyView.SetName("spotify-view")

	// Header: [Small Art] [Track Info] [Controls]
	header := gtk.NewBox(gtk.OrientationHorizontal, 12)
	header.SetName("spotify-header")

	// Small album art (64x64)
	spotifyArtSmall = gtk.NewImage()
	spotifyArtSmall.SetSizeRequest(64, 64)
	header.PackStart(spotifyArtSmall, false, false, 0)

	// Track info (vertical)
	infoBox := gtk.NewBox(gtk.OrientationVertical, 4)
	spotifyTitle = gtk.NewLabel("")
	spotifyTitle.SetName("spotify-title")
	spotifyTitle.SetXAlign(0)
	spotifyTitle.SetEllipsize(3) // PANGO_ELLIPSIZE_END

	spotifyArtist = gtk.NewLabel("")
	spotifyArtist.SetName("spotify-artist")
	spotifyArtist.SetXAlign(0)

	spotifyAlbum = gtk.NewLabel("")
	spotifyAlbum.SetName("spotify-album")
	spotifyAlbum.SetXAlign(0)

	spotifyStatus = gtk.NewLabel("")
	spotifyStatus.SetName("spotify-status")
	spotifyStatus.SetXAlign(0)

	infoBox.PackStart(spotifyTitle, false, false, 0)
	infoBox.PackStart(spotifyArtist, false, false, 0)
	infoBox.PackStart(spotifyAlbum, false, false, 0)
	infoBox.PackStart(spotifyStatus, false, false, 0)
	header.PackStart(infoBox, true, true, 0)

	// Playback controls
	controls := gtk.NewBox(gtk.OrientationHorizontal, 8)
	controls.SetHAlign(gtk.AlignEnd)
	controls.SetVAlign(gtk.AlignCenter)

	prevBtn := gtk.NewButton()
	prevBtn.SetName("spotify-control")
	prevBtn.SetLabel("⏮")
	prevBtn.Connect("clicked", func() {
		exec.Command("playerctl", "previous").Run()
		glib.TimeoutAdd(300, func() bool { refreshSpotifyInfo(); return false })
	})

	playBtn := gtk.NewButton()
	playBtn.SetName("spotify-control")
	playBtn.SetLabel("⏯")
	playBtn.Connect("clicked", func() {
		exec.Command("playerctl", "play-pause").Run()
		glib.TimeoutAdd(300, func() bool { refreshSpotifyInfo(); return false })
	})

	nextBtn := gtk.NewButton()
	nextBtn.SetName("spotify-control")
	nextBtn.SetLabel("⏭")
	nextBtn.Connect("clicked", func() {
		exec.Command("playerctl", "next").Run()
		glib.TimeoutAdd(500, func() bool { refreshSpotifyInfo(); return false })
	})

	controls.PackStart(prevBtn, false, false, 0)
	controls.PackStart(playBtn, false, false, 0)
	controls.PackStart(nextBtn, false, false, 0)
	header.PackEnd(controls, false, false, 0)

	spotifyView.PackStart(header, false, false, 0)

	// Content: [List] [Big Art]
	content := gtk.NewBox(gtk.OrientationHorizontal, 12)

	// Control list
	spotifyList = gtk.NewListBox()
	spotifyList.SetName("spotify-list")
	spotifyList.SetSelectionMode(gtk.SelectionSingle)

	// Add control options
	for _, ctrl := range modules.SpotifyControls() {
		row := createSpotifyControlRow(ctrl)
		spotifyList.Add(row)
		row.ShowAll()
	}

	spotifyList.Connect("row-activated", func(_ *gtk.ListBox, row *gtk.ListBoxRow) {
		idx := row.Index()
		ctrls := modules.SpotifyControls()
		if idx >= 0 && idx < len(ctrls) {
			ctrls[idx].Action()
			glib.TimeoutAdd(300, func() bool { refreshSpotifyInfo(); return false })
		}
	})

	content.PackStart(spotifyList, true, true, 0)

	// Big album art (150x150)
	artFrame := gtk.NewBox(gtk.OrientationVertical, 0)
	artFrame.SetHAlign(gtk.AlignCenter)
	artFrame.SetVAlign(gtk.AlignCenter)
	spotifyArtBig = gtk.NewImage()
	spotifyArtBig.SetSizeRequest(150, 150)
	artFrame.PackStart(spotifyArtBig, false, false, 0)
	content.PackEnd(artFrame, false, false, 0)

	spotifyView.PackStart(content, false, false, 0)
}

func createSpotifyControlRow(r modules.Result) *gtk.ListBoxRow {
	row := gtk.NewListBoxRow()
	hbox := gtk.NewBox(gtk.OrientationHorizontal, 8)
	hbox.SetMarginStart(8)
	hbox.SetMarginEnd(8)
	hbox.SetMarginTop(6)
	hbox.SetMarginBottom(6)

	if icon := loadIcon(r.Icon, 20); icon != nil {
		hbox.PackStart(icon, false, false, 0)
	}

	label := gtk.NewLabel(r.Title)
	label.SetXAlign(0)
	hbox.PackStart(label, true, true, 0)

	if r.Desc != "" {
		desc := gtk.NewLabel(r.Desc)
		desc.SetXAlign(1)
		desc.SetName("spark-desc")
		hbox.PackEnd(desc, false, false, 0)
	}

	row.Add(hbox)
	return row
}

func showSpotifyView() {
	inSpotifyMode = true
	resultsScroll.Hide()
	hidePreview()

	refreshSpotifyInfo()

	// Show all children then the view itself
	spotifyView.Show()
	spotifyView.ShowAll()

	// Select first row
	if first := spotifyList.RowAtIndex(0); first != nil {
		spotifyList.SelectRow(first)
	}
}

func hideSpotifyView() {
	inSpotifyMode = false
	spotifyView.Hide()
}

func refreshSpotifyInfo() {
	info := modules.GetSpotifyInfo()
	if info == nil {
		spotifyTitle.SetText("No player detected")
		spotifyArtist.SetText("")
		spotifyAlbum.SetText("")
		spotifyStatus.SetText("Start Spotify or another player")
		return
	}

	spotifyTitle.SetText(info.Title)
	spotifyArtist.SetText(info.Artist)
	spotifyAlbum.SetText(info.Album)

	statusIcon := "▶"
	if info.Status == "Paused" {
		statusIcon = "⏸"
	} else if info.Status == "Stopped" {
		statusIcon = "⏹"
	}
	spotifyStatus.SetText(statusIcon + " " + info.Status)

	// Load album art
	if info.ArtPath != "" {
		// Small art (64x64)
		if pb, err := gdkpixbuf.NewPixbufFromFileAtSize(info.ArtPath, 64, 64); err == nil {
			spotifyArtSmall.SetFromPixbuf(pb)
		}
		// Big art (150x150)
		if pb, err := gdkpixbuf.NewPixbufFromFileAtSize(info.ArtPath, 150, 150); err == nil {
			spotifyArtBig.SetFromPixbuf(pb)
		}
	}
}
