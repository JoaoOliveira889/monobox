package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monobox/internal/domain"
)

// Update is the main bubbletea message dispatcher.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	oldStatus := m.statusMsg
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cmd = m.handleResize(msg)
	case spinnerTickMsg:
		m.spinnerFrame++
		cmd = spinnerTickCmd()
	case splashTickMsg:
		if m.showSplash {
			m.splashFrame++
			cmd = splashTickCmd()
		}
	case tickMsg:
		cmd = m.loadContainersCmd()
	case containersLoadedMsg:
		cmd = m.handleContainersLoaded(msg)
	case containerActionDoneMsg:
		cmd = m.handleActionDone(msg)
	case containerLogsClearedMsg:
		cmd = m.handleContainerLogsCleared(msg)
	case logStreamOpenedMsg:
		cmd = m.handleLogStreamOpened(msg)
	case logLineMsg:
		cmd = m.handleLogLine(msg)
	case logStreamDoneMsg:
		m.handleLogStreamDone(msg)
	case clearStatusMsg:
		if m.statusMsgID == msg.id {
			m.statusMsg = ""
		}
	case tea.KeyMsg:
		cmd = m.handleKeys(msg)
	}

	// Schedule auto-clear for new status messages.
	if m.statusMsg != "" && m.statusMsg != oldStatus {
		m.statusMsgID++
		cmd = tea.Batch(cmd, clearStatusCmd(m.statusMsgID))
	}

	return m, cmd
}

func (m *Model) handleResize(msg tea.WindowSizeMsg) tea.Cmd {
	m.width = msg.Width
	m.height = msg.Height

	lpW := m.leftPanelWidth()
	rpW := m.rightPanelWidth()
	headerHeight := 3
	footerHeight := 1
	bodyHeight := m.height - headerHeight - footerHeight
	if bodyHeight < 5 {
		bodyHeight = 5
	}

	listInnerW := lpW - 3
	if listInnerW < 0 {
		listInnerW = 0
	}
	logsInnerW := rpW - 4
	if logsInnerW < 0 {
		logsInnerW = 0
	}

	innerHeight := bodyHeight - 2
	if innerHeight < 0 {
		innerHeight = 0
	}

	if m.listViewport.Width == 0 {
		m.listViewport = viewport.New(listInnerW, innerHeight)
	} else {
		m.listViewport.Width = listInnerW
		m.listViewport.Height = innerHeight
	}
	if m.logViewport.Width == 0 {
		m.logViewport = viewport.New(logsInnerW, innerHeight)
	} else {
		m.logViewport.Width = logsInnerW
		m.logViewport.Height = innerHeight
	}

	m.refreshListViewport()
	m.refreshLogViewportContent()
	return nil
}

func (m *Model) handleContainersLoaded(msg containersLoadedMsg) tea.Cmd {
	m.loading = false
	// Hide splash once containers arrive.
	m.showSplash = false

	if msg.err != nil {
		m.setStatus(fmt.Sprintf("✗ Error loading containers: %s", msg.err))
		return tickCmd()
	}

	var prevID string
	if c := m.selectedContainer(); c != nil {
		prevID = c.ID
	}

	m.containers = msg.containers
	sortContainerItems(m.containers)

	if prevID != "" {
		for i, c := range m.containers {
			if c.ID == prevID {
				m.cursor = i
				break
			}
		}
	}
	if m.cursor >= len(m.containers) {
		m.cursor = max(0, len(m.containers)-1)
	}

	m.refreshListViewport()
	if m.stream == nil {
		if c := m.selectedContainer(); c != nil {
			return tea.Batch(tickCmd(), m.startLogStream(c.ID))
		}
	}
	return tickCmd()
}

func (m *Model) handleActionDone(msg containerActionDoneMsg) tea.Cmd {
	for i, c := range m.containers {
		if c.ID == msg.id {
			m.containers[i].loading = false
			// Optimistic local update: flip status immediately so the UI
			// reflects the new state before loadContainersCmd returns.
			if msg.err == nil {
				switch msg.action {
				case "stop":
					m.containers[i].Status = domain.StatusExited
					m.containers[i].RunningFor = ""
				case "start", "restart":
					m.containers[i].Status = domain.StatusRunning
				}
			}
			break
		}
	}
	m.refreshListViewport()
	if msg.err != nil {
		m.setStatus(fmt.Sprintf("✗ %s failed: %s", msg.action, msg.err))
	} else {
		shortID := msg.id
		if len(shortID) > 12 {
			shortID = shortID[:12]
		}
		m.setStatus(fmt.Sprintf("✓ %s %s", msg.action, shortID))
	}
	// Immediate refresh to get accurate engine state, plus a short-delayed
	// second pass for engines that update status asynchronously.
	return tea.Batch(
		m.loadContainersCmd(),
		tea.Tick(800*time.Millisecond, func(time.Time) tea.Msg {
			return tickMsg(time.Now())
		}),
	)
}

func (m *Model) handleContainerLogsCleared(msg containerLogsClearedMsg) tea.Cmd {
	if msg.err != nil {
		m.setStatus(fmt.Sprintf("✗ Failed to clear logs: %s", msg.err))
	} else {
		m.logLines = nil
		m.refreshLogViewportContent()
		m.setStatus(fmt.Sprintf("✓ %s logs cleared", m.engine))
	}
	c := m.selectedContainer()
	if c != nil && c.ID == msg.containerID {
		return m.startLogStream(c.ID)
	}
	return nil
}

func (m *Model) handleLogStreamOpened(msg logStreamOpenedMsg) tea.Cmd {
	if msg.containerID != m.logContainerID {
		msg.cancel()
		_ = msg.reader.Close()
		return nil
	}
	m.cancelStream()
	m.stream = &logStream{
		containerID: msg.containerID,
		reader:      msg.reader,
		scanner:     msg.scanner,
		ctx:         msg.ctx,
		cancel:      msg.cancel,
	}
	m.refreshLogViewportContent()
	if m.logFollow {
		m.logViewport.GotoBottom()
	}
	return nextLogLineCmd(m.stream.ctx, msg.containerID, m.stream.scanner)
}

func (m *Model) handleLogLine(msg logLineMsg) tea.Cmd {
	if m.stream == nil || m.stream.containerID != msg.containerID {
		return nil
	}
	m.appendLogLine(msg.line)
	m.refreshLogViewportContent()
	return nextLogLineCmd(m.stream.ctx, msg.containerID, m.stream.scanner)
}

func (m *Model) handleLogStreamDone(msg logStreamDoneMsg) {
	if m.stream != nil && m.stream.containerID == msg.containerID {
		_ = m.stream.reader.Close()
		m.stream = nil
	}
}

