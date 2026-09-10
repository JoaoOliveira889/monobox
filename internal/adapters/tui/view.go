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

func (m *Model) View() string {
	if m.quitting {
		return ""
	}
	if m.showSplash {
		return m.renderSplash()
	}
	if m.confirmPortConflict {
		return m.renderCenteredModal(m.renderPortConflictModal())
	}
	if m.confirmClearLogs {
		return m.renderCenteredModal(m.renderClearLogsConfirmation())
	}
	if m.confirmRemove {
		return m.renderCenteredModal(m.renderRemoveConfirmation())
	}
	if m.confirmBatchAction != "" {
		return m.renderCenteredModal(m.renderBatchConfirmation())
	}
	if m.showHelp {
		return m.renderHelpOverlay()
	}
	if m.showPruneModal {
		return m.renderCenteredModal(m.renderPruneModal())
	}
	if m.showSettingsModal {
		return m.renderCenteredModal(m.renderSettingsModal())
	}
	if m.showThemeMenu {
		return m.renderCenteredModal(m.renderThemeMenuModal())
	}
	if m.showGraphModal {
		return m.renderCenteredModal(m.renderGraphModal())
	}
	if m.showEnvModal {
		return m.renderCenteredModal(m.renderEnvModal())
	}
	if m.showHealthModal {
		return m.renderCenteredModal(m.renderHealthModal())
	}
	if m.showInspect {
		return m.renderCenteredModal(m.renderInspectModal())
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
	footer := m.renderFooter()
	body := m.renderBody()

	view := lipgloss.JoinVertical(lipgloss.Left,
		header,
		body,
		footer,
	)

	return lipgloss.NewStyle().
		MaxWidth(m.width).
		MaxHeight(m.height).
		Render(view)
}

func (m *Model) renderHeader() string {
	var brand string
	switch {
	case m.width < 30:
		brand = ui.BrandMonoStyle.Render("MB")
	default:
		brand = renderBrandWordmark(true)
	}

	engine := strings.ToLower(strings.TrimSpace(m.engine))
	if engine == "" {
		engine = "docker"
	}

	summary := m.renderHeaderStatsSummary()
	sep := ui.SubtleStyle.Render(" │ ")

	leftHeader := " " + brand
	if m.width >= 45 && engine != "" {
		leftHeader += sep + ui.SubtleStyle.Render(engine)
	}
	if summary != "" {
		leftHeader += sep + summary
	}
	if m.loading {
		if m.width >= 60 {
			leftHeader += "  " + ui.SpinnerStyle.Render(m.spinnerView()+" loading…")
		} else {
			leftHeader += "  " + ui.SpinnerStyle.Render(m.spinnerView())
		}
	}

	rightHeader := ""
	if m.statusMsg != "" {
		rightHeader = m.renderHeaderStatusBar() + " "
	}

	headerLine := renderHeaderBetween(leftHeader, rightHeader, m.width)
	border := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorBorder)).
		Render(strings.Repeat("─", m.width))

	return headerLine + "\n" + border
}

func renderHeaderBetween(left, right string, totalWidth int) string {
	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	if right == "" {
		return left
	}
	spaces := totalWidth - leftW - rightW
	if spaces < 2 {
		maxRightW := totalWidth - leftW - 2
		if maxRightW > 6 {
			right = truncateRunes(right, maxRightW)
			rightW = lipgloss.Width(right)
			spaces = totalWidth - leftW - rightW
		} else {
			return left
		}
	}
	return left + strings.Repeat(" ", spaces) + right
}

func (m *Model) renderHeaderStatusBar() string {
	if m.statusMsg == "" {
		return ""
	}
	msg := m.statusMsg
	switch {
	case strings.HasPrefix(m.statusMsg, "✓"):
		return ui.StatusSuccessStyle.Render(msg)
	case strings.HasPrefix(m.statusMsg, "✗") || strings.HasPrefix(m.statusMsg, "Error"):
		return ui.StatusErrorStyle.Render(msg)
	case strings.HasPrefix(m.statusMsg, "⚠") || strings.HasPrefix(m.statusMsg, "Warn"):
		return ui.StatusWarningStyle.Render(msg)
	default:
		return ui.StatusInfoStyle.Render(msg)
	}
}

func (m *Model) renderHeaderStatsSummary() string {
	total := len(m.containers)
	if total == 0 {
		return ui.SubtleStyle.Render("● No containers")
	}

	running, stopped := 0, 0
	highLoadCount := 0
	var totalCPU float64
	var totalMemBytes float64

	for _, c := range m.containers {
		if c.IsRunning() {
			running++
			cpuVal := parseCPUVal(c.CPU)
			totalCPU += cpuVal
			totalMemBytes += parseMemBytesVal(c.Mem)
			if cpuVal >= 80.0 {
				highLoadCount++
			}
		} else {
			stopped++
		}
	}

	var parts []string
	filtered := len(m.FilteredContainers())
	dot := ui.SubtleStyle.Render("●")

	if m.filterQuery != "" || m.filtering {
		parts = append(parts, fmt.Sprintf("%s %d/%d containers", dot, filtered, total))
	} else {
		parts = append(parts, fmt.Sprintf("%s %d containers", dot, total))
	}

	if running > 0 {
		parts = append(parts, ui.RunningStyle.Render(fmt.Sprintf("%d running", running)))
	}
	if stopped > 0 {
		parts = append(parts, ui.StoppedStyle.Render(fmt.Sprintf("%d stopped", stopped)))
	}
	if highLoadCount > 0 {
		parts = append(parts, ui.WarningStyle.Render(fmt.Sprintf("⚡ %d high load", highLoadCount)))
	}
	if running > 0 {
		parts = append(parts, lipgloss.NewStyle().Foreground(ui.ColorHighlight).Bold(true).Render(fmt.Sprintf("%.2f%% CPU", totalCPU)))
		parts = append(parts, lipgloss.NewStyle().Foreground(ui.ColorCyan).Bold(true).Render(formatBytes(totalMemBytes)))
	}

	sep := ui.SubtleStyle.Render(" · ")
	maxW := m.width - 1
	if maxW < 10 {
		maxW = 10
	}
	return fitInlineParts(parts, sep, maxW)
}

func fitInlineParts(parts []string, separator string, maxWidth int) string {
	for len(parts) > 0 {
		content := strings.Join(parts, separator)
		if lipgloss.Width(content) <= maxWidth {
			return content
		}
		parts = parts[:len(parts)-1]
	}
	return ""
}

func parseCPUVal(cpuStr string) float64 {
	cpuStr = strings.TrimSuffix(strings.TrimSpace(cpuStr), "%")
	val, _ := strconv.ParseFloat(cpuStr, 64)
	return val
}

func parseMemBytesVal(memStr string) float64 {
	if memStr == "" {
		return 0
	}
	parts := strings.Split(memStr, "/")
	used := strings.TrimSpace(parts[0])
	usedFields := strings.Fields(used)
	if len(usedFields) == 0 {
		return 0
	}
	return parseSizeToBytes(usedFields[0])
}

func parseSizeToBytes(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "N/A" {
		return 0
	}
	split := strings.IndexFunc(s, func(r rune) bool {
		return !((r >= '0' && r <= '9') || r == '.')
	})
	var numStr, unitStr string
	if split < 0 {
		numStr = s
	} else {
		numStr = s[:split]
		unitStr = strings.TrimSpace(s[split:])
	}
	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0
	}
	unitUpper := strings.ToUpper(unitStr)
	switch {
	case strings.HasPrefix(unitUpper, "G"):
		return val * 1024 * 1024 * 1024
	case strings.HasPrefix(unitUpper, "M"):
		return val * 1024 * 1024
	case strings.HasPrefix(unitUpper, "K"):
		return val * 1024
	case strings.HasPrefix(unitUpper, "T"):
		return val * 1024 * 1024 * 1024 * 1024
	default:
		return val
	}
}

func formatBytes(bytes float64) string {
	if bytes <= 0 {
		return "0 B"
	}
	const (
		KiB = 1024.0
		MiB = KiB * 1024
		GiB = MiB * 1024
		TiB = GiB * 1024
	)
	switch {
	case bytes >= TiB:
		return fmt.Sprintf("%.2f TiB", bytes/TiB)
	case bytes >= GiB:
		return fmt.Sprintf("%.2f GiB", bytes/GiB)
	case bytes >= MiB:
		return fmt.Sprintf("%.1f MiB", bytes/MiB)
	case bytes >= KiB:
		return fmt.Sprintf("%.1f KiB", bytes/KiB)
	default:
		return fmt.Sprintf("%.0f B", bytes)
	}
}

func (m *Model) renderFooter() string {
	sep := ui.SubtleStyle.Render(" · ")
	var parts []string

	if m.filtering {
		parts = []string{
			m.fmtKey("esc", "clear"),
			m.fmtKey("enter", "apply"),
			m.fmtKey("↑↓/jk", "nav"),
		}
	} else if m.logSearching {
		parts = []string{
			m.fmtKey("esc", "clear"),
			m.fmtKey("enter", "done"),
			m.fmtKey("n/N", "next/prev"),
		}
	} else if m.activePanel == LogsPanel {
		followAction := "follow:ON"
		if !m.logFollow {
			followAction = "follow:OFF"
		}
		parts = []string{
			m.fmtKey("jk", "scroll"),
			m.fmtKey("ctrl+d/u", "page"),
			m.fmtKey("/", "search"),
			m.fmtKey("f", followAction),
			m.fmtKey("c", "clear"),
			m.fmtKey("1/esc", "back"),
			m.fmtKey("q", "quit"),
		}
	} else if node := m.selectedNode(); node != nil && node.Type == NodeProjectHeader {
		parts = []string{
			m.fmtKey("jk", "nav"),
			m.fmtKey("ctrl+d/u", "page"),
			m.fmtKey("space/enter", "toggle"),
			m.fmtKey("s", "batch start/stop"),
			m.fmtKey("r", "batch restart"),
			m.fmtKey("d", "batch remove"),
			m.fmtKey("/", "filter"),
			m.fmtKey("q", "quit"),
		}
	} else {
		parts = []string{
			m.fmtKey("jk", "nav"),
			m.fmtKey("ctrl+d/u", "page"),
			m.fmtKey("enter/2", "logs"),
			m.fmtKey("s", "start/stop"),
			m.fmtKey("r", "restart"),
			m.fmtKey("/", "filter"),
			m.fmtKey("q", "quit"),
		}
	}

	return m.renderResponsiveFooter(parts, sep)
}

func (m *Model) renderResponsiveFooter(parts []string, sep string) string {
	version := ui.SubtleStyle.Render(fmt.Sprintf("monobox %s", Version))
	help := m.fmtKey("?", "help")
	fixedRight := help + "  " + version

	contentWidth := m.width - 2
	if contentWidth < 10 {
		contentWidth = 10
	}

	rendered := strings.Join(parts, sep)
	maxLeftWidth := contentWidth - lipgloss.Width(fixedRight) - 1

	for len(parts) > 0 && lipgloss.Width(rendered) > maxLeftWidth {
		parts = parts[:len(parts)-1]
		rendered = strings.Join(parts, sep)
	}

	left := rendered
	spacerLen := contentWidth - lipgloss.Width(left) - lipgloss.Width(fixedRight)
	if spacerLen < 0 {
		spacerLen = 0
	}
	spacer := strings.Repeat(" ", spacerLen)

	footerText := " " + left + spacer + fixedRight
	if footerWidth := lipgloss.Width(footerText); footerWidth < contentWidth+1 {
		footerText += strings.Repeat(" ", contentWidth+1-footerWidth)
	}

	return ui.FooterStyle.Padding(0, 0).Render(footerText)
}

func (m *Model) fmtKey(k, action string) string {
	return ui.FooterKeyStyle.Render(k) + " " + ui.FooterActionStyle.Render(action)
}

func (m *Model) renderBody() string {
	bodyHeight := m.panelHeight()

	leftWidth := m.leftPanelWidth()
	rightWidth := m.rightPanelWidth()

	left := m.renderContainerList(leftWidth, bodyHeight)
	right := m.renderDetailPanel(rightWidth, bodyHeight)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m *Model) renderCenteredModal(content string) string {
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		ui.ActivePanelStyle.Padding(1, 2).Render(content),
	)
}

func (m *Model) renderClearLogsConfirmation() string {
	return lipgloss.JoinHorizontal(lipgloss.Center,
		ui.ValueStyle.Render("Clear logs?"),
		"  ",
		m.fmtKey("y", "yes"),
		"  ",
		m.fmtKey("n", "no"),
	)
}

func (m *Model) renderRemoveConfirmation() string {
	c := m.selectedContainer()
	name := "container"
	if c != nil {
		name = c.Name
	}
	return lipgloss.JoinHorizontal(lipgloss.Center,
		ui.ErrorStyle.Render("Remove container "+name+"?"),
		"  ",
		m.fmtKey("y", "yes"),
		"  ",
		m.fmtKey("n", "cancel"),
	)
}

func (m *Model) renderBatchConfirmation() string {
	action := strings.TrimPrefix(m.confirmBatchAction, "batch_")
	projName := m.batchProjectName
	msg := fmt.Sprintf("%s all containers in '%s'?", strings.Title(action), projName)
	var actionStyle lipgloss.Style
	if action == "stop" || action == "remove" {
		actionStyle = ui.ErrorStyle
	} else {
		actionStyle = ui.ValueStyle
	}
	return lipgloss.JoinHorizontal(lipgloss.Center,
		actionStyle.Render(msg),
		"  ",
		m.fmtKey("y", "yes"),
		"  ",
		m.fmtKey("n", "cancel"),
	)
}

func (m *Model) renderPruneModal() string {
	allBox := "[ ] Remove all unused images (--all)"
	if m.pruneAll {
		allBox = "[x] Remove all unused images (--all)"
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		ui.WarningStyle.Render(fmt.Sprintf("⚠️  System Prune (%s)", m.engine)),
		"",
		ui.SubtleStyle.Render("Remove all stopped containers, unused networks, and dangling images."),
		"",
		ui.HighlightStyle.Render(allBox),
		"",
		lipgloss.JoinHorizontal(lipgloss.Center,
			m.fmtKey("y", "prune now"),
			"   ",
			m.fmtKey("a", "toggle --all"),
			"   ",
			m.fmtKey("n/esc", "cancel"),
		),
	)
}

func (m *Model) renderSettingsModal() string {
	title := ui.BrandTitleStyle.Render("SETTINGS") + " " + ui.SubtleStyle.Render("(S/esc to close)")

	type settingItem struct {
		name  string
		value string
		desc  string
	}

	items := []settingItem{
		{"Theme", m.cfg.Theme, "Active color palette"},
		{"Metrics Interval", fmt.Sprintf("%ds", m.cfg.MetricsInterval), "Container resource stats polling rate"},
		{"Log Line Limit", fmt.Sprintf("%d", m.cfg.LogLineLimit), "Max log lines kept in memory"},
		{"Log Tail Limit", fmt.Sprintf("%d", m.cfg.LogTailLimit), "Initial log lines to fetch"},
		{"Show Timestamps", fmt.Sprintf("%t", m.showTimestamps), "Show ISO timestamps on log streams"},
	}

	var rows []string
	for i, item := range items {
		cursor := "  "
		if i == m.settingsCursor {
			cursor = ui.CursorStyle.Render("▸ ")
		}

		val := item.value
		if m.settingsEditing && i == m.settingsCursor {
			val = m.settingsInput.View()
		}

		nameStyle := ui.LabelStyle
		valStyle := ui.ValueStyle
		if i == m.settingsCursor {
			nameStyle = nameStyle.Bold(true)
			valStyle = valStyle.Bold(true)
		}

		row := fmt.Sprintf("%s%-18s %-14s %s",
			cursor,
			nameStyle.Render(item.name),
			valStyle.Render(val),
			ui.SubtleStyle.Render(item.desc),
		)
		rows = append(rows, row)
	}

	footer := ui.SubtleStyle.Render("↑↓/jk: Navigate  •  Enter: Edit/Select  •  Esc: Cancel")
	if m.settingsEditing {
		footer = ui.HighlightStyle.Render("Editing value — Enter: Save  •  Esc: Cancel")
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		strings.Join(rows, "\n"),
		"",
		footer,
	)
}

func (m *Model) renderInspectModal() string {
	c := m.selectedContainer()
	name := "Container"
	if c != nil {
		name = c.Name
	}
	header := ui.LabelStyle.Render("Inspect — "+name) + " " + ui.SubtleStyle.Render("(esc/q/i to close)")
	body := renderViewportWithScrollbar(m.inspectViewport, true)
	return lipgloss.JoinVertical(lipgloss.Left, header, "", body)
}

func (m *Model) clampModalSize(marginW, maxW, marginH, maxH int) (int, int) {
	w := m.width - marginW
	if w > maxW {
		w = maxW
	}
	if w < 40 {
		w = 40
	}
	h := m.height - marginH
	if h > maxH {
		h = maxH
	}
	if h < 10 {
		h = 10
	}
	return w, h
}

func padRight(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func (m *Model) renderHelpOverlay() string {
	modalOuterWidth := m.width - 4
	if modalOuterWidth > 120 {
		modalOuterWidth = 120
	}
	if modalOuterWidth < 40 {
		modalOuterWidth = 40
	}

	innerWidth := modalOuterWidth - 6
	if innerWidth < 30 {
		innerWidth = 30
	}

	title := lipgloss.JoinHorizontal(lipgloss.Bottom,
		renderBrandWordmark(true),
		" ",
		ui.BrandTitleStyle.Render("SHORTCUTS"),
	)

	maxViewportHeight := m.height - 8
	if maxViewportHeight < 6 {
		maxViewportHeight = 6
	}
	vpHeight := maxViewportHeight
	vpWidth := innerWidth - 2
	if vpWidth < 20 {
		vpWidth = 20
	}

	if m.helpViewport.Width != vpWidth || m.helpViewport.Height != vpHeight {
		m.helpViewport = viewport.New(vpWidth, vpHeight)
	} else {
		m.helpViewport.Width = vpWidth
		m.helpViewport.Height = vpHeight
	}

	body := m.renderHelpMenu(vpWidth, vpHeight)
	contentHeight := lipgloss.Height(body)
	if contentHeight < vpHeight {
		vpHeight = contentHeight
		if vpHeight < 5 {
			vpHeight = 5
		}
		m.helpViewport.Height = vpHeight
	}
	m.helpViewport.SetContent(body)

	titleBar := lipgloss.NewStyle().Align(lipgloss.Center).Width(innerWidth).Render(title)
	footerHint := lipgloss.NewStyle().Align(lipgloss.Center).Width(innerWidth).Render(
		ui.SubtleStyle.Render("esc / ? close  •  ↑/↓ / jk scroll  •  g / G top / bottom"),
	)

	var helpBody string
	if contentHeight <= vpHeight {
		helpBody = m.helpViewport.View()
	} else {
		helpBody = renderViewportWithScrollbar(m.helpViewport, true)
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		titleBar,
		"",
		helpBody,
		"",
		footerHint,
	)

	panelStyle := ui.ActivePanelStyle.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ui.ColorCyan)).
		Width(innerWidth).
		Padding(1, 2)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panelStyle.Render(content))
}

func (m *Model) renderHelpMenu(width, height int) string {
	type helpEntry struct {
		key    string
		action string
	}
	type helpSection struct {
		heading string
		entries []helpEntry
	}
	type helpColumn struct {
		sections []helpSection
	}

	secNav := helpSection{
		heading: "CURSOR & PANELS",
		entries: []helpEntry{
			{key: "jk | ↑↓", action: "Move cursor / scroll"},
			{key: "ctrl+d/u", action: "Half-page scroll"},
			{key: "g | G", action: "Jump top / bottom"},
			{key: "1 | 2 | tab", action: "Switch panels"},
			{key: "hl | ←→", action: "Focus / toggle group"},
			{key: "< | >", action: "Resize panel ratio"},
		},
	}

	secActions := helpSection{
		heading: "CONTAINER ACTIONS",
		entries: []helpEntry{
			{key: "s", action: "Start / Stop"},
			{key: "r", action: "Restart container"},
			{key: "p", action: "Pause / Unpause"},
			{key: "d | delete", action: "Remove container"},
			{key: "e | x", action: "Exec shell (/bin/sh)"},
			{key: "i", action: "Inspect JSON"},
			{key: "o", action: "Open host port URL"},
			{key: "g", action: "Metrics graph"},
			{key: "E | H", action: "Env vars / health logs"},
			{key: "y | Y", action: "Copy ID / info"},
		},
	}

	secStacks := helpSection{
		heading: "COMPOSE STACKS",
		entries: []helpEntry{
			{key: "space | enter", action: "Expand / collapse"},
			{key: "u", action: "Compose Up -d"},
			{key: "D", action: "Compose Down"},
			{key: "s (group)", action: "Batch Start / Stop"},
			{key: "r (group)", action: "Batch Restart"},
			{key: "d (group)", action: "Batch Remove"},
		},
	}

	secLogs := helpSection{
		heading: "LOGS & SEARCH",
		entries: []helpEntry{
			{key: "f", action: "Toggle live follow"},
			{key: "t", action: "Toggle timestamps"},
			{key: "c | ctrl+l", action: "Clear container logs"},
			{key: "/", action: "Search / filter"},
			{key: "n | N", action: "Next / prev match"},
			{key: "!", action: "Cycle severity filter"},
			{key: "ctrl+r", action: "Toggle regex search"},
			{key: "s | ctrl+s", action: "Export logs to file"},
		},
	}

	secSystem := helpSection{
		heading: "SYSTEM & MODALS",
		entries: []helpEntry{
			{key: "S", action: "Settings modal"},
			{key: "P", action: "System prune modal"},
			{key: "T", action: "Theme menu"},
			{key: "m", action: "Mask secrets (Env)"},
			{key: "? | ctrl+p", action: "Toggle shortcuts help"},
			{key: "esc", action: "Back / cancel modal"},
			{key: "q | ctrl+c", action: "Quit MonoBox"},
		},
	}

	var colDefs []helpColumn
	var sep string
	var sepWidth int

	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorBorder))

	if width >= 95 {
		colDefs = []helpColumn{
			{sections: []helpSection{secNav, secSystem}},
			{sections: []helpSection{secActions}},
			{sections: []helpSection{secStacks, secLogs}},
		}
		sep = "  " + sepStyle.Render("│") + "  "
		sepWidth = 5
	} else if width >= 60 {
		colDefs = []helpColumn{
			{sections: []helpSection{secNav, secActions, secSystem}},
			{sections: []helpSection{secStacks, secLogs}},
		}
		sep = " " + sepStyle.Render("│") + " "
		sepWidth = 3
	} else {
		colDefs = []helpColumn{
			{sections: []helpSection{secNav, secActions, secStacks, secLogs, secSystem}},
		}
		sep = ""
		sepWidth = 0
	}

	numCols := len(colDefs)
	totalSepWidth := (numCols - 1) * sepWidth
	availableWidth := width - totalSepWidth
	if availableWidth < numCols {
		availableWidth = numCols
	}
	baseColWidth := availableWidth / numCols
	remainder := availableWidth % numCols

	colWidths := make([]int, numCols)
	for i := 0; i < numCols; i++ {
		colWidths[i] = baseColWidth
		if i == numCols-1 {
			colWidths[i] += remainder
		}
	}

	headingStyle := lipgloss.NewStyle().Foreground(ui.ColorCyan).Bold(true)
	dividerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorBorder))
	keyStyle := ui.FooterKeyStyle
	actStyle := ui.ValueStyle

	renderedColumns := make([][]string, numCols)

	for colIdx, col := range colDefs {
		cWidth := colWidths[colIdx]

		maxKey := 0
		for _, sec := range col.sections {
			for _, e := range sec.entries {
				if w := lipgloss.Width(e.key); w > maxKey {
					maxKey = w
				}
			}
		}
		if maxKey < 6 {
			maxKey = 6
		}
		if maxKey > 14 {
			maxKey = 14
		}
		if maxKey > cWidth-4 && cWidth > 4 {
			maxKey = cWidth - 4
		}

		keyColW := maxKey
		actColW := cWidth - keyColW - 2
		if actColW < 1 {
			actColW = 1
		}

		var colLines []string
		for secIdx, sec := range col.sections {
			if secIdx > 0 {
				colLines = append(colLines, strings.Repeat(" ", cWidth))
			}

			titleStr := sec.heading
			if lipgloss.Width(titleStr) > cWidth {
				titleStr = truncateRunes(titleStr, cWidth)
			}
			titleRendered := headingStyle.Render(titleStr)
			if pad := cWidth - lipgloss.Width(titleRendered); pad > 0 {
				titleRendered += strings.Repeat(" ", pad)
			}
			colLines = append(colLines, titleRendered)

			divLine := ""
			if cWidth > 0 {
				divLine = dividerStyle.Render(strings.Repeat("─", cWidth))
			}
			colLines = append(colLines, divLine)

			for _, e := range sec.entries {
				kPadded := padRight(e.key, keyColW)
				actPadded := padRight(e.action, actColW)
				if lipgloss.Width(e.action) > actColW {
					actPadded = truncateRunes(e.action, actColW)
				}
				row := keyStyle.Render(kPadded) + "  " + actStyle.Render(actPadded)
				if pad := cWidth - lipgloss.Width(row); pad > 0 {
					row += strings.Repeat(" ", pad)
				}
				colLines = append(colLines, row)
			}
		}
		renderedColumns[colIdx] = colLines
	}

	maxLines := 0
	for _, lines := range renderedColumns {
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}

	for colIdx := 0; colIdx < numCols; colIdx++ {
		cWidth := colWidths[colIdx]
		for len(renderedColumns[colIdx]) < maxLines {
			renderedColumns[colIdx] = append(renderedColumns[colIdx], strings.Repeat(" ", cWidth))
		}
	}

	finalLines := make([]string, maxLines)
	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
		if numCols == 1 {
			finalLines[lineIdx] = renderedColumns[0][lineIdx]
		} else {
			parts := make([]string, numCols)
			for colIdx := 0; colIdx < numCols; colIdx++ {
				parts[colIdx] = renderedColumns[colIdx][lineIdx]
			}
			finalLines[lineIdx] = strings.Join(parts, sep)
		}
	}

	return strings.Join(finalLines, "\n")
}

func (m *Model) renderThemeMenuModal() string {
	title := ui.ModalTitleStyle.Render(" Select Theme ") + " " + ui.SubtleStyle.Render("(↑/↓ navigate • enter save • esc cancel)")

	var rows []string
	for i, t := range ui.Themes {
		isSelected := i == m.themeCursor
		isCurrentSaved := strings.EqualFold(t.Name, m.cfg.Theme)

		var prefix string
		if isSelected {
			prefix = ui.PointerStyle.Render("➔ ")
		} else {
			prefix = "  "
		}

		var itemText string
		if isSelected {
			itemText = ui.SelectedItemStyle.Render(" " + t.Name + " ")
		} else {
			itemText = ui.NormalItemStyle.Render(" " + t.Name + " ")
		}

		var badge string
		if isCurrentSaved {
			badge = " " + ui.SuccessStyle.Render("[active]")
		}

		sampleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Highlight)).Bold(true)
		preview := sampleStyle.Render(" ■ ")

		rows = append(rows, prefix+itemText+preview+badge)
	}

	body := strings.Join(rows, "\n")
	return lipgloss.JoinVertical(lipgloss.Left, title, "", body)
}

func (m *Model) fmtHelpRow(key, desc string) string {
	kStyled := ui.FooterKeyStyle.Width(18).Render(key)
	dStyled := ui.SubtleStyle.Render(desc)
	return " " + kStyled + " " + dStyled
}

func (m *Model) renderEnvModal() string {
	c := m.selectedContainer()
	name := "Container"
	if c != nil {
		name = c.Name
	}
	header := ui.LabelStyle.Render("Environment Variables — "+name) + " " + ui.SubtleStyle.Render("(esc/q/E to close)")

	var envList []string
	if c != nil {
		if details := m.inspectDetailsCache[c.ID]; details != nil && len(details.Env) > 0 {
			envList = details.Env
		}
	}

	w := m.width - 12
	if w < 30 {
		w = 30
	}
	h := m.height - 10
	if h < 6 {
		h = 6
	}

	if len(envList) == 0 {
		body := ui.SubtleStyle.Render("No environment variables found or loading inspect details...")
		return lipgloss.JoinVertical(lipgloss.Left, header, "", body)
	}

	var rows []string
	for _, env := range envList {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			val := parts[1]
			
			masked := false
			if m.envMaskSecrets && isSensitiveKey(key) {
				val = maskValue(val)
				masked = true
			}
			
			k := ui.FooterKeyStyle.Render(key)
			v := ui.ValueStyle.Render(val)
			if masked {
				v += " 🔒"
			}
			rows = append(rows, fmt.Sprintf("  %-30s = %s", k, v))
		} else {
			rows = append(rows, "  "+ui.ValueStyle.Render(env))
		}
	}

	m.envViewport = viewport.New(w, h)
	m.envViewport.SetContent(strings.Join(rows, "\n"))
	body := renderViewportWithScrollbar(m.envViewport, true)

	return lipgloss.JoinVertical(lipgloss.Left, header, "", body)
}

func (m *Model) renderHealthModal() string {
	c := m.selectedContainer()
	name := "Container"
	if c != nil {
		name = c.Name
	}
	header := ui.LabelStyle.Render("Healthcheck Logs — "+name) + " " + ui.SubtleStyle.Render("(esc/q/H to close)")

	var health *domain.HealthDetail
	if c != nil {
		if details := m.inspectDetailsCache[c.ID]; details != nil {
			health = details.Health
		}
	}

	w := m.width - 12
	if w < 30 {
		w = 30
	}
	h := m.height - 10
	if h < 6 {
		h = 6
	}

	if health == nil || len(health.Log) == 0 {
		body := ui.SubtleStyle.Render("No healthcheck history found for this container.")
		return lipgloss.JoinVertical(lipgloss.Left, header, "", body)
	}

	var rows []string
	statusStr := string(health.Status)
	if health.Status == domain.HealthHealthy {
		statusStr = ui.StatusSuccessStyle.Render("● HEALTHY")
	} else if health.Status == domain.HealthUnhealthy {
		statusStr = ui.StatusErrorStyle.Render("✖ UNHEALTHY")
	}
	rows = append(rows, fmt.Sprintf("  Status: %s  •  Failing Streak: %d", statusStr, health.FailingStreak))
	rows = append(rows, "")
	rows = append(rows, ui.LabelStyle.Render("  PROBE LOGS (Recent):"))

	for i, log := range health.Log {
		exitStr := ui.SuccessStyle.Render("Exit: 0 (Success)")
		if log.ExitCode != 0 {
			exitStr = ui.ErrorStyle.Render(fmt.Sprintf("Exit: %d (Failure)", log.ExitCode))
		}
		rows = append(rows, fmt.Sprintf("  [%d] %s  •  %s", i+1, log.Start, exitStr))
		if strings.TrimSpace(log.Output) != "" {
			outLines := strings.Split(strings.TrimSpace(log.Output), "\n")
			for _, l := range outLines {
				rows = append(rows, ui.SubtleStyle.Render("      | "+l))
			}
		}
		rows = append(rows, "")
	}

	m.healthViewport = viewport.New(w, h)
	m.healthViewport.SetContent(strings.Join(rows, "\n"))
	body := renderViewportWithScrollbar(m.healthViewport, true)

	return lipgloss.JoinVertical(lipgloss.Left, header, "", body)
}

func (m *Model) renderPortConflictModal() string {
	title := ui.WarningStyle.Render(" ⚠️ PORT CONFLICT DETECTED ")
	msg := fmt.Sprintf("Host port %s is already in use by running container '%s'.\nStarting this container may fail due to port binding conflict.",
		ui.ValueStyle.Render(m.conflictingPort),
		ui.LabelStyle.Render(m.conflictingContainer),
	)
	prompt := ui.SubtleStyle.Render("Do you want to start it anyway? [y/N]")
	return lipgloss.JoinVertical(lipgloss.Left, title, "", msg, "", prompt)
}

func (m *Model) renderGraphModal() string {
	c := m.selectedContainer()
	name := "Container"
	if c != nil {
		name = c.Name
	}
	header := ui.LabelStyle.Render("Historical Metrics Graph — "+name) + " " + ui.SubtleStyle.Render("(esc/q/g to close)")

	if c == nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, "", ui.SubtleStyle.Render("No container selected."))
	}

	key := c.ID
	if key == "" {
		key = c.Name
	}
	cpuHist, memHist := m.GetStatsHistory(key)

	w := m.width - 12
	if w < 40 {
		w = 40
	}
	h := m.height - 10
	if h < 10 {
		h = 10
	}

	var rows []string
	rows = append(rows, fmt.Sprintf("  Container: %s  •  Status: %s", ui.ValueStyle.Render(c.Name), string(c.Status)))
	rows = append(rows, fmt.Sprintf("  Current CPU: %s  •  Current Mem: %s", ui.ValueStyle.Render(c.CPU), ui.ValueStyle.Render(c.Mem)))
	rows = append(rows, "")

	chartWidth := w - 16
	if chartWidth < 20 {
		chartWidth = 20
	}

	rows = append(rows, ui.LabelStyle.Render("  📈 CPU UTILIZATION HISTORY (%):"))
	cpuGraph := renderDetailedHistoryGraph(cpuHist, chartWidth, 5, ui.ColorHighlight)
	for _, l := range strings.Split(cpuGraph, "\n") {
		rows = append(rows, "    "+l)
	}
	rows = append(rows, "")

	rows = append(rows, ui.LabelStyle.Render("  📊 MEMORY UTILIZATION HISTORY (%):"))
	memGraph := renderDetailedHistoryGraph(memHist, chartWidth, 5, ui.ColorCyan)
	for _, l := range strings.Split(memGraph, "\n") {
		rows = append(rows, "    "+l)
	}

	m.graphViewport = viewport.New(w, h)
	m.graphViewport.SetContent(strings.Join(rows, "\n"))
	body := renderViewportWithScrollbar(m.graphViewport, true)

	return lipgloss.JoinVertical(lipgloss.Left, header, "", body)
}

func renderDetailedHistoryGraph(history []float64, width, height int, barColor lipgloss.Color) string {
	if len(history) == 0 {
		return ui.SubtleStyle.Render("No stats history recorded yet.")
	}

	data := history
	if len(data) > width {
		data = data[len(data)-width:]
	}

	var maxVal float64 = 1.0
	for _, v := range data {
		if v > maxVal {
			maxVal = v
		}
	}

	blocks := []string{" ", " ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	style := lipgloss.NewStyle().Foreground(barColor)

	var lines []string
	for r := height - 1; r >= 0; r-- {
		var line strings.Builder
		levelThreshold := maxVal * float64(r) / float64(height)
		nextThreshold := maxVal * float64(r+1) / float64(height)

		for _, val := range data {
			if val >= nextThreshold {
				line.WriteString("█")
			} else if val > levelThreshold {
				fraction := (val - levelThreshold) / (nextThreshold - levelThreshold)
				idx := int(fraction * float64(len(blocks)-1))
				if idx < 0 {
					idx = 0
				}
				if idx >= len(blocks) {
					idx = len(blocks) - 1
				}
				line.WriteString(blocks[idx])
			} else {
				line.WriteString(" ")
			}
		}
		lines = append(lines, fmt.Sprintf("%5.1f%% │ %s", levelThreshold, style.Render(line.String())))
	}
	lines = append(lines, fmt.Sprintf("       └%s", strings.Repeat("─", len(data))))
	return strings.Join(lines, "\n")
}
