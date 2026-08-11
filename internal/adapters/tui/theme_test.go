package tui_test

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monobox/internal/adapters/tui"
	"github.com/JoaoOliveira889/monobox/internal/pkg/config"
	"github.com/JoaoOliveira889/monobox/internal/pkg/ui"
)

func TestThemeMenuNavigationAndSelection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "monobox-tui-theme-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Setenv("HOME", tmpDir)

	provider := &stubProvider{}
	m := tui.NewModel(provider, "docker")

	// Press 'T' to open Theme Menu
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'T'}})
	mPtr, ok := res.(*tui.Model)
	if !ok {
		t.Fatalf("expected *tui.Model return, got %T", res)
	}

	// Move cursor down
	res, _ = mPtr.Update(tea.KeyMsg{Type: tea.KeyDown})
	mPtr, ok = res.(*tui.Model)
	if !ok {
		t.Fatalf("expected *tui.Model return, got %T", res)
	}

	// Press Enter to select
	res, _ = mPtr.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_, ok = res.(*tui.Model)
	if !ok {
		t.Fatalf("expected *tui.Model return, got %T", res)
	}

	// Verify config saved on disk
	reloaded, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}
	if reloaded.Theme != ui.Themes[1].Name && reloaded.Theme != ui.Themes[0].Name {
		t.Errorf("expected valid theme from Themes slice saved, got %s", reloaded.Theme)
	}
}

func TestThemeMenuCancel(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "monobox-tui-theme-cancel-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Setenv("HOME", tmpDir)

	provider := &stubProvider{}
	m := tui.NewModel(provider, "docker")

	// Open Theme Menu
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'T'}})
	mPtr := res.(*tui.Model)

	// Move cursor down
	res, _ = mPtr.Update(tea.KeyMsg{Type: tea.KeyDown})
	mPtr = res.(*tui.Model)

	// Cancel with Esc
	res, _ = mPtr.Update(tea.KeyMsg{Type: tea.KeyEsc})
	_ = res.(*tui.Model)

	// Reload config to confirm nothing saved
	reloaded, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if reloaded.Theme != "Tokyo Night" {
		t.Errorf("expected theme Tokyo Night on cancel, got %s", reloaded.Theme)
	}
}
