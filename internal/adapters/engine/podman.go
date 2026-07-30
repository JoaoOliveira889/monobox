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
	return parseDockerJSON(out, domain.EnginePodman)
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
