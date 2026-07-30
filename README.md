# monobox

> Terminal UI for Docker and Podman containers — built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

```
mono box
```

## Features

- **Auto-detect engine** — Docker first, then Podman; clear error if neither available
- **Container list** — name, image, status, uptime, engine
- **Live logs** — last 200 lines + real-time follow with scrolling
- **Lifecycle actions** — start/stop toggle, restart
- Exit logs without interrupting the container

## Keybindings

| Key | Action |
|-----|--------|
| `↑`/`k`, `↓`/`j` | Navigate containers |
| `Enter` | Open logs |
| `s` | Start / stop toggle |
| `r` | Restart |
| `f` | Toggle log follow |
| `Esc` | Back to container list |
| `q` / `Ctrl+C` | Quit |

## Install

```bash
go install github.com/JoaoOliveira889/monobox/cmd/monobox@latest
```

Or via Homebrew (after first release):

```bash
brew install JoaoOliveira889/tap/monobox
```

## Build from source

```bash
git clone https://github.com/JoaoOliveira889/monobox
cd monobox
go build -o monobox ./cmd/monobox
go install ./cmd/monobox
```

## Requirements

- Docker **or** Podman with daemon/service running

## License

MIT
