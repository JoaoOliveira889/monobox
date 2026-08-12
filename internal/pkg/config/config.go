package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Theme           string `json:"theme" yaml:"theme"`
	MetricsInterval int    `json:"metrics_interval" yaml:"metrics_interval"` // in seconds
	LogLineLimit    int    `json:"log_line_limit" yaml:"log_line_limit"`
	LogTailLimit    int    `json:"log_tail_limit" yaml:"log_tail_limit"`
	ShowTimestamps  bool   `json:"show_timestamps" yaml:"show_timestamps"`
}

func DefaultConfig() Config {
	return Config{
		Theme:           "Tokyo Night",
		MetricsInterval: 5,
		LogLineLimit:    100,
		LogTailLimit:    100,
		ShowTimestamps:  false,
	}
}

func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "monobox", "config.yaml"), nil
}

func Load() (Config, error) {
	cfg := DefaultConfig()
	path, err := GetConfigPath()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			_ = Save(cfg)
			return cfg, nil
		}
		return cfg, err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, "\"'")

		switch key {
		case "theme":
			if val != "" {
				cfg.Theme = val
			}
		case "metrics_interval":
			if n, err := strconv.Atoi(val); err == nil && n > 0 {
				cfg.MetricsInterval = n
			}
		case "log_line_limit":
			if n, err := strconv.Atoi(val); err == nil && n > 0 {
				cfg.LogLineLimit = n
			}
		case "log_tail_limit":
			if n, err := strconv.Atoi(val); err == nil && n > 0 {
				cfg.LogTailLimit = n
			}
		case "show_timestamps":
			if b, err := strconv.ParseBool(val); err == nil {
				cfg.ShowTimestamps = b
			}
		}
	}

	return cfg, nil
}

func Save(cfg Config) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	content := fmt.Sprintf("# Monobox Configuration\n"+
		"theme: %s\n"+
		"metrics_interval: %d\n"+
		"log_line_limit: %d\n"+
		"log_tail_limit: %d\n"+
		"show_timestamps: %t\n",
		cfg.Theme,
		cfg.MetricsInterval,
		cfg.LogLineLimit,
		cfg.LogTailLimit,
		cfg.ShowTimestamps,
	)

	return os.WriteFile(path, []byte(content), 0644)
}
