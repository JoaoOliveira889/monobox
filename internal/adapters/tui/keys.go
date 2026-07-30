package tui

import tea "github.com/charmbracelet/bubbletea"

type keyMap struct {
	Up      []string
	Down    []string
	Enter   []string
	Esc     []string
	Quit    []string
	Toggle  []string // s — start/stop toggle
	Restart []string // r
	Follow  []string // f — toggle log follow
	Help    []string // ?
}

var keys = keyMap{
	Up:      []string{"up", "k"},
	Down:    []string{"down", "j"},
	Enter:   []string{"enter"},
	Esc:     []string{"esc"},
	Quit:    []string{"q", "ctrl+c"},
	Toggle:  []string{"s"},
	Restart: []string{"r"},
	Follow:  []string{"f"},
	Help:    []string{"?"},
}

func matchesKey(msg tea.KeyMsg, ks ...string) bool {
	s := msg.String()
	for _, k := range ks {
		if s == k {
			return true
		}
	}
	return false
}
