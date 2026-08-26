# monobox

<p align="center">
  <a href="https://github.com/JoaoOliveira889/monobox/releases/latest"><img src="https://img.shields.io/github/v/release/JoaoOliveira889/monobox?color=7aa2f7&label=tag&logo=github&style=flat-square" alt="Latest Tag"></a>
  <a href="https://github.com/JoaoOliveira889/monobox/releases/latest"><img src="https://img.shields.io/github/downloads/JoaoOliveira889/monobox/total?color=9ece6a&label=downloads&logo=github&style=flat-square" alt="Total Downloads"></a>
  <a href="https://goreportcard.com/report/github.com/JoaoOliveira889/monobox"><img src="https://goreportcard.com/badge/github.com/JoaoOliveira889/monobox?style=flat-square" alt="Go Report Card"></a>
  <a href="https://github.com/JoaoOliveira889/homebrew-tap"><img src="https://img.shields.io/badge/homebrew-v0.0.6-b8ff3d?logo=homebrew&style=flat-square" alt="Homebrew Version"></a>
</p>

<p align="center">
  <a href="https://github.com/JoaoOliveira889/monobox"><strong>MonoBox v0.0.6 · JoaoOliveira889/monobox</strong></a>
</p>

**Terminal UI for Docker and Podman containers.** A TUI tool that lists all your containers with live status, log streaming, and one-key lifecycle actions — with confirmation guards for every mutating command.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss), and [Bubbles](https://github.com/charmbracelet/bubbles).

---

## Features

- **Auto-detect engine** — Docker first, then Podman; clear error if neither is available
- **Focused dashboard** — compact engine summary, stable-width service icons, high-contrast selection, and contextual quick actions
- **Decoupled high-performance polling** — asynchronous container list and stats refreshes with instant UI response
- **Live log streaming** — real-time `docker logs -f` / `podman logs -f` with follow mode, log export (`s`), search match navigation (`n`/`N`), severity filtering (`!`), and Regex pattern search (`Ctrl+R`)
- **Lifecycle & Compose actions** — start/stop toggle, pause/unpause, restart, and batch Compose `up -d` (`u`) / `down` (`D`)
- **System Prune Modal (`P`)** — reclaim disk space by safely removing stopped containers, unused networks, and dangling/unused images
- **In-App Settings Editor (`S`)** — customize themes, refresh rates, log line & tail limits directly in the TUI
- **Clipboard Integration (`y` / `Y`)** — one-key copy of container IDs or full container metadata
- **Historical Metrics Graph (`g`)** — full-screen CPU & Memory timeline chart with min/max/current stats
- **Port Conflict Detection (`⚠️`)** — automatic detection and confirmation guard before starting containers with occupied host ports
- **Interactive Theme Menu (`T`)** — switch themes in real time (Tokyo Night, Dracula, Nord, Gruvbox, Monokai, One Dark)
- **User Config (`~/.config/monobox/config.yaml`)** — persist default theme, metrics refresh interval, log line & tail limits & timestamps
- **Container Inspection DX (`E` & `H`)** — interactive Environment Variables modal (`E`) with secret masking (`m`), Healthcheck probe logs modal (`H`), Mounts & Network details
- **Smart Shell Fallback (`e` / `x`)** — interactive shell execution with automatic `/bin/bash` → `/bin/sh` → `/bin/ash` fallback
- **Metrics & High-Load Alerts (`⚡` / `🔥`)** — visual warning badges when CPU/Memory usage exceeds 80% or 90%
- **Clean Port Mapping** — deduplicated host-to-container port mapping (e.g. `5115 ➔ 8080/tcp`) and host URL (`http://localhost:5115`)
- **Categorized Shortcuts (`?`)** — MonoGit-style 2-column shortcuts overlay with scrollable viewport
- **Security first** — exec with discrete arguments, no shell injection, log file at `0600`

---

## Keybindings

### Global

| Key | Action |
|-----|--------|
| `↑` / `k`, `↓` / `j` | Navigate containers / scroll modal |
| `←` / `h`, `→` / `l` | Switch panel focus |
| `1`, `2` | Jump to Containers / Logs panel |
| `tab` | Toggle panel focus |
| `<` / `,`, `>` / `.` | Resize panel divider |
| `S` | In-app Settings Editor modal |
| `P` | System Prune modal |
| `T` | Interactive Theme Menu modal |
| `?` | Shortcuts Help Overlay modal |
| `q` / `Ctrl+C` | Quit |

### Container List Panel

| Key | Action |
|-----|--------|
| `Enter` | Open live log stream |
| `/` | Real-time container search & filter |
| `e` / `x` | Exec interactive shell (bash / sh / ash fallback) |
| `E` | View Environment Variables modal (`m` to toggle secret masking) |
| `H` | View Healthcheck Probe Logs modal |
| `i` | Inspect container configuration (JSON) |
| `s` | Start / stop toggle |
| `p` | Pause / unpause container |
| `r` | Restart container |
| `d` / `Delete` | Remove container (confirms with `y` / `n`) |
| `y` | Copy Container ID to clipboard |
| `Y` | Copy Container info to clipboard |
| `o` | Open mapped host port in browser (`http://localhost:port`) |
| `u` *(on group)* | Compose Up (`docker compose up -d`) |
| `D` *(on group)* | Compose Down (`docker compose down`, confirms) |

### Log Panel

| Key | Action |
|-----|--------|
| `↑↓` / `jk` | Scroll logs |
| `PgUp` / `Ctrl+U` | Page up |
| `PgDn` / `Ctrl+D` | Page down |
| `End` / `G` | Jump to bottom |
| `f` | Toggle follow mode |
| `t` | Toggle log timestamps |
| `!` | Cycle log severity filter (`ALL` → `INFO+` → `WARN+` → `ERROR`) |
| `/` | Search / filter logs |
| `n` / `N` | Next / previous search match |
| `Ctrl+R` | Toggle Regex pattern search |
| `s` / `Ctrl+S` | Export logs to `.log` file |
| `c` / `Ctrl+L` | Clear engine log history (confirms with `y` / `n`) |
| `Esc` | Back to container list |

---

## Icon Legend

| Icon | Service |
|------|---------|
| 🐘 | Postgres / PostgreSQL |
| ⚡ | Redis |
| 🍃 | MongoDB |
| 🐬 | MySQL / MariaDB |
| 🌐 | Nginx / Caddy / Web server |
| 🔧 | gRPC / API service |
| 🔐 | Auth service (OpenFGA, Keycloak) |
| ☁️ | Cloud / LocalStack / MinIO |
| 📬 | Message queue (RabbitMQ, Kafka) |
| 🦭 | Podman container |
| 🐳 | Docker container |

---

## Architecture

```
monobox/
├── cmd/monobox/          # Entry point
├── internal/
│   ├── domain/           # Core entities: Container, ContainerProvider interface
│   ├── adapters/
│   │   ├── engine/       # Docker & Podman providers (exec-based, no shell injection)
│   │   └── tui/          # Bubble Tea UI: model, update, view, keys, commands
│   └── pkg/
│       ├── ui/           # Shared Lip Gloss styles and theme system
│       └── logging/      # File-based slog logger (0600 permissions)
└── .goreleaser.yaml      # Multi-platform release config
```

**Security note:** Container commands use `exec.Command` with discrete arguments — zero shell injection vectors. Log files are created with `0600` permissions. The Docker Desktop log-clear helper runs in a read-only, capability-dropped container with `--network none`.

---

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Commit your changes following [Conventional Commits](https://www.conventionalcommits.org/)
4. Push and open a Pull Request

---

## Requirements

- Docker **or** Podman with daemon/service running

---

## License

[MIT](LICENSE) © João Oliveira
