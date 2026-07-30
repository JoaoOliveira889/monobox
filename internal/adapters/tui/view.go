package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoOliveira889/monobox/internal/pkg/ui"
)

// View renders the full TUI frame.
func (m *Model) View() string {
	if m.quitting {
		return ""
	}
	if m.width < minTerminalWidth || m.height < minTerminalHeight {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			ui.ErrorStyle.Render(fmt.Sprintf(
				"Terminal too small.\nPlease resize to at least %d×%d.",
				minTerminalWidth, minTerminalHeight,
			)),
		)
	}

	header := m.renderHeader()
	body := m.renderBody()
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

// refreshViewports syncs viewport content after state changes.
func (m *Model) refreshViewports() {
	m.refreshListViewport()
	m.refreshLogViewportContent()
}

// renderCenteredModal centers a modal content string over the full screen.
func (m *Model) renderCenteredModal(content string) string {
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		ui.ActivePanelStyle.Padding(1, 2).Render(content),
	)
}

// renderHeader renders the top bar with brand and engine info.
func (m *Model) renderHeader() string {
	brand := lipgloss.JoinHorizontal(lipgloss.Bottom,
		ui.BrandMonoStyle.Render("mono"),
		ui.BrandBoxStyle.Render("box"),
	)

	engineTag := ui.SubtleStyle.Render(" [" + m.engine + "]")
	left := lipgloss.JoinHorizontal(lipgloss.Center, brand, engineTag)

	var right string
	if m.loading {
		right = ui.SpinnerStyle.Render(m.spinnerView() + " loading…")
	} else {
		count := fmt.Sprintf("%d containers", len(m.containers))
		right = ui.SubtleStyle.Render(count)
	}

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	return lipgloss.NewStyle().
		Background(lipgloss.Color(ui.ColorBg)).
		Foreground(lipgloss.Color(ui.ColorFg)).
		Width(m.width).
		Render(left + strings.Repeat(" ", gap) + right)
}

// renderFooter renders the bottom key-hint bar.
func (m *Model) renderFooter() string {
	var hints []string
	if m.activePanel == ListPanel {
		hints = []string{
			m.fkey("↑↓/jk", "navigate"),
			m.fkey("enter", "logs"),
			m.fkey("s", "start/stop"),
			m.fkey("r", "restart"),
			m.fkey("q", "quit"),
		}
	} else {
		hints = []string{
			m.fkey("↑↓", "scroll"),
			m.fkey("f", "follow"),
			m.fkey("esc", "back"),
			m.fkey("q", "quit"),
		}
	}

	bar := strings.Join(hints, "  ")

	// Status message on the right.
	right := ""
	if m.statusMsg != "" {
		right = ui.StatusInfoStyle.Render(m.statusMsg)
	}

	gap := m.width - lipgloss.Width(bar) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	return ui.FooterStyle.Width(m.width).Render(bar + strings.Repeat(" ", gap) + right)
}

func (m *Model) fkey(key, action string) string {
	return ui.FooterKeyStyle.Render(key) + " " + ui.FooterActionStyle.Render(action)
}

// renderBody renders the two-panel layout.
func (m *Model) renderBody() string {
	ph := m.panelHeight()

	left := m.renderTitledPanel(
		m.leftPanelWidth(), ph,
		"Containers",
		m.listViewport.View(),
		m.activePanel == ListPanel,
		ui.ColorHighlight,
	)

	var logsTitle string
	if c := m.selectedContainer(); c != nil {
		logsTitle = "Logs — " + c.Name
		if m.logFollow {
			logsTitle += " [follow]"
		}
	} else {
		logsTitle = "Logs"
	}

	right := m.renderTitledPanel(
		m.rightPanelWidth(), ph,
		logsTitle,
		m.logViewport.View(),
		m.activePanel == LogsPanel,
		ui.ColorBox,
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}
