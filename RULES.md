# Monobox Rules

## Project-specific
- All container engine operations must go through `internal/adapters/engine/` using `exec.Command` with discrete arguments.
- The TUI model lives in `internal/adapters/tui/`.
- Shared UI styles live in `internal/pkg/ui/`.
- Build: `go build -buildvcs=false -o monobox ./cmd/monobox`. No Makefile — use `go build` directly.
- Install: `go install -buildvcs=false ./cmd/monobox` to make binary available in PATH.
- Release: `.goreleaser.yaml` handles multi-platform builds and Homebrew tap.
