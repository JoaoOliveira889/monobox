package tui_test

import (
	"context"
	"io"
	"os/exec"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monobox/internal/adapters/tui"
	"github.com/JoaoOliveira889/monobox/internal/domain"
)

// stubProvider satisfies domain.ContainerProvider without real engine calls.
type stubProvider struct {
	containers []domain.Container
}

func (s *stubProvider) EngineName() domain.Engine { return domain.EngineDocker }
func (s *stubProvider) List() ([]domain.Container, error) {
	return s.containers, nil
}
func (s *stubProvider) Stats() (map[string]domain.ContainerStats, error) {
	return nil, nil
}
func (s *stubProvider) ClearLogs(_ string) error { return nil }
func (s *stubProvider) Start(_ string) error     { return nil }
func (s *stubProvider) Stop(_ string) error    { return nil }
func (s *stubProvider) Restart(_ string) error { return nil }
func (s *stubProvider) Pause(_ string) error   { return nil }
func (s *stubProvider) Unpause(_ string) error { return nil }
func (s *stubProvider) Remove(_ string, _ bool) error {
	return nil
}
func (s *stubProvider) Inspect(_ string) (string, error) { return "{}", nil }
func (s *stubProvider) ExecCmd(_ string) *exec.Cmd {
	return exec.Command("echo", "test")
}
func (s *stubProvider) Logs(_ context.Context, _ string, _ int, _ bool, _ bool) (io.ReadCloser, error) {
	return nil, nil
}
func (s *stubProvider) SystemPrune(_ bool) (string, error) { return "Total reclaimed space: 0B", nil }
func (s *stubProvider) ComposeUp(_ string) error           { return nil }
func (s *stubProvider) ComposeDown(_ string) error         { return nil }


// TestNewModel verifies the initial model state.
func TestNewModel(t *testing.T) {
	stub := &stubProvider{}
	m := tui.NewModel(stub, "docker")

	if m.ActivePanel() != tui.ListPanel {
		t.Errorf("ActivePanel = %v, want ListPanel", m.ActivePanel())
	}
	if m.Cursor() != 0 {
		t.Errorf("Cursor = %d, want 0", m.Cursor())
	}
	if !m.LogFollow() {
		t.Error("LogFollow should be true by default")
	}
}

// TestApplyContainersLoaded verifies cursor preservation on list refresh.
func TestApplyContainersLoaded(t *testing.T) {
	containers := []domain.Container{
		{ID: "aaa", Name: "web", Image: "nginx", Status: domain.StatusRunning, Engine: domain.EngineDocker},
		{ID: "bbb", Name: "db", Image: "postgres", Status: domain.StatusExited, Engine: domain.EngineDocker},
	}
	stub := &stubProvider{containers: containers}
	m := tui.NewModel(stub, "docker")
	m.SetCursor(1)

	// Simulate a container list reload while cursor is at index 1.
	m.ApplyContainersLoaded(containers)
	if m.Cursor() != 1 {
		t.Errorf("cursor should be preserved at 1, got %d", m.Cursor())
	}
}

// TestCursorBounds verifies cursor clamps when the list shrinks.
func TestCursorBounds(t *testing.T) {
	containers := []domain.Container{
		{ID: "aaa", Name: "web", Image: "nginx", Status: domain.StatusRunning, Engine: domain.EngineDocker},
	}
	stub := &stubProvider{containers: containers}
	m := tui.NewModel(stub, "docker")
	m.SetCursor(5) // out of bounds

	m.ApplyContainersLoaded(containers)
	if m.Cursor() != 0 {
		t.Errorf("cursor should clamp to 0, got %d", m.Cursor())
	}
}

func TestViewWithPortsAndStats(t *testing.T) {
	containers := []domain.Container{
		{
			ID:     "aaa",
			Name:   "web",
			Image:  "nginx",
			Status: domain.StatusRunning,
			Health: domain.HealthHealthy,
			Engine: domain.EngineDocker,
			CPU:    "1.50%",
			Mem:    "100.0MiB / 500MiB (2.0%)",
			Ports:  "8080->80/tcp",
		},
	}
	stub := &stubProvider{containers: containers}
	m := tui.NewModel(stub, "docker")
	m.ApplyContainersLoaded(containers)

	viewStr := m.View()
	_ = viewStr
}

func TestStatsHistoryAccumulation(t *testing.T) {
	stub := &stubProvider{}
	m := tui.NewModel(stub, "docker")

	step1 := []domain.Container{
		{ID: "c1", Name: "app", Status: domain.StatusRunning},
	}
	stats1 := map[string]domain.ContainerStats{
		"c1": {CPU: "5.0%", MemPerc: "10.0%"},
	}

	step2 := []domain.Container{
		{ID: "c1", Name: "app", Status: domain.StatusRunning},
	}
	stats2 := map[string]domain.ContainerStats{
		"c1": {CPU: "15.0%", MemPerc: "12.0%"},
	}

	m.ApplyContainersLoaded(step1)
	m.ApplyStats(stats1)
	
	m.ApplyContainersLoaded(step2)
	m.ApplyStats(stats2)

	cpu, mem := m.GetStatsHistory("c1")
	if len(cpu) != 2 || cpu[0] != 5.0 || cpu[1] != 15.0 {
		t.Errorf("unexpected CPU history: %v", cpu)
	}
	if len(mem) != 2 || mem[0] != 10.0 || mem[1] != 12.0 {
		t.Errorf("unexpected Mem history: %v", mem)
	}
}

func TestViewWithHealthBadges(t *testing.T) {
	containers := []domain.Container{
		{ID: "c1", Name: "healthy-app", Status: domain.StatusRunning, Health: domain.HealthHealthy},
		{ID: "c2", Name: "unhealthy-app", Status: domain.StatusRunning, Health: domain.HealthUnhealthy},
		{ID: "c3", Name: "starting-app", Status: domain.StatusRunning, Health: domain.HealthStarting},
	}
	stub := &stubProvider{containers: containers}
	m := tui.NewModel(stub, "docker")
	m.ApplyContainersLoaded(containers)

	viewStr := m.View()
	if viewStr == "" {
		t.Error("expected non-empty view string")
	}
}

func TestHelpModalToggle(t *testing.T) {
	stub := &stubProvider{}
	m := tui.NewModel(stub, "docker")

	// Simulate pressing '?' key to open help
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	viewStr := m.View()
	if viewStr == "" {
		t.Error("expected non-empty help modal view")
	}
}


