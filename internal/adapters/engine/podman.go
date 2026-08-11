package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/JoaoOliveira889/monobox/internal/domain"
)

type PodmanProvider struct{}

func NewPodmanProvider() *PodmanProvider { return &PodmanProvider{} }

func (p *PodmanProvider) EngineName() domain.Engine { return domain.EnginePodman }

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
		applyStats(containers, statsMap)
	}
	return containers, nil
}

func (p *PodmanProvider) Stats() (map[string]domain.ContainerStats, error) {
	return engineStats("podman")
}

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

func (p *PodmanProvider) Start(id string) error {
	return runCmd("podman", "start", id)
}

func (p *PodmanProvider) Stop(id string) error {
	return runCmd("podman", "stop", id)
}

func (p *PodmanProvider) Restart(id string) error {
	return runCmd("podman", "restart", id)
}

func (p *PodmanProvider) Pause(id string) error {
	return runCmd("podman", "pause", id)
}

func (p *PodmanProvider) Unpause(id string) error {
	return runCmd("podman", "unpause", id)
}

func (p *PodmanProvider) Remove(id string, force bool) error {
	if force {
		return runCmd("podman", "rm", "-f", id)
	}
	return runCmd("podman", "rm", id)
}

func (p *PodmanProvider) Inspect(id string) (string, error) {
	out, err := exec.Command("podman", "inspect", id).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("podman inspect %s: %s", id, strings.TrimSpace(string(out)))
	}
	var pretty bytes.Buffer
	if jsonErr := json.Indent(&pretty, out, "", "  "); jsonErr == nil {
		return pretty.String(), nil
	}
	return string(out), nil
}

func (p *PodmanProvider) ExecCmd(id string) *exec.Cmd {
	return exec.Command("podman", "exec", "-it", id, "/bin/sh")
}

func (p *PodmanProvider) Logs(ctx context.Context, id string, tail int, follow bool, timestamps bool) (io.ReadCloser, error) {
	return dockerLogs(ctx, "podman", id, tail, follow, timestamps)
}

func runCmd(binary string, args ...string) error {
	out, err := exec.Command(binary, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %s", binary, strings.Join(args, " "), strings.TrimSpace(string(out)))
	}
	return nil
}
