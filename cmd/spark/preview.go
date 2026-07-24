package main

import (
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/diamondburned/gotk4/pkg/gdkpixbuf/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/diamondburned/gotk4/pkg/pango"
	"github.com/tapiaw38/spark/internal/modules"
)

func buildPreviewPane() *gtk.Box {
	previewBox = gtk.NewBox(gtk.OrientationVertical, 8)
	previewBox.SetName("spark-preview")
	previewBox.SetSizeRequest(-1, 232)
	previewBox.SetNoShowAll(true)

	previewToolbar = gtk.NewBox(gtk.OrientationHorizontal, 6)
	previewToolbar.SetName("spark-preview-toolbar")
	previewToolbar.SetNoShowAll(true)
	previewMeta = gtk.NewLabel("")
	previewMeta.SetXAlign(0)
	previewMeta.SetName("spark-desc")
	for _, b := range []struct {
		label string
		fn    func()
	}{
		{"‹", func() { changePreviewPage(-1) }},
		{"›", func() { changePreviewPage(1) }},
		{"-", func() { changePreviewZoom(-60) }},
		{"+", func() { changePreviewZoom(60) }},
	} {
		btn := gtk.NewButtonWithLabel(b.label)
		btn.Connect("clicked", b.fn)
		previewToolbar.PackStart(btn, false, false, 0)
	}
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
	previewLabel.SetSizeRequest(300, -1)
	previewLabel.SetNoShowAll(true)

	previewBox.PackStart(previewToolbar, false, false, 0)
	previewBox.PackStart(previewImage, false, false, 0)
	previewBox.PackStart(previewLabel, false, false, 0)
	return previewBox
}

func updatePreview(row *gtk.ListBoxRow) {
	if isClearing {
		return
	}
	if inSpotifyMode {
		hidePreview()
		return
	}
	idx := -1
	if row != nil {
		idx = row.Index()
	}
	if idx < 0 || idx >= len(currentResults) {
		if quickLookActive {
			cancelPreviewLoad()
		} else {
			hidePreview()
		}
		return
	}

	r := currentResults[idx]
	updatePreviewToolbar(r)

	switch {
	case r.Type == "file":
		page, scale := previewPageScale()
		loadPreviewAsync("Loading preview... page "+strconv.Itoa(page)+" zoom "+strconv.Itoa(scale),
			func() (string, string) {
				text := modules.GetPreview(r)
				if text == "" {
					text = "No preview available"
				}
				return modules.GetPreviewImageAt(r, page, scale), text
			})
	case showPixbufPath(modules.GetPreviewImage(r)):
		return
	case r.PreviewImageURL != "":
		loadPreviewAsync(r.Preview, func() (string, string) {
			return modules.CacheYouTubeThumbnail(r.PreviewImageURL), ""
		})
	case r.Type == "clipboard" && r.Data != "":
		loadPreviewAsync(r.Title, func() (string, string) {
			return modules.GetClipboardPreviewImage(r), ""
		})
	default:
		if text := modules.GetPreview(r); text != "" {
			showPreviewText(text)
		} else if quickLookActive {
			showPreviewText("No preview available")
		} else {
			hidePreview()
		}
	}
}

func previewPageScale() (page, scale int) {
	page, scale = previewPage, previewScale
	if page < 1 {
		page = 1
	}
	if scale == 0 {
		scale = 360
	}
	return page, scale
}

func showPixbufPath(path string) bool {
	if path == "" {
		return false
	}
	pb, err := gdkpixbuf.NewPixbufFromFileAtScale(path, 220, 180, true)
	if err != nil {
		return false
	}
	showPreviewPixbuf(pb)
	return true
}

func loadPreviewAsync(loading string, fetch func() (imgPath, text string)) {
	version := atomic.AddUint64(&previewVersion, 1)
	showPreviewLoading(loading)
	go func() {
		img, text := fetch()
		glib.IdleAdd(func() {
			if atomic.LoadUint64(&previewVersion) != version {
				return
			}
			if showPixbufPath(img) {
				return
			}
			if text != "" {
				showPreviewText(text)
			}
		})
	}()
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
		previewMeta.SetText(r.Title + "  page " + strconv.Itoa(page) + "  zoom " + strconv.Itoa(scale))
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
