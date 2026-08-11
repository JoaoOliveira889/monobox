package tui

import (
	"testing"
)

func TestFormatCleanPorts(t *testing.T) {
	tests := []struct {
		name             string
		raw              string
		wantCleanPorts   string
		wantOfficialPort string
	}{
		{
			name:             "empty ports",
			raw:              "",
			wantCleanPorts:   "none",
			wantOfficialPort: "",
		},
		{
			name:             "docker duplicate ipv4 and ipv6",
			raw:              "0.0.0.0:5115->8080/tcp, [::]:5115->8080/tcp",
			wantCleanPorts:   "5115 ➔ 8080/tcp",
			wantOfficialPort: "5115",
		},
		{
			name:             "multiple ports mapped",
			raw:              "0.0.0.0:8080->80/tcp, 0.0.0.0:8443->443/tcp",
			wantCleanPorts:   "8080 ➔ 80/tcp, 8443 ➔ 443/tcp",
			wantOfficialPort: "8080",
		},
		{
			name:             "unmapped internal port",
			raw:              "5432/tcp",
			wantCleanPorts:   "5432/tcp",
			wantOfficialPort: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotClean, gotOfficial := formatCleanPorts(tt.raw)
			if gotClean != tt.wantCleanPorts {
				t.Errorf("formatCleanPorts() gotClean = %q, want %q", gotClean, tt.wantCleanPorts)
			}
			if gotOfficial != tt.wantOfficialPort {
				t.Errorf("formatCleanPorts() gotOfficial = %q, want %q", gotOfficial, tt.wantOfficialPort)
			}
		})
	}
}

func TestParseMemPercVal(t *testing.T) {
	val := parseMemPercVal("24.93MiB / 256MiB (9.74%)")
	if val != 9.74 {
		t.Errorf("expected 9.74, got %f", val)
	}
}
