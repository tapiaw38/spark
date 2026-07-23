# Spark

Alfred-like app launcher for Linux (Wayland/MangoWM).

## Stack

- **Language:** Go
- **UI:** GTK3 (gotk4 bindings)
- **Wayland:** gtk-layer-shell
- **Config:** YAML (~/.config/spark/config.yaml)

## Build

```bash
go build -buildvcs=false -o spark ./cmd/spark/
```

## Run

```bash
./spark
```

## Setup Hotkey

```bash
./spark --setup  # Updates ~/.config/mango/bind.conf with hotkey from config
```

## Project Structure

```
spark/
├── cmd/spark/main.go       # Entry point, GTK window, UI logic
├── internal/
│   ├── apps/apps.go        # .desktop file parsing, app search
│   ├── config/config.go    # YAML config, CSS generation, hotkey setup
│   ├── history/history.go  # App launch frequency tracking
│   └── modules/
│       ├── modules.go      # Result struct definition
│       ├── calc.go         # Calculator (2+2)
│       ├── web.go          # Web shortcuts (g, yt, gh, etc.)
│       ├── system.go       # System commands (lock, shutdown, etc.)
│       ├── shell.go        # Shell execution (> command)
│       ├── files.go        # File search (f prefix)
│       ├── file_actions.go # File actions + buffer
│       ├── file_ops.go     # Rename/copy/move operations
│       ├── navigation.go   # Folder navigation
│       ├── clipboard.go    # Clipboard history (clip/cb prefix)
│       ├── snippets.go     # Text expansion (;keyword)
│       ├── dictionary.go   # Word definitions (define/def)
│       ├── spell.go        # Spelling suggestions (spell prefix)
│       ├── recent.go       # Recent documents
│       ├── large_type.go   # Large Type overlay
│       ├── help.go         # Built-in help
│       ├── contacts.go     # Local vCard contacts
│       ├── email.go        # Email compose helpers
│       ├── stats.go        # Usage stats
│       ├── sync.go         # Settings sync helpers
│       ├── preview.go      # Preview pane content
│       ├── music.go        # Local music search (m prefix)
│       ├── youtube.go      # YouTube video search + thumbnails (yt prefix)
│       └── spotify.go      # Music control (sp prefix)
```

## Features & Prefixes

| Prefix | Feature | Example |
|--------|---------|---------|
| (none) | App search | `firefox` |
| `>` | Shell command | `> htop` |
| `;` | Snippet | `;email` |
| `f` | File search | `f readme` |
| `Tab` | File actions | select file, press `Tab` |
| file op window | Visual file ops | Tab -> Rename/Copy/Move |
| `nav` | Folder navigation | `nav ~/Downloads` |
| `pick` | Destination picker | `pick copy source | ~/Downloads` |
| `rename` | Rename file | `rename source | new-name` |
| `copy` | Copy file | `copy source | target` |
| `move` | Move file | `move source | target` |
| `undo` | Undo last file operation | `undo` |
| `status` | Last action/error | `status` |
| `buffer` | File buffer | `buffer` |
| `recent` | Recent documents | `recent invoice`, `recent app firefox` |
| `large` | Large Type | `large 555-1234` |
| `large all` | Large Type all monitors | `large all 555-1234` |
| `help` | Help | `help` |
| `contact` | Contacts | `contact Ada` |
| `email` | Email | `email contact | Subject | Body` |
| `stats` | Usage Stats | `stats` |
| `sync` | Sync Settings | `sync`, `sync import ~/spark-settings.zip` |
| `clip` | Clipboard | `clip` |
| `define` | Dictionary | `define word` |
| `spell` | Spelling | `spell recieve` |
| `sp` | Spotify/Music | `sp` |
| `m` | Local music | `m song`, `m artist name`, `m album name` |
| `mq` | Music queue | `mq` |
| `g` | Google search | `g query` |
| `yt` | YouTube videos | `yt video` |
| `gh` | GitHub | `gh repo` |
| `lock` | Lock screen | `lock` |
| `shutdown` | Shutdown | `shutdown` |

## Config

Location: `~/.config/spark/config.yaml`

```yaml
width: 600
max_results: 6
background_color: "30, 30, 40"
background_alpha: 0.95
border_radius: 12
font_size: 18
text_color: white
selection_color: "100, 150, 255"
show_icons: true
icon_size: 24
margin_top: 100
history_boost: 3
hotkey: "SUPER,s"
spell_language: "en"
web_shortcuts:
  g:
    name: Google
    url: "https://www.google.com/search?q=%s"
    icon: web-browser
```

## Key Files to Modify

### Adding a new module

1. Create `internal/modules/newmodule.go`
2. Implement `func NewModuleSearch(query string) []Result`
3. Add call in `cmd/spark/main.go` in `updateResults()`

### Result struct (modules/modules.go)

```go
type Result struct {
    Type         string   // Module identifier
    Title        string   // Main text
    Desc         string   // Secondary text
    Icon         string   // Icon name or path
    Preview      string   // Preview pane text
    PreviewImage string   // Preview pane image path
    Action       func()   // Execute on Enter
}
```

### Spotify view

Special UI in `cmd/spark/main.go`:
- `createSpotifyView()` - Build layout
- `showSpotifyView()` / `hideSpotifyView()` - Toggle
- `refreshSpotifyInfo()` - Update track info

Uses `playerctl` for MPRIS control.

## Dependencies

System packages (Arch):
```bash
pacman -S gtk3 gtk-layer-shell playerctl yt-dlp aspell hunspell
```

Go modules:
```bash
go mod tidy
```

## Debug

Run directly to see GTK errors:
```bash
./spark 2>&1
```

Test playerctl:
```bash
playerctl metadata
playerctl status
```
