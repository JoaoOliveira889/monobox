package tui

import tea "github.com/charmbracelet/bubbletea"

type keyMap struct {
	Up          []string
	Down        []string
	Enter       []string
	Esc         []string
	Quit        []string
	Toggle      []string
	Restart     []string
	Follow      []string
	ClearLogs   []string
	PageUp      []string
	PageDown    []string
	End         []string
	Help        []string
	Panel1      []string
	Panel2      []string
	Tab         []string
	ResizeLeft  []string
	ResizeRight []string
	Filter      []string
	Exec        []string
	Inspect     []string
	Remove      []string
	Pause            []string
	OpenPort         []string
	SaveLogs         []string
	ToggleTimestamps []string
	ThemeMenu        []string
	EnvModal         []string
	HealthModal      []string
	Graph            []string
	ToggleRegex      []string
	CopyID           []string
	CopyInfo         []string
	Prune            []string
	ComposeUp        []string
	ComposeDown      []string
	NextMatch        []string
	PrevMatch        []string
	SettingsModal    []string
	SeverityFilter   []string
}

var keys = keyMap{
	Up:               []string{"up", "k"},
	Down:             []string{"down", "j"},
	Enter:            []string{"enter"},
	Esc:              []string{"esc"},
	Quit:             []string{"q", "ctrl+c"},
	Toggle:           []string{"s"},
	Restart:          []string{"r"},
	Follow:           []string{"f"},
	ClearLogs:        []string{"c", "ctrl+l"},
	PageUp:           []string{"pgup", "ctrl+u"},
	PageDown:         []string{"pgdown", "ctrl+d"},
	End:              []string{"end", "G"},
	Help:             []string{"?"},
	Panel1:           []string{"1", "h", "left"},
	Panel2:           []string{"2", "l", "right"},
	Tab:              []string{"tab"},
	ResizeLeft:       []string{"<", ","},
	ResizeRight:      []string{">", "."},
	Filter:           []string{"/"},
	Exec:             []string{"e", "x"},
	Inspect:          []string{"i"},
	Remove:           []string{"d", "delete"},
	Pause:            []string{"p"},
	OpenPort:         []string{"o"},
	SaveLogs:         []string{"ctrl+s", "s"},
	ToggleTimestamps: []string{"t"},
	ThemeMenu:        []string{"T"},
	EnvModal:         []string{"E"},
	HealthModal:      []string{"H"},
	Graph:            []string{"g"},
	ToggleRegex:      []string{"ctrl+r"},
	CopyID:           []string{"y"},
	CopyInfo:         []string{"Y"},
	Prune:            []string{"P"},
	ComposeUp:        []string{"u"},
	ComposeDown:      []string{"D"},
	NextMatch:        []string{"n"},
	PrevMatch:        []string{"N"},
	SettingsModal:    []string{"S"},
	SeverityFilter:   []string{"!"},
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
