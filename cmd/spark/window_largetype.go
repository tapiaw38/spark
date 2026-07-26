package main

import (
	"github.com/diamondburned/gotk4/pkg/gdk/v3"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/diamondburned/gotk4/pkg/pango"
)

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
