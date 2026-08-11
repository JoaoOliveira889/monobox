package domain

import (
	"testing"
)

func TestParseInspectDetails(t *testing.T) {
	raw := `[
		{
			"Config": {
				"Env": ["PATH=/usr/bin", "PORT=8080"]
			},
			"Mounts": [
				{
					"Type": "bind",
					"Source": "/host/app",
					"Destination": "/app",
					"Mode": "rw"
				}
			],
			"NetworkSettings": {
				"Networks": {
					"bridge": {
						"IPAddress": "172.17.0.2",
						"Gateway": "172.17.0.1"
					}
				}
			},
			"State": {
				"Health": {
					"Status": "unhealthy",
					"FailingStreak": 2,
					"Log": [
						{
							"Start": "2026-08-11T12:00:00Z",
							"ExitCode": 1,
							"Output": "Healthcheck failed"
						}
					]
				}
			}
		}
	]`

	details, err := ParseInspectDetails(raw)
	if err != nil {
		t.Fatalf("ParseInspectDetails error: %v", err)
	}

	if len(details.Env) != 2 {
		t.Errorf("expected 2 env vars, got %d", len(details.Env))
	}
	if len(details.Mounts) != 1 || details.Mounts[0].Destination != "/app" {
		t.Errorf("expected mount destination /app, got %v", details.Mounts)
	}
	if len(details.Networks) != 1 || details.Networks[0].IPAddress != "172.17.0.2" {
		t.Errorf("expected IPAddress 172.17.0.2, got %v", details.Networks)
	}
	if details.Health == nil || details.Health.Status != HealthUnhealthy || details.Health.FailingStreak != 2 {
		t.Errorf("expected unhealthy status with streak 2, got %v", details.Health)
	}
}
