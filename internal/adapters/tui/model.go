package tui

import (
	"bufio"
	"context"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/JoaoOliveira889/monobox/internal/domain"
	"github.com/JoaoOliveira889/monobox/internal/pkg/config"
	"github.com/JoaoOliveira889/monobox/internal/pkg/ui"
)

var Version = "0.0.3"

const (
	minTerminalWidth  = 40
	minTerminalHeight = 10
	footerOverhead    = 4

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

type Panel int

const (
	ListPanel Panel = iota
	LogsPanel
)

type NodeType int

const (
	NodeProjectHeader NodeType = iota
	NodeContainerItem
)

type TreeNode struct {
	Type          NodeType
	ProjectName   string
	Expanded      bool
	Container     *containerItem
	TotalCount    int
	RunningCount  int
	IsLastInGroup bool
}

type containerItem struct {
	domain.Container
	loading  bool
	starting bool
}

type logStream struct {
	containerID string
	reader      io.Closer
	scanner     *bufio.Scanner
	ctx         context.Context
	cancel      context.CancelFunc
}

const maxStatsHistoryLen = 20

type StatsHistory struct {
	CPU []float64
	Mem []float64
}

type Model struct {
	provider domain.ContainerProvider
	engine   string

	cfg             config.Config
	logLineLimit    int
	refreshInterval time.Duration

	containers       []containerItem
	expandedProjects map[string]bool
	statsHistory     map[string]*StatsHistory
	cursor           int
	activePanel      Panel

	filterInput textinput.Model
	filtering   bool
	filterQuery string

	logLines       []string
	logFollow      bool
	stream         *logStream
	logContainerID string
	logViewport    viewport.Model
	listViewport   viewport.Model

	logSearchInput textinput.Model
	logSearching   bool
	logSearchQuery string
	showTimestamps bool

	confirmClearLogs   bool
	confirmRemove      bool
	confirmBatchAction string
	batchProjectName   string

	showInspect     bool
	inspectContent  string
	inspectViewport viewport.Model
	helpViewport    viewport.Model
	envViewport     viewport.Model
	healthViewport  viewport.Model

	showEnvModal    bool
	showHealthModal bool
	showHelp        bool
	showThemeMenu   bool
	themeCursor     int
	initialTheme    string

	inspectDetailsCache map[string]*domain.ContainerInspectDetails

	width          int
	height         int
	leftPanelRatio float64

	showSplash  bool
	splashFrame int

	statusMsg   string
	statusMsgID int

	spinnerFrame int
	loading      bool

	quitting bool
}

func NewModel(provider domain.ContainerProvider, engineName string) Model {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}
	ui.ApplyTheme(cfg.Theme)

	interval := time.Duration(cfg.MetricsInterval) * time.Second
	if interval <= 0 {
		interval = refreshInterval
	}
	limit := cfg.LogLineLimit
	if limit <= 0 {
		limit = logLineLimit
	}

	ti := textinput.New()
	ti.Placeholder = "filter containers..."
	ti.Prompt = "🔍 "
	ti.CharLimit = 50

	tiSearch := textinput.New()
	tiSearch.Placeholder = "search logs..."
	tiSearch.Prompt = "🔍 "
	tiSearch.CharLimit = 50

	return Model{
		provider:            provider,
		engine:              engineName,
		cfg:                 cfg,
		logLineLimit:        limit,
		refreshInterval:     interval,
		showTimestamps:      cfg.ShowTimestamps,
		activePanel:         ListPanel,
		leftPanelRatio:      defaultRatio,
		filterInput:         ti,
		logSearchInput:      tiSearch,
		logViewport:         viewport.New(0, 0),
		listViewport:        viewport.New(0, 0),
		inspectViewport:     viewport.New(0, 0),
		helpViewport:        viewport.New(0, 0),
		envViewport:         viewport.New(0, 0),
		healthViewport:      viewport.New(0, 0),
		loading:             true,
		logFollow:           true,
		showSplash:          true,
		expandedProjects:    make(map[string]bool),
		statsHistory:        make(map[string]*StatsHistory),
		inspectDetailsCache: make(map[string]*domain.ContainerInspectDetails),
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadContainersCmd(),
		spinnerTickCmd(),
		splashTickCmd(),
	)
}

func matchesContainer(c domain.Container, query string) bool {
	if strings.Contains(strings.ToLower(c.Name), query) {
		return true
	}
	if strings.Contains(strings.ToLower(c.Image), query) {
		return true
	}
	if strings.Contains(strings.ToLower(c.Ports), query) {
		return true
	}
	if strings.Contains(strings.ToLower(string(c.Status)), query) {
		return true
	}
	if strings.Contains(strings.ToLower(c.ComposeProject), query) {
		return true
	}
	icon, label := containerIconAndLabel(c)
	if strings.Contains(strings.ToLower(icon), query) || strings.Contains(strings.ToLower(label), query) {
		return true
	}
	return false
}

func (m *Model) FilteredContainers() []containerItem {
	query := strings.ToLower(strings.TrimSpace(m.filterQuery))
	if query == "" {
		return m.containers
	}
	var res []containerItem
	for _, c := range m.containers {
		if matchesContainer(c.Container, query) {
			res = append(res, c)
		}
	}
	return res
}

func (m *Model) isProjectExpanded(proj string) bool {
	if m.expandedProjects == nil {
		return true
	}
	expanded, ok := m.expandedProjects[proj]
	if !ok {
		return true
	}
	return expanded
}

func (m *Model) toggleProjectExpanded(proj string) {
	if m.expandedProjects == nil {
		m.expandedProjects = make(map[string]bool)
	}
	m.expandedProjects[proj] = !m.isProjectExpanded(proj)
}

func (m *Model) VisibleTreeNodes() []TreeNode {
	filtered := m.FilteredContainers()
	if len(filtered) == 0 {
		return nil
	}

	projectContainers := make(map[string][]containerItem)
	var projectOrder []string
	var standalone []containerItem

	for _, item := range filtered {
		proj := item.ComposeProject
		if proj != "" {
			if _, exists := projectContainers[proj]; !exists {
				projectOrder = append(projectOrder, proj)
			}
			projectContainers[proj] = append(projectContainers[proj], item)
		} else {
			standalone = append(standalone, item)
		}
	}

	sort.Strings(projectOrder)

	var nodes []TreeNode
	for _, proj := range projectOrder {
		items := projectContainers[proj]
		running := 0
		for _, c := range items {
			if c.IsRunning() {
				running++
			}
		}

		isExpanded := m.isProjectExpanded(proj)
		if m.filterQuery != "" {
			isExpanded = true
		}

		nodes = append(nodes, TreeNode{
			Type:         NodeProjectHeader,
			ProjectName:  proj,
			Expanded:     isExpanded,
			TotalCount:   len(items),
			RunningCount: running,
		})

		if isExpanded {
			for i, c := range items {
				cCopy := c
				nodes = append(nodes, TreeNode{
					Type:          NodeContainerItem,
					ProjectName:   proj,
					Container:     &cCopy,
					IsLastInGroup: i == len(items)-1,
				})
			}
		}
	}

	for _, c := range standalone {
		cCopy := c
		nodes = append(nodes, TreeNode{
			Type:        NodeContainerItem,
			ProjectName: "",
			Container:   &cCopy,
		})
	}

	return nodes
}

func (m *Model) selectedNode() *TreeNode {
	nodes := m.VisibleTreeNodes()
	if len(nodes) == 0 || m.cursor < 0 || m.cursor >= len(nodes) {
		return nil
	}
	return &nodes[m.cursor]
}

func (m *Model) selectedContainer() *containerItem {
	node := m.selectedNode()
	if node == nil || node.Type != NodeContainerItem {
		return nil
	}
	return node.Container
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

func (m *Model) cancelStream() {
	if m.stream != nil {
		m.stream.cancel()
		_ = m.stream.reader.Close()
		m.stream = nil
	}
}

func (m *Model) appendLogLine(line string) {
	limit := m.logLineLimit
	if limit <= 0 {
		limit = logLineLimit
	}
	m.logLines = append(m.logLines, line)
	if len(m.logLines) > limit {
		m.logLines = m.logLines[len(m.logLines)-limit:]
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
	nodes := m.VisibleTreeNodes()
	if m.listViewport.Height <= 0 || len(nodes) == 0 {
		return
	}
	if m.cursor >= len(nodes) {
		m.cursor = max(0, len(nodes)-1)
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
	query := strings.TrimSpace(m.logSearchQuery)
	queryLower := strings.ToLower(query)
	var formatted []string
	for _, line := range m.logLines {
		if queryLower != "" && !strings.Contains(strings.ToLower(line), queryLower) {
			continue
		}
		formatted = append(formatted, highlightLogLine(line, query))
	}
	if len(formatted) == 0 && query != "" {
		m.logViewport.SetContent(ui.SubtleStyle.Render(" No logs matching filter: " + m.logSearchQuery))
		return
	}
	m.logViewport.SetContent(strings.Join(formatted, "\n"))
	if m.logFollow {
		m.logViewport.GotoBottom()
	}
}

func highlightSubstring(s string, sub string, style lipgloss.Style) string {
	if sub == "" {
		return s
	}
	lowerS := strings.ToLower(s)
	lowerSub := strings.ToLower(sub)
	if !strings.Contains(lowerS, lowerSub) {
		return s
	}

	var sb strings.Builder
	idx := 0
	subLen := len(sub)
	for {
		i := strings.Index(lowerS[idx:], lowerSub)
		if i == -1 {
			sb.WriteString(s[idx:])
			break
		}
		matchStart := idx + i
		matchEnd := matchStart + subLen
		sb.WriteString(s[idx:matchStart])
		sb.WriteString(style.Render(s[matchStart:matchEnd]))
		idx = matchEnd
	}
	return sb.String()
}

func highlightKeywords(line string) string {
	keywordsErr := []string{"ERROR", "panic", "FATAL", "500", "ERR"}
	keywordsWarn := []string{"WARN", "WARNING"}

	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e")).Bold(true)
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#e0af68")).Bold(true)

	for _, kw := range keywordsErr {
		line = highlightSubstring(line, kw, errStyle)
	}
	for _, kw := range keywordsWarn {
		line = highlightSubstring(line, kw, warnStyle)
	}
	return line
}

func highlightLogLine(line string, query string) string {
	line = highlightKeywords(line)
	if query != "" {
		matchStyle := lipgloss.NewStyle().Background(lipgloss.Color("#e0af68")).Foreground(lipgloss.Color("#1a1b26")).Bold(true)
		line = highlightSubstring(line, query, matchStyle)
	}
	return line
}

var _ tea.Model = &Model{}

func (m Model) ActivePanel() Panel { return m.activePanel }
func (m Model) Cursor() int        { return m.cursor }
func (m Model) LogFollow() bool    { return m.logFollow }
func (m *Model) SetCursor(i int)   { m.cursor = i }

func statusRank(item containerItem) int {
	if item.starting {
		return -1
	}
	switch item.Status {
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
		iRank := statusRank(items[i])
		jRank := statusRank(items[j])
		if iRank != jRank {
			return iRank < jRank
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})
}

func (m *Model) ApplyContainersLoaded(list []domain.Container) {
	if m.statsHistory == nil {
		m.statsHistory = make(map[string]*StatsHistory)
	}
	for _, c := range list {
		key := c.ID
		if key == "" {
			key = c.Name
		}
		if key == "" {
			continue
		}
		h, ok := m.statsHistory[key]
		if !ok {
			h = &StatsHistory{}
			m.statsHistory[key] = h
		}
		if cpuVal, ok := ui.ParsePercent(c.CPU); ok {
			h.CPU = append(h.CPU, cpuVal)
			if len(h.CPU) > maxStatsHistoryLen {
				h.CPU = h.CPU[len(h.CPU)-maxStatsHistoryLen:]
			}
		}
		if memVal, ok := ui.ParsePercent(c.Mem); ok {
			h.Mem = append(h.Mem, memVal)
			if len(h.Mem) > maxStatsHistoryLen {
				h.Mem = h.Mem[len(h.Mem)-maxStatsHistoryLen:]
			}
		}
	}

	var prevID string
	var prevProj string
	if node := m.selectedNode(); node != nil {
		if node.Type == NodeContainerItem && node.Container != nil {
			prevID = node.Container.ID
		} else if node.Type == NodeProjectHeader {
			prevProj = node.ProjectName
		}
	}

	items := make([]containerItem, len(list))
	for i, c := range list {
		items[i] = containerItem{Container: c}
	}
	sortContainerItems(items)
	m.containers = items

	nodes := m.VisibleTreeNodes()
	if prevID != "" {
		for i, n := range nodes {
			if n.Type == NodeContainerItem && n.Container != nil && n.Container.ID == prevID {
				m.cursor = i
				return
			}
		}
	}
	if prevProj != "" {
		for i, n := range nodes {
			if n.Type == NodeProjectHeader && n.ProjectName == prevProj {
				m.cursor = i
				return
			}
		}
	}
	if m.cursor >= len(nodes) {
		m.cursor = max(0, len(nodes)-1)
	}
}

func (m *Model) GetStatsHistory(key string) (cpu []float64, mem []float64) {
	if m.statsHistory == nil {
		return nil, nil
	}
	if h, ok := m.statsHistory[key]; ok {
		return h.CPU, h.Mem
	}
	return nil, nil
}

func extractHostPort(portsStr string) string {
	if portsStr == "" {
		return ""
	}
	parts := strings.Split(portsStr, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.Contains(p, "->") {
			hostPart := strings.Split(p, "->")[0]
			hostPart = strings.TrimSpace(hostPart)
			if hostPart != "" {
				return hostPart
			}
		} else if strings.Contains(p, ":") {
			hostPart := strings.Split(p, ":")[0]
			hostPart = strings.TrimSpace(hostPart)
			if hostPart != "" {
				return hostPart
			}
		}
	}
	return ""
}
