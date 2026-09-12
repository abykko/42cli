package tui

import (
    "fmt"
    "os"
    tea "charm.land/bubbletea/v2"
    
    "42tui/conf"
    "42tui/deployment"
)

func Tui() {

    defer func() {
		deployment.Stop()
        deployment.StopByImage()
        conf.Set("logged_with", "")
    }()

    p := tea.NewProgram(InitialModel())
    if _, err := p.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v", err)
        os.Exit(1)
    }
}