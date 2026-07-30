package tui

import "time"

type errMsg struct{ Err error }

func (e errMsg) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return ""
}

// containersLoadedMsg is sent when the container list is refreshed.
type containersLoadedMsg struct {
	containers []containerItem
	err        error
}

// containerActionDoneMsg is sent after start/stop/restart completes.
type containerActionDoneMsg struct {
	id     string
	action string
	err    error
}

// logLineMsg is sent for each new log line arriving from the stream.
type logLineMsg struct {
	containerID string
	line        string
}

// logStreamDoneMsg signals that the log stream has ended.
type logStreamDoneMsg struct {
	containerID string
}

// spinnerTickMsg drives the loading animation.
type spinnerTickMsg struct{}

// clearStatusMsg auto-clears the status bar after a delay.
type clearStatusMsg struct{ id int }

// tickMsg drives periodic container list refresh.
type tickMsg time.Time
