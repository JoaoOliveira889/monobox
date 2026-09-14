package tui_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monobox/internal/adapters/tui"
	"github.com/JoaoOliveira889/monobox/internal/domain"
)

func TestHelpSearchAndEscHandling(t *testing.T) {
	containers := []domain.Container{
		{ID: "1", Name: "api-gateway", Image: "nginx:alpine", Status: domain.StatusRunning, Engine: domain.EngineDocker},
	}
	stub := &stubProvider{containers: containers}
	m := tui.NewModel(stub, "docker")
	m.ApplyContainersLoaded(containers)

	// Set window size
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// Open help modal via "?"
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})

	view := m.View()
	if !strings.Contains(view, "SHORTCUTS") {
		t.Fatalf("expected help modal to be open, view: %s", view)
	}
	if !strings.Contains(view, "shortcuts") {
		t.Errorf("expected count badge in help view, got: %s", view)
	}

	// Filter shortcuts for "exec"
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e', 'x', 'e', 'c'}})
	view = m.View()
	if !strings.Contains(view, "CONTAINER ACTIONS") {
		t.Errorf("expected filtered help to contain CONTAINER ACTIONS, got: %s", view)
	}
	if !strings.Contains(view, "Exec shell") {
		t.Errorf("expected filtered help to contain 'Exec shell', got: %s", view)
	}
	if strings.Contains(view, "COMPOSE STACKS") {
		t.Errorf("expected filtered help to exclude 'COMPOSE STACKS' when searching 'exec'")
	}

	// Filter with non-matching query
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z', 'z', 'z', 'q'}})
	view = m.View()
	if !strings.Contains(view, "No shortcuts matching") {
		t.Errorf("expected empty search state message, got: %s", view)
	}

	// First Esc should clear the search input, but keep help modal open
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	view = m.View()
	if !strings.Contains(view, "SHORTCUTS") {
		t.Errorf("expected help modal to remain open on first Esc")
	}
	if strings.Contains(view, "No shortcuts matching") {
		t.Errorf("expected search to be cleared on first Esc")
	}

	// Second Esc should close the help modal
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	view = m.View()
	if strings.Contains(view, "SHORTCUTS") {
		t.Errorf("expected help modal to close on second Esc")
	}

	// Re-open with ctrl+p and close with "?"
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	view = m.View()
	if !strings.Contains(view, "SHORTCUTS") {
		t.Errorf("expected help modal to open via ctrl+p")
	}

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	view = m.View()
	if strings.Contains(view, "SHORTCUTS") {
		t.Errorf("expected help modal to close via '?'")
	}
}

func TestContainerFilterNoDoubleEllipsisOrBrokenLayout(t *testing.T) {
	containers := []domain.Container{
		{ID: "aea8c58641c5", Name: "openfga-postgres", Image: "postgres:17", Status: domain.StatusExited, Engine: domain.EngineDocker},
		{ID: "b9c1d2e3f4a5", Name: "openfga-service", Image: "openfga/openfga", Status: domain.StatusRunning, Engine: domain.EngineDocker},
		{ID: "c1a2b3c4d5e6", Name: "redis-cache", Image: "redis:7", Status: domain.StatusRunning, Engine: domain.EngineDocker},
	}
	stub := &stubProvider{containers: containers}
	m := tui.NewModel(stub, "docker")
	m.ApplyContainersLoaded(containers)

	// Simulate window size as in screenshot (~80x24)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Activate container filter "/" and type "po"
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p', 'o'}})

	view := m.View()

	// Should not have double ellipsis "……"
	if strings.Contains(view, "……") {
		t.Errorf("view contains double ellipsis '……' indicating broken rune truncation:\n%s", view)
	}

	// Should correctly display the matching container
	if !strings.Contains(view, "openfga-postgres") {
		t.Errorf("expected filtered view to contain 'openfga-postgres', got:\n%s", view)
	}

	// Should not show redis-cache
	if strings.Contains(view, "redis-cache") {
		t.Errorf("expected filtered view to exclude non-matching 'redis-cache'")
	}
}
