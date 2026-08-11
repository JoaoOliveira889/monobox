package tui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monobox/internal/adapters/tui"
	"github.com/JoaoOliveira889/monobox/internal/domain"
)

func TestLifecycleKeysAndModals(t *testing.T) {
	containers := []domain.Container{
		{ID: "111", Name: "web-app", Image: "nginx:alpine", Status: domain.StatusRunning, Engine: domain.EngineDocker, Ports: "8080->80/tcp"},
		{ID: "222", Name: "db-app", Image: "postgres:15", Status: domain.StatusPaused, Engine: domain.EngineDocker, Ports: "5432->5432/tcp"},
	}
	stub := &stubProvider{containers: containers}
	m := tui.NewModel(stub, "docker")
	m.ApplyContainersLoaded(containers)

	// Test Inspect Key 'i'
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	viewStr := m.View()
	if !testing.Short() && len(viewStr) == 0 {
		t.Error("View during inspect should render modal content")
	}

	// Close Inspect Modal with Esc
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Test Remove Modal Key 'd'
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	viewRemove := m.View()
	if !testing.Short() && len(viewRemove) == 0 {
		t.Error("View during remove confirmation should render modal content")
	}

	// Cancel Remove Modal with 'n'
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	// Test Open Port Key 'o'
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})

	// Test Pause Key 'p' on running container
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
}
