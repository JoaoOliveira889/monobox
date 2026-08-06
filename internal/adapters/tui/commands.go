package tui

import (
	"bufio"
	"context"
	"io"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monobox/internal/domain"
	"github.com/JoaoOliveira889/monobox/internal/pkg/logging"
)

// loadContainersCmd fetches all containers from the engine.
func (m Model) loadContainersCmd() tea.Cmd {
	p := m.provider
	eng := m.engine
	return func() tea.Msg {
		list, err := p.List()
		if err != nil {
			logging.Error("list containers", "engine", eng, "err", err)
			return containersLoadedMsg{err: err}
		}
		items := make([]containerItem, len(list))
		for i, c := range list {
			items[i] = containerItem{Container: c}
		}
		return containersLoadedMsg{containers: items}
	}
}

// containerActionCmd runs a container lifecycle action.
func (m Model) containerActionCmd(id, action string) tea.Cmd {
	p := m.provider
	return func() tea.Msg {
		var err error
		switch action {
		case "start":
			err = p.Start(id)
		case "stop":
			err = p.Stop(id)
		case "restart":
			err = p.Restart(id)
		}
		return containerActionDoneMsg{id: id, action: action, err: err}
	}
}

// clearContainerLogsCmd truncates the container log file on the engine daemon.
func clearContainerLogsCmd(p domain.ContainerProvider, containerID string) tea.Cmd {
	return func() tea.Msg {
		err := p.ClearLogs(containerID)
		return containerLogsClearedMsg{containerID: containerID, err: err}
	}
}

// openLogStreamTailCmd opens an engine log stream. It never pre-reads the
// stream: a quiet container would otherwise leave the tab waiting forever.
func openLogStreamTailCmd(p domain.ContainerProvider, containerID string, tail int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		rc, err := p.Logs(ctx, containerID, tail, true)
		if err != nil {
			cancel()
			logging.Error("open log stream", "container", containerID, "err", err)
			return logStreamDoneMsg{containerID: containerID}
		}

		return logStreamOpenedMsg{
			containerID: containerID,
			reader:      rc,
			scanner:     newLogScanner(rc),
			ctx:         ctx,
			cancel:      cancel,
		}
	}
}

// nextLogLineCmd reads the next line from the scanner and emits it.
// Uses a large buffer to handle long log lines (e.g. JSON logs).
func nextLogLineCmd(ctx context.Context, containerID string, sc *bufio.Scanner) tea.Cmd {
	return func() tea.Msg {
		select {
		case <-ctx.Done():
			return logStreamDoneMsg{containerID: containerID}
		default:
		}
		if sc.Scan() {
			return logLineMsg{containerID: containerID, line: sc.Text()}
		}
		if err := sc.Err(); err != nil {
			logging.Error("scanner error", "container", containerID, "err", err)
		}
		return logStreamDoneMsg{containerID: containerID}
	}
}

// newLogScanner creates a bufio.Scanner with a 1MB token buffer for large log lines.
func newLogScanner(r io.Reader) *bufio.Scanner {
	sc := bufio.NewScanner(r)
	buf := make([]byte, 1024*1024) // 1 MB buffer
	sc.Buffer(buf, 1024*1024)
	return sc
}

// spinnerTickCmd schedules the next spinner frame.
func spinnerTickCmd() tea.Cmd {
	return tea.Tick(spinnerTickInterval, func(time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

// splashTickCmd drives the splash screen animation.
func splashTickCmd() tea.Cmd {
	return tea.Tick(90*time.Millisecond, func(time.Time) tea.Msg {
		return splashTickMsg{}
	})
}

// tickCmd schedules periodic container list refresh.
func tickCmd() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// clearStatusCmd clears the status bar after statusClearDuration.
func clearStatusCmd(id int) tea.Cmd {
	return tea.Tick(statusClearDuration, func(time.Time) tea.Msg {
		return clearStatusMsg{id: id}
	})
}
