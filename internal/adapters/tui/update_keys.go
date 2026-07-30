package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleKeys routes keyboard events to the correct panel handler.
func (m *Model) handleKeys(msg tea.KeyMsg) tea.Cmd {
	// Global quit — always available.
	if matchesKey(msg, keys.Quit...) {
		m.quitting = true
		m.cancelStream()
		return tea.Quit
	}

	if m.activePanel == LogsPanel {
		return m.handleLogsKeys(msg)
	}
	return m.handleListKeys(msg)
}

// handleListKeys handles keys when the container list is focused.
func (m *Model) handleListKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case matchesKey(msg, keys.Up...):
		if m.cursor > 0 {
			m.cursor--
			m.refreshListViewport()
		}

	case matchesKey(msg, keys.Down...):
		if m.cursor < len(m.containers)-1 {
			m.cursor++
			m.refreshListViewport()
		}

	case matchesKey(msg, keys.Enter...):
		// Open logs for the selected container.
		c := m.selectedContainer()
		if c == nil {
			return nil
		}
		m.activePanel = LogsPanel
		m.logLines = nil
		m.refreshLogViewportContent()
		return openLogStreamCmd(m.provider, c.ID)

	case matchesKey(msg, keys.Toggle...):
		// s: start if stopped, stop if running.
		c := m.selectedContainer()
		if c == nil || c.loading {
			return nil
		}
		action := "start"
		if c.IsRunning() {
			action = "stop"
		}
		m.markContainerLoading(m.cursor)
		m.setStatus("⟳ " + action + "…")
		return m.containerActionCmd(c.ID, action)

	case matchesKey(msg, keys.Restart...):
		c := m.selectedContainer()
		if c == nil || c.loading {
			return nil
		}
		m.markContainerLoading(m.cursor)
		m.setStatus("⟳ restarting…")
		return m.containerActionCmd(c.ID, "restart")

	case matchesKey(msg, keys.Help...):
		// TODO: help overlay
	}
	return nil
}

// handleLogsKeys handles keys when the log panel is focused.
func (m *Model) handleLogsKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case matchesKey(msg, keys.Esc...):
		// Return to list without stopping the container.
		m.cancelStream()
		m.activePanel = ListPanel
		m.logLines = nil
		m.refreshLogViewportContent()
		m.refreshListViewport()

	case matchesKey(msg, keys.Follow...):
		// f: toggle log follow mode.
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
	}
	return nil
}

func (m *Model) markContainerLoading(index int) {
	if index >= 0 && index < len(m.containers) {
		m.containers[index].loading = true
		m.refreshListViewport()
	}
}
