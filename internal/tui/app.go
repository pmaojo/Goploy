package tui

import (
	"github.com/rivo/tview"
    "allaboutapps.dev/aw/go-starter/internal/config"
)

type App struct {
	TviewApp *tview.Application
	Config   *config.GoployConfig
    Pages    *tview.Pages
}

func NewApp(cfg *config.GoployConfig) *App {
	app := &App{
		TviewApp: tview.NewApplication(),
		Config:   cfg,
        Pages:    tview.NewPages(),
	}

    // Initialize the UI
    app.setupUI()

	return app
}

func (a *App) setupUI() {
    // Create the project list
    projectList := NewProjectList(a.Config.Projects)

    // Using a Flex layout for future expansion (e.g. Logs on the right)
    flex := tview.NewFlex().
        AddItem(projectList, 0, 1, true)

    a.Pages.AddPage("main", flex, true, true)
    a.TviewApp.SetRoot(a.Pages, true)
}

func (a *App) Run() error {
	return a.TviewApp.Run()
}
