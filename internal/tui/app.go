package tui

import (
	"fmt"
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
	Deployer deployment.Deployer
}

func NewApp(cfg *config.GoployConfig) *App {
	app := &App{
		TviewApp: tview.NewApplication(),
		Config:   cfg,
        Pages:    tview.NewPages(),
		Deployer: deployment.NewSSHDeployer(),
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
	a.LogView.SetBorder(true).SetTitle("Deployment Logs (FR4)")

    // Create the project list
    projectList := NewProjectList(a.Config.Projects, func(project config.Project) {
		a.handleDeployment(project)
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

func (a *App) handleDeployment(project config.Project) {
	a.LogView.Clear()
	fmt.Fprintf(a.LogView, "[yellow]Starting deployment for %s...[white]\n", project.Name)

	go func() {
		writer := &ThreadSafeWriter{
			App:  a.TviewApp,
			View: a.LogView,
		}
		err := a.Deployer.Deploy(project, writer)
		if err != nil {
			fmt.Fprintf(writer, "[red]Deployment failed: %v[white]\n", err)
		} else {
			fmt.Fprintf(writer, "[green]Deployment finished successfully.[white]\n")
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
