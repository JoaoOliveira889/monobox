package engine

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/JoaoOliveira889/monobox/internal/domain"
)

// DetectEngine probes for Docker first, then Podman.
// Returns the provider, the engine name string, and any error.
// Returns a descriptive error when neither is available.
func DetectEngine() (domain.ContainerProvider, string, error) {
	if p, err := probeDocker(); err == nil {
		return p, string(domain.EngineDocker), nil
	}
	if p, err := probePodman(); err == nil {
		return p, string(domain.EnginePodman), nil
	}
	return nil, "", errors.New(
		"monobox: no container engine found — install Docker or Podman and make sure the daemon is running",
	)
}

// probeDocker checks whether docker is in PATH and the daemon is reachable.
func probeDocker() (domain.ContainerProvider, error) {
	path, err := exec.LookPath("docker")
	if err != nil {
		return nil, fmt.Errorf("docker not in PATH: %w", err)
	}
	if out, err := exec.Command(path, "info", "--format", "{{.ServerVersion}}").CombinedOutput(); err != nil {
		return nil, fmt.Errorf("docker daemon not reachable: %s", out)
	}
	return NewDockerProvider(), nil
}

// probePodman checks whether podman is in PATH and the service is reachable.
func probePodman() (domain.ContainerProvider, error) {
	path, err := exec.LookPath("podman")
	if err != nil {
		return nil, fmt.Errorf("podman not in PATH: %w", err)
	}
	if out, err := exec.Command(path, "info", "--format", "{{.Version.Version}}").CombinedOutput(); err != nil {
		return nil, fmt.Errorf("podman not reachable: %s", out)
	}
	return NewPodmanProvider(), nil
}
