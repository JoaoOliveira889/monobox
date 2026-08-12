package ui_test

import (
	"regexp"
	"testing"

	"github.com/JoaoOliveira889/monobox/internal/pkg/ui"
)

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func stripANSI(s string) string {
	return ansiRegexp.ReplaceAllString(s, "")
}

func TestRenderSparkline(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		got := ui.RenderSparkline(nil, 10)
		runes := []rune(stripANSI(got))
		if len(runes) != 10 {
			t.Errorf("expected 10 padded track runes, got %d (%q)", len(runes), got)
		}
	})

	t.Run("increasing series", func(t *testing.T) {
		data := []float64{0, 10, 25, 50, 75, 90, 100}
		got := ui.RenderSparkline(data, 7)
		clean := stripANSI(got)
		if len([]rune(clean)) != 7 {
			t.Errorf("expected 7 runes, got %d (%q)", len([]rune(clean)), clean)
		}
	})
}

func TestParsePercent(t *testing.T) {
	tests := []struct {
		input   string
		wantVal float64
		wantOk  bool
	}{
		{"12.35%", 12.35, true},
		{"0.0%", 0.0, true},
		{"100MiB / 500MiB (2.10%)", 2.10, true},
		{"invalid string", 0, false},
		{"", 0, false},
	}

	for _, tt := range tests {
		val, ok := ui.ParsePercent(tt.input)
		if ok != tt.wantOk {
			t.Errorf("ParsePercent(%q) ok = %v, want %v", tt.input, ok, tt.wantOk)
		}
		if ok && val != tt.wantVal {
			t.Errorf("ParsePercent(%q) val = %v, want %v", tt.input, val, tt.wantVal)
		}
	}
}
