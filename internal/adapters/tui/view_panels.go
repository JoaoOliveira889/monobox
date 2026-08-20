package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoOliveira889/monobox/internal/domain"
	"github.com/JoaoOliveira889/monobox/internal/pkg/ui"
)

const iconWidth = 3

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

	if innerWidth > 0 {
		content = truncateLineWidths(content, innerWidth)
	}
	content = clampLines(content, innerHeight)

	panelStyle := lipgloss.NewStyle().
		Border(border, false, true, true, true).
		BorderForeground(borderColor).
		Width(innerWidth).
		Height(innerHeight)

	panel := panelStyle.Render(content)
	return lipgloss.JoinVertical(lipgloss.Left, topLine, panel)
}

func (m *Model) renderRepoList(width, height int) string {
	filtered := m.FilteredContainers()
	content := m.listViewport.View()

	if m.loading {
		content = ui.SpinnerStyle.Render(m.spinnerView() + " Loading containers…")
	} else if len(m.containers) == 0 {
		content = ui.SubtleStyle.Render(" No containers found.\n Make sure Docker or Podman is running.")
	} else if len(filtered) == 0 {
		content = ui.SubtleStyle.Render(fmt.Sprintf(" No containers matching %q", m.filterQuery))
	}

	if m.filtering || m.filterQuery != "" {
		m.filterInput.Width = width - 4
		filterView := m.filterInput.View()
		if m.filtering {
			filterView = lipgloss.NewStyle().Foreground(ui.ColorHighlight).Bold(true).Render(filterView)
		} else {
			filterView = ui.SubtleStyle.Render(filterView)
		}
		content = lipgloss.JoinVertical(lipgloss.Left, "  "+filterView, content)
	}

	title := "1 Containers"
	if m.filterQuery != "" || m.filtering {
		title = fmt.Sprintf("1 Containers (%d/%d)", len(filtered), len(m.containers))
	}
	accent := lipgloss.Color(ui.ColorMono)
	return m.renderTitledPanel(width, height, title, content, m.activePanel == ListPanel, accent)
}

func (m *Model) renderContainerListContent() string {
	if m.loading {
		return ui.SpinnerStyle.Render(m.spinnerView() + " Loading containers…")
	}
	nodes := m.VisibleTreeNodes()
	if len(m.containers) == 0 {
		return ui.SubtleStyle.Render(" No containers found.\n Make sure Docker or Podman is running.")
	}
	if len(nodes) == 0 {
		return ui.SubtleStyle.Render(fmt.Sprintf(" No containers matching %q", m.filterQuery))
	}

	vpWidth := m.listViewport.Width
	if vpWidth < 10 {
		vpWidth = m.leftPanelWidth() - 3
	}

	var rows []string
	for i, node := range nodes {
		if node.Type == NodeProjectHeader {
			rows = append(rows, m.renderProjectHeaderRow(i, node, vpWidth))
		} else if node.Container != nil {
			rows = append(rows, m.renderContainerRow(i, node, vpWidth))
		}
	}
	return strings.Join(rows, "\n")
}

func (m *Model) renderProjectHeaderRow(index int, node TreeNode, maxWidth int) string {
	selected := index == m.cursor

	var prefix string
	if selected {
		prefix = "▍ "
	} else {
		prefix = "  "
	}

	toggleIcon := "[-] "
	if !node.Expanded {
		toggleIcon = "[+] "
	}
	if !selected {
		toggleIcon = lipgloss.NewStyle().Foreground(ui.ColorBox).Bold(true).Render(toggleIcon)
	}

	projName := node.ProjectName
	if !selected {
		projName = lipgloss.NewStyle().Foreground(ui.ColorFg).Bold(true).Render(projName)
	}

	badge := fmt.Sprintf("(%d containers)", node.TotalCount)
	if node.RunningCount > 0 {
		badge = fmt.Sprintf("(%d/%d running)", node.RunningCount, node.TotalCount)
	}
	if selected {
		badge = " " + badge
	} else {
		badge = ui.SubtleStyle.Render(" " + badge)
	}

	left := prefix + toggleIcon + projName + badge
	leftWidth := lipgloss.Width(left)

	if leftWidth < maxWidth {
		padding := strings.Repeat(" ", maxWidth-leftWidth)
		left += padding
	} else if leftWidth > maxWidth {
		left = truncateRunes(left, maxWidth)
	}
	if selected {
		return ui.SelectedItemStyle.Render(left)
	}
	return left
}

func (m *Model) renderContainerRow(index int, node TreeNode, maxWidth int) string {
	c := *node.Container
	selected := index == m.cursor

	var prefix string
	if selected {
		prefix = "▍ "
	} else {
		prefix = "  "
	}

	var treeBranch string
	if node.ProjectName != "" {
		if node.IsLastInGroup {
			treeBranch = "└─ "
		} else {
			treeBranch = "├─ "
		}
		if !selected {
			treeBranch = ui.SubtleStyle.Render(treeBranch)
		}
	}

	iconStr, _ := containerIconAndLabel(c.Container)
	if padding := iconWidth - lipgloss.Width(iconStr); padding > 0 {
		iconStr += strings.Repeat(" ", padding)
	}
	iconCellWidth := iconWidth

	var statusBadge string
	if selected {
		statusBadge = statusBadgeText(c, maxWidth)
	} else {
		statusBadge = statusBadgeStyled(c, maxWidth)
	}

	prefixWidth := lipgloss.Width(prefix)
	treeWidth := lipgloss.Width(treeBranch)
	iconCols := iconCellWidth
	statusWidth := lipgloss.Width(statusBadge)

	availForName := maxWidth - prefixWidth - treeWidth - iconCols - statusWidth - 1
	if availForName < 3 {
		availForName = 3
	}

	name := c.Name
	if lipgloss.Width(name) > availForName {
		name = truncateRunes(name, availForName)
	}

	var nameStr string
	if selected {
		nameStr = name
	} else {
		nameStr = lipgloss.NewStyle().Foreground(ui.ColorFg).Render(name)
	}

	leftContent := prefix + treeBranch + iconStr + " " + nameStr
	leftWidth := lipgloss.Width(leftContent)
	gapLen := maxWidth - leftWidth - statusWidth
	if gapLen < 1 {
		gapLen = 1
	}
	gap := strings.Repeat(" ", gapLen)

	row := leftContent + gap + statusBadge
	rowWidth := lipgloss.Width(row)
	if rowWidth < maxWidth {
		padding := strings.Repeat(" ", maxWidth-rowWidth)
		row += padding
	} else if rowWidth > maxWidth {
		row = truncateRunes(row, maxWidth)
	}
	if selected {
		return ui.SelectedItemStyle.Render(row)
	}

	return row
}

func containerIconAndLabel(c domain.Container) (icon, label string) {
	lower := strings.ToLower(c.Name + " " + c.Image)
	switch {
	case strings.Contains(lower, "postgres") || strings.Contains(lower, "postgre") || strings.Contains(lower, "pg"):
		return "🐘", "Postgres"
	case strings.Contains(lower, "redis"):
		return "⚡", "Redis"
	case strings.Contains(lower, "mongo"):
		return "🍃", "MongoDB"
	case strings.Contains(lower, "mysql") || strings.Contains(lower, "mariadb"):
		return "🐬", "MySQL"
	case strings.Contains(lower, "nginx") || strings.Contains(lower, "caddy") || strings.Contains(lower, "httpd") || strings.Contains(lower, "web"):
		return "🌐", "Web Server"
	case strings.Contains(lower, "grpc") || strings.Contains(lower, "api") || strings.Contains(lower, "service") || strings.Contains(lower, "node"):
		return "🔧", "gRPC / API Service"
	case strings.Contains(lower, "openfga") || strings.Contains(lower, "auth") || strings.Contains(lower, "keycloak"):
		return "🔐", "Auth Service"
	case strings.Contains(lower, "ministack") || strings.Contains(lower, "ministask") || strings.Contains(lower, "localstack") || strings.Contains(lower, "minio") || strings.Contains(lower, "aws"):
		return "☁️", "Cloud / LocalStack"
	case strings.Contains(lower, "queue") || strings.Contains(lower, "rabbitmq") || strings.Contains(lower, "kafka"):
		return "📬", "Message Queue"
	case strings.Contains(lower, "podman"):
		return "🦭", "Podman Container"
	case strings.Contains(lower, "docker"):
		return "🐳", "Docker Container"
	default:
		if c.Engine == domain.EnginePodman {
			return "🦭", "Podman Container"
		}
		return "🐳", "Docker Container"
	}
}

func statusBadgeText(c containerItem, maxWidth int) string {
	if c.starting {
		if maxWidth < 25 {
			return "⏳ STR..."
		}
		return "⏳ STARTING..."
	}
	if c.Status == domain.StatusRunning {
		switch c.Health {
		case domain.HealthHealthy:
			if maxWidth < 25 {
				return "● HLTH"
			}
			return "● HEALTHY"
		case domain.HealthUnhealthy:
			if maxWidth < 25 {
				return "✖ UNH"
			}
			return "✖ UNHEALTHY"
		case domain.HealthStarting:
			if maxWidth < 25 {
				return "⏳ STR..."
			}
			return "⏳ STARTING..."
		}
	}

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

func parseHostPorts(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	var result []string
	seen := make(map[string]bool)

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if idx := strings.Index(p, "->"); idx != -1 {
			hostPart := p[:idx]
			if colon := strings.LastIndex(hostPart, ":"); colon != -1 {
				port := hostPart[colon+1:]
				if port != "" && !seen[port] {
					seen[port] = true
					result = append(result, port)
				}
			}
		}
	}
	return result
}

func parseMemPercVal(memStr string) float64 {
	if idx := strings.Index(memStr, "("); idx != -1 {
		end := strings.Index(memStr[idx:], "%)")
		if end != -1 {
			sub := memStr[idx+1 : idx+end]
			val, _ := strconv.ParseFloat(strings.TrimSpace(sub), 64)
			return val
		}
	}
	return parseCPUVal(memStr)
}

func statusBadgeStyled(c containerItem, maxWidth int) string {
	text := statusBadgeText(c, maxWidth)
	if c.starting {
		return lipgloss.NewStyle().Foreground(ui.ColorWarning).Bold(true).Render(text)
	}
	if c.Status == domain.StatusRunning {
		cpuFloat := parseCPUVal(c.CPU)
		memPercFloat := parseMemPercVal(c.Mem)
		if cpuFloat >= 90.0 || memPercFloat >= 90.0 {
			return lipgloss.NewStyle().Foreground(ui.ColorError).Bold(true).Render("🔥 CRITICAL")
		}
		if cpuFloat >= 80.0 || memPercFloat >= 80.0 {
			return lipgloss.NewStyle().Foreground(ui.ColorWarning).Bold(true).Render("⚡ HIGH LOAD")
		}
		switch c.Health {
		case domain.HealthHealthy:
			return lipgloss.NewStyle().Foreground(ui.ColorSuccess).Bold(true).Render(text)
		case domain.HealthUnhealthy:
			return lipgloss.NewStyle().Foreground(ui.ColorError).Bold(true).Render(text)
		case domain.HealthStarting:
			return lipgloss.NewStyle().Foreground(ui.ColorWarning).Bold(true).Render(text)
		}
		return lipgloss.NewStyle().Foreground(ui.ColorSuccess).Bold(true).Render(text)
	}
	switch c.Status {
	case domain.StatusPaused:
		return lipgloss.NewStyle().Foreground(ui.ColorWarning).Bold(true).Render(text)
	default:
		return ui.SubtleStyle.Render(text)
	}
}

func (m *Model) renderDetailPanel(width, height int) string {
	if node := m.selectedNode(); node != nil && node.Type == NodeProjectHeader {
		title := "2 Project Stack — " + node.ProjectName
		var cardLines []string
		cardLines = append(cardLines, fmt.Sprintf("  %-12s %s", "PROJECT:", ui.ValueStyle.Render(node.ProjectName)))
		cardLines = append(cardLines, fmt.Sprintf("  %-12s %s", "TYPE:", ui.ValueStyle.Render("📦 Docker Compose Stack")))
		cardLines = append(cardLines, fmt.Sprintf("  %-12s %s", "CONTAINERS:", ui.ValueStyle.Render(fmt.Sprintf("%d (%d running)", node.TotalCount, node.RunningCount))))
		cardLines = append(cardLines, fmt.Sprintf("  %-12s %s", "ENGINE:", ui.SubtleStyle.Render(m.engine)))

		divWidth := width - 6
		if divWidth < 10 {
			divWidth = 10
		}
		cardLines = append(cardLines, "")
		cardLines = append(cardLines, ui.SubtleStyle.Render("  "+strings.Repeat("─", divWidth)))
		cardLines = append(cardLines, ui.LabelStyle.Render("  BATCH ACTIONS:"))
		cardLines = append(cardLines, ui.SubtleStyle.Render("   ▸ Press s to START/STOP all containers in stack"))
		cardLines = append(cardLines, ui.SubtleStyle.Render("   ▸ Press r to RESTART all containers in stack"))
		cardLines = append(cardLines, ui.SubtleStyle.Render("   ▸ Press d to REMOVE all containers in stack"))
		cardLines = append(cardLines, ui.SubtleStyle.Render("   ▸ Press Space / Enter to toggle collapse/expand"))

		cardLines = append(cardLines, "")
		cardLines = append(cardLines, ui.SubtleStyle.Render("  "+strings.Repeat("─", divWidth)))
		cardLines = append(cardLines, ui.LabelStyle.Render("  STACK CONTAINERS:"))

		for _, item := range m.containers {
			if item.ComposeProject == node.ProjectName {
				statusBadge := statusBadgeStyled(item, width)
				cardLines = append(cardLines, fmt.Sprintf("   • %-20s %s", item.Name, statusBadge))
			}
		}

		content := strings.Join(cardLines, "\n")
		accent := lipgloss.Color(ui.ColorBox)
		return m.renderTitledPanel(width, height, title, content, m.activePanel == LogsPanel, accent)
	}

	c := m.selectedContainer()
	if c == nil {
		content := ui.SubtleStyle.Render(" No container selected")
		return m.renderTitledPanel(width, height, "2 Logs", content, false, lipgloss.Color(ui.ColorBorder))
	}

	var title string
	var content string

	if m.activePanel == LogsPanel {
		followStatus := "[follow: ON]"
		if !m.logFollow {
			followStatus = "[follow: OFF]"
		}
		tsStatus := ""
		if m.showTimestamps {
			tsStatus = " [ts: ON]"
		}
		liveStatus := ""
		if m.stream != nil {
			liveStatus = " [live]"
		}
		title = fmt.Sprintf("2 Logs — %s %s%s%s", c.Name, followStatus, tsStatus, liveStatus)
		vpView := renderViewportWithScrollbar(m.logViewport, true)
		if len(m.logLines) == 0 {
			if m.stream != nil {
				vpView = ui.SubtleStyle.Render(" Waiting for container output…\n Live follow is enabled.")
			} else {
				vpView = ui.SubtleStyle.Render(" No log output received yet.")
			}
		}

		if m.logSearching || m.logSearchQuery != "" {
			m.logSearchInput.Width = width - 4
			if m.logSearchInput.Width < 10 {
				m.logSearchInput.Width = 10
			}
			searchView := m.logSearchInput.View()
			if m.logSearching {
				searchView = lipgloss.NewStyle().Foreground(ui.ColorHighlight).Bold(true).Render(searchView)
			} else {
				searchView = ui.SubtleStyle.Render(searchView)
			}
			content = lipgloss.JoinVertical(lipgloss.Left, "  "+searchView, vpView)
		} else {
			content = vpView
		}
	} else {
		title = "2 Container — " + c.Name

		var cardLines []string
		cardLines = append(cardLines, fmt.Sprintf("  %-10s %s", "NAME:", ui.ValueStyle.Render(c.Name)))
		cardLines = append(cardLines, fmt.Sprintf("  %-10s %s", "TYPE:", ui.ValueStyle.Render(func() string {
			icon, label := containerIconAndLabel(c.Container)
			return icon + "  " + label
		}())))

		cardLines = append(cardLines, fmt.Sprintf("  %-10s %s", "IMAGE:", ui.ValueStyle.Render(c.Image)))

		cleanPorts, officialPort := formatCleanPorts(c.Ports)
		portsMax := width - 14
		if portsMax < 10 {
			portsMax = 10
		}
		if lipgloss.Width(cleanPorts) > portsMax {
			cleanPorts = truncateRunes(cleanPorts, portsMax)
		}
		cardLines = append(cardLines, fmt.Sprintf("  %-10s %s", "PORTS:", ui.ValueStyle.Render(cleanPorts)))
		if officialPort != "" {
			hostUrl := fmt.Sprintf("http://localhost:%s", officialPort)
			styledUrl := lipgloss.NewStyle().Foreground(ui.ColorCyan).Bold(true).Render(hostUrl)
			hint := ui.SubtleStyle.Render(" (press 'o' to open)")
			cardLines = append(cardLines, fmt.Sprintf("  %-10s %s%s", "URL:", styledUrl, hint))
		}

		if details := m.inspectDetailsCache[c.ID]; details != nil {
			if len(details.Networks) > 0 {
				var netStrs []string
				for _, n := range details.Networks {
					if n.IPAddress != "" {
						netStrs = append(netStrs, fmt.Sprintf("%s (%s)", n.Name, n.IPAddress))
					} else {
						netStrs = append(netStrs, n.Name)
					}
				}
				cardLines = append(cardLines, fmt.Sprintf("  %-10s %s", "NETWORKS:", ui.ValueStyle.Render(strings.Join(netStrs, ", "))))
			}

			if len(details.Mounts) > 0 {
				var mountStrs []string
				for _, mnt := range details.Mounts {
					src := mnt.Source
					if len(src) > 25 {
						src = "…" + src[len(src)-24:]
					}
					mountStrs = append(mountStrs, fmt.Sprintf("%s ➔ %s (%s)", src, mnt.Destination, mnt.Type))
				}
				mountVal := strings.Join(mountStrs, ", ")
				cardLines = append(cardLines, fmt.Sprintf("  %-10s %s", "VOLUMES:", ui.ValueStyle.Render(mountVal)))
			}

			if len(details.Env) > 0 {
				envHint := ui.SubtleStyle.Render(" (press 'E' to view)")
				cardLines = append(cardLines, fmt.Sprintf("  %-10s %s%s", "ENV:", ui.ValueStyle.Render(fmt.Sprintf("%d variables", len(details.Env))), envHint))
			}

			if details.Health != nil {
				hBadge := string(details.Health.Status)
				if details.Health.Status == domain.HealthHealthy {
					hBadge = ui.StatusSuccessStyle.Render("● HEALTHY")
				} else if details.Health.Status == domain.HealthUnhealthy {
					hBadge = ui.StatusErrorStyle.Render(fmt.Sprintf("✖ UNHEALTHY (Failing streak: %d)", details.Health.FailingStreak))
				} else if details.Health.Status == domain.HealthStarting {
					hBadge = ui.StatusWarningStyle.Render("⏳ STARTING")
				}
				hHint := ui.SubtleStyle.Render(" (press 'H' for logs)")
				cardLines = append(cardLines, fmt.Sprintf("  %-10s %s%s", "HEALTH:", hBadge, hHint))
			}
		}

		statusStr := statusBadgeStyled(*c, width)
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
		if c.IsRunning() {
			cpuVal := c.CPU
			if cpuVal == "" {
				cpuVal = "0.0%"
			}
			memVal := c.Mem
			if memVal == "" {
				memVal = "N/A"
			}

			cpuFloat := parseCPUVal(cpuVal)
			memPercFloat := parseMemPercVal(memVal)

			var cpuStyled string
			if cpuFloat >= 90.0 {
				cpuStyled = lipgloss.NewStyle().Foreground(ui.ColorError).Bold(true).Render(fmt.Sprintf("🔥 %s (CRITICAL)", cpuVal))
			} else if cpuFloat >= 80.0 {
				cpuStyled = lipgloss.NewStyle().Foreground(ui.ColorWarning).Bold(true).Render(fmt.Sprintf("⚡ %s (HIGH CPU)", cpuVal))
			} else {
				cpuStyled = ui.ValueStyle.Render(cpuVal)
			}

			var memStyled string
			if memPercFloat >= 90.0 {
				memStyled = lipgloss.NewStyle().Foreground(ui.ColorError).Bold(true).Render(fmt.Sprintf("🔥 %s (CRITICAL)", memVal))
			} else if memPercFloat >= 80.0 {
				memStyled = lipgloss.NewStyle().Foreground(ui.ColorWarning).Bold(true).Render(fmt.Sprintf("⚡ %s (HIGH MEMORY)", memVal))
			} else {
				memStyled = ui.ValueStyle.Render(memVal)
			}

			key := c.ID
			if key == "" {
				key = c.Name
			}
			cpuHist, memHist := m.GetStatsHistory(key)
			cpuSpark := ui.RenderSparkline(cpuHist, 12)
			memSpark := ui.RenderSparkline(memHist, 12)

			cpuLine := fmt.Sprintf("  %-10s %s", "CPU:", cpuStyled)
			if cpuSpark != "" {
				cpuLine += "  " + ui.SubtleStyle.Render(cpuSpark)
			}
			memLine := fmt.Sprintf("  %-10s %s", "MEMORY:", memStyled)
			if memSpark != "" {
				memLine += "  " + ui.SubtleStyle.Render(memSpark)
			}

			cardLines = append(cardLines, cpuLine)
			cardLines = append(cardLines, memLine)
		}

		divWidth := width - 6
		if divWidth < 10 {
			divWidth = 10
		}
		cardLines = append(cardLines, "")
		cardLines = append(cardLines, ui.SubtleStyle.Render("  "+strings.Repeat("─", divWidth)))
		cardLines = append(cardLines, ui.LabelStyle.Render("  QUICK ACTIONS"))
		shortcuts := []string{
			renderDetailShortcut("enter", "logs"),
			renderDetailShortcut("i", "inspect"),
			renderDetailShortcut("r", "restart"),
		}
		if c.IsRunning() {
			shortcuts = append(shortcuts,
				renderDetailShortcut("e", "shell"),
				renderDetailShortcut("E", "environment"),
				renderDetailShortcut("H", "health logs"),
				renderDetailShortcut("s", "stop"),
				renderDetailShortcut("p", "pause"),
			)
		} else if c.Status == domain.StatusPaused {
			shortcuts = append(shortcuts, renderDetailShortcut("p", "unpause"))
		} else {
			shortcuts = append(shortcuts, renderDetailShortcut("s", "start"))
		}
		if extractHostPort(c.Ports) != "" {
			shortcuts = append(shortcuts, renderDetailShortcut("o", "open"))
		}
		shortcuts = append(shortcuts,
			renderDetailShortcut("d", "remove"),
			renderDetailShortcut("?", "all shortcuts"),
		)
		for _, line := range wrapDetailShortcuts(shortcuts, divWidth) {
			cardLines = append(cardLines, "  "+line)
		}

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

func renderDetailShortcut(key, label string) string {
	return ui.FooterKeyStyle.Render("["+key+"]") + " " + ui.SubtleStyle.Render(label)
}

func wrapDetailShortcuts(shortcuts []string, maxWidth int) []string {
	const separator = "  "
	var lines []string
	current := ""

	for _, shortcut := range shortcuts {
		candidate := shortcut
		if current != "" {
			candidate = current + separator + shortcut
		}
		if current != "" && lipgloss.Width(candidate) > maxWidth {
			lines = append(lines, current)
			current = shortcut
			continue
		}
		current = candidate
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

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

func clampLines(s string, maxLines int) string {
	if maxLines <= 0 {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}
	return strings.Join(lines[:maxLines], "\n")
}

func truncateLineWidths(s string, maxWidth int) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if lipgloss.Width(line) > maxWidth {
			lines[i] = truncateRunes(line, maxWidth-1) + "…"
		}
	}
	return strings.Join(lines, "\n")
}

type portBinding struct {
	hostPort      string
	containerPort string
	proto         string
}

func parsePortBindings(raw string) []portBinding {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	var bindings []portBinding
	seen := make(map[string]bool)

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		var hostPort, containerPortProto string
		if strings.Contains(p, "->") {
			sp := strings.SplitN(p, "->", 2)
			hostPart := strings.TrimSpace(sp[0])
			containerPortProto = strings.TrimSpace(sp[1])

			if idx := strings.LastIndex(hostPart, ":"); idx != -1 {
				hostPort = hostPart[idx+1:]
			} else {
				hostPort = hostPart
			}
		} else {
			containerPortProto = p
		}

		var containerPort, proto string
		if strings.Contains(containerPortProto, "/") {
			cp := strings.SplitN(containerPortProto, "/", 2)
			containerPort = cp[0]
			proto = cp[1]
		} else {
			containerPort = containerPortProto
			proto = "tcp"
		}

		key := hostPort + ":" + containerPort + "/" + proto
		if !seen[key] {
			seen[key] = true
			bindings = append(bindings, portBinding{
				hostPort:      hostPort,
				containerPort: containerPort,
				proto:         proto,
			})
		}
	}
	return bindings
}

func formatCleanPorts(raw string) (cleanPorts string, officialHostPort string) {
	bindings := parsePortBindings(raw)
	if len(bindings) == 0 {
		return "none", ""
	}

	var items []string
	for _, b := range bindings {
		if b.hostPort != "" {
			if officialHostPort == "" {
				officialHostPort = b.hostPort
			}
			items = append(items, fmt.Sprintf("%s ➔ %s/%s", b.hostPort, b.containerPort, b.proto))
		} else {
			items = append(items, fmt.Sprintf("%s/%s", b.containerPort, b.proto))
		}
	}

	return strings.Join(items, ", "), officialHostPort
}
