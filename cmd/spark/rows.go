package main

import (
	"net/url"
	"path/filepath"

	"github.com/diamondburned/gotk4/pkg/gdk/v3"
	"github.com/diamondburned/gotk4/pkg/gdkpixbuf/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/diamondburned/gotk4/pkg/pango"
	"github.com/tapiaw38/spark/internal/config"
	"github.com/tapiaw38/spark/internal/modules"
)

func preloadIcons() {
	theme := gtk.IconThemeGetDefault()
	for i, app := range allApps {
		if i >= 20 {
			break
		}
		if app.Icon == "" || app.Icon[0] == '/' {
			continue
		}
		if pb, err := theme.LoadIcon(app.Icon, 24, gtk.IconLookupForceSize); err == nil {
			iconCacheMu.Lock()
			iconCache[app.Icon] = pb
			iconCacheMu.Unlock()
		}
	}
}

func createResultRow(r modules.Result) *gtk.ListBoxRow {
	row := gtk.NewListBoxRow()
	row.SetName("spark-row")
	row.SetCanFocus(false)

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
