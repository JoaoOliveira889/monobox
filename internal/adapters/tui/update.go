package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monobox/internal/domain"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	oldStatus := m.statusMsg
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cmd = m.handleResize(msg)
	case splashTickMsg:
		if m.showSplash {
			m.splashFrame++
			cmd = splashTickCmd()
		}
	case tickMsg:
		cmd = tea.Batch(m.loadContainersCmd(), m.loadStatsCmd())
	case statsLoadedMsg:
		cmd = m.handleStatsLoaded(msg)
	case containersLoadedMsg:
		cmd = m.handleContainersLoaded(msg)
	case containerActionDoneMsg:
		cmd = m.handleActionDone(msg)
	case batchActionDoneMsg:
		cmd = m.handleBatchActionDone(msg)
	case containerLogsClearedMsg:
		cmd = m.handleContainerLogsCleared(msg)
	case pruneDoneMsg:
		cmd = m.handlePruneDone(msg)
	case composeActionDoneMsg:
		cmd = m.handleComposeActionDone(msg)
	case inspectDoneMsg:
		cmd = m.handleInspectDone(msg)
	case execDoneMsg:
		cmd = m.handleExecDone(msg)
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

	w := m.width - 8
	if w < 20 {
		w = 20
	}
	h := m.height - 8
	if h < 5 {
		h = 5
	}
	m.inspectViewport.Width = w
	m.inspectViewport.Height = h

	m.refreshListViewport()
	m.refreshLogViewportContent()
	return nil
}

func (m *Model) handleInspectDone(msg inspectDoneMsg) tea.Cmd {
	if msg.err != nil {
		m.setStatus(fmt.Sprintf("✗ Inspect failed: %s", msg.err))
		return nil
	}
	if details, err := domain.ParseInspectDetails(msg.content); err == nil {
		if m.inspectDetailsCache == nil {
			m.inspectDetailsCache = make(map[string]*domain.ContainerInspectDetails)
		}
		m.inspectDetailsCache[msg.containerID] = details
	}
	m.showInspect = true
	m.inspectContent = msg.content
	w := m.width - 8
	if w < 20 {
		w = 20
	}
	h := m.height - 8
	if h < 5 {
		h = 5
	}
	m.inspectViewport = viewport.New(w, h)
	m.inspectViewport.SetContent(msg.content)
	return nil
}

func (m *Model) handleExecDone(msg execDoneMsg) tea.Cmd {
	if msg.err != nil {
		m.setStatus(fmt.Sprintf("✗ Shell exited: %s", msg.err))
	} else {
		m.setStatus("✓ Shell session closed")
	}
	return tea.Batch(m.loadContainersCmd(), m.loadStatsCmd())
}

func (m *Model) handleStatsLoaded(msg statsLoadedMsg) tea.Cmd {
	if msg.err != nil {
		return nil
	}
	m.ApplyStats(msg.stats)
	m.refreshListViewport()
	return nil
}

func (m *Model) handleContainersLoaded(msg containersLoadedMsg) tea.Cmd {
	m.loading = false
	m.showSplash = false

	if msg.err != nil {
		m.setStatus(fmt.Sprintf("✗ Error loading containers: %s", msg.err))
		return m.tickCmd()
	}

	m.ApplyContainersLoaded(func() []domain.Container {
		list := make([]domain.Container, len(msg.containers))
		for i, item := range msg.containers {
			list[i] = item.Container
		}
		return list
	}())

	m.refreshListViewport()
	if m.stream == nil {
		if c := m.selectedContainer(); c != nil {
			return tea.Batch(m.tickCmd(), m.startLogStream(c.ID))
		}
	}
	return m.tickCmd()
}

func (m *Model) handleActionDone(msg containerActionDoneMsg) tea.Cmd {
	for i, c := range m.containers {
		if c.ID == msg.id {
			m.containers[i].loading = false
			m.containers[i].starting = false
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
	return tea.Batch(
		m.loadContainersCmd(),
		m.loadStatsCmd(),
		tea.Tick(800*time.Millisecond, func(time.Time) tea.Msg {
			return tickMsg(time.Now())
		}),
	)
}

func (m *Model) handleBatchActionDone(msg batchActionDoneMsg) tea.Cmd {
	for i, c := range m.containers {
		if c.ComposeProject == msg.projectName {
			m.containers[i].loading = false
		}
	}
	m.refreshListViewport()

	if msg.err != nil && msg.success == 0 {
		m.setStatus(fmt.Sprintf("✗ batch %s failed for %s: %s", msg.action, msg.projectName, msg.err))
	} else if msg.err != nil {
		m.setStatus(fmt.Sprintf("⚠ batch %s for %s (%d/%d succeeded): %s", msg.action, msg.projectName, msg.success, msg.count, msg.err))
	} else {
		m.setStatus(fmt.Sprintf("✓ batch %s completed for %s (%d containers)", msg.action, msg.projectName, msg.count))
	}

	return tea.Batch(
		m.loadContainersCmd(),
		m.loadStatsCmd(),
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

func (m *Model) handlePruneDone(msg pruneDoneMsg) tea.Cmd {
	if msg.err != nil {
		m.setStatus(fmt.Sprintf("✗ Prune failed: %s", msg.err))
	} else {
		m.setStatus("✓ System prune complete")
	}
	return tea.Batch(m.loadContainersCmd(), m.loadStatsCmd())
}

func (m *Model) handleComposeActionDone(msg composeActionDoneMsg) tea.Cmd {
	for i, c := range m.containers {
		if c.ComposeProject == msg.project {
			m.containers[i].loading = false
		}
	}
	m.refreshListViewport()
	if msg.err != nil {
		m.setStatus(fmt.Sprintf("✗ Compose %s for %s failed: %s", msg.action, msg.project, msg.err))
	} else {
		m.setStatus(fmt.Sprintf("✓ Compose %s %s", msg.action, msg.project))
	}
	return tea.Batch(
		m.loadContainersCmd(),
		m.loadStatsCmd(),
		tea.Tick(800*time.Millisecond, func(time.Time) tea.Msg {
			return tickMsg(time.Now())
		}),
	)
}

