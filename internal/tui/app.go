package tui

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/rivo/tview"

	"github.com/pmaojo/goploy/internal/config"
	"github.com/pmaojo/goploy/internal/deployment"
	"github.com/pmaojo/goploy/internal/proxy"
)

type App struct {
	TviewApp           *tview.Application
	Config             *config.GoployConfig
	Pages              *tview.Pages
	LogView            *tview.TextView
	DetailsView        *tview.TextView
	ProjectList        *tview.List
	Controller         deployment.Controller
	DomainConfigurator proxy.Configurator

	// State for managing running tasks
	logCancelCtx context.Context
	logCancel    context.CancelFunc
	statusCancel context.CancelFunc
	mu           sync.Mutex
}

func NewApp(cfg *config.GoployConfig) *App {
	return NewAppWithDependencies(cfg, deployment.NewSSHClient(nil), proxy.NewCaddyClient(nil))
}

// NewAppWithDependencies allows injecting collaborators for testing.
func NewAppWithDependencies(cfg *config.GoployConfig, controller deployment.Controller, domainConfigurator proxy.Configurator) *App {
	app := &App{
		TviewApp:           tview.NewApplication(),
		Config:             cfg,
		Pages:              tview.NewPages(),
		Controller:         controller,
		DomainConfigurator: domainConfigurator,
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

	// Create the details view
	a.DetailsView = tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(true)
	a.DetailsView.SetBorder(true).SetTitle("Status (FR7)")

	// Create the project list
	a.ProjectList = NewProjectList(a.Config.Projects, &ProjectListHandlers{
		OnDeploy:           func(p config.Project) { a.handleDeployment(p) },
		OnLogs:             func(p config.Project) { a.handleLogs(p) },
		OnRestart:          func(p config.Project) { a.handleRestart(p) },
		OnStop:             func(p config.Project) { a.handleStop(p) },
		OnShell:            func(p config.Project) { a.handleShell(p) },
		OnRefresh:          func(p config.Project) { a.handleRefresh(p) },
		OnConfigureDomains: func(p config.Project) { a.handleConfigureDomains(p) },
	})

	// Hook into list selection change
	a.ProjectList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(a.Config.Projects) {
			a.startMonitoring(a.Config.Projects[index])
		}
	})

	// Right side: Details (top) and Logs (bottom)
	rightFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.DetailsView, 10, 1, false). // Fixed height for details
		AddItem(a.LogView, 0, 3, false)

	// Main Flex layout
	flex := tview.NewFlex().
		AddItem(a.ProjectList, 0, 1, true).
		AddItem(rightFlex, 0, 2, false)

	a.Pages.AddPage("main", flex, true, true)
	a.TviewApp.SetRoot(a.Pages, true)

	// Trigger initial monitoring for the first project if exists
	if len(a.Config.Projects) > 0 {
		a.startMonitoring(a.Config.Projects[0])
	}
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

func (a *App) startMonitoring(project config.Project) {
	a.mu.Lock()
	// Cancel existing monitoring
	if a.statusCancel != nil {
		a.statusCancel()
	}
	// Create new context
	ctx, cancel := context.WithCancel(context.Background())
	a.statusCancel = cancel
	a.mu.Unlock()

	// Clear Details View
	a.DetailsView.Clear()
	fmt.Fprintf(a.DetailsView, "[yellow]Fetching status for %s...[white]\n", project.Name)

	go func() {
		// Immediate check
		a.updateStatus(ctx, project)

		// Start ticker
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.updateStatus(ctx, project)
			}
		}
	}()
}

func (a *App) updateStatus(ctx context.Context, project config.Project) {
	status, err := a.Controller.GetStatus(ctx, project)
	if ctx.Err() != nil {
		return // context cancelled
	}

	a.TviewApp.QueueUpdateDraw(func() {
		// Check if we are still selecting this project (UI consistency check)
		// We check if the current item in the list matches the project we just fetched.
		currentIdx := a.ProjectList.GetCurrentItem()
		if currentIdx >= 0 && currentIdx < len(a.Config.Projects) {
			if a.Config.Projects[currentIdx].Name != project.Name {
				return // Ignore update if selection changed
			}
		}

		if err != nil {
			a.DetailsView.Clear()
			fmt.Fprintf(a.DetailsView, "[red]Failed to fetch status: %v[white]\n", err)
			return
		}

		// Update Details View
		a.DetailsView.Clear()
		fmt.Fprintf(a.DetailsView, "[green]Project:[white] %s\n", status.Name)
		fmt.Fprintf(a.DetailsView, "[green]Branch:[white] %s\n", status.Branch)
		fmt.Fprintf(a.DetailsView, "[green]Last Deployed:[white] %s\n", status.LastDeployedAt.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(a.DetailsView, "[green]Status:[white] %s\n", status.Status)
		fmt.Fprintf(a.DetailsView, "\n[yellow]Containers:[white]\n")
		for _, c := range status.Containers {
			stateColor := "red"
			if strings.ToLower(c.State) == "running" {
				stateColor = "green"
			}
			fmt.Fprintf(a.DetailsView, "- %s: [%s]%s[white] (%s)\n", c.Name, stateColor, c.State, c.Status)
		}

		// Update Project List Item
		// Find the item index
		for i, p := range a.Config.Projects {
			if p.Name == project.Name {
				// Format: Status · Branch · Time
				summary := fmt.Sprintf("%s · %s · %s", status.Status, status.Branch, timeSince(status.LastDeployedAt))
				a.ProjectList.SetItemText(i, p.Name, summary)
				break
			}
		}
	})
}

func timeSince(t time.Time) string {
	if t.IsZero() {
		return "Never"
	}
	d := time.Since(t)
	if d < time.Minute {
		return "Just now"
	} else if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(d.Hours()/24))
}

func (a *App) handleDeployment(project config.Project) {
	a.cancelPreviousTask()
	a.LogView.SetTitle("Deployment Logs (FR4)")
	a.LogView.Clear()
	fmt.Fprintf(a.LogView, "[yellow]Starting deployment for %s...[white]\n", project.Name)

	go func() {
		writer := a.getWriter()
		// TUI deployment doesn't specify ref currently (uses default)
		err := a.Controller.Deploy(project, writer, "")
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

func (a *App) handleRefresh(project config.Project) {
	// Manual refresh of status
	go func() {
		a.updateStatus(context.Background(), project)
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

func (a *App) handleConfigureDomains(project config.Project) {
	a.cancelPreviousTask()
	a.LogView.SetTitle("Domain Management (Caddy)")
	a.LogView.Clear()

	form := a.buildDomainForm(project)
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 8, 1, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)

	a.Pages.AddPage("domains_modal", modal, true, true)
	a.TviewApp.SetFocus(form)
}

func (a *App) buildDomainForm(project config.Project) *tview.Form {
	existing := ""
	var configurator proxy.Configurator
	title := "Configure Domains"

	// Determine configurator based on project config
	if project.Caddy != nil {
		if len(project.Caddy.Domains) > 0 {
			existing = strings.Join(project.Caddy.Domains, ", ")
		}
		// Assuming a.DomainConfigurator is generic or we cast it?
		// Currently a.DomainConfigurator is initialized as CaddyClient in NewApp.
		// We need to support Nginx too.
		// For now, we will use the injected one if it matches, or create a new one?
		// The dependency injection in NewApp is static. We might need to make it dynamic or support both.
		// But let's assume for this specific method, we can determine which one to use.
		configurator = a.DomainConfigurator // Default to Caddy if injected
		title += " (Caddy)"
	} else if project.Nginx != nil {
		if len(project.Nginx.Domains) > 0 {
			existing = strings.Join(project.Nginx.Domains, ", ")
		}
		// Create Nginx client on the fly or use a factory?
		// Since we have Controller, we can create it.
		configurator = proxy.NewNginxClient(a.Controller)
		title += " (Nginx)"
	} else {
		// No proxy config
		// Maybe default to Caddy if unsure, or show error?
		// Or show a message "No proxy configured in yaml".
	}

	input := tview.NewInputField().
		SetLabel("Domains").
		SetText(existing).
		SetFieldWidth(60)

	form := tview.NewForm()

	if configurator == nil {
		form.AddTextView("Error", "No Caddy or Nginx configuration found in goploy.yaml for this project.", 40, 2, true, false)
		form.AddButton("Close", func() {
			a.Pages.RemovePage("domains_modal")
		})
	} else {
		form.AddFormItem(input).
		AddButton("Save", func() {
			domains := parseDomainsInput(input.GetText())
			if len(domains) == 0 {
				a.TviewApp.QueueUpdateDraw(func() {
					fmt.Fprintf(a.LogView, "[red]Please provide at least one domain.[white]\n")
				})
				return
			}

			a.Pages.RemovePage("domains_modal")
			a.LogView.Clear()
			fmt.Fprintf(a.LogView, "[yellow]Configuring domains for %s...[white]\n", project.Name)

			go func() {
				err := configurator.ConfigureDomains(context.Background(), project, domains)
				a.TviewApp.QueueUpdateDraw(func() {
					if err != nil {
						fmt.Fprintf(a.LogView, "[red]Failed to configure domains: %v[white]\n", err)
						return
					}

					fmt.Fprintf(a.LogView, "[green]Domains configured successfully.[white]\n")

					// Update cached project domains so subsequent edits reflect the new state
					for i, p := range a.Config.Projects {
						if p.Name == project.Name {
							if project.Caddy != nil {
								a.Config.Projects[i].Caddy.Domains = domains
							} else if project.Nginx != nil {
								a.Config.Projects[i].Nginx.Domains = domains
							}
						}
					}
				})
			}()
		}).
		AddButton("Cancel", func() {
			a.Pages.RemovePage("domains_modal")
		})
	}

	form.SetBorder(true).SetTitle(title)
	return form
}

func parseDomainsInput(input string) []string {
	fields := strings.FieldsFunc(input, func(r rune) bool {
		return r == ',' || r == '\n' || r == ';'
	})

	var domains []string
	for _, f := range fields {
		if trimmed := strings.TrimSpace(f); trimmed != "" {
			domains = append(domains, trimmed)
		}
	}

	return domains
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
