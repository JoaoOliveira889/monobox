package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
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

	// Exact top line width calculation:
	// TopLeft (1) + "─[" (2) + title (len) + "]" (1) + Top (repeatCount) + TopRight (1) = width
	// => repeatCount = width - len(title) - 5
	titleLen := lipgloss.Width(title)
	repeatCount := width - titleLen - 5
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

	title := "1 Containers"
	accent := lipgloss.Color(ui.ColorMono)
	return m.renderTitledPanel(width, height, title, content, m.activePanel == ListPanel, accent)
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
		vpWidth = m.leftPanelWidth() - 3
	}

	var rows []string
	for i, c := range m.containers {
		rows = append(rows, m.renderContainerRow(i, c, vpWidth))
	}
	return strings.Join(rows, "\n")
}

// renderContainerRow renders a single container row with tech icon + container name on left
// and status badge on right (no parenthesized image name).
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

	// Tech / engine icon (e.g. 🐘 postgres, ⚡ redis, 🔒 fga, ⚙ grpc/api, 🐳 docker)
	icon := containerIcon(c.Container)
	iconStr := icon + " "

	// Right-aligned status badge
	var statusBadge string
	if selected {
		statusBadge = bgStyle.Foreground(ui.ColorBg).Bold(true).Render(statusBadgeText(c.Container, maxWidth))
	} else {
		statusBadge = statusBadgeStyled(c.Container, maxWidth)
	}

	prefixWidth := lipgloss.Width(prefix)
	iconWidth := lipgloss.Width(iconStr)
	statusWidth := lipgloss.Width(statusBadge)

	availForName := maxWidth - prefixWidth - iconWidth - statusWidth - 1
	if availForName < 3 {
		availForName = 3
	}

	name := c.Name
	if lipgloss.Width(name) > availForName {
		name = truncateRunes(name, availForName)
	}

	var nameStr string
	if selected {
		nameStr = bgStyle.Foreground(ui.ColorBg).Bold(true).Render(name)
	} else {
		nameStr = lipgloss.NewStyle().Foreground(ui.ColorFg).Render(name)
	}

	var iconStyled string
	if selected {
		iconStyled = bgStyle.Foreground(ui.ColorBg).Render(iconStr)
	} else {
		iconStyled = iconStr
	}

	leftContent := prefix + iconStyled + nameStr
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
	} else if rowWidth > maxWidth {
		row = truncateRunes(row, maxWidth)
	}

	return row
}

// containerIcon determines a smart icon for the container based on its name/image/engine.
func containerIcon(c domain.Container) string {
	lower := strings.ToLower(c.Name + " " + c.Image)
	switch {
	case strings.Contains(lower, "postgres") || strings.Contains(lower, "pg"):
		return "🐘"
	case strings.Contains(lower, "redis"):
		return "⚡"
	case strings.Contains(lower, "mongo"):
		return "🍃"
	case strings.Contains(lower, "mysql") || strings.Contains(lower, "mariadb"):
		return "🐬"
	case strings.Contains(lower, "nginx") || strings.Contains(lower, "caddy") || strings.Contains(lower, "httpd") || strings.Contains(lower, "web"):
		return "🌐"
	case strings.Contains(lower, "grpc") || strings.Contains(lower, "api") || strings.Contains(lower, "service") || strings.Contains(lower, "node"):
		return "⚙"
	case strings.Contains(lower, "openfga") || strings.Contains(lower, "auth") || strings.Contains(lower, "keycloak"):
		return "🔒"
	case strings.Contains(lower, "queue") || strings.Contains(lower, "rabbitmq") || strings.Contains(lower, "kafka"):
		return "📬"
	default:
		if c.Engine == domain.EnginePodman {
			return "🦭"
		}
		return "🐳"
	}
}

// containerIconLabel returns an icon + text label for the detail card.
func containerIconLabel(c domain.Container) string {
	icon := containerIcon(c)
	lower := strings.ToLower(c.Name + " " + c.Image)
	switch {
	case strings.Contains(lower, "postgres") || strings.Contains(lower, "pg"):
		return icon + " Postgres"
	case strings.Contains(lower, "redis"):
		return icon + " Redis"
	case strings.Contains(lower, "mongo"):
		return icon + " MongoDB"
	case strings.Contains(lower, "mysql") || strings.Contains(lower, "mariadb"):
		return icon + " MySQL"
	case strings.Contains(lower, "nginx") || strings.Contains(lower, "caddy"):
		return icon + " Web Server"
	case strings.Contains(lower, "grpc") || strings.Contains(lower, "api") || strings.Contains(lower, "service"):
		return icon + " gRPC / API Service"
	case strings.Contains(lower, "openfga") || strings.Contains(lower, "auth"):
		return icon + " OpenFGA Auth"
	default:
		if c.Engine == domain.EnginePodman {
			return icon + " Podman Container"
		}
		return icon + " Docker Container"
	}
}

func statusBadgeText(c domain.Container, maxWidth int) string {
	if maxWidth < 25 {
		switch c.Status {
		case domain.StatusRunning:
			return "● RUN"
		case domain.StatusExited:
			return "○ STOPPED"
		case domain.StatusPaused:
			return "⏸ PAUS"
		default:
			return "? UNK"
		}
	}
	switch c.Status {
	case domain.StatusRunning:
		return "● RUNNING"
	case domain.StatusExited:
		return "○ STOPPED"
	case domain.StatusPaused:
		return "⏸ PAUSED"
	case domain.StatusCreated:
		return "○ CREATED"
	default:
		return "? UNKNOWN"
	}
}

func statusBadgeStyled(c domain.Container, maxWidth int) string {
	text := statusBadgeText(c, maxWidth)
	switch c.Status {
	case domain.StatusRunning:
		return lipgloss.NewStyle().Foreground(ui.ColorSuccess).Bold(true).Render(text)
	case domain.StatusExited:
		return ui.SubtleStyle.Render(text)
	case domain.StatusPaused:
		return lipgloss.NewStyle().Foreground(ui.ColorWarning).Bold(true).Render(text)
	case domain.StatusCreated:
		return ui.SubtleStyle.Render(text)
	default:
		return ui.SubtleStyle.Render(text)
	}
}

// renderDetailPanel renders the detail/log panel (right side).
func (m *Model) renderDetailPanel(width, height int) string {
	c := m.selectedContainer()
	if c == nil {
		content := ui.SubtleStyle.Render(" No container selected")
		return m.renderTitledPanel(width, height, "2 Logs", content, false, lipgloss.Color(ui.ColorBorder))
	}

	var title string
	var content string

	if m.activePanel == LogsPanel {
		title = "2 Logs — " + c.Name
		if m.logFollow {
			title += " [follow]"
		}
		content = renderViewportWithScrollbar(m.logViewport, true)
		if len(m.logLines) == 0 {
			if m.stream != nil {
				content = ui.SpinnerStyle.Render(m.spinnerView() + " Reading logs…")
			} else {
				content = ui.SubtleStyle.Render(" No log output received yet.")
			}
		}
	} else {
		// ListPanel is active: show container detail card + recent logs preview
		title = "2 Container — " + c.Name

		var cardLines []string
		cardLines = append(cardLines, fmt.Sprintf("  %-10s %s", "NAME:", ui.ValueStyle.Render(c.Name)))
		cardLines = append(cardLines, fmt.Sprintf("  %-10s %s", "TYPE:", ui.ValueStyle.Render(containerIconLabel(c.Container))))
		cardLines = append(cardLines, fmt.Sprintf("  %-10s %s", "IMAGE:", ui.ValueStyle.Render(c.Image)))

		statusStr := statusBadgeStyled(c.Container, width)
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

		divWidth := width - 6
		if divWidth < 10 {
			divWidth = 10
		}
		cardLines = append(cardLines, "")
		cardLines = append(cardLines, ui.SubtleStyle.Render("  "+strings.Repeat("─", divWidth)))
		cardLines = append(cardLines, ui.LabelStyle.Render("  ACTIONS & LOGS:"))
		cardLines = append(cardLines, ui.SubtleStyle.Render("   • Press Enter / l / 2 to open live log stream"))
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
				if lipgloss.Width(line) > divWidth {
					line = truncateRunes(line, divWidth)
				}
				cardLines = append(cardLines, ui.SubtleStyle.Render("   "+line))
			}
		}

		content = strings.Join(cardLines, "\n")
	}

	accent := lipgloss.Color(ui.ColorBox)
	return m.renderTitledPanel(width, height, title, content, m.activePanel == LogsPanel, accent)
}

// renderViewportWithScrollbar renders viewport content with a vertical scrollbar
// indicator on the right edge when content overflows.
func renderViewportWithScrollbar(vp viewport.Model, active bool) string {
	content := vp.View()
	lines := strings.Split(content, "\n")
	height := len(lines)
	if height == 0 {
		return content
	}

	percent := vp.ScrollPercent()
	if percent < 0 {
		percent = 0
	} else if percent > 1.0 {
		percent = 1.0
	}

	thumbPos := int(percent * float64(height-1))
	if thumbPos >= height {
		thumbPos = height - 1
	}

	trackColor := ui.ColorBorder
	thumbColor := ui.ColorSubtle
	if active {
		thumbColor = ui.ColorHighlight
	}

	thumbStyle := lipgloss.NewStyle().Foreground(thumbColor).Bold(true)
	trackStyle := lipgloss.NewStyle().Foreground(trackColor)

	for i := 0; i < height; i++ {
		if i == thumbPos {
			lines[i] += " " + thumbStyle.Render("█")
		} else {
			lines[i] += " " + trackStyle.Render("│")
		}
	}

	return strings.Join(lines, "\n")
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
