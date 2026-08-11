package ui

import (
	"regexp"
	"strconv"
)

var sparklineBlocks = []rune{' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// RenderSparkline generates an ASCII sparkline string (e.g. " ▂▃▄▅▆▇█")
// from a slice of float64 data points, capped at maxLen recent points.
func RenderSparkline(values []float64, maxLen int) string {
	if len(values) == 0 {
		return ""
	}
	if maxLen > 0 && len(values) > maxLen {
		values = values[len(values)-maxLen:]
	}

	minVal, maxVal := values[0], values[0]
	for _, v := range values {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	runes := make([]rune, len(values))
	if maxVal == minVal {
		block := sparklineBlocks[0]
		if maxVal > 0 {
			block = sparklineBlocks[1]
		}
		for i := range runes {
			runes[i] = block
		}
		return string(runes)
	}

	maxIdx := len(sparklineBlocks) - 1
	for i, v := range values {
		norm := (v - minVal) / (maxVal - minVal)
		idx := int(norm * float64(maxIdx))
		if idx < 0 {
			idx = 0
		} else if idx > maxIdx {
			idx = maxIdx
		}
		runes[i] = sparklineBlocks[idx]
	}

	return string(runes)
}

var percentRegex = regexp.MustCompile(`([0-9]+(?:\.[0-9]+)?)\s*%`)

// ParsePercent parses a numeric percentage float from strings like "12.35%" or "100MiB (2.10%)".
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
