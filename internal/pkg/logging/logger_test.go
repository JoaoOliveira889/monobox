package logging

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoggerInitAndLogging(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "monobox-log-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Setenv("HOME", tempDir)

	Init()
	defer Close()

	Info("test info message", "key", "value")
	Warn("test warning message", "status", 1)
	Error("test error message", "error", "some err")

	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = filepath.Join(tempDir, ".config")
	}
	logPath := filepath.Join(configDir, "monobox", logFileName)
	fi, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("expected log file to exist at %s: %v", logPath, err)
	}

	if fi.Size() == 0 {
		t.Errorf("expected log file to have content, got 0 bytes")
	}

	// Verify permissions (0600)
	perm := fi.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected file permissions 0600, got %o", perm)
	}
}
