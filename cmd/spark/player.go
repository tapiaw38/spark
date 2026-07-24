package main

import (
	"github.com/diamondburned/gotk4/pkg/gdkpixbuf/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/tapiaw38/spark/internal/modules"
)

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
