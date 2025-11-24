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
