package tui

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monobox/internal/domain"
	"github.com/JoaoOliveira889/monobox/internal/pkg/logging"
)

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
		case "pause":
			err = p.Pause(id)
		case "unpause":
			err = p.Unpause(id)
		case "remove":
			err = p.Remove(id, true)
		}
		return containerActionDoneMsg{id: id, action: action, err: err}
	}
}

func (m Model) batchActionCmd(projectName, action string, containerIDs []string) tea.Cmd {
	p := m.provider
	return func() tea.Msg {
		var lastErr error
		successCount := 0
		for _, id := range containerIDs {
			var err error
			switch action {
			case "start":
				err = p.Start(id)
			case "stop":
				err = p.Stop(id)
			case "restart":
				err = p.Restart(id)
			case "pause":
				err = p.Pause(id)
			case "unpause":
				err = p.Unpause(id)
			case "remove":
				err = p.Remove(id, true)
			}
			if err != nil {
				lastErr = err
			} else {
				successCount++
			}
		}
		return batchActionDoneMsg{
			projectName: projectName,
			action:      action,
			count:       len(containerIDs),
			success:     successCount,
			err:         lastErr,
		}
	}
}

func inspectCmd(p domain.ContainerProvider, containerID string) tea.Cmd {
	return func() tea.Msg {
		content, err := p.Inspect(containerID)
		return inspectDoneMsg{containerID: containerID, content: content, err: err}
	}
}

func openBrowserCmd(port string) tea.Cmd {
	return func() tea.Msg {
		url := "http://localhost:" + port
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		default:
			cmd = exec.Command("xdg-open", url)
		}
		_ = cmd.Run()
		return nil
	}
}

func clearContainerLogsCmd(p domain.ContainerProvider, containerID string) tea.Cmd {
	return func() tea.Msg {
		err := p.ClearLogs(containerID)
		return containerLogsClearedMsg{containerID: containerID, err: err}
	}
}

func openLogStreamTailCmd(p domain.ContainerProvider, containerID string, tail int, timestamps bool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		rc, err := p.Logs(ctx, containerID, tail, true, timestamps)
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

func newLogScanner(r io.Reader) *bufio.Scanner {
	sc := bufio.NewScanner(r)
	buf := make([]byte, 1024*1024)
	sc.Buffer(buf, 1024*1024)
	return sc
}

func splashTickCmd() tea.Cmd {
	return tea.Tick(90*time.Millisecond, func(time.Time) tea.Msg {
		return splashTickMsg{}
	})
}

func (m Model) tickCmd() tea.Cmd {
	interval := m.refreshInterval
	if interval <= 0 {
		interval = refreshInterval
	}
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func clearStatusCmd(id int) tea.Cmd {
	return tea.Tick(statusClearDuration, func(time.Time) tea.Msg {
		return clearStatusMsg{id: id}
	})
}
