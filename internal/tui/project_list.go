package tui

import (
	"fmt"
	"github.com/rivo/tview"
	"github.com/pmaojo/goploy/internal/config"
    "github.com/gdamore/tcell/v2"
)

type ProjectListHandlers struct {
	OnDeploy  func(config.Project)
	OnLogs    func(config.Project)
	OnRestart func(config.Project)
	OnStop    func(config.Project)
	OnShell   func(config.Project)
	OnRefresh func(config.Project)
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
		case 'r', 'R': // Refresh Status (Override Restart?)
			// Wait, 'r' is taken by Restart.
			// Let's use 'u' for Update/Refresh or 'F5' if possible, or shift-R?
			// User said: "R = refresh status".
			// But previously 'r' was restart.
			// Let's check FR6: "Basic Container Control: Restart...".
			// If 'r' is restart, we need another key for refresh.
			// Maybe 'ctrl-r'? Or just re-map Restart to 'T' (restarT) or something?
			// Or check capitalization? 'r' vs 'R'?
			// The current code handles 'r' and 'R' as Restart.

			// Let's assume user wants 'R' (Shift+R) for Refresh and 'r' for Restart?
			// The current code: case 'r', 'R': // Restart

			// I'll make 'r' = Restart, 'R' (Shift+r) = Refresh?
			// tcell event.Rune() distinguishes case.

			// Wait, the block below has case 'r', 'R' falling through to same logic.
			// I need to split them.

			// But commonly 'r' is refresh in browsers.
			// Restart is destructive. Maybe Restart should be 'R' (Harder)?
			// Or 'ctrl-r'?

			// Let's check user request: "Other projects can be refreshed on demand via a keyboard shortcut (e.g. R = refresh status for the highlighted project)."
			// I will map 'R' (Shift+r) to Refresh, and 'r' to Restart.

			if event.Rune() == 'r' {
				if handlers.OnRestart != nil {
					handlers.OnRestart(p)
				}
				return nil
			} else if event.Rune() == 'R' {
				if handlers.OnRefresh != nil {
					handlers.OnRefresh(p)
				}
				return nil
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
