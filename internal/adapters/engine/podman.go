package engine

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/JoaoOliveira889/monobox/internal/domain"
)

// PodmanProvider implements domain.ContainerProvider for the Podman engine.
// Podman is CLI-compatible with Docker, so we reuse the same JSON parsing and
// log-streaming logic — only the binary name differs.
type PodmanProvider struct{}

// NewPodmanProvider returns a PodmanProvider.
func NewPodmanProvider() *PodmanProvider { return &PodmanProvider{} }

func (p *PodmanProvider) EngineName() domain.Engine { return domain.EnginePodman }

// List returns all containers (running + stopped) via "podman ps -a".
func (p *PodmanProvider) List() ([]domain.Container, error) {
	out, err := exec.Command(
		"podman", "ps", "-a",
		"--no-trunc",
		"--format", "{{json .}}",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("podman ps: %w", err)
	}
	containers, err := parseDockerJSON(out, domain.EnginePodman)
	if err != nil {
		return nil, err
	}

	if statsMap, err := p.Stats(); err == nil {
		for i, c := range containers {
			if st, ok := statsMap[c.ID]; ok {
				containers[i].CPU = st.CPU
				if st.MemPerc != "" {
					containers[i].Mem = fmt.Sprintf("%s (%s)", st.Mem, st.MemPerc)
				} else {
					containers[i].Mem = st.Mem
				}
			} else if st, ok := statsMap[c.Name]; ok {
				containers[i].CPU = st.CPU
				if st.MemPerc != "" {
					containers[i].Mem = fmt.Sprintf("%s (%s)", st.Mem, st.MemPerc)
				} else {
					containers[i].Mem = st.Mem
				}
			}
		}
	}
	return containers, nil
}

// Stats returns metrics for active Podman containers.
func (p *PodmanProvider) Stats() (map[string]domain.ContainerStats, error) {
	return engineStats("podman")
}

// ClearLogs truncates a local Podman log file when its configured log driver
// exposes a writable path.
func (p *PodmanProvider) ClearLogs(id string) error {
	out, err := exec.Command("podman", "inspect", "--format", "{{.HostConfig.LogConfig.Path}}", id).Output()
	if err != nil {
		return fmt.Errorf("inspect log path: %w", err)
	}
	logPath := strings.TrimSpace(string(out))
	if logPath == "" || logPath == "<no value>" {
		out, err = exec.Command("podman", "inspect", "--format", "{{.LogPath}}", id).Output()
		if err != nil {
			return fmt.Errorf("inspect log path: %w", err)
		}
		logPath = strings.TrimSpace(string(out))
	}
	if logPath == "" || logPath == "<no value>" {
		return fmt.Errorf("no local log path found for container %s", id)
	}
	return clearLogFile(logPath)
}

// Start starts the container with the given ID.
func (p *PodmanProvider) Start(id string) error {
	return runCmd("podman", "start", id)
}

// Stop stops the container with the given ID.
func (p *PodmanProvider) Stop(id string) error {
	return runCmd("podman", "stop", id)
}

// Restart restarts the container with the given ID.
func (p *PodmanProvider) Restart(id string) error {
	return runCmd("podman", "restart", id)
}

// Logs streams container logs for Podman.
func (p *PodmanProvider) Logs(ctx context.Context, id string, tail int, follow bool) (io.ReadCloser, error) {
	return dockerLogs(ctx, "podman", id, tail, follow)
}

// runCmd runs a binary subcommand against a container ID.
func runCmd(binary, subcmd, id string) error {
	out, err := exec.Command(binary, subcmd, id).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s %s: %s", binary, subcmd, id, strings.TrimSpace(string(out)))
	}
	return nil
}
