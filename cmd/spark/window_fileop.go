package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/tapiaw38/spark/internal/modules"
)

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
