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

type CLIProvider struct {
	binary string
	engine domain.Engine
}

func (c *CLIProvider) EngineName() domain.Engine {
	return c.engine
}

func (c *CLIProvider) List() ([]domain.Container, error) {
	out, err := exec.Command(
		c.binary, "ps", "-a",
		"--no-trunc",
		"--format", "{{json .}}",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("%s ps: %w", c.binary, err)
	}
	return parseDockerJSON(out, c.engine)
}

func (c *CLIProvider) Stats() (map[string]domain.ContainerStats, error) {
	return engineStats(c.binary)
}

func (c *CLIProvider) Start(id string) error {
	return runCmd(c.binary, "start", id)
}

func (c *CLIProvider) Stop(id string) error {
	return runCmd(c.binary, "stop", id)
}

func (c *CLIProvider) Restart(id string) error {
	return runCmd(c.binary, "restart", id)
}

func (c *CLIProvider) Pause(id string) error {
	return runCmd(c.binary, "pause", id)
}

func (c *CLIProvider) Unpause(id string) error {
	return runCmd(c.binary, "unpause", id)
}

func (c *CLIProvider) Remove(id string, force bool) error {
	if force {
		return runCmd(c.binary, "rm", "-f", id)
	}
	return runCmd(c.binary, "rm", id)
}

func (c *CLIProvider) Inspect(id string) (string, error) {
	out, err := exec.Command(c.binary, "inspect", id).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s inspect %s: %s", c.binary, id, strings.TrimSpace(string(out)))
	}
	var pretty bytes.Buffer
	if jsonErr := json.Indent(&pretty, out, "", "  "); jsonErr == nil {
		return pretty.String(), nil
	}
	return string(out), nil
}

func (c *CLIProvider) ExecCmd(id string) *exec.Cmd {
	return exec.Command(c.binary, "exec", "-it", id, "sh", "-c", "[ -x /bin/bash ] && exec /bin/bash || ([ -x /bin/sh ] && exec /bin/sh || exec /bin/ash)")
}

func (c *CLIProvider) Logs(ctx context.Context, id string, tail int, follow bool, timestamps bool) (io.ReadCloser, error) {
	return dockerLogs(ctx, c.binary, id, tail, follow, timestamps)
}

func (c *CLIProvider) SystemPrune(all bool) (string, error) {
	args := []string{"system", "prune", "-f"}
	if all {
		args = append(args, "--all")
	}
	out, err := exec.Command(c.binary, args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s system prune: %s (%w)", c.binary, strings.TrimSpace(string(out)), err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (c *CLIProvider) ComposeUp(project string) error {
	return runCmd(c.binary, "compose", "-p", project, "up", "-d")
}

func (c *CLIProvider) ComposeDown(project string) error {
	return runCmd(c.binary, "compose", "-p", project, "down")
}

