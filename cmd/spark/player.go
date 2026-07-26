package main

import (
	"sync/atomic"
	"time"

	"github.com/diamondburned/gotk4/pkg/gdkpixbuf/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/diamondburned/gotk4/pkg/pango"
	"github.com/tapiaw38/spark/internal/modules"
)

const (
	playerArtSmallSize = 64
	playerArtBigSize   = 150

	playerRefreshDelay = 300 * time.Millisecond
)

func createSpotifyView() {
	spotifyView = gtk.NewBox(gtk.OrientationVertical, 8)
	spotifyView.SetName("spotify-view")

	header := gtk.NewBox(gtk.OrientationHorizontal, 12)
	header.SetName("spotify-header")

	spotifyArtSmall = gtk.NewImage()
	spotifyArtSmall.SetSizeRequest(playerArtSmallSize, playerArtSmallSize)
	header.PackStart(spotifyArtSmall, false, false, 0)

	infoBox := gtk.NewBox(gtk.OrientationVertical, 4)
	spotifyTitle = gtk.NewLabel("")
	spotifyTitle.SetName("spotify-title")
	spotifyTitle.SetXAlign(0)
	spotifyTitle.SetEllipsize(pango.EllipsizeEnd)

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

	header.PackEnd(buildPlayerControls(), false, false, 0)

	spotifyView.PackStart(header, false, false, 0)

	content := gtk.NewBox(gtk.OrientationHorizontal, 12)

	spotifyList = gtk.NewListBox()
	spotifyList.SetName("spotify-list")
	spotifyList.SetSelectionMode(gtk.SelectionSingle)

	for _, ctrl := range playerMode.Controls() {
		row := createSpotifyControlRow(ctrl)
		spotifyList.Add(row)
		row.ShowAll()
	}

	spotifyList.Connect("row-activated", func(_ *gtk.ListBox, row *gtk.ListBoxRow) {
		idx := row.Index()
		ctrls := playerMode.Controls()
		if idx >= 0 && idx < len(ctrls) {
			runPlayerResultAction(ctrls[idx])
		}
	})

	content.PackStart(spotifyList, true, true, 0)

	artFrame := gtk.NewBox(gtk.OrientationVertical, 0)
	artFrame.SetHAlign(gtk.AlignCenter)
	artFrame.SetVAlign(gtk.AlignCenter)
	spotifyArtBig = gtk.NewImage()
	spotifyArtBig.SetSizeRequest(playerArtBigSize, playerArtBigSize)
	artFrame.PackStart(spotifyArtBig, false, false, 0)
	content.PackEnd(artFrame, false, false, 0)

	spotifyView.PackStart(content, false, false, 0)
}

func buildPlayerControls() *gtk.Box {
	controls := gtk.NewBox(gtk.OrientationHorizontal, 8)
	controls.SetHAlign(gtk.AlignEnd)
	controls.SetVAlign(gtk.AlignCenter)

	buttons := []struct {
		label string
		index int
	}{
		{"⏮", 2},
		{"⏯", 0},
		{"⏭", 1},
	}
	for _, b := range buttons {
		btn := gtk.NewButton()
		btn.SetName("spotify-control")
		btn.SetLabel(b.label)
		idx := b.index
		btn.Connect("clicked", func() {
			runPlayerResultAction(playerMode.Controls()[idx])
		})
		controls.PackStart(btn, false, false, 0)
	}
	return controls
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

func showSpotifyView() {
	showPlayerView(modules.PlayerSpotify)
}

func showPlayerView(kind modules.PlayerKind) {
	if inSpotifyMode && playerMode == kind {
		return
	}
	playerMode = kind
	inSpotifyMode = true
	resultsScroll.Hide()
	hidePreview()
	refreshPlayerControls()

	refreshSpotifyInfo()

	spotifyView.Show()
	spotifyView.ShowAll()

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
	for _, ctrl := range playerMode.Controls() {
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
	mode := playerMode
	version := atomic.AddUint64(&playerInfoVersion, 1)
	go func() {
		info := modules.GetPlayerInfo(mode)
		glib.IdleAdd(func() {
			if atomic.LoadUint64(&playerInfoVersion) != version {
				return
			}
			applyPlayerInfo(info)
		})
	}()
}

func applyPlayerInfo(info *modules.PlayerInfo) {
	if info == nil {
		showPlayerDisconnected()
		return
	}

	spotifyTitle.SetText(info.Title)
	spotifyArtist.SetText(info.Artist)
	spotifyAlbum.SetText(info.Album)
	spotifyStatus.SetText(playerStatusIcon(info.Status) + " " + string(info.Status))

	if info.ArtCachePath != "" {
		if pb, err := gdkpixbuf.NewPixbufFromFileAtSize(info.ArtCachePath, playerArtSmallSize, playerArtSmallSize); err == nil {
			spotifyArtSmall.SetFromPixbuf(pb)
		}
		if pb, err := gdkpixbuf.NewPixbufFromFileAtSize(info.ArtCachePath, playerArtBigSize, playerArtBigSize); err == nil {
			spotifyArtBig.SetFromPixbuf(pb)
		}
	}
}

func runPlayerResultAction(result modules.Result) {
	go func() {
		executeActionSpec(result.ActionSpec)
		time.Sleep(playerRefreshDelay)
		glib.IdleAdd(func() { refreshSpotifyInfo() })
	}()
}

func showPlayerDisconnected() {
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
}

func playerStatusIcon(status modules.PlaybackStatus) string {
	switch status {
	case modules.StatusPaused:
		return "⏸"
	case modules.StatusStopped:
		return "⏹"
	default:
		return "▶"
	}
}
