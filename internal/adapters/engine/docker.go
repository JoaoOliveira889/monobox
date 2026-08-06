package engine

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/JoaoOliveira889/monobox/internal/domain"
)

// dockerContainer is the JSON shape returned by "docker ps --format json".
type dockerContainer struct {
	ID         string `json:"ID"`
	Names      string `json:"Names"`
	Image      string `json:"Image"`
	State      string `json:"State"`
	Status     string `json:"Status"`
	RunningFor string `json:"RunningFor"`
	Ports      string `json:"Ports"`
}

// DockerProvider implements domain.ContainerProvider for the Docker engine.
type DockerProvider struct{}

// NewDockerProvider returns a DockerProvider.
func NewDockerProvider() *DockerProvider { return &DockerProvider{} }

func (d *DockerProvider) EngineName() domain.Engine { return domain.EngineDocker }

// List returns all containers (running + stopped) via "docker ps -a" and fetches stats.
func (d *DockerProvider) List() ([]domain.Container, error) {
	out, err := exec.Command(
		"docker", "ps", "-a",
		"--no-trunc",
		"--format", "{{json .}}",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("docker ps: %w", err)
	}
	containers, err := parseDockerJSON(out, domain.EngineDocker)
	if err != nil {
		return nil, err
	}

	// Fetch stats for active containers
	if statsMap, err := d.Stats(); err == nil {
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

type dockerStatsJSON struct {
	ID       string `json:"ID"`
	Name     string `json:"Name"`
	CPUPerc  string `json:"CPUPerc"`
	MemUsage string `json:"MemUsage"`
	MemPerc  string `json:"MemPerc"`
}

// Stats queries "docker stats --no-stream" and maps metrics by container ID and Name.
func (d *DockerProvider) Stats() (map[string]domain.ContainerStats, error) {
	return engineStats("docker")
}

func engineStats(binary string) (map[string]domain.ContainerStats, error) {
	out, err := exec.Command(binary, "stats", "--no-stream", "--format", "{{json .}}").Output()
	if err != nil {
		return nil, err
	}
	res := make(map[string]domain.ContainerStats)
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var s dockerStatsJSON
		if err := json.Unmarshal([]byte(line), &s); err == nil {
			stats := domain.ContainerStats{
				CPU:     s.CPUPerc,
				Mem:     s.MemUsage,
				MemPerc: s.MemPerc,
			}
			if s.ID != "" {
				res[s.ID] = stats
			}
			if s.Name != "" {
				res[s.Name] = stats
			}
		}
	}
	return res, nil
}

// ClearLogs truncates a local Docker log file when its driver exposes one.
// Remote engines and non-file log drivers fail safely instead of pretending to
// clear output that docker logs would still show.
func (d *DockerProvider) ClearLogs(id string) error {
	out, err := exec.Command("docker", "inspect", "--format", "{{.LogPath}}", id).Output()
	if err != nil {
		return fmt.Errorf("inspect log path: %w", err)
	}
	logPath := strings.TrimSpace(string(out))
	if logPath == "" {
		return fmt.Errorf("no log path found for container %s", id)
	}
	if err := clearLogFile(logPath); err == nil {
		return nil
	} else if !isDockerJSONLogPath(logPath) {
		return err
	}

	// Docker Desktop keeps /var/lib/docker inside its Linux VM. A constrained
	// helper mounts only that directory, has no network or capabilities, and
	// executes truncate directly (no shell interpolation).
	return clearDockerDesktopLog(logPath)
}

func clearDockerDesktopLog(logPath string) error {
	cmd := dockerLogClearHelperCmd(logPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("clear Docker Desktop log: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

func dockerLogClearHelperCmd(logPath string) *exec.Cmd {
	return exec.Command(
		"docker", "run", "--rm",
		"--network", "none",
		"--read-only",
		"--cap-drop", "ALL",
		"--security-opt", "no-new-privileges",
		"-v", "/var/lib/docker:/var/lib/docker",
		"alpine:3.22",
		"truncate", "-s", "0", logPath,
	)
}

func isDockerJSONLogPath(logPath string) bool {
	cleanPath := filepath.Clean(logPath)
	return strings.HasPrefix(cleanPath, "/var/lib/docker/containers/") &&
		strings.HasSuffix(cleanPath, "-json.log")
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

// Logs streams container logs.
// tail is the number of historic lines (0 = start now).
// follow controls whether the stream stays open for new output.
func (d *DockerProvider) Logs(ctx context.Context, id string, tail int, follow bool) (io.ReadCloser, error) {
	return dockerLogs(ctx, "docker", id, tail, follow)
}

// dockerLogs runs the engine's logs command. Its stdout is passed through
// untouched, so displayed lines are exactly the same bytes as engine logs.
func dockerLogs(ctx context.Context, binary, id string, tail int, follow bool) (io.ReadCloser, error) {
	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	if tail >= 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", tail))
	} else {
		args = append(args, "--tail", "all")
	}
	args = append(args, id)

	cmd := exec.CommandContext(ctx, binary, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("%s logs %s stdout: %w", binary, id, err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("%s logs %s: %w", binary, id, err)
	}

	return &commandLogReader{ReadCloser: stdout, cmd: cmd}, nil
}

type commandLogReader struct {
	io.ReadCloser
	cmd  *exec.Cmd
	once sync.Once
}

func (r *commandLogReader) Close() error {
	_ = r.ReadCloser.Close()
	r.once.Do(func() {
		go func() { _ = r.cmd.Wait() }()
	})
	return nil
}

func clearLogFile(logPath string) error {
	cleanPath := filepath.Clean(logPath)
	if !filepath.IsAbs(cleanPath) || cleanPath != logPath {
		return fmt.Errorf("invalid engine log path")
	}
	info, err := os.Lstat(cleanPath)
	if err != nil {
		return fmt.Errorf("access log path: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("engine log path is not a regular file")
	}
	if err := os.Truncate(cleanPath, 0); err != nil {
		return fmt.Errorf("truncate engine log: %w", err)
	}
	return nil
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
			continue
		}
		containers = append(containers, domain.Container{
			ID:         dc.ID,
			Name:       strings.TrimPrefix(dc.Names, "/"),
			Image:      dc.Image,
			Status:     parseDockerState(dc.State),
			RunningFor: parseRunningFor(dc.Status),
			Engine:     eng,
			Ports:      cleanPorts(dc.Ports),
		})
	}
	return containers, nil
}

func cleanPorts(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parts := strings.Split(raw, ",")
	var cleaned []string
	seen := make(map[string]bool)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		pClean := strings.TrimPrefix(p, "0.0.0.0:")
		pClean = strings.TrimPrefix(pClean, ":::")
		if !seen[pClean] {
			seen[pClean] = true
			cleaned = append(cleaned, pClean)
		}
	}
	return strings.Join(cleaned, ", ")
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

// parseRunningFor extracts a short human-readable uptime from Docker's Status field.
// "Up 5 minutes" → "5 minutes", "Exited (0) 2 hours ago" → "2 hours ago"
func parseRunningFor(status string) string {
	status = strings.TrimSpace(status)
	lower := strings.ToLower(status)
	if strings.HasPrefix(lower, "up ") {
		after := strings.TrimPrefix(status, "Up ")
		if after == "" {
			after = strings.TrimPrefix(status, "up ")
		}
		return after
	}
	parts := strings.Fields(status)
	if len(parts) >= 3 {
		return strings.Join(parts[len(parts)-3:], " ")
	}
	return status
}
