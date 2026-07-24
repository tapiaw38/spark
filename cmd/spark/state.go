package main

import (
	"context"
	"sync"
	"time"

	"github.com/diamondburned/gotk4/pkg/gdkpixbuf/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/tapiaw38/spark/internal/apps"
	"github.com/tapiaw38/spark/internal/modules"
)

var (
	allApps         []apps.App
	currentResults  []modules.Result
	listBox         *gtk.ListBox
	resultsScroll   *gtk.ScrolledWindow
	searchEntry     *gtk.Entry
	previewBox      *gtk.Box
	previewToolbar  *gtk.Box
	previewMeta     *gtk.Label
	previewImage    *gtk.Image
	previewLabel    *gtk.Label
	searchMu        sync.Mutex
	searchVersion   uint64
	previewVersion  uint64
	fileSearchMu    sync.Mutex
	fileSearchStop  context.CancelFunc
	quickLookActive bool
	previewPage     int
	previewScale    int
	inActionMode    bool
	debounceTimer   *time.Timer
	iconCache       = make(map[string]*gdkpixbuf.Pixbuf)
	iconCacheMu     sync.RWMutex

	// Spotify view widgets
	spotifyView     *gtk.Box
	spotifyArtBig   *gtk.Image
	spotifyArtSmall *gtk.Image
	spotifyTitle    *gtk.Label
	spotifyArtist   *gtk.Label
	spotifyAlbum    *gtk.Label
	spotifyStatus   *gtk.Label
	spotifyList     *gtk.ListBox
	inSpotifyMode   bool
	playerMode      modules.PlayerKind
	mainBox         *gtk.Box
	isClearing      bool
)

const (
	dragInfoURI uint = iota + 1
	dragInfoText
)
