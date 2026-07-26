package main

import (
	"strconv"

	"github.com/diamondburned/gotk4/pkg/gdk/v3"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/tapiaw38/spark/internal/history"
)

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
