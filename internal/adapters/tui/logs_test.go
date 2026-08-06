package tui

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monobox/internal/domain"
)

type logProvider struct {
	clearCalls int
	clearErr   error
	logsTail   int
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
func (p *logProvider) Start(string) error   { return nil }
func (p *logProvider) Stop(string) error    { return nil }
func (p *logProvider) Restart(string) error { return nil }
func (p *logProvider) Logs(_ context.Context, _ string, tail int, _ bool) (io.ReadCloser, error) {
	p.logsTail = tail
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

func keyMsg(key string) tea.KeyMsg {
	if key == "enter" {
		return tea.KeyMsg{Type: tea.KeyEnter}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
}
