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

// openLogStreamCmd opens a log stream for a container.
// Returns logStreamOpenedMsg on success, logStreamDoneMsg on failure.
func openLogStreamCmd(p domain.ContainerProvider, containerID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		rc, err := p.Logs(ctx, containerID, logTailLines, true)
		if err != nil {
			cancel()
			logging.Error("open log stream", "container", containerID, "err", err)
			return logStreamDoneMsg{containerID: containerID}
		}
		return logStreamOpenedMsg{
			containerID: containerID,
			reader:      rc,
			ctx:         ctx,
			cancel:      cancel,
		}
	}
}

// nextLogLineCmd reads one line from scanner and emits it as logLineMsg.
// On EOF or context cancellation, emits logStreamDoneMsg.
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
		return logStreamDoneMsg{containerID: containerID}
	}
}

// spinnerTickCmd schedules the next spinner frame.
func spinnerTickCmd() tea.Cmd {
	return tea.Tick(spinnerTickInterval, func(time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

// tickCmd schedules the next periodic container list refresh.
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

// logStreamOpenedMsg carries the opened stream back to the model.
type logStreamOpenedMsg struct {
	containerID string
	reader      io.ReadCloser
	ctx         context.Context
	cancel      context.CancelFunc
}
