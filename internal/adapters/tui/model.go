package tui

import (
	"bufio"
	"context"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monobox/internal/domain"
)

// Version is set at build time via ldflags.
var Version = "0.0.1"

const (
	minTerminalWidth  = 40
	minTerminalHeight = 10
	// headerHeight = 1 (brand) + 1 (status bar) + 1 (border line)
	footerOverhead = 4

	logTailLines    = 100
	logLineLimit    = 100
	refreshInterval = 5 * time.Second

	statusClearDuration = 3 * time.Second
	spinnerTickInterval = 80 * time.Millisecond
	splashDuration      = 1500 * time.Millisecond

	defaultRatio      = 0.40
	minLeftPanelRatio = 0.20
	maxLeftPanelRatio = 0.80
	resizeStep        = 0.05

	minPanelWidth  = 20
	minPanelHeight = 5

	containerRowHeight = 1
)

// Panel identifies which panel is currently focused.
type Panel int

const (
	ListPanel Panel = iota
	LogsPanel
)

// containerItem is the TUI view-model for a single container row.
type containerItem struct {
	domain.Container
	loading bool // true while a lifecycle action is in progress
}

// logStream holds everything needed to read from a live log stream.
type logStream struct {
	containerID string
	reader      io.Closer
	scanner     *bufio.Scanner
	ctx         context.Context
	cancel      context.CancelFunc
}

// Model is the root bubbletea model for monobox.
type Model struct {
	provider domain.ContainerProvider
	engine   string

	containers  []containerItem
	cursor      int
	activePanel Panel

	// log panel state
	logLines       []string
	logFollow      bool
	stream         *logStream
	logContainerID string
	logViewport    viewport.Model
	listViewport   viewport.Model

	confirmClearLogs bool

	// layout
	width          int
	height         int
	leftPanelRatio float64

	// splash
	showSplash  bool
	splashFrame int

	// status bar
	statusMsg   string
	statusMsgID int

	// animation
	spinnerFrame int
	loading      bool

	quitting bool
}

// NewModel returns an initialized Model ready for bubbletea.
func NewModel(provider domain.ContainerProvider, engineName string) Model {
	return Model{
		provider:       provider,
		engine:         engineName,
		activePanel:    ListPanel,
		leftPanelRatio: defaultRatio,
		logViewport:    viewport.New(0, 0),
		listViewport:   viewport.New(0, 0),
		loading:        true,
		logFollow:      true,
		showSplash:     true,
	}
}

// Init fires the first batch of commands.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadContainersCmd(),
		spinnerTickCmd(),
		splashTickCmd(),
	)
}

// selectedContainer returns a pointer to the focused container, or nil.
func (m *Model) selectedContainer() *containerItem {
	if len(m.containers) == 0 || m.cursor < 0 || m.cursor >= len(m.containers) {
		return nil
	}
	return &m.containers[m.cursor]
}

func (m Model) leftPanelWidth() int {
	ratio := m.leftPanelRatio
	if ratio <= 0 {
		ratio = defaultRatio
	}
	w := int(float64(m.width) * ratio)
	if w < minPanelWidth {
		w = minPanelWidth
	}
	return w
}

func (m Model) rightPanelWidth() int {
	return m.width - m.leftPanelWidth()
}

func (m Model) panelHeight() int {
	h := m.height - footerOverhead
	if h < minPanelHeight {
		h = minPanelHeight
	}
	return h
}

func (m *Model) setStatus(msg string) {
	m.statusMsg = msg
}

// cancelStream cancels the active log stream, if any.
func (m *Model) cancelStream() {
	if m.stream != nil {
		m.stream.cancel()
		_ = m.stream.reader.Close()
		m.stream = nil
	}
}

func (m *Model) appendLogLine(line string) {
	m.logLines = append(m.logLines, line)
	if len(m.logLines) > logLineLimit {
		m.logLines = m.logLines[len(m.logLines)-logLineLimit:]
	}
}

var spinnerFrames = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

func (m *Model) spinnerView() string {
	return spinnerFrames[m.spinnerFrame%len(spinnerFrames)]
}

func (m *Model) refreshListViewport() {
	m.listViewport.SetContent(m.renderContainerListContent())
	m.ensureSelectedContainerVisible()
}

func (m *Model) ensureSelectedContainerVisible() {
	if m.listViewport.Height <= 0 || len(m.containers) == 0 {
		return
	}
	rowTop := m.cursor * containerRowHeight
	rowBottom := rowTop + containerRowHeight - 1
	viewTop := m.listViewport.YOffset
	viewBottom := viewTop + m.listViewport.Height - 1
	if rowTop < viewTop {
		m.listViewport.SetYOffset(rowTop)
	} else if rowBottom > viewBottom {
		m.listViewport.SetYOffset(rowBottom - m.listViewport.Height + 1)
	}
}

func (m *Model) refreshLogViewportContent() {
	if len(m.logLines) == 0 {
		m.logViewport.SetContent("")
		return
	}
	m.logViewport.SetContent(strings.Join(m.logLines, "\n"))
	if m.logFollow {
		m.logViewport.GotoBottom()
	}
}

var _ tea.Model = &Model{}

// ── Test accessors ─────────────────────────────────────────────────────────────

func (m Model) ActivePanel() Panel { return m.activePanel }
func (m Model) Cursor() int        { return m.cursor }
func (m Model) LogFollow() bool    { return m.logFollow }
func (m *Model) SetCursor(i int)   { m.cursor = i }

func statusRank(status domain.ContainerStatus) int {
	switch status {
	case domain.StatusRunning:
		return 0
	case domain.StatusPaused:
		return 1
	case domain.StatusCreated:
		return 2
	case domain.StatusExited:
		return 3
	default:
		return 4
	}
}

func sortContainerItems(items []containerItem) {
	sort.SliceStable(items, func(i, j int) bool {
		iRank := statusRank(items[i].Status)
		jRank := statusRank(items[j].Status)
		if iRank != jRank {
			return iRank < jRank
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})
}

func (m *Model) ApplyContainersLoaded(list []domain.Container) {
	var prevID string
	if c := m.selectedContainer(); c != nil {
		prevID = c.ID
	}
	items := make([]containerItem, len(list))
	for i, c := range list {
		items[i] = containerItem{Container: c}
	}
	sortContainerItems(items)
	m.containers = items
	if prevID != "" {
		for i, c := range m.containers {
			if c.ID == prevID {
				m.cursor = i
				return
			}
		}
	}
	if m.cursor >= len(m.containers) {
		m.cursor = max(0, len(m.containers)-1)
	}
}
