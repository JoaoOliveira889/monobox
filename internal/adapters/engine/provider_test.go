package engine_test

import (
	"os/exec"
	"testing"

	"github.com/JoaoOliveira889/monobox/internal/adapters/engine"
	"github.com/JoaoOliveira889/monobox/internal/domain"
)

func TestDockerProvider_EngineName(t *testing.T) {
	p := engine.NewDockerProvider()
	if p.EngineName() != domain.EngineDocker {
		t.Errorf("EngineName() = %q, want %q", p.EngineName(), domain.EngineDocker)
	}
}

func TestDockerProvider_List_ReturnsSlice(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not in PATH")
	}
	p := engine.NewDockerProvider()
	containers, err := p.List()
	if err != nil {
		t.Skipf("docker daemon not available: %v", err)
	}
	// Just asserting it's a slice (may be empty in CI).
	_ = containers
}

func TestPodmanProvider_EngineName(t *testing.T) {
	p := engine.NewPodmanProvider()
	if p.EngineName() != domain.EnginePodman {
		t.Errorf("EngineName() = %q, want %q", p.EngineName(), domain.EnginePodman)
	}
}
