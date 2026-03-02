# muster-tui (beta)

> **Early-stage Go TUI for [muster](https://github.com/ImJustRicky/muster).** This is a companion frontend — the primary bash TUI (`muster`) receives features first. muster-tui will catch up over time.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

## Status

muster-tui covers the core workflows but is missing some features available in the regular bash TUI:

**Working:**
- Dashboard with service health + skills display
- Deploy all / deploy single service with live progress
- Full-screen log viewer with colorized output
- Status, history, doctor screens
- Logs, rollback, cleanup via raw output viewer
- Service picker for targeted operations
- Interactive settings editor (TUI mode, colors, stack, retention, etc.)
- Double Ctrl+C quit protection

**Not yet implemented:**
- Deploy failure recovery menus (retry / rollback / skip / abort)
- Dev mode (deploy + health watch loop)
- Credential prompts during deploy
- Remote deployment configuration
- Skill marketplace browsing
- Project setup wizard

For the full feature set, use the bash TUI: just run `muster`.

## Prerequisites

- [muster](https://github.com/ImJustRicky/muster) installed and on your PATH
- Go 1.21+ (to build from source)
- A muster auth token (auto-created on first launch, or see [Setup](#setup))

## Install

```bash
go install github.com/ImJustRicky/muster-tui@latest
```

Or download a binary from [Releases](https://github.com/ImJustRicky/muster-tui/releases).

Or build from source:

```bash
git clone https://github.com/ImJustRicky/muster-tui.git
cd muster-tui
go build -o muster-tui .
```

## Setup

When you run `muster` with muster-tui installed, it auto-creates an auth token and launches the Go TUI. No manual setup needed.

To switch between TUIs, open **Settings** in either interface and change **TUI Mode** to `go` or `bash`.

### Manual token setup

```bash
# Create a token with admin scope
muster auth create my-tui --scope admin

# Save to muster-tui
muster-tui --set-token <your-token>
```

Or set as environment variable:

```bash
export MUSTER_TOKEN=<your-token>
```

## Usage

```bash
muster-tui
```

Or just run `muster` — it will launch muster-tui automatically when installed and TUI mode is set to `go`.

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `j/k` or `arrows` | Navigate |
| `enter` | Select |
| `r` | Force refresh (dashboard, status) |
| `ctrl+o` | Toggle log viewer (deploy) |
| `a` | Toggle show-all (history) |
| `esc` | Go back |
| `q` | Quit / go back |
| `ctrl+c` (x2) | Force quit |

## How It Works

muster-tui is a frontend shell — it doesn't deploy anything itself. It calls the muster CLI with `--json` flags and parses the structured output:

- `muster status --json` — service health
- `muster deploy --json` — deploy events (NDJSON stream)
- `muster history --json` — event history
- `muster doctor --json` — diagnostics

Commands without `--json` support (logs, rollback, cleanup) stream raw output into a scrollable viewport.

## License

Apache 2.0 — see [LICENSE](LICENSE).
