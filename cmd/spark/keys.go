package main

import (
	"github.com/diamondburned/gotk4/pkg/gdk/v3"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
)

const (
	previewScaleDefault = 360
	previewScaleStep    = 60
	previewScaleMin     = 180
	previewScaleMax     = 720
)

func onKeyPress(event *gdk.Event) bool {
	switch event.AsKey().Keyval() {
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
		toggleQuickLook()
		return true
	case gdk.KEY_Page_Down, gdk.KEY_Right:
		return quickLookActive && stepPreviewPage(1)
	case gdk.KEY_Page_Up, gdk.KEY_Left:
		return quickLookActive && stepPreviewPage(-1)
	case gdk.KEY_plus, gdk.KEY_KP_Add, gdk.KEY_equal:
		return quickLookActive && stepPreviewZoom(previewScaleStep)
	case gdk.KEY_minus, gdk.KEY_KP_Subtract:
		return quickLookActive && stepPreviewZoom(-previewScaleStep)
	}
	return false
}

func toggleQuickLook() {
	quickLookActive = !quickLookActive
	if !quickLookActive {
		hidePreview()
		return
	}
	previewPage = 1
	if previewScale == 0 {
		previewScale = previewScaleDefault
	}
	updatePreview(listBox.SelectedRow())
}

func stepPreviewPage(delta int) bool {
	previewPage += delta
	if previewPage < 1 {
		previewPage = 1
	}
	updatePreview(listBox.SelectedRow())
	return true
}

func stepPreviewZoom(delta int) bool {
	if previewScale == 0 {
		previewScale = previewScaleDefault
	}
	previewScale += delta
	if previewScale < previewScaleMin {
		previewScale = previewScaleMin
	}
	if previewScale > previewScaleMax {
		previewScale = previewScaleMax
	}
	updatePreview(listBox.SelectedRow())
	return true
}
