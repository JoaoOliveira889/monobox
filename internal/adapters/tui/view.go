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
	if m.showSplash {
		return m.renderSplash()
	}
	if m.width < minTerminalWidth || m.height < minTerminalHeight {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			ui.ErrorStyle.Render(fmt.Sprintf(
				"Terminal too small.\nResize to at least %d×%d.",
				minTerminalWidth, minTerminalHeight,
			)),
		)
	}

	header := m.renderHeader()
	body := m.renderBody()
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

// renderSplash renders the startup splash screen (monogit style).
func (m *Model) renderSplash() string {
	barWidth := 20
	filled := (m.splashFrame * 2) % (barWidth + 1)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	progressBar := ui.SpinnerStyle.Render(bar)

	scanStatus := ui.SpinnerStyle.Render(splashDots(m.splashFrame) + " detecting engine…")
	version := ui.SubtleStyle.Render("v" + Version)
	subtitle := ui.SubtleStyle.Render("Docker & Podman container manager")

	body := lipgloss.JoinVertical(lipgloss.Center,
		renderBrandWordmark(),
		"",
		subtitle,
		"",
		progressBar,
		"",
		scanStatus,
		"",
		version,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, body)
}

func splashDots(frame int) string {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	return frames[frame%len(frames)]
}

func renderBrandWordmark() string {
	return lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Bottom,
			ui.BrandMonoStyle.Render("mono"),
			ui.BrandBoxStyle.Render("box"),
		),
		ui.SubtleStyle.Render("container manager"),
	)
}

func renderBrandCompact() string {
	return lipgloss.JoinHorizontal(lipgloss.Bottom,
		ui.BrandMonoStyle.Render("mono"),
		ui.BrandBoxStyle.Render("box"),
	)
}

// renderHeader — matches monogit's header style:
// brand  •  spinner/status  •  container count
func (m *Model) renderHeader() string {
	brand := renderBrandCompact()

	// Status / spinner line
	var middle string
	if m.loading {
		middle = ui.SpinnerStyle.Render(m.spinnerView() + " loading containers…")
	} else if m.stream != nil {
		middle = ui.SpinnerStyle.Render(m.spinnerView() + " streaming logs…")
	}

	// Right: container count + engine
	stats := ""
	if m.width >= 50 {
		running := 0
		for _, c := range m.containers {
			if c.IsRunning() {
				running++
			}
		}
		stats = fmt.Sprintf("%d containers  %d running  [%s] ", len(m.containers), running, m.engine)
	}

	spacerLen := m.width - lipgloss.Width(brand) - lipgloss.Width(middle) - lipgloss.Width(stats)
	if spacerLen < 0 {
		spacerLen = 0
	}
	spacer := strings.Repeat(" ", spacerLen)

	headerLine := " " + lipgloss.JoinHorizontal(lipgloss.Bottom,
		brand,
		spacer,
		middle,
		ui.SubtleStyle.Render(stats),
	)

	border := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorBorder)).
		Render(strings.Repeat("─", m.width))

	// Status bar line (success/error/info messages)
	statusLine := m.renderStatusBar()

	return headerLine + "\n" + statusLine + "\n" + border
}

func (m *Model) renderStatusBar() string {
	if m.statusMsg == "" {
		// Show summary: running / stopped counts
		var parts []string
		running, stopped := 0, 0
		for _, c := range m.containers {
			if c.IsRunning() {
				running++
			} else {
				stopped++
			}
		}
		dot := lipgloss.NewStyle().Foreground(ui.ColorSuccess).Render("●")
		if len(m.containers) > 0 {
			parts = append(parts, fmt.Sprintf("%d running", running))
			if stopped > 0 {
				parts = append(parts, ui.SubtleStyle.Render(fmt.Sprintf("%d stopped", stopped)))
			}
		} else if !m.loading {
			parts = append(parts, "No containers found")
		}

		sep := ui.SubtleStyle.Render("  •  ")
		barText := " " + dot + " " + strings.Join(parts, sep)
		return ui.SubtleStyle.Width(m.width).Render(barText)
	}

	switch {
	case strings.HasPrefix(m.statusMsg, "✓"):
		return ui.StatusSuccessStyle.Width(m.width).Render(" " + m.statusMsg)
	case strings.HasPrefix(m.statusMsg, "✗"):
		return ui.StatusErrorStyle.Width(m.width).Render(" " + m.statusMsg)
	default:
		return ui.StatusInfoStyle.Width(m.width).Render(" " + m.statusMsg)
	}
}

// renderFooter — matches monogit's footer: key hints left, version right.
func (m *Model) renderFooter() string {
	sep := ui.SubtleStyle.Render(" • ")
	var parts []string

	if m.activePanel == LogsPanel {
		parts = []string{
			m.fmtKey("↑↓/jk", "scroll"),
			m.fmtKey("f", "follow"),
			m.fmtKey("esc", "back"),
			m.fmtKey("q", "quit"),
		}
	} else {
		parts = []string{
			m.fmtKey("↑↓/jk", "navigate"),
			m.fmtKey("enter", "logs"),
			m.fmtKey("s", "start/stop"),
			m.fmtKey("r", "restart"),
			m.fmtKey("q", "quit"),
		}
	}

	version := ui.SubtleStyle.Render(fmt.Sprintf("monobox %s", Version))

	contentWidth := m.width - 2
	if contentWidth < 10 {
		contentWidth = 10
	}

	rendered := strings.Join(parts, sep)
	maxLeftWidth := contentWidth - lipgloss.Width(version) - 1

	for len(parts) > 0 && lipgloss.Width(rendered) > maxLeftWidth {
		parts = parts[:len(parts)-1]
		rendered = strings.Join(parts, sep)
	}

	spacerLen := contentWidth - lipgloss.Width(rendered) - lipgloss.Width(version)
	if spacerLen < 0 {
		spacerLen = 0
	}
	spacer := strings.Repeat(" ", spacerLen)

	footerText := " " + rendered + spacer + version
	if w := lipgloss.Width(footerText); w < contentWidth+1 {
		footerText += strings.Repeat(" ", contentWidth+1-w)
	}

	return ui.FooterStyle.Padding(0, 0).Render(footerText)
}

func (m *Model) fmtKey(k, action string) string {
	return ui.FooterKeyStyle.Render(k) + " " + ui.FooterActionStyle.Render(action)
}

// renderBody renders the two-panel layout.
func (m *Model) renderBody() string {
	ph := m.panelHeight()

	// Container list panel — "Containers" title, mono accent when active
	listAccent := lipgloss.Color(ui.ColorMono)
	left := m.renderTitledPanel(
		m.leftPanelWidth(), ph,
		"Containers",
		renderViewportWithScrollbar(m.listViewport, m.activePanel == ListPanel),
		m.activePanel == ListPanel,
		listAccent,
	)

	// Logs panel — show container name in title, box accent when active
	logsTitle := "Logs"
	if c := m.selectedContainer(); c != nil {
		logsTitle = "Logs — " + c.Name
		if m.logFollow {
			logsTitle += " [follow]"
		}
	}

	logsContent := m.logViewport.View()
	if len(m.logLines) == 0 && m.activePanel == LogsPanel {
		if m.stream != nil {
			logsContent = ui.SpinnerStyle.Render(m.spinnerView() + " Reading logs…")
		}
	} else if len(m.logLines) == 0 {
		logsContent = ui.SubtleStyle.Render("Select a container and press Enter to view logs.")
	}

	right := m.renderTitledPanel(
		m.rightPanelWidth(), ph,
		logsTitle,
		logsContent,
		m.activePanel == LogsPanel,
		lipgloss.Color(ui.ColorBox),
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

// refreshViewports syncs viewport content after state changes.
func (m *Model) refreshViewports() {
	m.refreshListViewport()
	m.refreshLogViewportContent()
}

// renderCenteredModal centers content on screen (for future modals).
func (m *Model) renderCenteredModal(content string) string {
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		ui.ActivePanelStyle.Padding(1, 2).Render(content),
	)
}

// renderViewportWithScrollbar renders viewport content with a scrollbar indicator
// when content overflows — matches monogit's pattern.
func renderViewportWithScrollbar(vp interface {
	View() string
	ScrollPercent() float64
}, active bool) string {
	return vp.View()
}
