package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoOliveira889/monobox/internal/domain"
	"github.com/JoaoOliveira889/monobox/internal/pkg/ui"
)

// renderTitledPanel renders a bordered panel with a decorative title in the
// top border — identical pattern to monogit.
func (m *Model) renderTitledPanel(width, height int, title, content string, active bool, accent lipgloss.Color) string {
	borderColor := lipgloss.Color(ui.ColorBorder)
	if active {
		borderColor = accent
	}

	border := lipgloss.RoundedBorder()
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	if active {
		borderStyle = borderStyle.Bold(true)
	}

	maxTitleWidth := width - 6
	if maxTitleWidth < 3 {
		maxTitleWidth = 3
	}
	titleRunes := []rune(title)
	if len(titleRunes) > maxTitleWidth {
		title = string(titleRunes[:maxTitleWidth-1]) + "…"
	}

	var titleStyled string
	if active {
		titleStyled = lipgloss.NewStyle().Foreground(accent).Bold(true).Render(title)
	} else {
		titleStyled = ui.SubtleStyle.Render(title)
	}

	titleText := "─[" + titleStyled + "]─"
	titleWidth := lipgloss.Width(titleText)

	repeatCount := width - titleWidth - 2
	if repeatCount < 0 {
		repeatCount = 0
	}

	topLine := borderStyle.Render(border.TopLeft) +
		borderStyle.Render("─[") + titleStyled + borderStyle.Render("]") +
		borderStyle.Render(strings.Repeat(border.Top, repeatCount)+border.TopRight)

	innerWidth := width - 2
	if innerWidth < 0 {
		innerWidth = 0
	}
	innerHeight := height - 2
	if innerHeight < 0 {
		innerHeight = 0
	}

	panelStyle := lipgloss.NewStyle().
		Border(border, false, true, true, true).
		BorderForeground(borderColor).
		Width(innerWidth).
		Height(innerHeight)

	panel := panelStyle.Render(content)
	return lipgloss.JoinVertical(lipgloss.Left, topLine, panel)
}

// renderRepoList renders the container list panel (left side).
func (m *Model) renderRepoList(width, height int) string {
	content := m.listViewport.View()
	if m.loading {
		content = ui.SpinnerStyle.Render(m.spinnerView() + " Loading containers…")
	} else if len(m.containers) == 0 {
		content = ui.SubtleStyle.Render(" No containers found.\n Make sure Docker or Podman is running.")
	}

	accent := lipgloss.Color(ui.ColorMono)
	return m.renderTitledPanel(width, height, "Containers", content, m.activePanel == ListPanel, accent)
}

// renderContainerListContent renders all container rows inside listViewport.
func (m *Model) renderContainerListContent() string {
	if m.loading {
		return ui.SpinnerStyle.Render(m.spinnerView() + " Loading containers…")
	}
	if len(m.containers) == 0 {
		return ui.SubtleStyle.Render(" No containers found.\n Make sure Docker or Podman is running.")
	}

	vpWidth := m.listViewport.Width
	if vpWidth < 10 {
		vpWidth = m.leftPanelWidth() - 4
	}

	var rows []string
	for i, c := range m.containers {
		rows = append(rows, m.renderContainerRow(i, c, vpWidth))
	}
	return strings.Join(rows, "\n")
}

// renderContainerRow renders a single container row in monogit repoLine style.
func (m *Model) renderContainerRow(index int, c containerItem, maxWidth int) string {
	selected := index == m.cursor
	var bgStyle lipgloss.Style
	if selected {
		bgStyle = lipgloss.NewStyle().Background(ui.ColorHighlight)
	}

	var prefix string
	if selected {
		prefix = bgStyle.Foreground(ui.ColorBg).Render("▌ ")
	} else {
		prefix = "  "
	}

	// Right-aligned status badge
	var statusBadge string
	if selected {
		statusBadge = bgStyle.Foreground(ui.ColorBg).Bold(true).Render(statusBadgeText(c.Container))
	} else {
		statusBadge = statusBadgeStyled(c.Container)
	}

	prefixWidth := lipgloss.Width(prefix)
	statusWidth := lipgloss.Width(statusBadge)

	availForText := maxWidth - prefixWidth - statusWidth - 1
	if availForText < 5 {
		availForText = 5
	}

	name := c.Name
	nameWidth := lipgloss.Width(name)

	var nameStr, imageStr string
	if nameWidth >= availForText {
		name = truncateRunes(name, availForText)
		if selected {
			nameStr = bgStyle.Foreground(ui.ColorBg).Bold(true).Render(name)
		} else {
			nameStr = lipgloss.NewStyle().Foreground(ui.ColorFg).Render(name)
		}
	} else {
		if selected {
			nameStr = bgStyle.Foreground(ui.ColorBg).Bold(true).Render(name)
		} else {
			nameStr = lipgloss.NewStyle().Foreground(ui.ColorFg).Render(name)
		}

		availForImage := availForText - nameWidth - 1
		if availForImage >= 4 {
			maxImgLen := availForImage - 3
			img := c.Image
			if lipgloss.Width(img) > maxImgLen {
				img = truncateRunes(img, maxImgLen)
			}
			if img != "" {
				if selected {
					imageStr = bgStyle.Foreground(ui.ColorBg).Render(" (") +
						bgStyle.Foreground(ui.ColorSelected).Bold(true).Render(img) +
						bgStyle.Foreground(ui.ColorBg).Render(")")
				} else {
					imageStr = lipgloss.NewStyle().Foreground(ui.ColorSubtle).Render(" (") +
						lipgloss.NewStyle().Foreground(ui.ColorCyan).Render(img) +
						lipgloss.NewStyle().Foreground(ui.ColorSubtle).Render(")")
				}
			}
		}
	}

	leftContent := prefix + nameStr
	if imageStr != "" {
		midSp := " "
		if selected {
			midSp = bgStyle.Render(" ")
		}
		leftContent += midSp + imageStr
	}

	leftWidth := lipgloss.Width(leftContent)
	gapLen := maxWidth - leftWidth - statusWidth
	if gapLen < 1 {
		gapLen = 1
	}
	gap := strings.Repeat(" ", gapLen)
	if selected {
		gap = bgStyle.Render(gap)
	}

	row := leftContent + gap + statusBadge
	rowWidth := lipgloss.Width(row)
	if rowWidth < maxWidth {
		padding := strings.Repeat(" ", maxWidth-rowWidth)
		if selected {
			row += bgStyle.Render(padding)
		} else {
			row += padding
		}
	}

	return row
}

// statusBadgeText produces unstyled text for selected row background.
func statusBadgeText(c domain.Container) string {
	switch c.Status {
	case domain.StatusRunning:
		return "● RUNNING"
	case domain.StatusExited:
		return "○ EXITED"
	case domain.StatusPaused:
		return "⏸ PAUSED"
	case domain.StatusCreated:
		return "○ CREATED"
	default:
		return "? UNKNOWN"
	}
}

// statusBadgeStyled produces colored badge text for non-selected rows.
func statusBadgeStyled(c domain.Container) string {
	switch c.Status {
	case domain.StatusRunning:
		return lipgloss.NewStyle().Foreground(ui.ColorSuccess).Bold(true).Render("● RUNNING")
	case domain.StatusExited:
		return ui.SubtleStyle.Render("○ EXITED")
	case domain.StatusPaused:
		return lipgloss.NewStyle().Foreground(ui.ColorWarning).Bold(true).Render("⏸ PAUSED")
	case domain.StatusCreated:
		return ui.SubtleStyle.Render("○ CREATED")
	default:
		return ui.SubtleStyle.Render("? UNKNOWN")
	}
}

// renderDetailPanel renders the detail/log panel (right side).
func (m *Model) renderDetailPanel(width, height int) string {
	c := m.selectedContainer()
	if c == nil {
		content := ui.SubtleStyle.Render(" No container selected")
		return m.renderTitledPanel(width, height, "Logs", content, false, lipgloss.Color(ui.ColorBorder))
	}

	var title string
	var content string

	if m.activePanel == LogsPanel {
		title = "Logs — " + c.Name
		if m.logFollow {
			title += " [follow]"
		}
		content = m.logViewport.View()
		if len(m.logLines) == 0 {
			if m.stream != nil {
				content = ui.SpinnerStyle.Render(m.spinnerView() + " Reading logs…")
			} else {
				content = ui.SubtleStyle.Render(" No log output received yet.")
			}
		}
	} else {
		// ListPanel is active: show container detail card + recent logs preview
		title = "Container — " + c.Name

		var cardLines []string
		cardLines = append(cardLines, fmt.Sprintf("  %-10s %s", "NAME:", ui.ValueStyle.Render(c.Name)))
		cardLines = append(cardLines, fmt.Sprintf("  %-10s %s", "IMAGE:", ui.ValueStyle.Render(c.Image)))

		statusStr := statusBadgeStyled(c.Container)
		if c.RunningFor != "" {
			statusStr += ui.SubtleStyle.Render(" (" + c.RunningFor + ")")
		}
		cardLines = append(cardLines, fmt.Sprintf("  %-10s %s", "STATUS:", statusStr))
		cardLines = append(cardLines, fmt.Sprintf("  %-10s %s", "ENGINE:", ui.SubtleStyle.Render(string(c.Engine))))
		if c.ID != "" {
			shortID := c.ID
			if len(shortID) > 12 {
				shortID = shortID[:12]
			}
			cardLines = append(cardLines, fmt.Sprintf("  %-10s %s", "ID:", ui.SubtleStyle.Render(shortID)))
		}

		cardLines = append(cardLines, "")
		cardLines = append(cardLines, ui.SubtleStyle.Render("  " + strings.Repeat("─", width-6)))
		cardLines = append(cardLines, ui.LabelStyle.Render("  ACTIONS & LOGS:"))
		cardLines = append(cardLines, ui.SubtleStyle.Render("   • Press Enter to open live log stream"))
		if c.IsRunning() {
			cardLines = append(cardLines, ui.SubtleStyle.Render("   • Press s to stop container"))
		} else {
			cardLines = append(cardLines, ui.SubtleStyle.Render("   • Press s to start container"))
		}
		cardLines = append(cardLines, ui.SubtleStyle.Render("   • Press r to restart container"))

		if len(m.logLines) > 0 {
			cardLines = append(cardLines, "")
			cardLines = append(cardLines, ui.SubtleStyle.Render("  RECENT LOGS:"))
			start := len(m.logLines) - 5
			if start < 0 {
				start = 0
			}
			for _, line := range m.logLines[start:] {
				cardLines = append(cardLines, ui.SubtleStyle.Render("   "+line))
			}
		}

		content = strings.Join(cardLines, "\n")
	}

	accent := lipgloss.Color(ui.ColorBox)
	return m.renderTitledPanel(width, height, title, content, m.activePanel == LogsPanel, accent)
}

func truncateRunes(s string, maxLen int) string {
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return "…"
	}
	return string(r[:maxLen-1]) + "…"
}
