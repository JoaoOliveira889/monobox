package tui

import (
	"testing"

	"github.com/JoaoOliveira889/monobox/internal/domain"
)

func TestDetectPortConflict(t *testing.T) {
	c1 := containerItem{
		Container: domain.Container{
			ID:     "c1",
			Name:   "web-app-1",
			Status: domain.StatusRunning,
			Ports:  "0.0.0.0:8080->80/tcp",
		},
	}
	c2 := containerItem{
		Container: domain.Container{
			ID:     "c2",
			Name:   "web-app-2",
			Status: domain.StatusExited,
			Ports:  "0.0.0.0:8080->80/tcp",
		},
	}

	m := Model{
		containers: []containerItem{c1, c2},
	}

	hasConflict, otherName, port := m.detectPortConflict(&c2)
	if !hasConflict {
		t.Fatal("expected port conflict between c2 and running c1")
	}
	if otherName != "web-app-1" || port != "8080" {
		t.Errorf("expected conflict with web-app-1 on port 8080, got %s on port %s", otherName, port)
	}
}
