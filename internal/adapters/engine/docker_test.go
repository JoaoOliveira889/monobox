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

func TestParseDockerJSON_WithComposeLabels(t *testing.T) {
	jsonInput := `{"ID":"abc123456789","Names":"my-stack_web_1","Image":"web:latest","State":"running","Status":"Up 2 hours","Labels":"com.docker.compose.project=my-stack,com.docker.compose.service=web"}`
	containers, err := parseDockerJSON([]byte(jsonInput), domain.EngineDocker)
	if err != nil {
		t.Fatalf("parseDockerJSON error: %v", err)
	}
	if len(containers) != 1 {
		t.Fatalf("got %d containers, want 1", len(containers))
	}
	c := containers[0]
	if c.ComposeProject != "my-stack" {
		t.Errorf("ComposeProject = %q, want %q", c.ComposeProject, "my-stack")
	}
	if c.Labels["com.docker.compose.service"] != "web" {
		t.Errorf("service label = %q, want %q", c.Labels["com.docker.compose.service"], "web")
	}
}

func TestParseHealthStatus(t *testing.T) {
	tests := []struct {
		status string
		want   domain.HealthStatus
	}{
		{"Up 2 hours (healthy)", domain.HealthHealthy},
		{"Up 5 minutes (unhealthy)", domain.HealthUnhealthy},
		{"Up 10 seconds (health: starting)", domain.HealthStarting},
		{"Up 10 seconds (starting)", domain.HealthStarting},
		{"Up 3 hours", domain.HealthNone},
		{"Exited (0) 5 minutes ago", domain.HealthNone},
	}

	for _, tt := range tests {
		got := parseHealthStatus(tt.status)
		if got != tt.want {
			t.Errorf("parseHealthStatus(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestParseDockerJSON_WithHealthStatus(t *testing.T) {
	jsonInput := `{"ID":"1234567890ab","Names":"test-app","Image":"nginx:latest","State":"running","Status":"Up 5 minutes (healthy)","Ports":""}`
	containers, err := parseDockerJSON([]byte(jsonInput), domain.EngineDocker)
	if err != nil {
		t.Fatalf("parseDockerJSON error: %v", err)
	}
	if len(containers) != 1 {
		t.Fatalf("got %d containers, want 1", len(containers))
	}
	if containers[0].Health != domain.HealthHealthy {
		t.Errorf("Health = %q, want %q", containers[0].Health, domain.HealthHealthy)
	}
}

