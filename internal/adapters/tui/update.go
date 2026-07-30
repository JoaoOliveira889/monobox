package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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

	listInner := m.leftPanelWidth() - 2
	if listInner < 0 {
		listInner = 0
	}
	logsInner := m.rightPanelWidth() - 2
	if logsInner < 0 {
		logsInner = 0
	}
	ph := m.panelHeight()

	if m.listViewport.Width == 0 {
		m.listViewport = viewport.New(listInner, ph)
	} else {
		m.listViewport.Width = listInner
		m.listViewport.Height = ph
	}
	if m.logViewport.Width == 0 {
		m.logViewport = viewport.New(logsInner, ph)
	} else {
		m.logViewport.Width = logsInner
		m.logViewport.Height = ph
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
	return tickCmd()
}

func (m *Model) handleActionDone(msg containerActionDoneMsg) tea.Cmd {
	for i, c := range m.containers {
		if c.ID == msg.id {
			m.containers[i].loading = false
			break
		}
	}
	if msg.err != nil {
		m.setStatus(fmt.Sprintf("✗ %s failed: %s", msg.action, msg.err))
	} else {
		shortID := msg.id
		if len(shortID) > 12 {
			shortID = shortID[:12]
		}
		m.setStatus(fmt.Sprintf("✓ %s %s", msg.action, shortID))
	}
	return m.loadContainersCmd()
}

func (m *Model) handleLogStreamOpened(msg logStreamOpenedMsg) tea.Cmd {
	m.cancelStream()
	m.logLines = nil
	sc := newLogScanner(msg.reader)
	m.stream = &logStream{
		containerID: msg.containerID,
		scanner:     sc,
		ctx:         msg.ctx,
		cancel:      msg.cancel,
	}
	// Clear log viewport content to show "Reading logs…" placeholder.
	m.logViewport.SetContent("")
	return nextLogLineCmd(m.stream.ctx, msg.containerID, sc)
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
		m.stream = nil
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
