package ui_test

import (
	"testing"

	"github.com/JoaoOliveira889/monobox/internal/pkg/ui"
)

func TestRenderSparkline(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		got := ui.RenderSparkline(nil, 10)
		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("single value zero", func(t *testing.T) {
		got := ui.RenderSparkline([]float64{0}, 10)
		if got != " " {
			t.Errorf("expected ' ', got %q", got)
		}
	})

	t.Run("single value positive", func(t *testing.T) {
		got := ui.RenderSparkline([]float64{5.5}, 10)
		if got != "▂" {
			t.Errorf("expected '▂', got %q", got)
		}
	})

	t.Run("increasing series", func(t *testing.T) {
		data := []float64{0, 1, 2, 3, 4, 5, 6, 7}
		got := ui.RenderSparkline(data, 10)
		want := " ▂▃▄▅▆▇█"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("maxLen cap", func(t *testing.T) {
		data := []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		got := ui.RenderSparkline(data, 4)
		if len([]rune(got)) != 4 {
			t.Errorf("expected 4 runes, got %d (%q)", len([]rune(got)), got)
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
