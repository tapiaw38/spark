# Spark

App launcher for Linux (Wayland/MangoWM).

## Stack

- **Language:** Go
- **UI:** GTK3 (gotk4 bindings)
- **Wayland:** gtk-layer-shell
- **Config:** YAML (~/.config/spark/config.yaml)

## Build

```bash
go build -buildvcs=false -o spark ./cmd/spark/
```

## Development

Format:

```bash
gofmt -w $(find . -name '*.go' -not -path './vendor/*')
```

Test:

```bash
go test ./...
```

## Run

```bash
./spark
```

## Setup Hotkey

```bash
./spark --setup  # Updates ~/.config/mango/bind.conf with hotkey from config
```

## Architecture

Spark uses a lightweight hexagonal style:

- `internal/modules` owns search logic and returns declarative `Result` values.
- `Result.ActionSpec` describes what should happen when the user presses Enter.
- `cmd/spark/action_executor.go` is the UI adapter that executes actions against the desktop/session.
- `internal/platform/commands` is the only package that wraps `os/exec`.

Modules should prefer `ActionSpec` helpers such as `OpenAction`, `CopyAction`, `TerminalAction`, `EmailAction`, `FileAction`, `MusicAction`, `SyncAction`, and `SystemAction` instead of running shell commands directly.

## Project Structure

```bash
spark/
├── cmd/spark/
│   ├── main.go             # Entry point and command dispatch
│   ├── launcher.go         # GTK launcher window
│   ├── action_executor.go  # Executes modules.ActionSpec values
│   └── window_*.go         # Focused GTK windows/dialogs
├── internal/
│   ├── apps/apps.go        # .desktop file parsing, app search
│   ├── config/config.go    # YAML config, CSS generation, hotkey setup
│   ├── history/history.go  # App launch frequency tracking
│   ├── platform/
│   │   └── commands/       # Wrapper around os/exec
│   └── modules/
│       ├── modules.go      # Result, ActionSpec, registries
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

| Prefix            | Feature                       | Example                                                        |
| ----------------- | ----------------------------- | -------------------------------------------------------------- |
| (none)            | App search                    | `firefox`                                                      |
| `>`               | Shell command                 | `> htop`                                                       |
| `;`               | Snippet                       | `;email`                                                       |
| `f`               | File search                   | `f readme`                                                     |
| `Tab`             | File actions                  | select file, press `Tab`                                       |
| file op window    | Visual file ops               | Tab -> Rename/Copy/Move                                        |
| `nav`             | Folder navigation             | `nav ~/Downloads`                                              |
| `pick`            | Destination picker            | `pick copy source \| ~/Downloads`                              |
| `rename`          | Rename file                   | `rename source \| new-name`                                    |
| `copy`            | Copy file                     | `copy source \| target`                                        |
| `move`            | Move file                     | `move source \| target`                                        |
| `undo`            | Undo last file operation      | `undo`                                                         |
| `status`          | Last action/error             | `status`                                                       |
| `buffer`          | File buffer                   | `buffer`                                                       |
| `recent`          | Recent documents              | `recent invoice`, `recent app firefox`                         |
| `large`           | Large Type                    | `large 555-1234`                                               |
| `large all`       | Large Type all monitors       | `large all 555-1234`                                           |
| `help`            | Help                          | `help`                                                         |
| `contact`         | Contacts                      | `contact Ada`, `contact carddav`                               |
| `email`           | Email composer + attachments  | `email contact \| Subject \| Body`                             |
| `stats`           | Usage Stats                   | `stats`                                                        |
| `sync`            | Sync Settings                 | `sync`, `sync import ~/spark-settings.zip`                     |
| `clip`            | Clipboard                     | `clip`                                                         |
| `define`          | Dictionary                    | `define word`                                                  |
| `spell`           | Spelling                      | `spell recieve`                                                |
| `sp`              | Spotify/Music                 | `sp`                                                           |
| `yp`              | YouTube player controls       | `yp`                                                           |
| `m`               | Local music                   | `m song`, `m artists`, `m artist name`, `m albums`, `m genres` |
| `mq`              | Music queue                   | `mq`, play with mpv                                            |
| `g`               | Google search                 | `g query`                                                      |
| `yt`              | YouTube videos                | `yt video`                                                     |
| `gh`              | GitHub                        | `gh repo`                                                      |
| `lock`            | Lock screen                   | `lock`                                                         |
| `shutdown`        | Shutdown                      | `shutdown`                                                     |
| `emoji`           | Emoji picker                  | `emoji fire`                                                   |
| `b64`/`b64d`      | Base64 encode/decode          | `b64 hello`                                                    |
| `url`/`urld`      | URL encode/decode             | `url a b&c`                                                    |
| `hash`            | SHA-256                       | `hash secret`                                                  |
| `uuid`            | Generate UUID v4              | `uuid`                                                         |
| `epoch`           | Unix time / date              | `epoch`, `epoch 1700000000`                                    |
| (units)           | Unit conversion               | `100 km to mi`, `50f to c`, `5 gb to mb`                       |
| `ssh`             | SSH hosts from ~/.ssh/config  | `ssh`, `ssh prod`                                              |
| `kill`            | Kill process                  | `kill firefox`                                                 |
| `screenshot`/`ss` | Screenshot (grim/slurp)       | `ss`                                                           |
| `w`               | Windows or MangoWM workspaces | `w`, `w firefox`, `w 2`                                        |
| `timer`           | Countdown -> notification     | `timer 5m`                                                     |
| `wt`              | Weather from wttr.in          | `wt`, `wt Berlin`                                              |
| `pass`            | Password store (pass)         | `pass github`                                                  |
| `bm`              | Browser bookmarks             | `bm docs`                                                      |

Quick Look: press `Shift`; PDF/doc previews show page/zoom controls, and `PageUp/PageDown` plus `+/-` also work.
File ops: Tab on a file, then Rename/Copy/Move opens a visual picker with breadcrumbs and folder browsing.

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
3. Return `Result` values with `ActionSpec` for Enter behavior
4. Register the module in `internal/modules/registry.go`

### Result struct (modules/modules.go)

```go
type Result struct {
    Type            ResultType // Module/result identifier
    Title           string     // Main text
    Desc            string     // Secondary text
    Icon            string     // Icon name or path
    IconText        string     // Text fallback for icon display
    Preview         string     // Preview pane text
    PreviewImage    string     // Preview pane image path
    PreviewImageURL string     // Remote image to cache for preview
    Data            string     // Module-specific data
    KeepOpen        bool       // Keep launcher open after action
    Confirm         bool       // Ask before executing action
    NavigateQuery   string     // Replace query instead of executing
    ActionSpec      ActionSpec // Declarative action executed by cmd/spark
    Action          func()     // Legacy escape hatch; avoid for new modules
}
```

Common action helpers:

```go
OpenAction(pathOrURL)
CopyAction(text)
TerminalAction(command)
EmailAction(to, subject, body, attachments...)
FileAction(op, args...)
MusicAction(op, args...)
SyncAction(op, args...)
SystemAction(op)
```

### Spotify view

Special UI in `cmd/spark/player.go`:

- `createSpotifyView()` - Build layout
- `showSpotifyView()` / `hideSpotifyView()` - Toggle
- `refreshSpotifyInfo()` - Update track info

Uses `playerctl` for MPRIS metadata and `ActionSpec` for controls.

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

## License

MIT. See [LICENSE](LICENSE).
