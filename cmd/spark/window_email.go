package main

import (
	"strings"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/tapiaw38/spark/internal/modules"
)

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
		executeActionSpec(modules.EmailAction(to, subject, body, splitPaths(attachmentsEntry.Text())...))
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
