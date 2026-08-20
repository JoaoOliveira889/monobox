package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/JoaoOliveira889/monobox/internal/domain"
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

func TestParseHostPorts(t *testing.T) {
	ports := parseHostPorts("0.0.0.0:8080->80/tcp, [::]:8080->80/tcp, 0.0.0.0:8443->443/tcp")
	if len(ports) != 2 || ports[0] != "8080" || ports[1] != "8443" {
		t.Errorf("expected [8080 8443], got %v", ports)
	}
}

func TestDetailPanelUsesCompactContextualActions(t *testing.T) {
	m := NewModel(nil, "docker")
	m.width = 120
	m.height = 35
	m.showSplash = false
	m.containers = []containerItem{{
		Container: domain.Container{
			ID:     "postgres-id",
			Name:   "postgres",
			Image:  "postgres:17",
			Status: domain.StatusRunning,
			Health: domain.HealthHealthy,
			Ports:  "5432->5432/tcp",
		},
	}}

	panel := m.renderDetailPanel(m.rightPanelWidth(), m.panelHeight())
	for _, expected := range []string{"QUICK ACTIONS", "[enter]", "logs", "[?]", "all shortcuts"} {
		if !strings.Contains(panel, expected) {
			t.Errorf("detail panel should contain %q, got %q", expected, panel)
		}
	}
	if strings.Contains(panel, "ACTIONS & LOGS") {
		t.Error("detail panel should not render the long legacy actions list")
	}
}

func TestWrapDetailShortcutsKeepsRowsWithinPanel(t *testing.T) {
	shortcuts := []string{
		renderDetailShortcut("enter", "logs"),
		renderDetailShortcut("E", "environment"),
		renderDetailShortcut("H", "health logs"),
	}

	for _, line := range wrapDetailShortcuts(shortcuts, 24) {
		if lipgloss.Width(line) > 24 {
			t.Errorf("shortcut row width = %d, want <= 24", lipgloss.Width(line))
		}
	}
}

func TestHeaderSummaryDropsMetricsBeforeContainerState(t *testing.T) {
	m := NewModel(nil, "docker")
	m.width = 48
	m.containers = []containerItem{
		{Container: domain.Container{Name: "api", Status: domain.StatusRunning, CPU: "12.50%", Mem: "100MiB / 500MiB"}},
		{Container: domain.Container{Name: "worker", Status: domain.StatusExited}},
	}

	status := m.renderHeaderStatusBar()
	for _, expected := range []string{"2 containers", "1 running", "1 stopped"} {
		if !strings.Contains(status, expected) {
			t.Errorf("header summary should preserve %q, got %q", expected, status)
		}
	}
	if strings.Contains(status, "12.50% CPU") {
		t.Error("header summary should drop metrics before container state on narrow terminals")
	}
}

func TestSelectedContainerRowUsesOneBackgroundSpan(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	m := NewModel(nil, "docker")
	row := m.renderContainerRow(0, TreeNode{
		Type:        NodeContainerItem,
		ProjectName: "stack",
		Container: &containerItem{Container: domain.Container{
			Name:   "postgres",
			Image:  "postgres:17",
			Status: domain.StatusRunning,
			Health: domain.HealthHealthy,
		}},
	}, 48)

	if spans := strings.Count(row, "\x1b["); spans > 2 {
		t.Errorf("selected row has %d ANSI spans, want one continuous styled row", spans)
	}
}

func TestFooterDoesNotApplyABackground(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	m := NewModel(nil, "docker")
	m.width = 120

	footer := m.renderFooter()
	if strings.Contains(footer, ";48;") {
		t.Errorf("footer should not render a background color, got %q", footer)
	}
}
