package tui

import (
	"bufio"
	"context"
	"io"
	"time"

	"github.com/JoaoOliveira889/monobox/internal/domain"
)

type containersLoadedMsg struct {
	containers []containerItem
	err        error
}

type statsLoadedMsg struct {
	stats map[string]domain.ContainerStats
	err   error
}

type containerActionDoneMsg struct {
	id     string
	action string
	err    error
}

type batchActionDoneMsg struct {
	projectName string
	action      string
	count       int
	success     int
	err         error
}

type containerLogsClearedMsg struct {
	containerID string
	err         error
}

type logStreamOpenedMsg struct {
	containerID string
	reader      io.Closer
	scanner     *bufio.Scanner
	ctx         context.Context
	cancel      context.CancelFunc
}

type logLineMsg struct {
	containerID string
	line        string
}

type logStreamDoneMsg struct {
	containerID string
}

type splashTickMsg struct{}

type execDoneMsg struct {
	err error
}

type inspectDoneMsg struct {
	containerID string
	content     string
	err         error
}

type clearStatusMsg struct{ id int }

type tickMsg time.Time

type pruneDoneMsg struct {
	output string
	err    error
}

type composeActionDoneMsg struct {
	project string
	action  string
	err     error
}
