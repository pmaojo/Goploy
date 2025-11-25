package projects_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pmaojo/goploy/internal/api"
	"github.com/pmaojo/goploy/internal/api/handlers/projects"
	"github.com/pmaojo/goploy/internal/config"
	"github.com/pmaojo/goploy/internal/deployment"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type MockDeployment struct {
	DeployFunc func(project config.Project, output io.Writer, ref string) error
}

func (m *MockDeployment) Deploy(project config.Project, output io.Writer, ref string) error {
	if m.DeployFunc != nil {
		return m.DeployFunc(project, output, ref)
	}
	return nil
}
func (m *MockDeployment) StreamLogs(ctx context.Context, project config.Project, output io.Writer) error {
	return nil
}
func (m *MockDeployment) Restart(project config.Project, output io.Writer) error { return nil }
func (m *MockDeployment) Stop(project config.Project, output io.Writer) error    { return nil }
func (m *MockDeployment) ListServices(project config.Project) ([]string, error)  { return nil, nil }
func (m *MockDeployment) RunShell(project config.Project, service string) error  { return nil }
func (m *MockDeployment) GetStatus(ctx context.Context, project config.Project) (deployment.ProjectStatus, error) {
	return deployment.ProjectStatus{}, nil
}

func TestTriggerDeploy_RefParsing(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/test/deploy", strings.NewReader(`{"ref":"feature/new-branch"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/projects/:name/deploy")
	c.SetParamNames("name")
	c.SetParamValues("test-project")

	mockDep := &MockDeployment{
		DeployFunc: func(project config.Project, output io.Writer, ref string) error {
			assert.Equal(t, "feature/new-branch", ref)
			return nil
		},
	}

	s := &api.Server{
		GoployConfig: &config.GoployConfig{
			Projects: []config.Project{
				{Name: "test-project"},
			},
		},
		Deployment: mockDep,
	}

	h := projects.TriggerDeploy(s)

	// Execute
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}
