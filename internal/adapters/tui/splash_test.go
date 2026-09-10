package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSplashInitialStateAndRendering(t *testing.T) {
	val := NewModel(nil, "docker")
	m := &val
	m.width = 100
	m.height = 30

	if !m.showSplash {
		t.Fatal("expected showSplash true upon initialization")
	}
	if m.splashReady {
		t.Fatal("expected splashReady false initially")
	}

	splashContent := m.renderSplash()
	expectedSubstrings := []string{
		"Mono",
		"Box",
		"container dashboard",
		"Docker & Podman container manager",
		"scanning",
		Version,
	}

	for _, sub := range expectedSubstrings {
		if !strings.Contains(splashContent, sub) {
			t.Errorf("renderSplash should contain %q, but got:\n%s", sub, splashContent)
		}
	}
}

func TestSplashStaysVisibleUntilMinDuration(t *testing.T) {
	val := NewModel(nil, "docker")
	m := &val
	m.width = 100
	m.height = 30

	// Handle containers loaded immediately
	m.Update(containersLoadedMsg{})

	// Should still be showing splash because 650ms haven't elapsed
	if !m.showSplash {
		t.Error("showSplash should remain true immediately after containers loaded to preserve branding")
	}
	if !m.splashReady {
		t.Error("splashReady should be true after containers loaded")
	}

	// Ticking before minDuration still keeps splash
	m.Update(splashTickMsg{})
	if !m.showSplash {
		t.Error("showSplash should remain true before min duration expires")
	}

	// Simulating elapsed duration past splashMinDuration
	m.splashStartedAt = time.Now().Add(-700 * time.Millisecond)
	m.Update(splashTickMsg{})

	if m.showSplash {
		t.Error("showSplash should be false after minDuration has passed and splashReady is true")
	}
}

func TestSplashDismissOnKeyWhenReady(t *testing.T) {
	val := NewModel(nil, "docker")
	m := &val
	m.width = 100
	m.height = 30

	// If not ready, key press should not dismiss
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !m.showSplash {
		t.Error("key press should not dismiss splash before containers are ready")
	}

	// When containers are loaded (ready), key press should dismiss immediately
	m.Update(containersLoadedMsg{})
	if !m.splashReady {
		t.Fatal("expected splashReady true")
	}

	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.showSplash {
		t.Error("key press should dismiss splash immediately when splashReady is true")
	}
}
