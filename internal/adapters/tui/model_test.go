package tui_test

import (
	"context"
	"io"
	"testing"

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
func (s *stubProvider) Logs(_ context.Context, _ string, _ int, _ bool) (io.ReadCloser, error) {
	return nil, nil
}

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
			Engine: domain.EngineDocker,
			CPU:    "1.50%",
			Mem:    "100.0MiB / 500MiB",
			Ports:  "8080->80/tcp",
		},
	}
	stub := &stubProvider{containers: containers}
	m := tui.NewModel(stub, "docker")
	m.ApplyContainersLoaded(containers)

	viewStr := m.View()
	_ = viewStr
}
