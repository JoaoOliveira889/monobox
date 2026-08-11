package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadSave(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "monobox-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override Home directory for testing
	t.Setenv("HOME", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Theme != "Tokyo Night" {
		t.Errorf("expected default theme Tokyo Night, got %s", cfg.Theme)
	}
	if cfg.MetricsInterval != 5 {
		t.Errorf("expected default metrics interval 5, got %d", cfg.MetricsInterval)
	}
	if cfg.LogLineLimit != 100 {
		t.Errorf("expected default log line limit 100, got %d", cfg.LogLineLimit)
	}
	if cfg.ShowTimestamps != false {
		t.Errorf("expected default show_timestamps false, got %t", cfg.ShowTimestamps)
	}

	// Modify and save
	cfg.Theme = "Dracula"
	cfg.MetricsInterval = 10
	cfg.LogLineLimit = 250
	cfg.ShowTimestamps = true

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	expectedPath := filepath.Join(tmpDir, ".config", "monobox", "config.yaml")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("expected config file at %s, but does not exist", expectedPath)
	}

	// Reload
	reloaded, err := Load()
	if err != nil {
		t.Fatalf("Load() reloaded error = %v", err)
	}
	if reloaded.Theme != "Dracula" {
		t.Errorf("expected theme Dracula, got %s", reloaded.Theme)
	}
	if reloaded.MetricsInterval != 10 {
		t.Errorf("expected metrics interval 10, got %d", reloaded.MetricsInterval)
	}
	if reloaded.LogLineLimit != 250 {
		t.Errorf("expected log line limit 250, got %d", reloaded.LogLineLimit)
	}
	if reloaded.ShowTimestamps != true {
		t.Errorf("expected show_timestamps true, got %t", reloaded.ShowTimestamps)
	}
}
