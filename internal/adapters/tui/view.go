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
	if m.showThemeMenu {
		return m.renderCenteredModal(m.renderThemeMenuModal())
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

func (m *Model) renderSplash() string {
	barWidth := 20
	filled := (m.splashFrame * 2) % (barWidth + 1)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	progressBar := ui.SpinnerStyle.Render(bar)

	scanStatus := ui.SpinnerStyle.Render(splashDots(m.splashFrame) + " detecting engine…")
	version := ui.SubtleStyle.Render("v" + Version)
	subtitle := ui.SubtleStyle.Render("Docker & Podman container manager")

	body := lipgloss.JoinVertical(lipgloss.Center,
		renderBrandWordmark(false),
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

func renderBrandWordmark(compact bool) string {
	mono := ui.BrandMonoStyle
	box := ui.BrandBoxStyle
	subtle := ui.SubtleStyle

	if compact {
		return lipgloss.JoinHorizontal(lipgloss.Bottom,
			mono.Render("Mono"),
			box.Render("Box"),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Top,
			mono.Render("Mono"),
			box.Render("Box"),
		),
		subtle.Render("container dashboard"),
	)
}

func (m *Model) renderHeader() string {
	var brand string
	switch {
	case m.width < 30:
		brand = ui.BrandMonoStyle.Render("MB")
	default:
		brand = renderBrandWordmark(true)
	}

	var stats string
	if m.width >= 50 {
		stats = fmt.Sprintf("%d containers ", len(m.containers))
	}

	loading := ""
	if m.loading {
		if m.width >= 40 {
			loading = ui.SpinnerStyle.Render(m.spinnerView() + " Loading...")
		} else {
			loading = ui.SpinnerStyle.Render(m.spinnerView())
		}
	}

	brandW := lipgloss.Width(brand)
	statsW := lipgloss.Width(stats)
	loadingW := lipgloss.Width(loading)

	maxW := m.width - 1
	if maxW < 10 {
		maxW = 10
	}

	spacerLen := maxW - brandW - statsW - loadingW - 2
	if spacerLen < 0 {
		spacerLen = 0
	}
	spacer := strings.Repeat(" ", spacerLen)

	headerLine := " " + lipgloss.JoinHorizontal(lipgloss.Bottom,
		brand,
		spacer,
		loading,
		ui.SubtleStyle.Render(stats),
	) + " "

	if lipgloss.Width(headerLine) > maxW {
		headerLine = truncateRunes(headerLine, maxW)
	}

	border := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ui.ColorBorder)).
		Render(strings.Repeat("─", maxW))

	styledStatus := m.renderHeaderStatusBar()

	return headerLine + "\n" + styledStatus + "\n" + border
}

func (m *Model) renderHeaderStatusBar() string {
	maxW := m.width - 1
	if maxW < 10 {
		maxW = 10
	}

	if m.statusMsg != "" {
		msg := " " + m.statusMsg
		switch {
		case strings.HasPrefix(m.statusMsg, "✓"):
			return ui.StatusSuccessStyle.MaxWidth(maxW).Render(msg)
		case strings.HasPrefix(m.statusMsg, "✗"):
			return ui.StatusErrorStyle.MaxWidth(maxW).Render(msg)
		default:
			return ui.StatusInfoStyle.MaxWidth(maxW).Render(msg)
		}
	}

	var parts []string
	total := len(m.containers)
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

	dot := lipgloss.NewStyle().Foreground(ui.ColorSuccess).Render("●")

	if total > 0 {
		parts = append(parts, fmt.Sprintf("%d containers", total))
		parts = append(parts, ui.RunningStyle.Render(fmt.Sprintf("%d running", running)))
		if stopped > 0 {
			parts = append(parts, ui.StoppedStyle.Render(fmt.Sprintf("%d stopped", stopped)))
		}
		if highLoadCount > 0 {
			parts = append(parts, ui.WarningStyle.Render(fmt.Sprintf("⚡ %d high load", highLoadCount)))
		}
		if running > 0 {
			engName := strings.ToLower(m.engine)
			if engName == "" {
				engName = "engine"
			}
			engineMetric := fmt.Sprintf("%s: %.2f%% CPU • %s", engName, totalCPU, formatBytes(totalMemBytes))
			parts = append(parts, lipgloss.NewStyle().Foreground(ui.ColorHighlight).Bold(true).Render(engineMetric))
		}
	} else if !m.loading {
		parts = append(parts, "No containers found")
	}

	sep := ui.SubtleStyle.Render("  •  ")
	barText := " " + dot + " " + strings.Join(parts, sep)

	return ui.SubtleStyle.MaxWidth(maxW).Render(barText)
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
	sep := ui.SubtleStyle.Render(" • ")
	var parts []string

	if m.filtering {
		parts = []string{
			m.fmtKey("?", "help"),
			m.fmtKey("esc", "clear filter"),
			m.fmtKey("enter", "done"),
			m.fmtKey("↑↓", "nav"),
		}
	} else if m.logSearching {
		parts = []string{
			m.fmtKey("?", "help"),
			m.fmtKey("esc", "clear search"),
			m.fmtKey("enter", "done"),
		}
	} else if m.activePanel == LogsPanel {
		followAction := "follow:ON"
		if !m.logFollow {
			followAction = "follow:OFF"
		}
		tsAction := "ts:OFF"
		if m.showTimestamps {
			tsAction = "ts:ON"
		}
		parts = []string{
			m.fmtKey("↑↓/jk", "scroll"),
			m.fmtKey("?", "help"),
			m.fmtKey("/", "search"),
			m.fmtKey("f", followAction),
			m.fmtKey("t", tsAction),
			m.fmtKey("s/ctrl+s", "export"),
			m.fmtKey("<>", "resize"),
			m.fmtKey("c", "clear"),
			m.fmtKey("esc", "back"),
			m.fmtKey("q", "quit"),
		}
	} else if node := m.selectedNode(); node != nil && node.Type == NodeProjectHeader {
		parts = []string{
			m.fmtKey("↑↓/jk", "nav"),
			m.fmtKey("?", "help"),
			m.fmtKey("space/enter", "toggle group"),
			m.fmtKey("s", "batch start/stop"),
			m.fmtKey("r", "batch restart"),
			m.fmtKey("d", "batch remove"),
			m.fmtKey("/", "filter"),
			m.fmtKey("q", "quit"),
		}
	} else {
		parts = []string{
			m.fmtKey("↑↓/jk", "nav"),
			m.fmtKey("?", "help"),
			m.fmtKey("/", "filter"),
			m.fmtKey("e", "exec"),
			m.fmtKey("i", "inspect"),
			m.fmtKey("s", "start/stop"),
			m.fmtKey("p", "pause"),
			m.fmtKey("d", "remove"),
			m.fmtKey("o", "open"),
			m.fmtKey("r", "restart"),
			m.fmtKey("T", "theme"),
			m.fmtKey("q", "quit"),
		}
	}

	return m.renderResponsiveFooter(parts, sep)
}

func (m *Model) renderResponsiveFooter(parts []string, sep string) string {
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

	left := rendered
	spacerLen := contentWidth - lipgloss.Width(left) - lipgloss.Width(version)
	if spacerLen < 0 {
		spacerLen = 0
	}
	spacer := strings.Repeat(" ", spacerLen)

	footerText := " " + left + spacer + version
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

	left := m.renderRepoList(leftWidth, bodyHeight)
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

func (m *Model) renderHelpOverlay() string {
	panelWidth, panelHeight := m.clampModalSize(4, 90, 2, 26)

	innerWidth := panelWidth - 6
	if innerWidth < 64 {
		innerWidth = 64
	}
	innerHeight := panelHeight - 6
	if innerHeight < 12 {
		innerHeight = 12
	}

	title := lipgloss.JoinHorizontal(lipgloss.Bottom,
		renderBrandWordmark(true),
		" ",
		ui.BrandTitleStyle.Render("SHORTCUTS"),
	)

	vpHeight := innerHeight - 3
	if vpHeight < 5 {
		vpHeight = 5
	}
	if m.helpViewport.Width != innerWidth-1 || m.helpViewport.Height != vpHeight {
		m.helpViewport = viewport.New(innerWidth-1, vpHeight)
	} else {
		m.helpViewport.Width = innerWidth - 1
		m.helpViewport.Height = vpHeight
	}

	body := m.renderHelpMenu(innerWidth-1, 999)
	m.helpViewport.SetContent(body)

	content := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Align(lipgloss.Center).Width(innerWidth).Render(title),
		"",
		renderViewportWithScrollbar(m.helpViewport, true),
		"",
		lipgloss.NewStyle().Align(lipgloss.Center).Width(innerWidth).Render(ui.SubtleStyle.Render("Press ESC, q or ? to close")),
	)

	panelStyle := ui.ActivePanelStyle.
		BorderStyle(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(ui.ColorHighlight)).
		Width(panelWidth).
		Height(panelHeight).
		Padding(1, 2)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panelStyle.Render(content))
}

func (m *Model) renderHelpMenu(width, height int) string {
	type helpEntry struct {
		key    string
		action string
	}
	type helpSection struct {
		title   string
		entries []helpEntry
	}

	sections := []helpSection{
		{
			title: "NAVIGATION & PANELS",
			entries: []helpEntry{
				{key: "1 | 2 | tab", action: "Switch panels"},
				{key: "↑↓ / jk", action: "Navigate / scroll"},
				{key: "< > / , .", action: "Resize panel width"},
				{key: "/", action: "Filter containers"},
				{key: "?", action: "Toggle help modal"},
				{key: "T", action: "Theme menu"},
				{key: "esc", action: "Back / clear filter"},
				{key: "q | ctrl+c", action: "Quit Monobox"},
			},
		},
		{
			title: "CONTAINER ACTIONS",
			entries: []helpEntry{
				{key: "s", action: "Start / Stop"},
				{key: "r", action: "Restart container"},
				{key: "p", action: "Pause / Unpause"},
				{key: "d | delete", action: "Remove container"},
				{key: "e | x", action: "Exec shell (/bin/sh)"},
				{key: "i", action: "Inspect JSON"},
				{key: "o", action: "Open host port"},
			},
		},
		{
			title: "COMPOSE STACKS",
			entries: []helpEntry{
				{key: "space | enter", action: "Expand / collapse"},
				{key: "s (group)", action: "Batch Start / Stop"},
				{key: "r (group)", action: "Batch Restart"},
				{key: "d (group)", action: "Batch Remove"},
			},
		},
		{
			title: "LOGS VIEW",
			entries: []helpEntry{
				{key: "↑↓ / jk", action: "Scroll log history"},
				{key: "pgup | pgdn", action: "Page up / down"},
				{key: "end | G", action: "Jump to bottom"},
				{key: "f", action: "Toggle live follow"},
				{key: "t", action: "Toggle timestamps"},
				{key: "c | ctrl+l", action: "Clear container logs"},
				{key: "/", action: "Search / filter logs"},
				{key: "s | ctrl+s", action: "Export logs to file"},
				{key: "esc", action: "Back to list"},
			},
		},
		{
			title: "THEME MENU",
			entries: []helpEntry{
				{key: "↑↓ / jk", action: "Live preview theme"},
				{key: "enter", action: "Save preference"},
				{key: "esc | q | T", action: "Cancel / restore"},
			},
		},
		{
			title: "CONFIRMATION MODALS",
			entries: []helpEntry{
				{key: "y | Y", action: "Confirm action"},
				{key: "n | N | esc", action: "Cancel action"},
			},
		},
	}

	colWidth := (width - 4) / 2
	if colWidth < 30 {
		colWidth = 30
	}

	var col1Blocks []string
	var col2Blocks []string

	for i, sec := range sections {
		var lines []string
		lines = append(lines, ui.LabelStyle.Bold(true).Render(sec.title))
		for _, e := range sec.entries {
			kStyled := ui.FooterKeyStyle.Width(14).Render(e.key)
			dStyled := ui.SubtleStyle.Render(e.action)
			lines = append(lines, " "+kStyled+" "+dStyled)
		}
		block := strings.Join(lines, "\n")
		if i%2 == 0 {
			col1Blocks = append(col1Blocks, block)
		} else {
			col2Blocks = append(col2Blocks, block)
		}
	}

	col1 := strings.Join(col1Blocks, "\n\n")
	col2 := strings.Join(col2Blocks, "\n\n")

	return lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(colWidth).Render(col1),
		"  ",
		lipgloss.NewStyle().Width(colWidth).Render(col2),
	)
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
			k := ui.FooterKeyStyle.Render(parts[0])
			v := ui.ValueStyle.Render(parts[1])
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
