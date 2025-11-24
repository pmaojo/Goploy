package tui

import (
	"fmt"
	"github.com/rivo/tview"
	"allaboutapps.dev/aw/go-starter/internal/config"
    "github.com/gdamore/tcell/v2"
)

// NewProjectList creates a tview.List populated with the given projects.
func NewProjectList(projects []config.Project, onSelect func(project config.Project)) *tview.List {
	list := tview.NewList().
		ShowSecondaryText(true)

    list.SetBorder(true).SetTitle("Projects (FR2)")

	for _, p := range projects {
		// Capture variable for closure
		p := p
		list.AddItem(p.Name, fmt.Sprintf("%s (%s)", p.Host, p.Path), 0, func() {
            if onSelect != nil {
                onSelect(p)
            }
		})
	}

    // Set some basic keyboard navigation help
    list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        return event
    })

	return list
}
