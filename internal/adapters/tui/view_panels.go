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
	borderColor := ui.ColorBorder
	if active {
		borderColor = accent
	}

	border := lipgloss.RoundedBorder()
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	if active {
		borderStyle = borderStyle.Bold(true)
	}

	// Truncate title if needed.
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

	// Build top border with embedded title: ╭─[Title]──────╮
	repeatCount := width - lipgloss.Width("─["+title+"]─") - 2
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

// renderContainerListContent renders the rows inside the container list viewport.
// Layout: ▌ prefix | status icon | name (branch-style) | image | uptime/status right-aligned
func (m *Model) renderContainerListContent() string {
	if m.loading {
		return ui.SpinnerStyle.Render(m.spinnerView() + " Loading containers…")
	}
	if len(m.containers) == 0 {
		return ui.SubtleStyle.Render("  No containers found.\n  Make sure Docker or Podman is running.")
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

// renderContainerRow renders a single container row — monogit's renderRepoLine style.
func (m *Model) renderContainerRow(index int, c containerItem, maxWidth int) string {
	selected := index == m.cursor

	// Background style for selected rows.
	var bgStyle lipgloss.Style
	if selected {
		bgStyle = lipgloss.NewStyle().Background(ui.ColorHighlight)
	}

	// Left prefix: ▌ (cursor) or two spaces.
	var prefix string
	if selected {
		prefix = bgStyle.Foreground(ui.ColorBg).Render("▌ ")
	} else {
		prefix = "  "
	}

	// Status icon + indicator string (right-aligned cluster).
	statusIcon, statusStyle := containerStatusDisplay(c.Container)
	if c.loading {
		statusIcon = m.spinnerView()
		statusStyle = ui.SpinnerStyle
	}

	// Build indicator cluster (right side).
	var indicators []string
	if selected {
		indicators = append(indicators, bgStyle.Foreground(ui.ColorBg).Bold(true).Render(statusIcon))
	} else {
		indicators = append(indicators, statusStyle.Render(statusIcon))
	}

	// Engine tag (subtle, right cluster).
	engineTag := string(c.Engine)
	if selected {
		indicators = append(indicators, bgStyle.Foreground(ui.ColorBg).Render(engineTag))
	} else {
		indicators = append(indicators, ui.SubtleStyle.Render(engineTag))
	}

	indicatorStr := strings.Join(indicators, " ")

	// Available width for name + image.
	prefixWidth := lipgloss.Width(prefix)
	indicatorWidth := lipgloss.Width(indicatorStr)
	availForText := maxWidth - prefixWidth - indicatorWidth - 1
	if availForText < 5 {
		availForText = 5
	}

	// Container name (primary, left).
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

		// Image name (secondary, branch-style).
		availForImage := availForText - nameWidth - 1
		if availForImage >= 4 {
			maxImageLen := availForImage - 3
			img := c.Image
			if lipgloss.Width(img) > maxImageLen {
				img = truncateRunes(img, maxImageLen)
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

	// Gap between name and indicators.
	leftWidth := lipgloss.Width(leftContent)
	gapLen := maxWidth - leftWidth - indicatorWidth
	if gapLen < 1 {
		gapLen = 1
	}
	gap := strings.Repeat(" ", gapLen)
	if selected {
		gap = bgStyle.Render(gap)
	}

	row := leftContent + gap + indicatorStr

	// Pad to full width for uniform selected background.
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

// containerStatusDisplay returns icon + style for a container's status.
func containerStatusDisplay(c domain.Container) (icon string, style lipgloss.Style) {
	switch c.Status {
	case domain.StatusRunning:
		return "▶", ui.RunningStyle
	case domain.StatusExited:
		return "■", ui.StoppedStyle
	case domain.StatusPaused:
		return "⏸", ui.WarningStyle
	case domain.StatusCreated:
		return "○", ui.SubtleStyle
	default:
		return "?", ui.SubtleStyle
	}
}

// truncateRunes truncates s to maxLen runes, adding "…" if needed.
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

// renderLogContent renders the log panel content with ANSI stripping.
func renderLogContent(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n")
}

// getPanelTitle returns a panel title with optional number prefix — unused for now
// but kept for future panel-number feature.
func getPanelTitle(title string) string {
	return title
}

// statusBadge renders a colored status badge for a container.
func statusBadge(c domain.Container, selected bool) string {
	var bg lipgloss.Style
	if selected {
		bg = lipgloss.NewStyle().Background(ui.ColorHighlight)
	}

	switch c.Status {
	case domain.StatusRunning:
		s := lipgloss.NewStyle().Foreground(ui.ColorSuccess).Bold(true)
		if selected {
			s = bg.Foreground(ui.ColorBg).Bold(true)
		}
		return s.Render("RUN")
	case domain.StatusExited:
		s := lipgloss.NewStyle().Foreground(ui.ColorSubtle)
		if selected {
			s = bg.Foreground(ui.ColorBg)
		}
		return s.Render("EXT")
	default:
		s := ui.SubtleStyle
		if selected {
			s = bg.Foreground(ui.ColorBg)
		}
		return s.Render(fmt.Sprintf("%-3s", strings.ToUpper(string(c.Status))[:3]))
	}
}
