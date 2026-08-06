package tui

import tea "github.com/charmbracelet/bubbletea"

type keyMap struct {
	Up          []string
	Down        []string
	Enter       []string
	Esc         []string
	Quit        []string
	Toggle      []string // s — start/stop toggle
	Restart     []string // r
	Follow      []string // f — toggle log follow
	ClearLogs   []string // c — clear log buffer
	PageUp      []string
	PageDown    []string
	End         []string
	Help        []string // ?
	Panel1      []string // 1, h, left — focus Containers panel
	Panel2      []string // 2, l, right — focus Logs panel
	Tab         []string // tab — toggle panel focus
	ResizeLeft  []string // < or , — move divider left
	ResizeRight []string // > or . — move divider right
}

var keys = keyMap{
	Up:          []string{"up", "k"},
	Down:        []string{"down", "j"},
	Enter:       []string{"enter"},
	Esc:         []string{"esc"},
	Quit:        []string{"q", "ctrl+c"},
	Toggle:      []string{"s"},
	Restart:     []string{"r"},
	Follow:      []string{"f"},
	ClearLogs:   []string{"c", "ctrl+l"},
	PageUp:      []string{"pgup", "ctrl+u"},
	PageDown:    []string{"pgdown", "ctrl+d"},
	End:         []string{"end", "G"},
	Help:        []string{"?"},
	Panel1:      []string{"1", "h", "left"},
	Panel2:      []string{"2", "l", "right"},
	Tab:         []string{"tab"},
	ResizeLeft:  []string{"<", ","},
	ResizeRight: []string{">", "."},
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
