package ui

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var sparklineBlocks = []rune{' ', ' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

func RenderSparkline(values []float64, maxLen int) string {
	return RenderStyledSparkline(values, maxLen)
}

func RenderStyledSparkline(values []float64, width int) string {
	if width <= 0 {
		width = 14
	}
	data := values
	if len(data) > width {
		data = data[len(data)-width:]
	}

	padded := make([]float64, width)
	offset := width - len(data)
	for i, v := range data {
		padded[offset+i] = v
	}

	var maxVal float64 = 100.0
	for _, v := range padded {
		if v > maxVal {
			maxVal = v
		}
	}

	var sb strings.Builder
	trackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBorder))

	for i, v := range padded {
		if i < offset {
			sb.WriteString(trackStyle.Render("░"))
			continue
		}

		ratio := v / maxVal
		if maxVal <= 100.0 && v <= 100.0 {
			ratio = v / 100.0
		}
		if ratio > 1.0 {
			ratio = 1.0
		}
		if ratio < 0 {
			ratio = 0
		}

		idx := int(ratio * float64(len(sparklineBlocks)-1))
		if idx == 0 && v > 0 {
			idx = 1
		}

		blockRune := sparklineBlocks[idx]

		var col lipgloss.Color
		switch {
		case v >= 90.0:
			col = ColorError
		case v >= 75.0:
			col = ColorWarning
		case v >= 40.0:
			col = ColorCyan
		default:
			col = ColorSuccess
		}

		sb.WriteString(lipgloss.NewStyle().Foreground(col).Render(string(blockRune)))
	}

	return sb.String()
}

var percentRegex = regexp.MustCompile(`([0-9]+(?:\.[0-9]+)?)\s*%`)

func ParsePercent(s string) (float64, bool) {
	matches := percentRegex.FindStringSubmatch(s)
	if len(matches) < 2 {
		return 0, false
	}
	val, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, false
	}
	return val, true
}
