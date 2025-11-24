package tui

import (
	"fmt"
	"github.com/rivo/tview"
	"allaboutapps.dev/aw/go-starter/internal/config"
    "github.com/gdamore/tcell/v2"
)

type ProjectListHandlers struct {
	OnDeploy  func(config.Project)
	OnLogs    func(config.Project)
	OnRestart func(config.Project)
	OnStop    func(config.Project)
	OnShell   func(config.Project)
}

// NewProjectList creates a tview.List populated with the given projects.
func NewProjectList(projects []config.Project, handlers *ProjectListHandlers) *tview.List {
	list := tview.NewList().
		ShowSecondaryText(true)

    list.SetBorder(true).SetTitle("Projects (FR2)")

	for _, p := range projects {
		// Capture variable for closure
		p := p
		list.AddItem(p.Name, fmt.Sprintf("%s (%s)", p.Host, p.Path), 0, func() {
            if handlers != nil && handlers.OnDeploy != nil {
                handlers.OnDeploy(p)
            }
		})
	}

    // Set some basic keyboard navigation help
    list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if handlers == nil {
			return event
		}

		index := list.GetCurrentItem()
		if index < 0 || index >= len(projects) {
			return event
		}
		p := projects[index]

		switch event.Rune() {
		case 'd', 'D': // Deploy
			if handlers.OnDeploy != nil {
				handlers.OnDeploy(p)
			}
			return nil
		case 'l', 'L': // Logs
			if handlers.OnLogs != nil {
				handlers.OnLogs(p)
			}
			return nil
		case 'r', 'R': // Restart
			if handlers.OnRestart != nil {
				handlers.OnRestart(p)
			}
			return nil
		case 's', 'S': // Stop
			if handlers.OnStop != nil {
				handlers.OnStop(p)
			}
			return nil
		case 'e', 'E': // Shell (Exec)
			if handlers.OnShell != nil {
				handlers.OnShell(p)
			}
			return nil
		}
        return event
    })

	return list
}
