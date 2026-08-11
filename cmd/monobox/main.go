package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JoaoOliveira889/monobox/internal/adapters/engine"
	"github.com/JoaoOliveira889/monobox/internal/adapters/tui"
	"github.com/JoaoOliveira889/monobox/internal/pkg/logging"
)

var (
	version = "0.0.3"
	commit  = "none"
	date    = "unknown"
)

func main() {
	tui.Version = version

	logging.Init()
	defer logging.Close()

	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("monobox %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built at: %s\n", date)
		return
	}

	provider, engineName, err := engine.DetectEngine()
	if err != nil {
		fmt.Fprintf(os.Stderr, "monobox: %v\n", err)
		os.Exit(1)
	}

	m := tui.NewModel(provider, engineName)
	p := tea.NewProgram(&m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		logging.Error("program exited with error", "error", err)
		fmt.Fprintf(os.Stderr, "monobox: %v\n", err)
		os.Exit(1)
	}
	logging.Info("program exited normally")
}
