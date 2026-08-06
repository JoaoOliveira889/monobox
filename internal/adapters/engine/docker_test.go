package engine

import (
	"testing"

	"github.com/JoaoOliveira889/monobox/internal/domain"
)

func TestCleanPorts(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"0.0.0.0:8080->80/tcp, :::8080->80/tcp", "8080->80/tcp"},
		{"0.0.0.0:5432->5432/tcp", "5432->5432/tcp"},
		{"80/tcp", "80/tcp"},
	}

	for _, tt := range tests {
		got := cleanPorts(tt.input)
		if got != tt.want {
			t.Errorf("cleanPorts(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseDockerJSON_WithPorts(t *testing.T) {
	jsonInput := `{"ID":"1234567890ab","Names":"test-app","Image":"nginx:latest","State":"running","Status":"Up 5 minutes","Ports":"0.0.0.0:8080->80/tcp, :::8080->80/tcp"}`
	containers, err := parseDockerJSON([]byte(jsonInput), domain.EngineDocker)
	if err != nil {
		t.Fatalf("parseDockerJSON error: %v", err)
	}
	if len(containers) != 1 {
		t.Fatalf("got %d containers, want 1", len(containers))
	}
	c := containers[0]
	if c.Ports != "8080->80/tcp" {
		t.Errorf("Ports = %q, want %q", c.Ports, "8080->80/tcp")
	}
}
