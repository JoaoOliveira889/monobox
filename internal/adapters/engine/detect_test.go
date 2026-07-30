package engine_test

import (
	"testing"

	"github.com/JoaoOliveira889/monobox/internal/adapters/engine"
	"github.com/JoaoOliveira889/monobox/internal/domain"
)

// TestDetectEngine_ReturnsProvider verifies that when docker is found in PATH and
// responds to "docker info", the engine returned is Docker.
func TestDetectEngine_ReturnsProvider(t *testing.T) {
	// Skip if neither docker nor podman is available or daemon is unreachable.
	provider, name, err := engine.DetectEngine()
	if err != nil {
		t.Skipf("DetectEngine() skipped: %v", err)
	}
	if provider == nil {
		t.Fatal("DetectEngine() returned nil provider")
	}
	if name != string(domain.EngineDocker) && name != string(domain.EnginePodman) {
		t.Errorf("DetectEngine() name = %q, want docker or podman", name)
	}
}

// TestDetectEngine_EngineName verifies that provider.EngineName() matches the
// name string returned by DetectEngine.
func TestDetectEngine_EngineName(t *testing.T) {
	provider, name, err := engine.DetectEngine()
	if err != nil {
		t.Skipf("DetectEngine() skipped: %v", err)
	}
	if string(provider.EngineName()) != name {
		t.Errorf("provider.EngineName() = %q, want %q", provider.EngineName(), name)
	}
}
