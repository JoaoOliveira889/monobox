package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monobox/internal/domain"
	"github.com/JoaoOliveira889/monobox/internal/pkg/config"
	"github.com/JoaoOliveira889/monobox/internal/pkg/ui"
)

func (m *Model) handleKeys(msg tea.KeyMsg) tea.Cmd {
	if m.showHelp {
		return m.handleHelpKeys(msg)
	}

	if m.showThemeMenu {
		return m.handleThemeMenuKeys(msg)
	}

	if m.showGraphModal {
		return m.handleGraphModalKeys(msg)
	}

	if m.showEnvModal {
		return m.handleEnvModalKeys(msg)
	}

	if m.showHealthModal {
		return m.handleHealthModalKeys(msg)
	}

	if m.showInspect {
		return m.handleInspectKeys(msg)
	}

	if m.confirmPortConflict {
		return m.handlePortConflictConfirmation(msg)
	}

	if m.confirmClearLogs {
		return m.handleClearLogsConfirmation(msg)
	}

	if m.confirmRemove {
		return m.handleRemoveConfirmation(msg)
	}

	if m.confirmBatchAction != "" {
		return m.handleBatchConfirmation(msg)
	}

	if m.filtering {
		return m.handleFilterKeys(msg)
	}

	if m.logSearching {
		return m.handleLogSearchKeys(msg)
	}

	if matchesKey(msg, keys.Quit...) {
		m.quitting = true
		m.cancelStream()
		return tea.Quit
	}

	if matchesKey(msg, keys.Help...) {
		m.showHelp = true
		return nil
	}

	if matchesKey(msg, keys.EnvModal...) {
		c := m.selectedContainer()
		if c == nil {
			return nil
		}
		m.showEnvModal = true
		if m.inspectDetailsCache == nil || m.inspectDetailsCache[c.ID] == nil {
			return inspectCmd(m.provider, c.ID)
		}
		return nil
	}

	if matchesKey(msg, keys.HealthModal...) {
		c := m.selectedContainer()
		if c == nil {
			return nil
		}
		m.showHealthModal = true
		if m.inspectDetailsCache == nil || m.inspectDetailsCache[c.ID] == nil {
			return inspectCmd(m.provider, c.ID)
		}
		return nil
	}

	if matchesKey(msg, keys.ThemeMenu...) {
		m.showThemeMenu = true
		m.initialTheme = m.cfg.Theme
		m.themeCursor = 0
		for i, t := range ui.Themes {
			if strings.EqualFold(t.Name, m.cfg.Theme) {
				m.themeCursor = i
				break
			}
		}
		return nil
	}

	if matchesKey(msg, keys.Panel1...) {
		m.activePanel = ListPanel
		m.refreshListViewport()
		return nil
	}

	if matchesKey(msg, keys.Panel2...) {
		m.activePanel = LogsPanel
		m.logFollow = true
		m.logViewport.GotoBottom()
		c := m.selectedContainer()
		if c != nil && m.stream == nil {
			return m.selectAndStreamContainerLogs()
		}
		return nil
	}

	if matchesKey(msg, keys.Tab...) {
		if m.activePanel == ListPanel {
			m.activePanel = LogsPanel
			m.logFollow = true
			m.logViewport.GotoBottom()
			c := m.selectedContainer()
			if c != nil && m.stream == nil {
				return m.selectAndStreamContainerLogs()
			}
		} else {
			m.activePanel = ListPanel
		}
		return nil
	}

	if matchesKey(msg, keys.ResizeLeft...) {
		m.leftPanelRatio -= resizeStep
		if m.leftPanelRatio < minLeftPanelRatio {
			m.leftPanelRatio = minLeftPanelRatio
		}
		return m.handleResize(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	}
	if matchesKey(msg, keys.ResizeRight...) {
		m.leftPanelRatio += resizeStep
		if m.leftPanelRatio > maxLeftPanelRatio {
			m.leftPanelRatio = maxLeftPanelRatio
		}
		return m.handleResize(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	}

	if m.activePanel == LogsPanel && matchesKey(msg, keys.ClearLogs...) {
		c := m.selectedContainer()
		if c == nil {
			return nil
		}
		m.confirmClearLogs = true
		return nil
	}

	if m.activePanel == LogsPanel {
		return m.handleLogsKeys(msg)
	}
	return m.handleListKeys(msg)
}

func (m *Model) handleInspectKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q", "i":
		m.showInspect = false
		m.inspectContent = ""
		return nil
	case "up", "k":
		m.inspectViewport.LineUp(1)
	case "down", "j":
		m.inspectViewport.LineDown(1)
	case "pgup", "ctrl+u":
		m.inspectViewport.PageUp()
	case "pgdown", "ctrl+d":
		m.inspectViewport.PageDown()
	}
	return nil
}

func (m *Model) handleRemoveConfirmation(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "y", "Y":
		c := m.selectedContainer()
		m.confirmRemove = false
		if c == nil {
			return nil
		}
		m.markContainerLoading(m.cursor)
		m.setStatus("⟳ removing " + c.Name + "…")
		return m.containerActionCmd(c.ID, "remove")
	case "n", "N", "esc":
		m.confirmRemove = false
	}
	return nil
}

func (m *Model) handleBatchConfirmation(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "y", "Y":
		action := m.confirmBatchAction
		projName := m.batchProjectName
		m.confirmBatchAction = ""
		m.batchProjectName = ""

		var ids []string
		for i, c := range m.containers {
			if c.ComposeProject == projName {
				ids = append(ids, c.ID)
				m.containers[i].loading = true
			}
		}
		if len(ids) == 0 {
			return nil
		}

		realAction := strings.TrimPrefix(action, "batch_")
		m.setStatus(fmt.Sprintf("⟳ %s all containers in %s…", realAction, projName))
		m.refreshListViewport()
		return m.batchActionCmd(projName, realAction, ids)

	case "n", "N", "esc":
		m.confirmBatchAction = ""
		m.batchProjectName = ""
	}
	return nil
}

func (m *Model) handleFilterKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.filtering = false
		m.filterInput.Blur()
		m.filterInput.SetValue("")
		m.filterQuery = ""
		m.cursor = 0
		m.refreshListViewport()
		return nil

	case "enter":
		m.filtering = false
		m.filterInput.Blur()
		m.refreshListViewport()
		return nil

	case "up", "down":
		m.filtering = false
		m.filterInput.Blur()
		return m.handleListKeys(msg)
	}

	var cmd tea.Cmd
	m.filterInput, cmd = m.filterInput.Update(msg)
	m.filterQuery = m.filterInput.Value()

	nodes := m.VisibleTreeNodes()
	if m.cursor >= len(nodes) {
		m.cursor = max(0, len(nodes)-1)
	}
	m.refreshListViewport()
	return cmd
}

func (m *Model) selectAndStreamContainerLogs() tea.Cmd {
	m.cancelStream()
	m.logLines = nil
	m.refreshLogViewportContent()
	c := m.selectedContainer()
	if c != nil {
		return m.startLogStream(c.ID)
	}
	return nil
}

func (m *Model) handleListKeys(msg tea.KeyMsg) tea.Cmd {
	nodes := m.VisibleTreeNodes()
	node := m.selectedNode()

	switch {
	case matchesKey(msg, keys.Filter...):
		m.filtering = true
		m.filterInput.Focus()
		return textinput.Blink

	case msg.String() == " " || msg.String() == "left" || msg.String() == "right" || msg.String() == "h" || msg.String() == "l":
		if node != nil && node.Type == NodeProjectHeader {
			m.toggleProjectExpanded(node.ProjectName)
			m.refreshListViewport()
			return nil
		}

	case matchesKey(msg, keys.Exec...):
		c := m.selectedContainer()
		if c == nil {
			return nil
		}
		if !c.IsRunning() {
			m.setStatus("✗ Container must be running to exec shell")
			return nil
		}
		return tea.ExecProcess(m.provider.ExecCmd(c.ID), func(err error) tea.Msg {
			return execDoneMsg{err: err}
		})

	case matchesKey(msg, keys.Inspect...):
		c := m.selectedContainer()
		if c == nil {
			return nil
		}
		m.setStatus("⟳ inspecting " + c.Name + "…")
		return inspectCmd(m.provider, c.ID)

	case matchesKey(msg, keys.Remove...):
		if node != nil && node.Type == NodeProjectHeader {
			m.confirmBatchAction = "batch_remove"
			m.batchProjectName = node.ProjectName
			return nil
		}
		c := m.selectedContainer()
		if c == nil {
			return nil
		}
		m.confirmRemove = true
		return nil

	case matchesKey(msg, keys.Pause...):
		c := m.selectedContainer()
		if c == nil || c.loading {
			return nil
		}
		if c.Status == domain.StatusPaused {
			m.markContainerLoading(m.cursor)
			m.setStatus("⟳ unpausing " + c.Name + "…")
			return m.containerActionCmd(c.ID, "unpause")
		} else if c.IsRunning() {
			m.markContainerLoading(m.cursor)
			m.setStatus("⟳ pausing " + c.Name + "…")
			return m.containerActionCmd(c.ID, "pause")
		} else {
			m.setStatus("✗ Container must be running to pause")
			return nil
		}

	case matchesKey(msg, keys.OpenPort...):
		c := m.selectedContainer()
		if c == nil {
			return nil
		}
		port := extractHostPort(c.Ports)
		if port == "" {
			m.setStatus("✗ No exposed host ports found for " + c.Name)
			return nil
		}
		m.setStatus("✓ Opening http://localhost:" + port + "…")
		return openBrowserCmd(port)

	case matchesKey(msg, keys.Up...):
		if m.cursor > 0 {
			m.cursor--
			m.refreshListViewport()
			return m.selectAndStreamContainerLogs()
		}

	case matchesKey(msg, keys.Down...):
		if m.cursor < len(nodes)-1 {
			m.cursor++
			m.refreshListViewport()
			return m.selectAndStreamContainerLogs()
		}

	case matchesKey(msg, keys.Enter...):
		if node != nil && node.Type == NodeProjectHeader {
			m.toggleProjectExpanded(node.ProjectName)
			m.refreshListViewport()
			return nil
		}
		c := m.selectedContainer()
		if c == nil {
			return nil
		}
		m.activePanel = LogsPanel
		m.logFollow = true
		return m.selectAndStreamContainerLogs()

	case matchesKey(msg, keys.Graph...):
		c := m.selectedContainer()
		if c == nil {
			return nil
		}
		m.showGraphModal = true
		return nil

	case matchesKey(msg, keys.Toggle...):
		if node != nil && node.Type == NodeProjectHeader {
			action := "batch_start"
			if node.RunningCount == node.TotalCount {
				action = "batch_stop"
			}
			m.confirmBatchAction = action
			m.batchProjectName = node.ProjectName
			return nil
		}
		c := m.selectedContainer()
		if c == nil || c.loading {
			return nil
		}
		if c.IsRunning() {
			m.markContainerLoading(m.cursor)
			m.setStatus("⟳ stop…")
			return m.containerActionCmd(c.ID, "stop")
		}
		if hasConflict, otherName, port := m.detectPortConflict(c); hasConflict {
			m.confirmPortConflict = true
			m.conflictingContainer = otherName
			m.conflictingPort = port
			m.pendingStartTarget = c.ID
			return nil
		}
		return m.startContainerOptimistic(c.ID, c.Name)

	case matchesKey(msg, keys.Restart...):
		if node != nil && node.Type == NodeProjectHeader {
			m.confirmBatchAction = "batch_restart"
			m.batchProjectName = node.ProjectName
			return nil
		}
		c := m.selectedContainer()
		if c == nil || c.loading {
			return nil
		}
		m.markContainerLoading(m.cursor)
		m.setStatus("⟳ restarting…")
		return m.containerActionCmd(c.ID, "restart")
	}
	return nil
}

func (m *Model) handleLogSearchKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.logSearching = false
		m.logSearchInput.Blur()
		m.logSearchInput.SetValue("")
		m.logSearchQuery = ""
		m.refreshLogViewportContent()
		return nil

	case "enter":
		m.logSearching = false
		m.logSearchInput.Blur()
		m.refreshLogViewportContent()
		return nil
	}

	var cmd tea.Cmd
	m.logSearchInput, cmd = m.logSearchInput.Update(msg)
	m.logSearchQuery = m.logSearchInput.Value()
	m.refreshLogViewportContent()
	return cmd
}

func (m *Model) exportLogsToFile() tea.Cmd {
	if len(m.logLines) == 0 {
		m.setStatus("✗ No logs to export")
		return nil
	}
	c := m.selectedContainer()
	containerName := "app"
	if c != nil && c.Name != "" {
		containerName = strings.TrimPrefix(c.Name, "/")
		containerName = strings.ReplaceAll(containerName, " ", "-")
	}
	dateStr := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("./%s-logs-%s.log", containerName, dateStr)

	query := strings.TrimSpace(m.logSearchQuery)
	queryLower := strings.ToLower(query)
	var linesToSave []string
	for _, l := range m.logLines {
		if queryLower != "" && !strings.Contains(strings.ToLower(l), queryLower) {
			continue
		}
		linesToSave = append(linesToSave, l)
	}

	content := strings.Join(linesToSave, "\n") + "\n"
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		m.setStatus(fmt.Sprintf("✗ Failed to export logs: %v", err))
	} else {
		m.setStatus(fmt.Sprintf("✓ Saved %d log lines to %s", len(linesToSave), filename))
	}
	return nil
}

func (m *Model) handleLogsKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case matchesKey(msg, keys.Filter...):
		m.logSearching = true
		m.logSearchInput.Focus()
		return textinput.Blink

	case matchesKey(msg, keys.SaveLogs...):
		return m.exportLogsToFile()

	case matchesKey(msg, keys.ToggleTimestamps...):
		m.showTimestamps = !m.showTimestamps
		if m.showTimestamps {
			m.setStatus("Timestamps: ON")
		} else {
			m.setStatus("Timestamps: OFF")
		}
		return m.selectAndStreamContainerLogs()

	case matchesKey(msg, keys.Esc...):
		if m.logSearchQuery != "" {
			m.logSearchQuery = ""
			m.logSearchInput.SetValue("")
			m.refreshLogViewportContent()
			return nil
		}
		m.cancelStream()
		m.activePanel = ListPanel
		m.logLines = nil
		m.refreshLogViewportContent()
		m.refreshListViewport()

	case matchesKey(msg, keys.Follow...):
		m.logFollow = !m.logFollow
		if m.logFollow {
			m.logViewport.GotoBottom()
			m.setStatus("Follow: ON")
		} else {
			m.setStatus("Follow: OFF")
		}

	case matchesKey(msg, keys.Up...):
		m.logViewport.LineUp(1)
		m.logFollow = false

	case matchesKey(msg, keys.Down...):
		m.logViewport.LineDown(1)
		if m.logViewport.AtBottom() {
			m.logFollow = true
		}

	case matchesKey(msg, keys.PageUp...):
		m.logViewport.PageUp()
		m.logFollow = false

	case matchesKey(msg, keys.PageDown...):
		m.logViewport.PageDown()
		if m.logViewport.AtBottom() {
			m.logFollow = true
		}

	case matchesKey(msg, keys.End...):
		m.logViewport.GotoBottom()
		m.logFollow = true
	}
	return nil
}

func (m *Model) startLogStream(containerID string) tea.Cmd {
	m.cancelStream()
	m.logContainerID = containerID
	tail := m.logViewport.Height
	if tail <= 0 {
		tail = logTailLines
	}
	if tail > logTailLines {
		tail = logTailLines
	}
	return openLogStreamTailCmd(m.provider, containerID, tail, m.showTimestamps)
}

func (m *Model) handleClearLogsConfirmation(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "y", "Y":
		c := m.selectedContainer()
		m.confirmClearLogs = false
		if c == nil {
			return nil
		}
		m.cancelStream()
		m.setStatus("⟳ Clearing " + m.engine + " logs for " + c.Name + "…")
		return clearContainerLogsCmd(m.provider, c.ID)
	case "n", "N", "esc":
		m.confirmClearLogs = false
	}
	return nil
}

func (m *Model) markContainerLoading(index int) {
	filtered := m.FilteredContainers()
	if index >= 0 && index < len(filtered) {
		targetID := filtered[index].ID
		for i := range m.containers {
			if m.containers[i].ID == targetID {
				m.containers[i].loading = true
				break
			}
		}
		m.refreshListViewport()
	}
}

func (m *Model) handleHelpKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q", "?":
		m.showHelp = false
		return nil
	case "up", "k":
		m.helpViewport.LineUp(1)
	case "down", "j":
		m.helpViewport.LineDown(1)
	case "pgup", "ctrl+u":
		m.helpViewport.PageUp()
	case "pgdown", "ctrl+d":
		m.helpViewport.PageDown()
	}
	return nil
}

func (m *Model) startContainerOptimistic(id, name string) tea.Cmd {
	for i := range m.containers {
		if m.containers[i].ID == id {
			m.containers[i].loading = true
			m.containers[i].starting = true
			break
		}
	}
	sortContainerItems(m.containers)

	nodes := m.VisibleTreeNodes()
	for i, n := range nodes {
		if n.Type == NodeContainerItem && n.Container != nil && n.Container.ID == id {
			m.cursor = i
			break
		}
	}

	m.refreshListViewport()
	m.setStatus("⟳ starting " + name + "…")
	return m.containerActionCmd(id, "start")
}

func (m *Model) handleThemeMenuKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case matchesKey(msg, keys.Esc...) || msg.String() == "q" || msg.String() == "T":
		m.showThemeMenu = false
		ui.ApplyTheme(m.initialTheme)
		m.refreshListViewport()
		m.refreshLogViewportContent()
		return nil

	case matchesKey(msg, keys.Up...):
		if m.themeCursor > 0 {
			m.themeCursor--
		} else {
			m.themeCursor = len(ui.Themes) - 1
		}
		ui.ApplyTheme(ui.Themes[m.themeCursor].Name)
		m.refreshListViewport()
		m.refreshLogViewportContent()
		return nil

	case matchesKey(msg, keys.Down...):
		if m.themeCursor < len(ui.Themes)-1 {
			m.themeCursor++
		} else {
			m.themeCursor = 0
		}
		ui.ApplyTheme(ui.Themes[m.themeCursor].Name)
		m.refreshListViewport()
		m.refreshLogViewportContent()
		return nil

	case matchesKey(msg, keys.Enter...):
		selectedTheme := ui.Themes[m.themeCursor].Name
		m.cfg.Theme = selectedTheme
		m.showThemeMenu = false
		if err := config.Save(m.cfg); err != nil {
			m.setStatus("✗ Failed to save config: " + err.Error())
		} else {
			m.setStatus("✓ Theme set to " + selectedTheme)
		}
		ui.ApplyTheme(selectedTheme)
		m.refreshListViewport()
		m.refreshLogViewportContent()
		return nil
	}
	return nil
}

func (m *Model) handleEnvModalKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q", "E", "e":
		m.showEnvModal = false
		return nil
	case "up", "k":
		m.envViewport.LineUp(1)
	case "down", "j":
		m.envViewport.LineDown(1)
	case "pgup", "ctrl+u":
		m.envViewport.PageUp()
	case "pgdown", "ctrl+d":
		m.envViewport.PageDown()
	}
	return nil
}

func (m *Model) handleHealthModalKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q", "H", "h":
		m.showHealthModal = false
		return nil
	case "up", "k":
		m.healthViewport.LineUp(1)
	case "down", "j":
		m.healthViewport.LineDown(1)
	case "pgup", "ctrl+u":
		m.healthViewport.PageUp()
	case "pgdown", "ctrl+d":
		m.healthViewport.PageDown()
	}
	return nil
}

func (m *Model) handleGraphModalKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q", "g", "G":
		m.showGraphModal = false
		return nil
	case "up", "k":
		m.graphViewport.LineUp(1)
	case "down", "j":
		m.graphViewport.LineDown(1)
	case "pgup", "ctrl+u":
		m.graphViewport.PageUp()
	case "pgdown", "ctrl+d":
		m.graphViewport.PageDown()
	}
	return nil
}

func (m *Model) handlePortConflictConfirmation(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "y", "Y", "enter":
		m.confirmPortConflict = false
		targetID := m.pendingStartTarget
		m.pendingStartTarget = ""
		c := m.containerByID(targetID)
		if c != nil {
			return m.startContainerOptimistic(c.ID, c.Name)
		}
		return nil
	case "n", "N", "esc":
		m.confirmPortConflict = false
		m.pendingStartTarget = ""
		m.setStatus("Port conflict start cancelled")
		return nil
	}
	return nil
}
