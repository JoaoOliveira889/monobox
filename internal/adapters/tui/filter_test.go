package tui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monobox/internal/adapters/tui"
	"github.com/JoaoOliveira889/monobox/internal/domain"
)

func TestContainerFiltering(t *testing.T) {
	containers := []domain.Container{
		{ID: "1", Name: "api-gateway", Image: "nginx:alpine", Status: domain.StatusRunning, Engine: domain.EngineDocker, Ports: "8080->80/tcp"},
		{ID: "2", Name: "auth-db", Image: "postgres:15", Status: domain.StatusRunning, Engine: domain.EngineDocker, Ports: "5432->5432/tcp"},
		{ID: "3", Name: "cache-redis", Image: "redis:7", Status: domain.StatusExited, Engine: domain.EngineDocker, Ports: "6379->6379/tcp"},
	}
	stub := &stubProvider{containers: containers}
	m := tui.NewModel(stub, "docker")
	m.ApplyContainersLoaded(containers)

	if len(m.FilteredContainers()) != 3 {
		t.Fatalf("expected 3 containers without filter, got %d", len(m.FilteredContainers()))
	}

	// Test filter by name
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a', 'p', 'i'}})

	filtered := m.FilteredContainers()
	if len(filtered) != 1 || filtered[0].Name != "api-gateway" {
		t.Errorf("expected 1 container (api-gateway), got %d (%v)", len(filtered), filtered)
	}

	// Test clear filter with Esc
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if len(m.FilteredContainers()) != 3 {
		t.Errorf("expected filter to be cleared with Esc, got %d containers", len(m.FilteredContainers()))
	}

	// Test filter by port
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5', '4', '3', '2'}})
	filtered = m.FilteredContainers()
	if len(filtered) != 1 || filtered[0].Name != "auth-db" {
		t.Errorf("expected 1 container matching port 5432, got %d", len(filtered))
	}

	// Test filter by status (exited)
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e', 'x', 'i', 't'}})
	filtered = m.FilteredContainers()
	if len(filtered) != 1 || filtered[0].Name != "cache-redis" {
		t.Errorf("expected 1 container with exited status, got %d", len(filtered))
	}
}
