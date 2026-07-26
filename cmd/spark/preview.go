package main

import (
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/diamondburned/gotk4/pkg/gdkpixbuf/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/tapiaw38/spark/internal/modules"
)

const (
	previewThumbWidth  = 220
	previewThumbHeight = 180
)

const noPreview = "No preview available"

func updatePreview(row *gtk.ListBoxRow) {
	if isClearing {
		return
	}
	if inSpotifyMode {
		hidePreview()
		return
	}

	r, ok := selectedResult(row)
	if !ok {
		abortPreview()
		return
	}
	updatePreviewToolbar(r)

	switch {
	case r.Type == "file":
		previewFile(r)
	case previewLocalImage(r):
	case r.PreviewImageURL != "":
		showPreviewLoading(r.Preview)
		loadPreviewImageAsync(func() string {
			return modules.CacheYouTubeThumbnail(r.PreviewImageURL)
		})
	case r.Type == "clipboard" && r.Data != "":
		showPreviewLoading(r.Title)
		loadPreviewImageAsync(func() string {
			return modules.GetClipboardPreviewImage(r)
		})
	default:
		previewTextOnly(r)
	}
}

func selectedResult(row *gtk.ListBoxRow) (modules.Result, bool) {
	if row == nil {
		return modules.Result{}, false
	}
	idx := row.Index()
	if idx < 0 || idx >= len(currentResults) {
		return modules.Result{}, false
	}
	return currentResults[idx], true
}

func abortPreview() {
	if quickLookActive {
		cancelPreviewLoad()
		return
	}
	hidePreview()
}

func previewFile(r modules.Result) {
	page, scale := currentPreviewPage(), currentPreviewScale()
	version := atomic.AddUint64(&previewVersion, 1)
	showPreviewLoading("Loading preview... page " + strconv.Itoa(page) + " zoom " + strconv.Itoa(scale))

	go func() {
		imagePath := modules.GetPreviewImageAt(r, page, scale)
		text := ""
		if imagePath == "" {
			text = modules.GetPreview(r)
		}
		glib.IdleAdd(func() {
			if atomic.LoadUint64(&previewVersion) != version {
				return
			}
			if showPixbufFromFile(imagePath) {
				return
			}
			if text == "" {
				text = noPreview
			}
			showPreviewText(text)
		})
	}()
}

func previewLocalImage(r modules.Result) bool {
	return showPixbufFromFile(modules.GetPreviewImage(r))
}

func previewTextOnly(r modules.Result) {
	preview := modules.GetPreview(r)
	if preview == "" {
		if quickLookActive {
			showPreviewText(noPreview)
			return
		}
		hidePreview()
		return
	}
	showPreviewText(preview)
}

func loadPreviewImageAsync(resolve func() string) {
	version := atomic.AddUint64(&previewVersion, 1)
	go func() {
		path := resolve()
		if path == "" {
			return
		}
		glib.IdleAdd(func() {
			if atomic.LoadUint64(&previewVersion) != version {
				return
			}
			showPixbufFromFile(path)
		})
	}()
}

func showPixbufFromFile(path string) bool {
	if path == "" {
		return false
	}
	pb, err := gdkpixbuf.NewPixbufFromFileAtScale(path, previewThumbWidth, previewThumbHeight, true)
	if err != nil {
		return false
	}
	showPreviewPixbuf(pb)
	return true
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
	previewScale = clamp(currentPreviewScale()+delta, previewScaleMin, previewScaleMax)
	updatePreview(listBox.SelectedRow())
}

func currentPreviewScale() int {
	if previewScale == 0 {
		return previewScaleDefault
	}
	return previewScale
}

func currentPreviewPage() int {
	if previewPage < 1 {
		return 1
	}
	return previewPage
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func updatePreviewToolbar(r modules.Result) {
	if previewToolbar == nil || previewMeta == nil {
		return
	}
	if r.Type != "file" {
		previewToolbar.Hide()
		return
	}
	page := currentPreviewPage()
	scale := currentPreviewScale()
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
