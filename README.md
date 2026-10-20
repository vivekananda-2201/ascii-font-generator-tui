# ascii-font-generator-tui (agen)

[![Go Reference](https://pkg.go.dev/badge/github.com/vivekananda-2201/ascii-font-generator-tui.svg)](https://pkg.go.dev/github.com/vivekananda-2201/ascii-font-generator-tui)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)

```text
    _   ___  ___ _____ ___   ___ ___ _  _ 
   /_\ / __|/ __|_   _|_ _| | __/ _ \ \| |
  / _ \\__ \ (__  | |  | |  | _| (_) | .` |
 /_/ \_\___/\___| |_| |___| |_| \___/|_|\_|
  Interactive Terminal FIGlet ASCII Art Studio
```

**ascii-font-generator-tui** (binary: `agen`) is an interactive Terminal User Interface (TUI) application designed for real-time browsing, fuzzy searching, previewing, customizing, and generating FIGlet ASCII art text directly from your terminal.

Say goodbye to the trial-and-error cycle of running `figlet -f <fontname>` dozens of times or relying on ad-heavy web generators. With instant keystroke feedback, automatic font discovery across your system, variant grouping, and seamless clipboard and `stdout` integration, generating ASCII banners and terminal art has never been easier.

Built with [Charm](https://charm.sh)'s **Bubble Tea** and **Lip Gloss** alongside **figlet4go**.

---

## ⚡ Highlights & Purpose

- **Live Split-Pane Preview**: Edit text in the top input and instantly observe the rendered FIGlet ASCII art in the right viewport, alongside a responsive font list on the left.
- **Instant Fuzzy Search (`Ctrl+S`)**: Rapidly filter through hundreds of fonts (e.g. *Slant*, *Shadow*, *Doom*, *Bloody*, *Standard*) using fuzzy string matching.
- **Variant & Spacing Selector (`Enter`)**:
  - Automatically identifies and groups related font families (e.g., `ANSI Regular`, `ANSI Shadow`, `Small Shadow`).
  - Adjust character spacing and layout modes on the fly (`Default`, `Wide`, `Extra Wide`, `Compact`) within an interactive centered modal dialog.
- **Massive Font Discovery**: Automatically detects `.flf` font files across standard paths:
  - Bundled FIGlet defaults (`standard`, `larry3d`)
  - User font directory (`~/.local/share/figlet-fonts/`)
  - System FIGlet paths (`/usr/share/figlet/fonts/`, `/usr/share/figlet/`, `/usr/share/fonts/figlet/`)
  - Custom directories supplied at runtime via `-d` / `--font-dir`
- **Unix Pipeline Friendly (`stdout`)**: Pressing `Enter` to confirm cleanly exits the TUI and prints the resulting ASCII art directly to `stdout`. Perfect for redirecting to a file (`agen > banner.txt`) or piping to tools like `lolcat`, `cowsay`, or MOTD scripts.
- **Clipboard Integration**:
  - Copy active art preview on demand at any time with `Ctrl+Y`.
  - Pass `-c` / `--clip` to automatically copy the chosen banner to the system clipboard upon exit.
- **Terminal State Safety**: Runs in an alternate screen buffer (`WithAltScreen`), handles window resizing dynamically, and safely restores your cursor and terminal state upon exit.

---

## 📦 Installation

### From Source (using Go)

```bash
# Install directly via Go
go install github.com/vivekananda-2201/ascii-font-generator-tui@latest
```

Make sure your `$GOPATH/bin` or `$HOME/go/bin` is in your `$PATH`:
```bash
export PATH="$HOME/go/bin:$PATH"
```

### Clone and Build Locally

```bash
# Clone the repository
git clone https://github.com/vivekananda-2201/ascii-font-generator-tui.git
cd ascii-font-generator-tui

# Build the binary (named 'agen')
go build -o agen .

# Optionally install into /usr/local/bin or ~/.local/bin
sudo mv agen /usr/local/bin/
```

Verify your installation:
```bash
agen --version
```

---

## 🚀 Quick Start & Usage

### 1. Launch Interactive TUI
Simply run `agen`:
```bash
agen
```

### 2. Copy Automatically to Clipboard on Exit
```bash
agen --clip
# or shorthand
agen -c
```

### 3. Load Additional Fonts from Custom Directory
```bash
agen --font-dir /path/to/my/figlet-fonts
# or shorthand
agen -d /path/to/my/figlet-fonts
```

### 4. Direct Pipeline & Redirection
Generate your banner interactively, confirm with `Enter`, and stream the output directly to a file:
```bash
# Save generated art to a header file
agen > header.txt

# Pipe to lolcat for rainbow-colored terminal art
agen | lolcat

# Embed into your shell login / MOTD
agen > ~/.config/shell-banner.txt
```

---

## ⌨️ Keybindings

| Key | Context | Action |
|---|---|---|
| `Tab` | Main Screen | Toggle focus between **Text Input** and **Font List** |
| `↑` / `k` | Font List | Navigate to previous font |
| `↓` / `j` | Font List | Navigate to next font |
| `PgUp` / `PgDn` | Font List | Jump 10 fonts up or down |
| `Ctrl+S` | Main Screen | Open inline fuzzy font search overlay |
| `Esc` | Search Mode | Cancel search and restore previous selection |
| `←` / `→` | Main Screen | Pan wide ASCII art horizontally |
| `Ctrl+Y` | Any | Copy current preview directly to system clipboard |
| `Enter` | Main Screen | Open **Size & Variant Modal** for selected font |
| `↑` / `↓` | Modal | Navigate size variants & spacing styles |
| `Enter` | Modal | Confirm selection, exit TUI, and write art to `stdout` |
| `Esc` | Modal | Dismiss modal and return to font browser |
| `Esc` / `Ctrl+C` | Main Screen | Exit cleanly without emitting output to `stdout` |

---

## 🔤 Adding Custom FIGlet Fonts

Want hundreds more fonts? You can easily download or drop any `.flf` (FIGlet Font) files:

1. **User Font Directory** (auto-scanned on startup):
   ```bash
   mkdir -p ~/.local/share/figlet-fonts
   # Example: clone xero's comprehensive FIGlet font collection
   git clone https://github.com/xero/figlet-fonts.git /tmp/figlet-fonts
   cp /tmp/figlet-fonts/*.flf ~/.local/share/figlet-fonts/
   ```

2. **Package Manager** (Debian/Ubuntu/Arch/Fedora):
   ```bash
   # Debian / Ubuntu
   sudo apt-get install figlet-fonts

   # Arch Linux
   sudo pacman -S figlet
   ```

3. **Runtime Flag**:
   ```bash
   agen --font-dir ~/Downloads/custom-fonts/
   ```

`ascii-font-generator-tui` automatically sanitizes, validates, and indexes fonts without crashing on corrupt files or non-standard glyph headers.

---

## 🛠️ CLI Options

```text
Usage: agen [options]

Options:
  -c, --clip       Copy output directly to clipboard upon exit
  -d, --font-dir   Scan additional directory for .flf fonts
  -v, --version    Print version and exit
  -h, --help       Show help message
```

---

## 🏛️ Project Architecture

```
ascii-font-generator-tui/
├── main.go                     # Entrypoint, command-line flags, stdout emission
├── go.mod                      # Module definition & dependency management
├── go.sum                      # Checksums for direct and indirect dependencies
├── internal/
│   ├── app/                    # Charm Bubble Tea model, view & update logic
│   │   ├── model.go            # State machine (Browse, Search, Variant modal)
│   │   ├── update.go           # Key event routing, search filtering & viewport
│   │   ├── view.go             # Lip Gloss styling, dual-pane layouts & modal UI
│   │   ├── styles.go           # Colors, border styles, typography definitions
│   │   ├── app_test.go         # Terminal bounds, navigation & layout tests
│   │   └── variant_modal_test.go # Variant modal stress tests across resolutions
│   ├── fonts/                  # Font indexing and FIGlet rendering engine
│   │   ├── loader.go           # Multi-directory .flf discovery & sanitization
│   │   ├── renderer.go         # figlet4go wrapper with spacing & cache support
│   │   ├── families.go         # Family grouping and variant options detector
│   │   └── fonts_test.go       # Font discovery, glyph order & rendering tests
│   └── clipboard/              # System clipboard provider
│       └── clipboard.go        # Cross-platform clipboard handler (wl-copy, xclip, pbcopy)
└── LICENSE                     # MIT License
```

---

## 🧪 Testing

Run all unit, modal geometry, and font engine tests:

```bash
go test -v ./...
```

Run race detector and benchmarks:

```bash
go test -race -v ./...
```

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).
