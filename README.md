# monobox

<p align="center">
  <a href="https://github.com/JoaoOliveira889/monobox/releases/latest"><img src="https://img.shields.io/github/v/release/JoaoOliveira889/monobox?color=7aa2f7&label=tag&logo=github&style=flat-square" alt="Latest Tag"></a>
  <a href="https://github.com/JoaoOliveira889/monobox/releases/latest"><img src="https://img.shields.io/github/downloads/JoaoOliveira889/monobox/total?color=9ece6a&label=downloads&logo=github&style=flat-square" alt="Total Downloads"></a>
  <a href="https://goreportcard.com/report/github.com/JoaoOliveira889/monobox"><img src="https://goreportcard.com/badge/github.com/JoaoOliveira889/monobox?style=flat-square" alt="Go Report Card"></a>
  <a href="https://github.com/JoaoOliveira889/homebrew-tap"><img src="https://img.shields.io/badge/homebrew-v0.0.1-b8ff3d?logo=homebrew&style=flat-square" alt="Homebrew Version"></a>
</p>

<p align="center">
  <a href="https://github.com/JoaoOliveira889/monobox"><strong>MonoBox v0.0.1 · JoaoOliveira889/monobox</strong></a>
</p>

**Terminal UI for Docker and Podman containers.** A TUI tool that lists all your containers with live status, log streaming, and one-key lifecycle actions — with confirmation guards for every mutating command.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss), and [Bubbles](https://github.com/charmbracelet/bubbles).

---

## Features

- **Auto-detect engine** — Docker first, then Podman; clear error if neither is available
- **Container list** — icon per service type, name, status badge, and uptime
- **Live log streaming** — real-time `docker logs -f` / `podman logs -f` with follow mode
- **Lifecycle actions** — start/stop toggle and restart with optimistic UI updates
- **Clear logs** — truncate engine log file with confirmation guard
- **Responsive layout** — resizable split panels, adapts to small terminals
- **Tokyo Night theme** — and 5 more built-in themes (Dracula, Nord, Gruvbox, Monokai, One Dark)
- **Security first** — exec with discrete arguments, no shell injection, log file at `0600`

---

## Layout

```
MonoBox                                                    13 containers
● 13 containers  •  10 running  •  3 stopped  •  docker: 1.08% CPU • 299.0 MiB
────────────────────────────────────────────────────────────────────────────────
╭─[1 Containers]──────────╮╭─[2 Container — openfga-migrate]────────────────────╮
│ 🔧 idez-account-grpc-1  ●RUNNING ││  NAME:      openfga-migrate                       │
│ 🔧 idez-holder-grpc-1   ●RUNNING ││  TYPE:      🔐  Auth Service                      │
│ 🐘 idez-postgres        ●RUNNING ││  IMAGE:     openfga/openfga:latest                │
│ 🔧 idez-tenant-grpc-1   ●RUNNING ││  PORTS:     none                                  │
│ 🔧 idez-user-grpc-1     ●RUNNING ││  STATUS:    ○ STOPPED (5 weeks ago)               │
│ 🔧 idez-users-api-1     ●RUNNING ││  ENGINE:    docker                                │
│ ☁️ ministack            ●RUNNING ││  ID:        f699679e5fa6                           │
│ 🔐 openfga              ●RUNNING ││                                                    │
│ 🐘 openfga-postgres     ●RUNNING ││  ACTIONS & LOGS:                                   │
│ ⚡ redis-local           ●RUNNING ││   ▸ Press Enter / l / 2 to open live log stream   │
│ 🔧 idez-device-grpc-1   ○STOPPED ││   ▸ Press s to start container                    │
│ 🔧 idez-shortcode-grpc-1 ○STOPPED ││   ▸ Press r to restart container                 │
│▌🔐 openfga-migrate       ○STOPPED │╰───────────────────────────────────────────────────╯
╰──────────────────────────────────╯
↑↓/jk nav  •  s start/stop  •  <> resize  •  enter logs  •  r restart  •  q quit   monobox 0.0.1
```

---

## Installation

### Option 1 — Homebrew (macOS & Linux)

```bash
brew tap JoaoOliveira889/tap
brew install monobox
```

### Option 2 — Pre-built binary

Download the latest release from the [Releases page](https://github.com/JoaoOliveira889/monobox/releases/latest).

```bash
# macOS (Apple Silicon)
curl -LO https://github.com/JoaoOliveira889/monobox/releases/latest/download/monobox_Darwin_arm64.tar.gz
tar -xzf monobox_Darwin_arm64.tar.gz
sudo mv monobox /usr/local/bin/

# Linux (amd64)
curl -LO https://github.com/JoaoOliveira889/monobox/releases/latest/download/monobox_Linux_x86_64.tar.gz
tar -xzf monobox_Linux_x86_64.tar.gz
sudo mv monobox /usr/local/bin/
```

### Option 3 — `go install`

```bash
go install github.com/JoaoOliveira889/monobox/cmd/monobox@latest
```

> Requires Go 1.24 or later.

### Option 4 — Build from source

```bash
git clone https://github.com/JoaoOliveira889/monobox
cd monobox
go build -o monobox ./cmd/monobox
go install ./cmd/monobox
```

---

## Usage

```bash
monobox
```

No flags required. MonoBox auto-detects Docker or Podman on startup.

---

## Keybindings

### Global

| Key | Action |
|-----|--------|
| `↑` / `k`, `↓` / `j` | Navigate containers |
| `←` / `h`, `→` / `l` | Switch panel focus |
| `1`, `2` | Jump to Containers / Logs panel |
| `tab` | Toggle panel focus |
| `<` / `,`, `>` / `.` | Resize panel divider |
| `q` / `Ctrl+C` | Quit |

### Container List Panel

| Key | Action |
|-----|--------|
| `Enter` | Open live log stream |
| `s` | Start / stop toggle |
| `r` | Restart container |

### Log Panel

| Key | Action |
|-----|--------|
| `↑↓` / `jk` | Scroll logs |
| `PgUp` / `Ctrl+U` | Page up |
| `PgDn` / `Ctrl+D` | Page down |
| `End` / `G` | Jump to bottom |
| `f` | Toggle follow mode |
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
