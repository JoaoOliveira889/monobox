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
	} else if m.stream != nil {
		loading = ui.SpinnerStyle.Render(m.spinnerView() + " Streaming...")
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
	for _, c := range m.containers {
		if c.IsRunning() {
			running++
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
		parts = append(parts, ui.SubtleStyle.Render("["+m.engine+"]"))
	} else if !m.loading {
		parts = append(parts, "No containers found")
	}

	helpHint := ui.LabelStyle.Render("Press q to quit")
	parts = append(parts, helpHint)

	sep := ui.SubtleStyle.Render("  •  ")
	barText := " " + dot + " " + strings.Join(parts, sep)

	return ui.SubtleStyle.MaxWidth(maxW).Render(barText)
}

// renderFooter — matches monogit's footer: key hints left, version right.
func (m *Model) renderFooter() string {
	sep := ui.SubtleStyle.Render(" • ")
	var parts []string

	if m.activePanel == LogsPanel {
		parts = []string{
			m.fmtKey("1/2/tab", "focus"),
			m.fmtKey("<>", "resize"),
			m.fmtKey("↑↓/jk", "scroll"),
			m.fmtKey("f", "follow"),
			m.fmtKey("esc", "back"),
			m.fmtKey("q", "quit"),
		}
	} else {
		parts = []string{
			m.fmtKey("1/2/tab", "focus"),
			m.fmtKey("<>", "resize"),
			m.fmtKey("↑↓/jk", "nav"),
			m.fmtKey("enter", "logs"),
			m.fmtKey("s", "start/stop"),
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
	headerHeight := 3
	footerHeight := 1
	bodyHeight := m.height - headerHeight - footerHeight
	if bodyHeight < 5 {
		bodyHeight = 5
	}

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
