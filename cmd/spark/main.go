package main

import (
	"context"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	allApps         []apps.App
	currentResults  []modules.Result
	listBox         *gtk.ListBox
	resultsScroll   *gtk.ScrolledWindow
	searchEntry     *gtk.Entry
	previewBox      *gtk.Box
	previewToolbar  *gtk.Box
	previewMeta     *gtk.Label
	previewImage    *gtk.Image
	previewLabel    *gtk.Label
	searchMu        sync.Mutex
	searchVersion   uint64
	previewVersion  uint64
	fileSearchMu    sync.Mutex
	fileSearchStop  context.CancelFunc
	quickLookActive bool
	previewPage     int
	previewScale    int
	inActionMode    bool
	debounceTimer   *time.Timer
	iconCache       = make(map[string]*gdkpixbuf.Pixbuf)
	iconCacheMu     sync.RWMutex

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
	playerMode      modules.PlayerKind
	mainBox         *gtk.Box
	isClearing      bool
)

const (
	dragInfoURI uint = iota + 1
	dragInfoText
)

func main() {
	if len(os.Args) > 2 && os.Args[1] == "--large-type" {
		gtk.Init()
		showLargeType(os.Args[2], -1)
		gtk.Main()
		os.Exit(0)
	}
	if len(os.Args) > 2 && os.Args[1] == "--large-type-all" {
		gtk.Init()
		showLargeTypeAll(os.Args[2])
		gtk.Main()
		os.Exit(0)
	}
	if len(os.Args) > 1 && os.Args[1] == "--stats-window" {
		gtk.Init()
		showStatsWindow()
		gtk.Main()
		os.Exit(0)
	}
	if len(os.Args) > 1 && os.Args[1] == "--email-window" {
		gtk.Init()
		to, subject, body := "", "", ""
		if len(os.Args) > 2 {
			to = os.Args[2]
		}
		if len(os.Args) > 3 {
			subject = os.Args[3]
		}
		if len(os.Args) > 4 {
			body = os.Args[4]
		}
		showEmailWindow(to, subject, body)
		gtk.Main()
		os.Exit(0)
	}
	if len(os.Args) > 4 && os.Args[1] == "--file-op-window" {
		gtk.Init()
		showFileOpWindow(os.Args[2], os.Args[3], os.Args[4])
		gtk.Main()
		os.Exit(0)
	}

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
	searchEntry = entry
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
	previewBox.SetSizeRequest(-1, 232)
	previewBox.SetNoShowAll(true)

	previewToolbar = gtk.NewBox(gtk.OrientationHorizontal, 6)
	previewToolbar.SetName("spark-preview-toolbar")
	previewToolbar.SetNoShowAll(true)
	prevPageBtn := gtk.NewButtonWithLabel("‹")
	nextPageBtn := gtk.NewButtonWithLabel("›")
	zoomOutBtn := gtk.NewButtonWithLabel("-")
	zoomInBtn := gtk.NewButtonWithLabel("+")
	previewMeta = gtk.NewLabel("")
	previewMeta.SetXAlign(0)
	previewMeta.SetName("spark-desc")
	prevPageBtn.Connect("clicked", func() { changePreviewPage(-1) })
	nextPageBtn.Connect("clicked", func() { changePreviewPage(1) })
	zoomOutBtn.Connect("clicked", func() { changePreviewZoom(-60) })
	zoomInBtn.Connect("clicked", func() { changePreviewZoom(60) })
	previewToolbar.PackStart(prevPageBtn, false, false, 0)
	previewToolbar.PackStart(nextPageBtn, false, false, 0)
	previewToolbar.PackStart(zoomOutBtn, false, false, 0)
	previewToolbar.PackStart(zoomInBtn, false, false, 0)
	previewToolbar.PackStart(previewMeta, true, true, 0)

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
	previewBox.PackStart(previewToolbar, false, false, 0)
	previewBox.PackStart(previewImage, false, false, 0)
	previewBox.PackStart(previewLabel, false, false, 0)

	// Update preview on selection change
	listBox.Connect("row-selected", func(_ *gtk.ListBox, row *gtk.ListBoxRow) {
		if quickLookActive {
			updatePreview(row)
		} else {
			hidePreview()
		}
	})

	// Search on typing with debounce
	entry.Connect("changed", func() {
		query := entry.Text()
		inActionMode = false

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
			if inActionMode {
				inActionMode = false
				updateResults(searchEntry.Text())
				return true
			}
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
		case gdk.KEY_Tab:
			showSelectedFileActions()
			return true
		case gdk.KEY_Shift_L, gdk.KEY_Shift_R:
			quickLookActive = !quickLookActive
			if quickLookActive {
				previewPage = 1
				if previewScale == 0 {
					previewScale = 360
				}
				updatePreview(listBox.SelectedRow())
			} else {
				hidePreview()
			}
			return true
		case gdk.KEY_Page_Down, gdk.KEY_Right:
			if quickLookActive {
				if previewPage < 999 {
					previewPage++
				}
				updatePreview(listBox.SelectedRow())
				return true
			}
		case gdk.KEY_Page_Up, gdk.KEY_Left:
			if quickLookActive {
				if previewPage > 1 {
					previewPage--
				}
				updatePreview(listBox.SelectedRow())
				return true
			}
		case gdk.KEY_plus, gdk.KEY_KP_Add, gdk.KEY_equal:
			if quickLookActive {
				if previewScale == 0 {
					previewScale = 360
				}
				if previewScale < 720 {
					previewScale += 60
				}
				updatePreview(listBox.SelectedRow())
				return true
			}
		case gdk.KEY_minus, gdk.KEY_KP_Subtract:
			if quickLookActive {
				if previewScale == 0 {
					previewScale = 360
				}
				if previewScale > 180 {
					previewScale -= 60
				}
				updatePreview(listBox.SelectedRow())
				return true
			}
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

	if results := modules.NavigationSearch(query); results != nil {
		setResults(results)
		return
	}

	if results := modules.DestinationPickerSearch(query); results != nil {
		setResults(results)
		return
	}

	if results := modules.FileOperationSearch(query); results != nil {
		setResults(results)
		return
	}

	// Collect results from all modules (priority order)
	currentResults = nil

	// 1. Shell commands (> prefix)
	currentResults = append(currentResults, modules.ShellSearch(query)...)

	// 2. Help
	currentResults = append(currentResults, modules.HelpSearch(query)...)

	// 3. Large Type
	currentResults = append(currentResults, modules.LargeTypeSearch(query)...)

	// 4. Recent documents
	currentResults = append(currentResults, modules.RecentSearch(query)...)

	// 5. Contacts
	currentResults = append(currentResults, modules.ContactsSearch(query)...)

	// 6. Email
	currentResults = append(currentResults, modules.EmailSearch(query)...)

	// 7. Usage stats
	currentResults = append(currentResults, modules.StatsSearch(query)...)

	// 8. Settings sync helpers
	currentResults = append(currentResults, modules.SyncSearch(query)...)

	// 9. Last action/error status
	currentResults = append(currentResults, modules.StatusSearch(query)...)

	// 10. Snippets (; prefix or "snip")
	currentResults = append(currentResults, modules.SnippetSearch(query)...)

	// 11. Dictionary (define/def prefix)
	currentResults = append(currentResults, modules.DictionarySearch(query)...)

	// 12. Spelling (spell/spelling prefix)
	currentResults = append(currentResults, modules.SpellSearch(query)...)

	// 13. Calculator
	currentResults = append(currentResults, modules.CalcSearch(query)...)

	// 14. Clipboard history (clip/cb prefix)
	currentResults = append(currentResults, modules.ClipboardSearch(query)...)

	// 15. Web shortcuts (g, gh, etc.)
	currentResults = append(currentResults, modules.WebSearch(query)...)

	// 16. System commands
	currentResults = append(currentResults, modules.SystemSearch(query)...)

	// 17. Spotify/music control (sp prefix)
	currentResults = append(currentResults, modules.SpotifySearch(query)...)

	// 18. Local music search (m prefix)
	currentResults = append(currentResults, modules.MusicSearch(query)...)

	// 19. File buffer actions
	currentResults = append(currentResults, modules.FileBufferSearch(query)...)

	// 20. File search (explicit f prefix)
	currentResults = append(currentResults, modules.FileSearch(query)...)

	// 21. Apps (limit search for short queries)
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

func updatePreview(row *gtk.ListBoxRow) {
	if isClearing {
		return
	}
	if inSpotifyMode {
		hidePreview()
		return
	}
	if row == nil {
		if quickLookActive {
			cancelPreviewLoad()
			return
		}
		hidePreview()
		return
	}

	idx := row.Index()
	if idx < 0 || idx >= len(currentResults) {
		if quickLookActive {
			cancelPreviewLoad()
			return
		}
		hidePreview()
		return
	}

	r := currentResults[idx]
	updatePreviewToolbar(r)

	if r.Type == "file" {
		version := atomic.AddUint64(&previewVersion, 1)
		page := previewPage
		scale := previewScale
		if page < 1 {
			page = 1
		}
		if scale == 0 {
			scale = 360
		}
		showPreviewLoading("Loading preview... page " + stringIntLocal(page) + " zoom " + stringIntLocal(scale))
		go func(res modules.Result, v uint64) {
			imagePath := modules.GetPreviewImageAt(res, page, scale)
			preview := ""
			if imagePath == "" {
				preview = modules.GetPreview(res)
			}
			glib.IdleAdd(func() {
				if atomic.LoadUint64(&previewVersion) != v {
					return
				}
				if imagePath != "" {
					if pb, err := gdkpixbuf.NewPixbufFromFileAtScale(imagePath, 220, 180, true); err == nil {
						showPreviewPixbuf(pb)
						return
					}
				}
				if preview == "" {
					showPreviewText("No preview available")
					return
				}
				showPreviewText(preview)
			})
		}(r, version)
		return
	}

	imagePath := modules.GetPreviewImage(r)
	if imagePath != "" {
		if pb, err := gdkpixbuf.NewPixbufFromFileAtScale(imagePath, 220, 180, true); err == nil {
			showPreviewPixbuf(pb)
			return
		}
	}

	if r.PreviewImageURL != "" {
		version := atomic.AddUint64(&previewVersion, 1)
		showPreviewLoading(r.Preview)
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
					showPreviewPixbuf(pb)
				}
			})
		}(r.PreviewImageURL, version)
		return
	}

	if r.Type == "clipboard" && r.Data != "" {
		version := atomic.AddUint64(&previewVersion, 1)
		showPreviewLoading(r.Title)
		go func(res modules.Result, v uint64) {
			path := modules.GetClipboardPreviewImage(res)
			if path == "" {
				return
			}
			glib.IdleAdd(func() {
				if atomic.LoadUint64(&previewVersion) != v {
					return
				}
				if pb, err := gdkpixbuf.NewPixbufFromFileAtScale(path, 220, 180, true); err == nil {
					showPreviewPixbuf(pb)
				}
			})
		}(r, version)
		return
	}

	preview := modules.GetPreview(r)
	if preview == "" {
		if quickLookActive {
			showPreviewText("No preview available")
			return
		}
		hidePreview()
		return
	}

	showPreviewText(preview)
}

func changePreviewPage(delta int) {
	if !quickLookActive {
		return
	}
	previewPage += delta
	if previewPage < 1 {
		previewPage = 1
	}
	updatePreview(listBox.SelectedRow())
}

func changePreviewZoom(delta int) {
	if !quickLookActive {
		return
	}
	if previewScale == 0 {
		previewScale = 360
	}
	previewScale += delta
	if previewScale < 180 {
		previewScale = 180
	}
	if previewScale > 720 {
		previewScale = 720
	}
	updatePreview(listBox.SelectedRow())
}

func updatePreviewToolbar(r modules.Result) {
	if previewToolbar == nil || previewMeta == nil {
		return
	}
	if r.Type != "file" {
		previewToolbar.Hide()
		return
	}
	page := previewPage
	if page < 1 {
		page = 1
	}
	scale := previewScale
	if scale == 0 {
		scale = 360
	}
	ext := strings.ToLower(filepath.Ext(r.Title))
	if ext == ".pdf" || ext == ".docx" || ext == ".odt" {
		previewMeta.SetText(r.Title + "  page " + stringIntLocal(page) + "  zoom " + stringIntLocal(scale))
		previewToolbar.ShowAll()
		return
	}
	previewMeta.SetText(r.Title)
	previewToolbar.ShowAll()
}

func hidePreview() {
	cancelPreviewLoad()
	clearPreviewContent()
	previewBox.Hide()
}

func cancelPreviewLoad() {
	atomic.AddUint64(&previewVersion, 1)
}

func clearPreviewContent() {
	if previewToolbar != nil {
		previewToolbar.Hide()
	}
	previewLabel.Hide()
	previewLabel.SetText("")
	previewImage.Hide()
	previewImage.Clear()
}

func showPreviewLoading(text string) {
	if previewBox.Visible() && (previewImage.Visible() || previewLabel.Visible()) {
		return
	}
	if !previewBox.Visible() {
		previewBox.Show()
	}
	previewLabel.SetText(text)
	previewLabel.Show()
	if !previewImage.Visible() {
		previewImage.Clear()
	}
}

func showPreviewText(text string) {
	if !previewBox.Visible() {
		previewBox.Show()
	}
	previewImage.Hide()
	previewImage.Clear()
	previewLabel.SetText(text)
	previewLabel.Show()
}

func showPreviewPixbuf(pb *gdkpixbuf.Pixbuf) {
	if !previewBox.Visible() {
		previewBox.Show()
	}
	previewImage.SetFromPixbuf(pb)
	previewImage.Show()
	previewLabel.Hide()
	previewLabel.SetText("")
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
		modules.PlayerControls(playerMode)[2].Action()
		glib.TimeoutAdd(300, func() bool { refreshSpotifyInfo(); return false })
	})

	playBtn := gtk.NewButton()
	playBtn.SetName("spotify-control")
	playBtn.SetLabel("⏯")
	playBtn.Connect("clicked", func() {
		modules.PlayerControls(playerMode)[0].Action()
		glib.TimeoutAdd(300, func() bool { refreshSpotifyInfo(); return false })
	})

	nextBtn := gtk.NewButton()
	nextBtn.SetName("spotify-control")
	nextBtn.SetLabel("⏭")
	nextBtn.Connect("clicked", func() {
		modules.PlayerControls(playerMode)[1].Action()
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
	for _, ctrl := range modules.PlayerControls(playerMode) {
		row := createSpotifyControlRow(ctrl)
		spotifyList.Add(row)
		row.ShowAll()
	}

	spotifyList.Connect("row-activated", func(_ *gtk.ListBox, row *gtk.ListBoxRow) {
		idx := row.Index()
		ctrls := modules.PlayerControls(playerMode)
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
	showPlayerView(modules.PlayerSpotify)
}

func showPlayerView(kind modules.PlayerKind) {
	playerMode = kind
	inSpotifyMode = true
	resultsScroll.Hide()
	hidePreview()
	refreshPlayerControls()

	refreshSpotifyInfo()

	// Show all children then the view itself
	spotifyView.Show()
	spotifyView.ShowAll()

	// Select first row
	if first := spotifyList.RowAtIndex(0); first != nil {
		spotifyList.SelectRow(first)
	}
}

func refreshPlayerControls() {
	if spotifyList == nil {
		return
	}
	for {
		row := spotifyList.RowAtIndex(0)
		if row == nil {
			break
		}
		spotifyList.Remove(row)
	}
	for _, ctrl := range modules.PlayerControls(playerMode) {
		row := createSpotifyControlRow(ctrl)
		spotifyList.Add(row)
		row.ShowAll()
	}
}

func hideSpotifyView() {
	inSpotifyMode = false
	spotifyView.Hide()
}

func refreshSpotifyInfo() {
	info := modules.GetPlayerInfo(playerMode)
	if info == nil {
		if playerMode == modules.PlayerYouTube {
			spotifyTitle.SetText("No YouTube player detected")
			spotifyStatus.SetText("Open YouTube in browser")
		} else {
			spotifyTitle.SetText("No Spotify player detected")
			spotifyStatus.SetText("Start Spotify")
		}
		spotifyArtist.SetText("")
		spotifyAlbum.SetText("")
		spotifyArtSmall.Clear()
		spotifyArtBig.Clear()
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

func showLargeTypeAll(text string) {
	screen := gdk.ScreenGetDefault()
	count := 1
	if screen != nil {
		count = screen.NMonitors()
	}
	for i := 0; i < count; i++ {
		showLargeType(text, i)
	}
}

func showLargeType(text string, monitor int) {
	window := gtk.NewWindow(gtk.WindowToplevel)
	window.SetTitle("Spark Large Type")
	window.SetDecorated(false)
	if monitor >= 0 {
		window.FullscreenOnMonitor(gdk.ScreenGetDefault(), monitor)
	} else {
		window.Fullscreen()
	}

	label := gtk.NewLabel(text)
	label.SetName("large-type-label")
	label.SetLineWrap(true)
	label.SetLineWrapMode(pango.WrapWordChar)
	label.SetJustify(gtk.JustifyCenter)
	label.SetXAlign(0.5)
	label.SetYAlign(0.5)

	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.SetName("large-type-window")
	box.SetHAlign(gtk.AlignFill)
	box.SetVAlign(gtk.AlignFill)
	box.SetMarginStart(60)
	box.SetMarginEnd(60)
	box.SetMarginTop(60)
	box.SetMarginBottom(60)
	box.PackStart(label, true, true, 0)
	window.Add(box)

	css := gtk.NewCSSProvider()
	css.LoadFromData(`
		#large-type-window {
			background: rgba(0, 0, 0, 0.92);
		}
		#large-type-label {
			color: white;
			font-size: ` + largeTypeFontSize(text) + `px;
			font-weight: bold;
		}
	`)
	screen := gdk.ScreenGetDefault()
	gtk.StyleContextAddProviderForScreen(screen, css, uint(gtk.STYLE_PROVIDER_PRIORITY_APPLICATION))

	window.Connect("key-press-event", func(_ *gtk.Window, _ *gdk.Event) bool {
		gtk.MainQuit()
		return true
	})
	window.Connect("button-press-event", func() bool {
		gtk.MainQuit()
		return true
	})
	window.Connect("destroy", func() {
		gtk.MainQuit()
	})
	window.ShowAll()
}

func largeTypeFontSize(text string) string {
	switch {
	case len(text) > 120:
		return "38"
	case len(text) > 80:
		return "48"
	case len(text) > 40:
		return "64"
	default:
		return "96"
	}
}

func showStatsWindow() {
	window := gtk.NewWindow(gtk.WindowToplevel)
	window.SetTitle("Spark Usage Stats")
	window.SetDefaultSize(520, 420)

	box := gtk.NewBox(gtk.OrientationVertical, 10)
	box.SetMarginStart(18)
	box.SetMarginEnd(18)
	box.SetMarginTop(18)
	box.SetMarginBottom(18)

	counts := history.Snapshot()
	type stat struct {
		name  string
		count int
	}
	var stats []stat
	max := 0
	for name, count := range counts {
		stats = append(stats, stat{name, count})
		if count > max {
			max = count
		}
	}
	for i := 0; i < len(stats)-1; i++ {
		for j := i + 1; j < len(stats); j++ {
			if stats[j].count > stats[i].count {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}
	if len(stats) > 10 {
		stats = stats[:10]
	}

	for _, s := range stats {
		row := gtk.NewBox(gtk.OrientationHorizontal, 10)
		name := gtk.NewLabel(s.name)
		name.SetXAlign(0)
		name.SetSizeRequest(160, -1)
		bar := gtk.NewProgressBar()
		if max > 0 {
			bar.SetFraction(float64(s.count) / float64(max))
		}
		bar.SetText(stringIntLocal(s.count))
		bar.SetShowText(true)
		row.PackStart(name, false, false, 0)
		row.PackStart(bar, true, true, 0)
		box.PackStart(row, false, false, 0)
	}
	if len(stats) == 0 {
		label := gtk.NewLabel("No usage stats yet")
		box.PackStart(label, true, true, 0)
	}

	window.Add(box)
	window.Connect("key-press-event", func(_ *gtk.Window, _ *gdk.Event) bool {
		gtk.MainQuit()
		return true
	})
	window.Connect("destroy", func() { gtk.MainQuit() })
	window.ShowAll()
}

func stringIntLocal(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

func showEmailWindow(toValue, subjectValue, bodyValue string) {
	window := gtk.NewWindow(gtk.WindowToplevel)
	window.SetTitle("Spark Email")
	window.SetDefaultSize(620, 320)

	box := gtk.NewBox(gtk.OrientationVertical, 10)
	box.SetMarginStart(16)
	box.SetMarginEnd(16)
	box.SetMarginTop(16)
	box.SetMarginBottom(16)

	toEntry := gtk.NewEntry()
	toEntry.SetPlaceholderText("To")
	toEntry.SetText(toValue)
	subjectEntry := gtk.NewEntry()
	subjectEntry.SetPlaceholderText("Subject")
	subjectEntry.SetText(subjectValue)
	bodyEntry := gtk.NewEntry()
	bodyEntry.SetPlaceholderText("Body")
	bodyEntry.SetText(bodyValue)
	attachmentsEntry := gtk.NewEntry()
	attachmentsEntry.SetPlaceholderText("Attachments, separated by |")

	buttons := gtk.NewBox(gtk.OrientationHorizontal, 8)
	bufferBtn := gtk.NewButtonWithLabel("Attach Buffer")
	bufferBtn.Connect("clicked", func() {
		attachmentsEntry.SetText(strings.Join(modules.FileBuffer(), "|"))
	})
	chooseBtn := gtk.NewButtonWithLabel("Choose File")
	chooseBtn.Connect("clicked", func() {
		go func() {
			if path := choosePath(false); path != "" {
				glib.IdleAdd(func() {
					current := strings.TrimSpace(attachmentsEntry.Text())
					if current != "" {
						current += "|"
					}
					attachmentsEntry.SetText(current + path)
				})
			}
		}()
	})
	send := gtk.NewButtonWithLabel("Send")
	send.Connect("clicked", func() {
		to := toEntry.Text()
		subject := subjectEntry.Text()
		body := bodyEntry.Text()
		modules.SendEmailFull(to, subject, body, splitPaths(attachmentsEntry.Text()))
		gtk.MainQuit()
	})
	buttons.PackStart(bufferBtn, false, false, 0)
	buttons.PackStart(chooseBtn, false, false, 0)
	buttons.PackEnd(send, false, false, 0)

	box.PackStart(toEntry, false, false, 0)
	box.PackStart(subjectEntry, false, false, 0)
	box.PackStart(bodyEntry, false, false, 0)
	box.PackStart(attachmentsEntry, false, false, 0)
	box.PackStart(buttons, false, false, 0)
	window.Add(box)
	window.Connect("destroy", func() { gtk.MainQuit() })
	window.ShowAll()
	toEntry.GrabFocus()
}

func splitPaths(raw string) []string {
	var out []string
	for _, part := range strings.FieldsFunc(raw, func(r rune) bool { return r == '|' || r == '\n' }) {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func showFileOpWindow(op, sourceValue, targetValue string) {
	window := gtk.NewWindow(gtk.WindowToplevel)
	window.SetTitle("Spark File Operation")
	window.SetDefaultSize(720, 520)

	box := gtk.NewBox(gtk.OrientationVertical, 10)
	box.SetMarginStart(16)
	box.SetMarginEnd(16)
	box.SetMarginTop(16)
	box.SetMarginBottom(16)

	crumbs := gtk.NewBox(gtk.OrientationHorizontal, 4)
	browser := gtk.NewListBox()
	browser.SetSelectionMode(gtk.SelectionSingle)
	browserScroll := gtk.NewScrolledWindow(nil, nil)
	browserScroll.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	browserScroll.SetSizeRequest(-1, 220)
	browserScroll.Add(browser)

	opEntry := gtk.NewEntry()
	opEntry.SetPlaceholderText("Operation")
	opEntry.SetText(op)
	sourceEntry := gtk.NewEntry()
	sourceEntry.SetPlaceholderText("Source")
	sourceEntry.SetText(sourceValue)
	targetEntry := gtk.NewEntry()
	targetEntry.SetPlaceholderText("Target name or folder")
	targetEntry.SetText(targetValue)

	currentDir := targetValue
	if currentDir == "" {
		currentDir = filepath.Dir(sourceValue)
	}
	if info, err := os.Stat(currentDir); err != nil || !info.IsDir() {
		currentDir = filepath.Dir(currentDir)
	}

	var refreshBrowser func(string)
	var setCrumbs func(string)
	var crumbWidgets []gtk.Widgetter
	setCrumbs = func(dir string) {
		for _, child := range crumbWidgets {
			crumbs.Remove(child)
		}
		crumbWidgets = nil
		clean := filepath.Clean(dir)
		home := os.Getenv("HOME")
		parts := []struct {
			label string
			path  string
		}{{label: "/", path: string(os.PathSeparator)}}
		if strings.HasPrefix(clean, home) {
			parts = append(parts, struct {
				label string
				path  string
			}{label: "~", path: home})
			rel, _ := filepath.Rel(home, clean)
			if rel != "." {
				cur := home
				for _, part := range strings.Split(rel, string(os.PathSeparator)) {
					cur = filepath.Join(cur, part)
					parts = append(parts, struct {
						label string
						path  string
					}{label: part, path: cur})
				}
			}
		} else {
			cur := string(os.PathSeparator)
			for _, part := range strings.Split(strings.TrimPrefix(clean, string(os.PathSeparator)), string(os.PathSeparator)) {
				if part == "" {
					continue
				}
				cur = filepath.Join(cur, part)
				parts = append(parts, struct {
					label string
					path  string
				}{label: part, path: cur})
			}
		}
		for i, part := range parts {
			p := part.path
			btn := gtk.NewButtonWithLabel(part.label)
			btn.Connect("clicked", func() { refreshBrowser(p) })
			crumbs.PackStart(btn, false, false, 0)
			crumbWidgets = append(crumbWidgets, btn)
			if i < len(parts)-1 {
				sep := gtk.NewLabel("/")
				crumbs.PackStart(sep, false, false, 0)
				crumbWidgets = append(crumbWidgets, sep)
			}
		}
		crumbs.ShowAll()
	}
	refreshBrowser = func(dir string) {
		if dir == "" {
			return
		}
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			dir = filepath.Dir(dir)
		}
		currentDir = filepath.Clean(dir)
		targetEntry.SetText(currentDir)
		setCrumbs(currentDir)
		for {
			row := browser.RowAtIndex(0)
			if row == nil {
				break
			}
			browser.Remove(row)
		}
		if parent := filepath.Dir(currentDir); parent != currentDir {
			browser.Add(fileOpBrowserRow("..", parent, true, targetEntry, refreshBrowser))
		}
		entries, err := os.ReadDir(currentDir)
		if err != nil {
			row := gtk.NewListBoxRow()
			row.Add(gtk.NewLabel(err.Error()))
			browser.Add(row)
			browser.ShowAll()
			return
		}
		count := 0
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			path := filepath.Join(currentDir, entry.Name())
			browser.Add(fileOpBrowserRow(entry.Name(), path, entry.IsDir(), targetEntry, refreshBrowser))
			count++
			if count >= 80 {
				break
			}
		}
		browser.ShowAll()
	}

	buttons := gtk.NewBox(gtk.OrientationHorizontal, 8)
	homeBtn := gtk.NewButtonWithLabel("Home")
	homeBtn.Connect("clicked", func() {
		refreshBrowser(os.Getenv("HOME"))
	})
	downloadsBtn := gtk.NewButtonWithLabel("Downloads")
	downloadsBtn.Connect("clicked", func() {
		refreshBrowser(filepath.Join(os.Getenv("HOME"), "Downloads"))
	})
	chooseBtn := gtk.NewButtonWithLabel("Choose")
	chooseBtn.Connect("clicked", func() {
		go func() {
			if path := choosePath(opEntry.Text() != "rename"); path != "" {
				glib.IdleAdd(func() {
					targetEntry.SetText(path)
					refreshBrowser(path)
				})
			}
		}()
	})
	cancelBtn := gtk.NewButtonWithLabel("Cancel")
	cancelBtn.Connect("clicked", func() { gtk.MainQuit() })
	runBtn := gtk.NewButtonWithLabel("Run")
	runBtn.Connect("clicked", func() {
		modules.RunFileOperation(opEntry.Text(), sourceEntry.Text(), targetEntry.Text())
		gtk.MainQuit()
	})

	buttons.PackStart(homeBtn, false, false, 0)
	buttons.PackStart(downloadsBtn, false, false, 0)
	buttons.PackStart(chooseBtn, false, false, 0)
	buttons.PackEnd(runBtn, false, false, 0)
	buttons.PackEnd(cancelBtn, false, false, 0)

	box.PackStart(crumbs, false, false, 0)
	box.PackStart(opEntry, false, false, 0)
	box.PackStart(sourceEntry, false, false, 0)
	box.PackStart(targetEntry, false, false, 0)
	box.PackStart(browserScroll, true, true, 0)
	box.PackStart(buttons, false, false, 0)
	window.Add(box)
	window.Connect("destroy", func() { gtk.MainQuit() })
	refreshBrowser(currentDir)
	window.ShowAll()
	targetEntry.GrabFocus()
}

func fileOpBrowserRow(name, path string, isDir bool, targetEntry *gtk.Entry, refresh func(string)) *gtk.ListBoxRow {
	row := gtk.NewListBoxRow()
	box := gtk.NewBox(gtk.OrientationHorizontal, 8)
	box.SetMarginStart(8)
	box.SetMarginEnd(8)
	box.SetMarginTop(6)
	box.SetMarginBottom(6)
	icon := "text-x-generic"
	if isDir {
		icon = "folder"
	}
	if img := loadIcon(icon, 20); img != nil {
		box.PackStart(img, false, false, 0)
	}
	label := gtk.NewLabel(name)
	label.SetXAlign(0)
	box.PackStart(label, true, true, 0)
	row.Add(box)
	row.Connect("activate", func() {
		if isDir {
			refresh(path)
			return
		}
		targetEntry.SetText(path)
	})
	return row
}

func fileOpBreadcrumb(path string) string {
	dir := filepath.Dir(path)
	home := os.Getenv("HOME")
	if strings.HasPrefix(dir, home) {
		dir = "~" + strings.TrimPrefix(dir, home)
	}
	parts := strings.Split(filepath.Clean(dir), string(os.PathSeparator))
	if len(parts) > 4 {
		parts = append([]string{"..."}, parts[len(parts)-3:]...)
	}
	return strings.Join(parts, " / ")
}

func choosePath(directory bool) string {
	if _, err := exec.LookPath("zenity"); err == nil {
		args := []string{"--file-selection"}
		if directory {
			args = append(args, "--directory")
		}
		out, err := exec.Command("zenity", args...).Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	}
	if _, err := exec.LookPath("kdialog"); err == nil {
		args := []string{"--getopenfilename"}
		if directory {
			args = []string{"--getexistingdirectory"}
		}
		out, err := exec.Command("kdialog", args...).Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	}
	return ""
}
