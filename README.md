# muster-tui

A rich TUI frontend for [muster](https://github.com/ImJustRicky/muster), the universal deploy framework.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

## Prerequisites

- [muster](https://github.com/ImJustRicky/muster) installed and on your PATH
- Go 1.21+ (to build from source)
- A muster auth token (see [Setup](#setup))

## Install

```bash
go install github.com/ImJustRicky/muster-tui@latest
```

Or build from source:

```bash
git clone https://github.com/ImJustRicky/muster-tui.git
cd muster-tui
go build -o muster-tui .
```

## Setup

muster-tui communicates with the muster CLI using token-based auth. Create a token in your muster project:

```bash
# Create a token with admin scope
muster auth create my-tui --scope admin
```

Save the token (shown only once) to muster-tui:

```bash
muster-tui --set-token <your-token>
```

Or set it as an environment variable:

```bash
export MUSTER_TOKEN=<your-token>
```

## Usage

Run from any directory containing a muster `muster.json`:

```bash
muster-tui
```

### Dashboard

The main screen shows:
- Service health status with colored indicators (green = healthy, red = unhealthy)
- Action menu: Deploy All, Deploy Service, History, Doctor, Quit
- Auto-refreshes health every 20 seconds (`r` to force refresh)

### Deploy View

- Progress bar showing current/total services
- Live streaming log output (last 6 lines)
- `Ctrl+O` to open full-screen log viewer
- Press any key to return to dashboard when complete

### Log Viewer

- Full-screen scrollable view of all deploy logs
- Auto-follows new output (disables on manual scroll)
- Colorized lines: errors in red, warnings in yellow, steps in gold
- `Ctrl+O` to close and return to deploy view

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `j/k` or `arrows` | Navigate menu |
| `enter` | Select |
| `r` | Force refresh (dashboard) |
| `ctrl+o` | Toggle log viewer (deploy) |
| `q` | Quit |
| `ctrl+c` | Force quit |

## Auth Scopes

Tokens have scopes that control access:

| Scope | Access |
|-------|--------|
| `read` | Status, history, doctor |
| `deploy` | Everything in read + deploy, rollback |
| `admin` | Full access |

## Project Structure

```
muster-tui/
├── main.go                    # Entry point
├── internal/
│   ├── auth/auth.go           # Token loading + storage
│   ├── config/config.go       # muster.json + settings.json reader
│   ├── engine/engine.go       # Calls muster CLI, parses JSON/NDJSON
│   └── tui/
│       ├── app.go             # Root model, screen routing
│       ├── styles.go          # Lip Gloss theme (gold accent)
│       ├── dashboard.go       # Services panel + action menu
│       ├── deploy.go          # Deploy progress view
│       ├── logviewer.go       # Full-screen scrollable log viewer
│       └── menu.go            # Reusable menu component
```

## How It Works

muster-tui is a frontend shell — it doesn't deploy anything itself. It calls the muster CLI with `--json` flags and parses the structured output:

- `muster status --json` — service health (JSON object)
- `muster deploy --json` — deploy events (NDJSON stream)
- `muster history --json` — event history (JSON array)
- `muster doctor --json` — diagnostics (JSON object)

All commands are authenticated via the `MUSTER_TOKEN` environment variable passed to the muster subprocess.

## License

Apache 2.0 — see [LICENSE](LICENSE).
