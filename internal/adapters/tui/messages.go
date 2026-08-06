package tui

import (
	"bufio"
	"context"
	"io"
	"time"
)

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

// containerLogsClearedMsg signals completion of engine log file truncation.
type containerLogsClearedMsg struct {
	containerID string
	err         error
}

// logStreamOpenedMsg carries an opened engine log stream.
type logStreamOpenedMsg struct {
	containerID string
	reader      io.Closer
	scanner     *bufio.Scanner
	ctx         context.Context
	cancel      context.CancelFunc
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

// splashTickMsg drives the splash screen animation.
type splashTickMsg struct{}

// clearStatusMsg auto-clears the status bar after a delay.
type clearStatusMsg struct{ id int }

// tickMsg drives periodic container list refresh.
type tickMsg time.Time
