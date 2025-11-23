package tui

import (
	"fmt"
	"github.com/rivo/tview"
	"allaboutapps.dev/aw/go-starter/internal/config"
    "github.com/gdamore/tcell/v2"
)

// NewProjectList creates a tview.List populated with the given projects.
func NewProjectList(projects []config.Project) *tview.List {
	list := tview.NewList().
		ShowSecondaryText(true)

    list.SetBorder(true).SetTitle("Projects (FR2)")

	for _, p := range projects {
		// Capture variable for closure
		p := p
		list.AddItem(p.Name, fmt.Sprintf("%s (%s)", p.Host, p.Path), 0, func() {
            // Placeholder for selection action
            // For now, we might just print to stdout or do nothing as per FR2
		})
	}

    // Add a Quit option or handle keys elsewhere.
    // Usually 'q' or Esc to quit is handled globally, but let's add a "Quit" item for usability if list is focused?
    // Actually, requirements say "Main TUI Navigation".
    // Let's stick to just the projects for now to be clean.

    // Set some basic keyboard navigation help
    list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        return event
    })

	return list
}
