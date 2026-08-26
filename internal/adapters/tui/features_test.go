package tui_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monobox/internal/adapters/tui"
	"github.com/JoaoOliveira889/monobox/internal/domain"
)

func TestNewFeaturesModalsAndShortcuts(t *testing.T) {
	containers := []domain.Container{
		{ID: "c1", Name: "web", Image: "nginx:latest", Status: domain.StatusRunning, ComposeProject: "my-stack", Ports: "8080->80/tcp"},
		{ID: "c2", Name: "db", Image: "postgres:15", Status: domain.StatusRunning, ComposeProject: "my-stack", Ports: "5432->5432/tcp"},
	}
	stub := &stubProvider{containers: containers}
	m := tui.NewModel(stub, "docker")
	m.ApplyContainersLoaded(containers)

	// Dismiss splash screen by passing WindowSize or containersLoaded
	res, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	mPtr := res.(*tui.Model)
	// Apply loaded containers to dismiss splash
	mPtr.ApplyContainersLoaded(containers)

	// 1. Test Prune Modal ('P')
	res, _ = mPtr.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'P'}})
	mPtr = res.(*tui.Model)
	viewPrune := mPtr.View()
	if !strings.Contains(viewPrune, "System Prune") {
		t.Errorf("expected view to contain 'System Prune', got %s", viewPrune)
	}

	// Toggle --all in prune modal ('a')
	res, _ = mPtr.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	mPtr = res.(*tui.Model)

	// Cancel Prune with 'n'
	res, _ = mPtr.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	mPtr = res.(*tui.Model)

	// 2. Test Settings Modal ('S')
	res, _ = mPtr.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	mPtr = res.(*tui.Model)
	viewSettings := mPtr.View()
	if !strings.Contains(viewSettings, "SETTINGS") {
		t.Errorf("expected view to contain 'SETTINGS', got %s", viewSettings)
	}

	// Close Settings with Esc
	res, _ = mPtr.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mPtr = res.(*tui.Model)

	// 3. Test Clipboard copy ('y' and 'Y')
	res, _ = mPtr.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	mPtr = res.(*tui.Model)
	res, _ = mPtr.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Y'}})
	mPtr = res.(*tui.Model)

	// 4. Test Compose Up and Down on Project Header
	mPtr.SetCursor(0)
	res, _ = mPtr.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	mPtr = res.(*tui.Model)
	res, _ = mPtr.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})
	mPtr = res.(*tui.Model)
	res, _ = mPtr.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}) // cancel compose down
	mPtr = res.(*tui.Model)

	// 5. Test Severity filter in logs ('!')
	res, _ = mPtr.Update(tea.KeyMsg{Type: tea.KeyTab})
	mPtr = res.(*tui.Model)
	res, _ = mPtr.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}}) // INFO+
	mPtr = res.(*tui.Model)
	res, _ = mPtr.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}}) // WARN+
	mPtr = res.(*tui.Model)
	res, _ = mPtr.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}}) // ERROR only
	mPtr = res.(*tui.Model)
	res, _ = mPtr.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}}) // ALL
	mPtr = res.(*tui.Model)

	// 6. Test Match Navigation ('n', 'N')
	res, _ = mPtr.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	mPtr = res.(*tui.Model)
	res, _ = mPtr.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})
	_ = res.(*tui.Model)
}
