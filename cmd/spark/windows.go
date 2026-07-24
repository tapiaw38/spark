package main

import (
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gdk/v3"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/diamondburned/gotk4/pkg/pango"
	"github.com/tapiaw38/spark/internal/history"
	"github.com/tapiaw38/spark/internal/modules"
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
	sort.Slice(stats, func(i, j int) bool { return stats[i].count > stats[j].count })
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
		bar.SetText(strconv.Itoa(s.count))
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
