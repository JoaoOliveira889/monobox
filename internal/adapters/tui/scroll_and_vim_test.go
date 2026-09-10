package tui

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monobox/internal/domain"
)

func createTestModelWithContainers(n int) *Model {
	val := NewModel(nil, "docker")
	m := &val
	m.width = 100
	m.height = 40
	m.showSplash = false

	var list []domain.Container
	for i := 0; i < n; i++ {
		list = append(list, domain.Container{
			ID:     fmt.Sprintf("c-%d", i),
			Name:   fmt.Sprintf("container-%d", i),
			Image:  "test:latest",
			Status: domain.StatusRunning,
		})
	}
	m.ApplyContainersLoaded(list)
	m.cursor = 0
	return m
}

func TestMouseScrollOneItemPerNotch(t *testing.T) {
	m := createTestModelWithContainers(10)

	if m.cursor != 0 {
		t.Fatalf("expected initial cursor 0, got %d", m.cursor)
	}

	// 1. Mouse wheel down (Press event) -> cursor moves exactly +1
	m.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonWheelDown,
		Action: tea.MouseActionPress,
	})
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after wheel down, got %d", m.cursor)
	}

	// 2. Mouse wheel release event MUST be ignored (should NOT advance cursor)
	m.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonWheelDown,
		Action: tea.MouseActionRelease,
	})
	if m.cursor != 1 {
		t.Errorf("expected cursor to remain 1 after wheel release event, got %d", m.cursor)
	}

	// 3. Rapid fire within 45ms debounce window MUST be ignored
	m.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonWheelDown,
		Action: tea.MouseActionPress,
	})
	if m.cursor != 1 {
		t.Errorf("expected cursor to remain 1 due to 45ms debounce, got %d", m.cursor)
	}

	// 4. After debounce elapsed, wheel down advances by 1
	time.Sleep(50 * time.Millisecond)
	m.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonWheelDown,
		Action: tea.MouseActionPress,
	})
	if m.cursor != 2 {
		t.Errorf("expected cursor 2 after debounce elapsed, got %d", m.cursor)
	}

	// 5. Wheel up advances cursor up by exactly 1
	time.Sleep(50 * time.Millisecond)
	m.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonWheelUp,
		Action: tea.MouseActionPress,
	})
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after wheel up, got %d", m.cursor)
	}
}

func TestVimKeysNavigation(t *testing.T) {
	m := createTestModelWithContainers(20)

	if m.cursor != 0 {
		t.Fatalf("expected initial cursor 0, got %d", m.cursor)
	}

	// ctrl+d (HalfPageDown): jumps 5 items down
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	if m.cursor != 5 {
		t.Errorf("expected cursor 5 after ctrl+d, got %d", m.cursor)
	}

	// ctrl+d again: jumps to 10
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	if m.cursor != 10 {
		t.Errorf("expected cursor 10 after second ctrl+d, got %d", m.cursor)
	}

	// ctrl+u (HalfPageUp): jumps 5 items up
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	if m.cursor != 5 {
		t.Errorf("expected cursor 5 after ctrl+u, got %d", m.cursor)
	}

	// G (End): jumps to last item (index 19)
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	if m.cursor != 19 {
		t.Errorf("expected cursor 19 after G, got %d", m.cursor)
	}

	// g (Top): jumps to first item (index 0)
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after g, got %d", m.cursor)
	}
}

func TestCtrlPTogglesHelp(t *testing.T) {
	m := createTestModelWithContainers(5)

	if m.showHelp {
		t.Fatal("expected showHelp false initially")
	}

	// ctrl+p opens help
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	if !m.showHelp {
		t.Errorf("expected showHelp true after ctrl+p")
	}

	// esc closes help
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.showHelp {
		t.Errorf("expected showHelp false after esc")
	}
}

func TestOpenBrowserCmdPortValidation(t *testing.T) {
	invalidPorts := []string{"", "abc", "-1", "0", "65536", "99999", "80; rm -rf /"}
	for _, port := range invalidPorts {
		cmd := openBrowserCmd(port)
		if cmd == nil {
			t.Fatalf("expected non-nil tea.Cmd for port %q", port)
		}
		msg := cmd()
		if msg != nil {
			t.Errorf("expected nil msg for invalid port %q, got %v", port, msg)
		}
	}
}
