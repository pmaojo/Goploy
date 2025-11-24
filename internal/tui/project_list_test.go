package tui

import (
	"testing"

	"allaboutapps.dev/aw/go-starter/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewProjectList(t *testing.T) {
	projects := []config.Project{
		{Name: "Project A", Host: "host1", Path: "/path/a", Repo: "repo1"},
		{Name: "Project B", Host: "host2", Path: "/path/b", Repo: "repo2"},
	}

	var selectedProject config.Project
	handlers := &ProjectListHandlers{
		OnDeploy: func(p config.Project) {
			selectedProject = p
		},
		OnShell: func(p config.Project) {
			selectedProject = p
		},
	}

	list := NewProjectList(projects, handlers)

	assert.NotNil(t, list)
	assert.Equal(t, 2, list.GetItemCount())

	mainText, secondaryText := list.GetItemText(0)
	assert.Equal(t, "Project A", mainText)
	assert.Contains(t, secondaryText, "host1")

	mainText, secondaryText = list.GetItemText(1)
	assert.Equal(t, "Project B", mainText)
	assert.Contains(t, secondaryText, "host2")

    // We can't easily trigger the callback without simulating tview events or exposing internal handlers,
    // but we can at least assert the captured variable `selectedProject` is empty initially.
    assert.Empty(t, selectedProject.Name)
}
