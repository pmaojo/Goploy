package tui

import (
	"context"
	"fmt"
	"io"
	"sync"
	// "io" // Removed unused import
	"github.com/rivo/tview"
    "allaboutapps.dev/aw/go-starter/internal/config"
	"allaboutapps.dev/aw/go-starter/internal/deployment"
)

type App struct {
	TviewApp *tview.Application
	Config   *config.GoployConfig
    Pages    *tview.Pages
	LogView  *tview.TextView
	Controller deployment.Controller

	// State for managing running tasks
	logCancelCtx context.Context
	logCancel    context.CancelFunc
	mu           sync.Mutex
}

func NewApp(cfg *config.GoployConfig) *App {
	app := &App{
		TviewApp: tview.NewApplication(),
		Config:   cfg,
        Pages:    tview.NewPages(),
		Controller: deployment.NewSSHClient(),
	}

    // Initialize the UI
    app.setupUI()

	return app
}

func (a *App) setupUI() {
	// Create the log view
	a.LogView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			a.TviewApp.Draw()
		})
	a.LogView.SetBorder(true).SetTitle("Logs")

    // Create the project list
    projectList := NewProjectList(a.Config.Projects, &ProjectListHandlers{
		OnDeploy: func(p config.Project) { a.handleDeployment(p) },
		OnLogs:   func(p config.Project) { a.handleLogs(p) },
		OnRestart: func(p config.Project) { a.handleRestart(p) },
		OnStop:    func(p config.Project) { a.handleStop(p) },
		OnShell:   func(p config.Project) { a.handleShell(p) },
	})

    // Using a Flex layout for future expansion (e.g. Logs on the right)
    flex := tview.NewFlex().
        AddItem(projectList, 0, 1, true).
		AddItem(a.LogView, 0, 2, false)

    a.Pages.AddPage("main", flex, true, true)
    a.TviewApp.SetRoot(a.Pages, true)
}

func (a *App) Run() error {
	return a.TviewApp.Run()
}

func (a *App) getWriter() io.Writer {
	return &ThreadSafeWriter{
		App:  a.TviewApp,
		View: a.LogView,
	}
}

func (a *App) cancelPreviousTask() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.logCancel != nil {
		a.logCancel()
		a.logCancel = nil
	}
}

func (a *App) handleDeployment(project config.Project) {
	a.cancelPreviousTask()
	a.LogView.SetTitle("Deployment Logs (FR4)")
	a.LogView.Clear()
	fmt.Fprintf(a.LogView, "[yellow]Starting deployment for %s...[white]\n", project.Name)

	go func() {
		writer := a.getWriter()
		err := a.Controller.Deploy(project, writer)
		if err != nil {
			fmt.Fprintf(writer, "[red]Deployment failed: %v[white]\n", err)
		} else {
			fmt.Fprintf(writer, "[green]Deployment finished successfully.[white]\n")
		}
	}()
}

func (a *App) handleLogs(project config.Project) {
	a.cancelPreviousTask()
	a.LogView.SetTitle("Monitoring Logs (FR5) - Press any other action to stop")
	a.LogView.Clear()
	fmt.Fprintf(a.LogView, "[yellow]Streaming logs for %s...[white]\n", project.Name)

	ctx, cancel := context.WithCancel(context.Background())
	a.mu.Lock()
	a.logCancel = cancel
	a.mu.Unlock()

	go func() {
		writer := a.getWriter()
		err := a.Controller.StreamLogs(ctx, project, writer)
		// If cancelled, err might be nil or Canceled depending on implementation.
		// We can check context error.
		if ctx.Err() == context.Canceled {
			fmt.Fprintf(writer, "[yellow]Log streaming stopped.[white]\n")
			return
		}

		if err != nil {
			fmt.Fprintf(writer, "[red]Log streaming failed: %v[white]\n", err)
		} else {
			fmt.Fprintf(writer, "[yellow]Log streaming ended.[white]\n")
		}
	}()
}

func (a *App) handleRestart(project config.Project) {
	a.cancelPreviousTask()
	a.LogView.SetTitle("Control (FR6)")
	a.LogView.Clear()
	fmt.Fprintf(a.LogView, "[yellow]Restarting %s...[white]\n", project.Name)

	go func() {
		writer := a.getWriter()
		err := a.Controller.Restart(project, writer)
		if err != nil {
			fmt.Fprintf(writer, "[red]Restart failed: %v[white]\n", err)
		} else {
			fmt.Fprintf(writer, "[green]Restart finished successfully.[white]\n")
		}
	}()
}

func (a *App) handleStop(project config.Project) {
	a.cancelPreviousTask()
	a.LogView.SetTitle("Control (FR6)")
	a.LogView.Clear()
	fmt.Fprintf(a.LogView, "[yellow]Stopping %s...[white]\n", project.Name)

	go func() {
		writer := a.getWriter()
		err := a.Controller.Stop(project, writer)
		if err != nil {
			fmt.Fprintf(writer, "[red]Stop failed: %v[white]\n", err)
		} else {
			fmt.Fprintf(writer, "[green]Stop finished successfully.[white]\n")
		}
	}()
}

func (a *App) handleShell(project config.Project) {
	a.cancelPreviousTask()
	a.LogView.SetTitle("Control (FR6) - Shell Access")
	a.LogView.Clear()
	fmt.Fprintf(a.LogView, "[yellow]Fetching services for %s...[white]\n", project.Name)

	// Run fetching in goroutine
	go func() {
		services, err := a.Controller.ListServices(project)
		if err != nil {
			a.TviewApp.QueueUpdateDraw(func() {
				fmt.Fprintf(a.LogView, "[red]Failed to fetch services: %v[white]\n", err)
			})
			return
		}

		if len(services) == 0 {
			a.TviewApp.QueueUpdateDraw(func() {
				fmt.Fprintf(a.LogView, "[red]No services found.[white]\n")
			})
			return
		}

		// Create and show the selection modal
		a.TviewApp.QueueUpdateDraw(func() {
			a.showServiceSelectionModal(project, services)
		})
	}()
}

func (a *App) showServiceSelectionModal(project config.Project, services []string) {
	list := tview.NewList()
	list.SetBorder(true).SetTitle("Select Service for Shell")

	for _, s := range services {
		// capture s
		s := s
		list.AddItem(s, "", 0, func() {
			// On Select:
			a.Pages.RemovePage("services_modal")

			// Suspend and Run Shell
			a.TviewApp.Suspend(func() {
				err := a.Controller.RunShell(project, s)
				if err != nil {
					// We are suspended, so we can print to stdout/stderr,
					// but better to log it when we return.
					// fmt.Printf("Error: %v\nPress Enter to continue...", err)
					// fmt.Scanln()
				}
			})

			// After resume
			a.LogView.Clear()
			fmt.Fprintf(a.LogView, "[yellow]Shell session ended.[white]\n")
		})
	}

	// Add a cancel option
	list.AddItem("Cancel", "Return to main menu", 'c', func() {
		a.Pages.RemovePage("services_modal")
	})

	// Center the list
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(list, 10, 1, true). // Fixed height for the list
			AddItem(nil, 0, 1, false), 0, 1, true).
		AddItem(nil, 0, 1, false)

	a.Pages.AddPage("services_modal", modal, true, true)
	a.TviewApp.SetFocus(list)
}

// ThreadSafeWriter allows writing to a tview.TextView from a goroutine
type ThreadSafeWriter struct {
	App  *tview.Application
	View *tview.TextView
}

func (w *ThreadSafeWriter) Write(p []byte) (n int, err error) {
	// Make a copy of the slice because the backing array might be reused
	// before the queued function is executed.
	data := make([]byte, len(p))
	copy(data, p)

	// We must use QueueUpdateDraw to safely update the UI from another goroutine
	w.App.QueueUpdateDraw(func() {
		w.View.Write(data)
	})
	return len(p), nil
}
