package tui

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monobox/internal/domain"
)

type logProvider struct {
	clearCalls int
	clearErr   error
	logsTail   int
	timestamps bool
}

func (p *logProvider) EngineName() domain.Engine { return domain.EngineDocker }
func (p *logProvider) List() ([]domain.Container, error) {
	return nil, nil
}
func (p *logProvider) Stats() (map[string]domain.ContainerStats, error) { return nil, nil }
func (p *logProvider) ClearLogs(string) error {
	p.clearCalls++
	return p.clearErr
}
func (p *logProvider) Start(string) error            { return nil }
func (p *logProvider) Stop(string) error             { return nil }
func (p *logProvider) Restart(string) error          { return nil }
func (p *logProvider) Pause(string) error            { return nil }
func (p *logProvider) Unpause(string) error          { return nil }
func (p *logProvider) Remove(string, bool) error    { return nil }
func (p *logProvider) Inspect(string) (string, error) { return "{}", nil }
func (p *logProvider) ExecCmd(string) *exec.Cmd      { return exec.Command("echo", "test") }
func (p *logProvider) Logs(_ context.Context, _ string, tail int, _, timestamps bool) (io.ReadCloser, error) {
	p.logsTail = tail
	p.timestamps = timestamps
	return io.NopCloser(strings.NewReader("")), nil
}

func TestClearLogsRequiresConfirmation(t *testing.T) {
	p := &logProvider{}
	m := NewModel(p, "docker")
	m.showSplash = false
	m.activePanel = LogsPanel
	m.ApplyContainersLoaded([]domain.Container{{
		ID: "container-1", Name: "redis", Image: "redis", Status: domain.StatusRunning,
	}})
	m.logLines = []string{"existing log"}

	if cmd := m.handleKeys(keyMsg("c")); cmd != nil {
		t.Fatal("clear key must wait for confirmation")
	}
	if !m.confirmClearLogs {
		t.Fatal("clear confirmation modal was not opened")
	}
	if p.clearCalls != 0 {
		t.Fatal("engine clear ran before confirmation")
	}

	cmd := m.handleKeys(keyMsg("y"))
	if cmd == nil {
		t.Fatal("confirmation did not start clear command")
	}
	msg := cmd()
	if p.clearCalls != 1 {
		t.Fatalf("ClearLogs calls = %d, want 1", p.clearCalls)
	}
	if _, ok := msg.(containerLogsClearedMsg); !ok {
		t.Fatalf("clear command message = %T", msg)
	}
}

func TestClearFailureKeepsVisibleLogs(t *testing.T) {
	p := &logProvider{clearErr: errors.New("permission denied")}
	m := NewModel(p, "podman")
	m.showSplash = false
	m.activePanel = LogsPanel
	m.ApplyContainersLoaded([]domain.Container{{ID: "container-1", Name: "db"}})
	m.logLines = []string{"existing log"}

	m.confirmClearLogs = true
	msg := m.handleKeys(keyMsg("y"))()
	m.handleContainerLogsCleared(msg.(containerLogsClearedMsg))

	if got := strings.Join(m.logLines, "\n"); got != "existing log" {
		t.Fatalf("log buffer = %q, want preserved logs", got)
	}
}

func TestContainerIconsAreReadableByService(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"docker", "🐳"},
		{"ministask", "☁️"},
		{"postgre", "🐘"},
		{"openfga", "🔐"},
		{"redis", "⚡"},
		{"podman", "🦭"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := containerIconAndLabel(domain.Container{Name: tt.name})
			if got != tt.want {
				t.Fatalf("containerIconAndLabel(%q) icon = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestLogTailFitsViewport(t *testing.T) {
	p := &logProvider{}
	m := NewModel(p, "docker")
	m.logViewport.Height = 22

	cmd := m.startLogStream("container-1")
	if cmd == nil {
		t.Fatal("startLogStream() returned nil command")
	}
	msg := cmd()
	opened := msg.(logStreamOpenedMsg)
	defer opened.reader.Close()
	if p.logsTail != 22 {
		t.Fatalf("logs tail = %d, want viewport height 22", p.logsTail)
	}
	if m.logContainerID != "container-1" {
		t.Fatalf("stream container = %q, want container-1", m.logContainerID)
	}
}

func TestLogSearchAndHighlight(t *testing.T) {
	p := &logProvider{}
	m := NewModel(p, "docker")
	m.showSplash = false
	m.activePanel = LogsPanel
	m.logViewport.Width = 80
	m.logViewport.Height = 20
	m.logLines = []string{
		"2026-08-11 INFO server starting",
		"2026-08-11 ERROR database connection failed",
		"2026-08-11 WARN retry attempt 1",
	}

	m.refreshLogViewportContent()
	contentAll := m.logViewport.View()
	if !strings.Contains(contentAll, "ERROR") {
		t.Fatalf("expected log viewport to contain ERROR")
	}

	m.logSearchQuery = "database"
	m.refreshLogViewportContent()
	contentFiltered := m.logViewport.View()
	if !strings.Contains(contentFiltered, "database") {
		t.Fatalf("expected filtered content to contain database")
	}
	if strings.Contains(contentFiltered, "retry attempt") {
		t.Fatalf("filtered content should not contain non-matching log lines")
	}
}

func TestExportLogsToFile(t *testing.T) {
	p := &logProvider{}
	m := NewModel(p, "docker")
	m.showSplash = false
	m.activePanel = LogsPanel
	m.ApplyContainersLoaded([]domain.Container{{
		ID: "container-1", Name: "my-app", Status: domain.StatusRunning,
	}})
	m.logLines = []string{"line 1", "line 2", "ERROR failed"}

	m.exportLogsToFile()
	if !strings.HasPrefix(m.statusMsg, "✓ Saved") {
		t.Fatalf("statusMsg = %q, want saved success message", m.statusMsg)
	}

	dateStr := time.Now().Format("2006-01-02")
	expectedFile := "./my-app-logs-" + dateStr + ".log"
	data, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("failed to read exported file %s: %v", expectedFile, err)
	}
	defer os.Remove(expectedFile)

	if !strings.Contains(string(data), "ERROR failed") {
		t.Fatalf("exported log file content = %q, missing expected lines", string(data))
	}
}

func TestToggleTimestamps(t *testing.T) {
	p := &logProvider{}
	m := NewModel(p, "docker")
	m.showSplash = false
	m.activePanel = LogsPanel
	m.ApplyContainersLoaded([]domain.Container{{
		ID: "container-1", Name: "app", Status: domain.StatusRunning,
	}})

	if m.showTimestamps {
		t.Fatal("expected showTimestamps to be false initially")
	}

	cmd := m.handleLogsKeys(keyMsg("t"))
	if !m.showTimestamps {
		t.Fatal("expected showTimestamps to be true after pressing t")
	}
	if cmd == nil {
		t.Fatal("expected command to restart log stream with timestamps")
	}
	msg := cmd()
	if opened, ok := msg.(logStreamOpenedMsg); ok {
		opened.reader.Close()
	}
	if !p.timestamps {
		t.Fatal("expected logs provider to receive timestamps = true")
	}
}

func keyMsg(key string) tea.KeyMsg {
	if key == "enter" {
		return tea.KeyMsg{Type: tea.KeyEnter}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
}
