package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoOliveira889/monobox/internal/domain"
	"github.com/JoaoOliveira889/monobox/internal/pkg/ui"
)

// renderTitledPanel renders a panel with a decorative titled border (monogit style).
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
	repeatCount := width - lipgloss.Width(titleText) - 2
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

// renderContainerListContent renders the rows inside the container list panel.
func (m *Model) renderContainerListContent() string {
	if m.loading {
		return ui.SpinnerStyle.Render(m.spinnerView() + " Loading containers…")
	}
	if len(m.containers) == 0 {
		return ui.SubtleStyle.Render("No containers found.")
	}

	// Calculate column widths.
	maxNameLen := 4   // "Name"
	maxImageLen := 5  // "Image"
	maxUptimeLen := 6 // "Uptime"
	for _, c := range m.containers {
		if l := len(c.Name); l > maxNameLen {
			maxNameLen = l
		}
		if l := len(c.Image); l > maxImageLen {
			maxImageLen = l
		}
		if l := len(c.RunningFor); l > maxUptimeLen {
			maxUptimeLen = l
		}
	}
	// Cap widths to avoid overflow.
	if maxNameLen > 30 {
		maxNameLen = 30
	}
	if maxImageLen > 35 {
		maxImageLen = 35
	}
	if maxUptimeLen > 20 {
		maxUptimeLen = 20
	}

	// Header row.
	header := ui.SubtleStyle.Render(
		fmt.Sprintf("  %-*s  %-*s  %-*s  %s",
			maxNameLen, "NAME",
			maxImageLen, "IMAGE",
			maxUptimeLen, "UPTIME",
			"STATUS",
		),
	)

	var rows []string
	rows = append(rows, header)

	for i, c := range m.containers {
		selected := i == m.cursor

		// Status icon + color.
		statusIcon, statusStyle := containerStatusDisplay(c.Container)
		if c.loading {
			statusIcon = m.spinnerView()
			statusStyle = ui.SpinnerStyle
		}

		name := truncate(c.Name, maxNameLen)
		image := truncate(c.Image, maxImageLen)
		uptime := truncate(c.RunningFor, maxUptimeLen)

		engineTag := ui.SubtleStyle.Render("[" + string(c.Engine) + "]")
		row := fmt.Sprintf("  %-*s  %-*s  %-*s  %s %s",
			maxNameLen, name,
			maxImageLen, image,
			maxUptimeLen, uptime,
			statusStyle.Render(statusIcon),
			engineTag,
		)

		if selected {
			row = ui.PointerStyle.Render("▶") + row[1:]
			row = ui.SelectedItemStyle.Render(row)
		} else {
			row = ui.NormalItemStyle.Render(row)
		}
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

func containerStatusDisplay(c domain.Container) (icon string, style lipgloss.Style) {
	switch c.Status {
	case domain.StatusRunning:
		return ui.IconRunning, ui.RunningStyle
	case domain.StatusExited, domain.StatusCreated:
		return ui.IconStopped, ui.StoppedStyle
	case domain.StatusPaused:
		return "⏸", ui.WarningStyle
	default:
		return "?", ui.SubtleStyle
	}
}

func truncate(s string, maxLen int) string {
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	return string(r[:maxLen-1]) + "…"
}
