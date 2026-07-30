package engine

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/JoaoOliveira889/monobox/internal/domain"
)

// dockerContainer is the JSON shape returned by "docker ps --format json".
type dockerContainer struct {
	ID      string `json:"ID"`
	Names   string `json:"Names"`
	Image   string `json:"Image"`
	State   string `json:"State"`
	Status  string `json:"Status"`
	RunningFor string `json:"RunningFor"`
}

// DockerProvider implements domain.ContainerProvider for the Docker engine.
type DockerProvider struct{}

// NewDockerProvider returns a DockerProvider.
func NewDockerProvider() *DockerProvider { return &DockerProvider{} }

func (d *DockerProvider) EngineName() domain.Engine { return domain.EngineDocker }

// List returns all containers (running + stopped) via "docker ps -a".
func (d *DockerProvider) List() ([]domain.Container, error) {
	out, err := exec.Command(
		"docker", "ps", "-a",
		"--no-trunc",
		"--format", "{{json .}}",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("docker ps: %w", err)
	}
	return parseDockerJSON(out, domain.EngineDocker)
}

// Start starts the container with the given ID.
func (d *DockerProvider) Start(id string) error {
	return runCmd("docker", "start", id)
}

// Stop stops the container with the given ID.
func (d *DockerProvider) Stop(id string) error {
	return runCmd("docker", "stop", id)
}

// Restart restarts the container with the given ID.
func (d *DockerProvider) Restart(id string) error {
	return runCmd("docker", "restart", id)
}

// Logs streams container logs. tail=0 means "all". follow controls -f flag.
func (d *DockerProvider) Logs(ctx context.Context, id string, tail int, follow bool) (io.ReadCloser, error) {
	return dockerLogs(ctx, "docker", id, tail, follow)
}

// dockerLogs builds the exec.Cmd for log streaming and returns its stdout pipe.
func dockerLogs(ctx context.Context, binary, id string, tail int, follow bool) (io.ReadCloser, error) {
	args := []string{"logs", "--timestamps"}
	if follow {
		args = append(args, "-f")
	}
	if tail > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", tail))
	} else {
		args = append(args, "--tail", "all")
	}
	args = append(args, id)

	cmd := exec.CommandContext(ctx, binary, args...)
	// Docker writes logs to stderr as well; combine both.
	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw
	if err := cmd.Start(); err != nil {
		pr.Close()
		pw.Close()
		return nil, fmt.Errorf("%s logs: %w", binary, err)
	}
	go func() {
		cmd.Wait()
		pw.Close()
	}()
	return pr, nil
}

// parseDockerJSON decodes newline-delimited JSON from "docker ps --format {{json .}}".
func parseDockerJSON(raw []byte, eng domain.Engine) ([]domain.Container, error) {
	var containers []domain.Container
	scanner := bufio.NewScanner(strings.NewReader(string(raw)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var dc dockerContainer
		if err := json.Unmarshal([]byte(line), &dc); err != nil {
			continue // skip malformed lines
		}
		containers = append(containers, domain.Container{
			ID:         dc.ID,
			Name:       strings.TrimPrefix(dc.Names, "/"),
			Image:      dc.Image,
			Status:     parseDockerState(dc.State),
			RunningFor: parseRunningFor(dc.Status),
			Engine:     eng,
		})
	}
	return containers, nil
}

func parseDockerState(state string) domain.ContainerStatus {
	switch strings.ToLower(state) {
	case "running":
		return domain.StatusRunning
	case "exited":
		return domain.StatusExited
	case "paused":
		return domain.StatusPaused
	case "created":
		return domain.StatusCreated
	default:
		return domain.StatusUnknown
	}
}

// parseRunningFor extracts a short human-readable uptime from Docker's Status field
// e.g. "Up 5 minutes" -> "5 minutes" or "Exited (0) 2 hours ago" -> "2 hours ago".
func parseRunningFor(status string) string {
	status = strings.TrimSpace(status)
	if strings.HasPrefix(strings.ToLower(status), "up ") {
		return strings.TrimPrefix(status, "Up ")
	}
	parts := strings.Split(status, " ")
	if len(parts) >= 3 {
		// "Exited (0) 2 hours ago" → last 3 words
		return strings.Join(parts[len(parts)-3:], " ")
	}
	return status
}
