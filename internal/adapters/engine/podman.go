package engine

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/JoaoOliveira889/monobox/internal/domain"
)

type PodmanProvider struct {
	CLIProvider
}

func NewPodmanProvider() *PodmanProvider {
	return &PodmanProvider{
		CLIProvider: CLIProvider{
			binary: "podman",
			engine: domain.EnginePodman,
		},
	}
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



func runCmd(binary string, args ...string) error {
	out, err := exec.Command(binary, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %s", binary, strings.Join(args, " "), strings.TrimSpace(string(out)))
	}
	return nil
}
