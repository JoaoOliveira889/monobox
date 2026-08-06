package tui

import (
	"fmt"
	"strconv"
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
	if m.confirmClearLogs {
		return m.renderCenteredModal(m.renderClearLogsConfirmation())
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

// renderHeader matches monogit's header style:
// Line 1: Brand wordmark (MonoBox 2-color) • spinner/loading • stats
// Line 2: Status bar line (● N containers • N running • [engine] • Press q to quit)
// Line 3: Border line (─)
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
	var totalCPU float64
	var totalMemBytes float64

	for _, c := range m.containers {
		if c.IsRunning() {
			running++
			totalCPU += parseCPUVal(c.CPU)
			totalMemBytes += parseMemBytesVal(c.Mem)
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
	var numStr string
	var unitStr string
	for i, r := range s {
		if (r >= '0' && r <= '9') || r == '.' {
			numStr += string(r)
		} else {
			unitStr = strings.TrimSpace(s[i:])
			break
		}
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

// renderFooter — key hints left, version right.
func (m *Model) renderFooter() string {
	sep := ui.SubtleStyle.Render(" • ")
	var parts []string

	if m.activePanel == LogsPanel {
		followAction := "follow:ON"
		if !m.logFollow {
			followAction = "follow:OFF"
		}
		parts = []string{
			m.fmtKey("↑↓/jk", "scroll"),
			m.fmtKey("pg↑↓/end", "scroll/bottom"),
			m.fmtKey("f", followAction),
			m.fmtKey("<>", "resize"),
			m.fmtKey("c", "clear"),
			m.fmtKey("esc", "back"),
			m.fmtKey("q", "quit"),
		}
	} else {
		parts = []string{
			m.fmtKey("↑↓/jk", "nav"),
			m.fmtKey("s", "start/stop"),
			m.fmtKey("<>", "resize"),
			m.fmtKey("enter", "logs"),
			m.fmtKey("r", "restart"),
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

// renderBody renders the two-panel layout.
func (m *Model) renderBody() string {
	bodyHeight := m.panelHeight()

	leftWidth := m.leftPanelWidth()
	rightWidth := m.rightPanelWidth()

	left := m.renderRepoList(leftWidth, bodyHeight)
	right := m.renderDetailPanel(rightWidth, bodyHeight)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

// refreshViewports syncs viewport content after state changes.
func (m *Model) refreshViewports() {
	m.refreshListViewport()
	m.refreshLogViewportContent()
}

// renderCenteredModal centers content on screen.
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
