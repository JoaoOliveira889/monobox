package domain

import (
	"context"
	"io"
)

// Engine identifies which container runtime is in use.
type Engine string

const (
	EngineDocker Engine = "docker"
	EnginePodman Engine = "podman"
)

// ContainerStatus mirrors the high-level state reported by the engine.
type ContainerStatus string

const (
	StatusRunning ContainerStatus = "running"
	StatusExited  ContainerStatus = "exited"
	StatusPaused  ContainerStatus = "paused"
	StatusCreated ContainerStatus = "created"
	StatusUnknown ContainerStatus = "unknown"
)

// Container holds the information displayed in the list panel.
type Container struct {
	ID         string
	Name       string
	Image      string
	Status     ContainerStatus
	RunningFor string // human-readable uptime, e.g. "5 minutes ago"
	Engine     Engine
}

// IsRunning reports whether the container is currently running.
func (c Container) IsRunning() bool {
	return c.Status == StatusRunning
}

// ContainerProvider is the port (interface) that engine adapters must satisfy.
type ContainerProvider interface {
	// EngineName returns the engine identifier ("docker" or "podman").
	EngineName() Engine

	// List returns all containers (running and stopped).
	List() ([]Container, error)

	// Start starts the container with the given ID.
	Start(id string) error

	// Stop stops the container with the given ID.
	Stop(id string) error

	// Restart restarts the container with the given ID.
	Restart(id string) error

	// Logs returns a stream of log lines for the given container.
	// tail is the number of historic lines to include (0 = all).
	// follow controls whether the stream stays open for new output.
	Logs(ctx context.Context, id string, tail int, follow bool) (io.ReadCloser, error)
}
