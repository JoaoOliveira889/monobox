package domain

import (
	"context"
	"io"
	"os/exec"
)

type Engine string

const (
	EngineDocker Engine = "docker"
	EnginePodman Engine = "podman"
)

type ContainerStatus string

const (
	StatusRunning ContainerStatus = "running"
	StatusExited  ContainerStatus = "exited"
	StatusPaused  ContainerStatus = "paused"
	StatusCreated ContainerStatus = "created"
	StatusUnknown ContainerStatus = "unknown"
)

type HealthStatus string

const (
	HealthNone      HealthStatus = ""
	HealthHealthy   HealthStatus = "healthy"
	HealthUnhealthy HealthStatus = "unhealthy"
	HealthStarting  HealthStatus = "starting"
)

type ContainerStats struct {
	CPU     string
	Mem     string
	MemPerc string
}

type MountDetail struct {
	Type        string `json:"Type"`
	Source      string `json:"Source"`
	Destination string `json:"Destination"`
	Mode        string `json:"Mode"`
}

type NetworkDetail struct {
	Name      string `json:"Name"`
	IPAddress string `json:"IPAddress"`
	Gateway   string `json:"Gateway"`
}

type HealthLogEntry struct {
	Start    string `json:"Start"`
	End      string `json:"End"`
	ExitCode int    `json:"ExitCode"`
	Output   string `json:"Output"`
}

type HealthDetail struct {
	Status        HealthStatus     `json:"Status"`
	FailingStreak int              `json:"FailingStreak"`
	Log           []HealthLogEntry `json:"Log"`
}

type ContainerInspectDetails struct {
	Env      []string        `json:"Env"`
	Mounts   []MountDetail   `json:"Mounts"`
	Networks []NetworkDetail `json:"Networks"`
	Health   *HealthDetail   `json:"Health"`
}

type Container struct {
	ID             string
	Name           string
	Image          string
	Status         ContainerStatus
	Health         HealthStatus
	RunningFor     string
	Engine         Engine
	CPU            string
	Mem            string
	Ports          string
	Labels         map[string]string
	ComposeProject string
}

func (c Container) IsRunning() bool {
	return c.Status == StatusRunning
}

type ContainerLifecycleProvider interface {
	Start(id string) error
	Stop(id string) error
	Restart(id string) error
	Pause(id string) error
	Unpause(id string) error
	Remove(id string, force bool) error
}

type LogProvider interface {
	ClearLogs(id string) error
	Logs(ctx context.Context, id string, tail int, follow bool, timestamps bool) (io.ReadCloser, error)
}

type InspectProvider interface {
	Inspect(id string) (string, error)
	ExecCmd(id string) *exec.Cmd
}

type SystemProvider interface {
	SystemPrune(all bool) (string, error)
}

type ComposeProvider interface {
	ComposeUp(project string) error
	ComposeDown(project string) error
}

type ContainerProvider interface {
	EngineName() Engine
	List() ([]Container, error)
	Stats() (map[string]ContainerStats, error)
	ContainerLifecycleProvider
	LogProvider
	InspectProvider
	SystemProvider
	ComposeProvider
}
