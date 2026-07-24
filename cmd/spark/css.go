package main

import (
	"github.com/diamondburned/gotk4/pkg/gdk/v3"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
	"github.com/tapiaw38/spark/internal/config"
)

func loadCSS() {
	css := gtk.NewCSSProvider()
	css.LoadFromData(config.GetCSS() + `
		#spark-preview {
			background: rgba(0, 0, 0, 0.3);
			padding: 8px;
			border-radius: 6px;
		}
		#spark-preview-label {
			color: rgba(255, 255, 255, 0.8);
			font-family: monospace;
			font-size: 11px;
		}
		#spark-preview-image {
			background: rgba(255, 255, 255, 0.08);
			border: 1px solid rgba(255, 255, 255, 0.14);
			padding: 6px;
		}
		#spotify-view {
			background: rgba(0, 0, 0, 0.2);
			border-radius: 8px;
			padding: 12px;
		}
		#spotify-header {
			background: rgba(0, 0, 0, 0.3);
			border-radius: 8px;
			padding: 12px;
		}
		#spotify-title {
			color: white;
			font-size: 16px;
			font-weight: bold;
		}
		#spotify-artist {
			color: rgba(255, 255, 255, 0.7);
			font-size: 13px;
		}
		#spotify-album {
			color: rgba(255, 255, 255, 0.5);
			font-size: 12px;
		}
		#spotify-status {
			color: #1DB954;
			font-size: 11px;
		}
		#spotify-control {
			background: rgba(255, 255, 255, 0.1);
			border-radius: 50%;
			padding: 8px;
			min-width: 36px;
			min-height: 36px;
		}
		#spotify-control:hover {
			background: rgba(255, 255, 255, 0.2);
		}
		#spotify-list {
			background: transparent;
		}
		#spotify-list row {
			background: transparent;
			border-radius: 6px;
			padding: 6px;
		}
		#spotify-list row:selected {
			background: rgba(100, 150, 255, 0.3);
		}
	`)
	screen := gdk.ScreenGetDefault()
	gtk.StyleContextAddProviderForScreen(screen, css, uint(gtk.STYLE_PROVIDER_PRIORITY_APPLICATION))
}
