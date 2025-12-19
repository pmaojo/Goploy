package proxy

import (
	"context"
	"io"
	"testing"

	"github.com/pmaojo/goploy/internal/config"
	"github.com/pmaojo/goploy/internal/deployment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockController is a mock for deployment.Controller
type MockController struct {
	mock.Mock
}

func (m *MockController) Deploy(project config.Project, output io.Writer, ref string) error {
	// args := m.Called(project, output, ref)
	// return args.Error(0)
	return nil
}

func (m *MockController) StreamLogs(ctx context.Context, project config.Project, output io.Writer) error {
	// args := m.Called(ctx, project, output)
	// return args.Error(0)
	return nil
}

func (m *MockController) Restart(project config.Project, output io.Writer) error {
	return nil
}
func (m *MockController) Stop(project config.Project, output io.Writer) error {
	return nil
}
func (m *MockController) ListServices(project config.Project) ([]string, error) {
	return nil, nil
}
func (m *MockController) RunShell(project config.Project, service string) error {
	return nil
}
func (m *MockController) GetStatus(ctx context.Context, project config.Project) (deployment.ProjectStatus, error) {
	return deployment.ProjectStatus{}, nil
}

func (m *MockController) UploadFile(project config.Project, content []byte, remotePath string) error {
	args := m.Called(project, content, remotePath)
	return args.Error(0)
}

func (m *MockController) RunCommand(project config.Project, cmd string) error {
	args := m.Called(project, cmd)
	return args.Error(0)
}


func TestNginxClient_ConfigureDomains(t *testing.T) {
	mockCtrl := new(MockController)
	client := NewNginxClient(mockCtrl)

	project := config.Project{
		Name: "Test Project",
		Nginx: &config.NginxConfig{
			Upstream: "localhost:3000",
		},
	}
	domains := []string{"example.com", "www.example.com"}

	// Expectations
	// 1. UploadFile
	mockCtrl.On("UploadFile", project, mock.Anything, "/tmp/test_project.nginx.conf").Return(nil)

	// 2. Move file
	mockCtrl.On("RunCommand", project, "sudo mv /tmp/test_project.nginx.conf /etc/nginx/sites-available/test_project").Return(nil)

	// 3. Symlink
	mockCtrl.On("RunCommand", project, "sudo ln -sf /etc/nginx/sites-available/test_project /etc/nginx/sites-enabled/test_project").Return(nil)

	// 4. Test config
	mockCtrl.On("RunCommand", project, "sudo nginx -t").Return(nil)

	// 5. Reload
	mockCtrl.On("RunCommand", project, "sudo systemctl reload nginx").Return(nil)

	err := client.ConfigureDomains(context.Background(), project, domains)
	assert.NoError(t, err)

	mockCtrl.AssertExpectations(t)
}

func TestNginxClient_MissingConfig(t *testing.T) {
	client := NewNginxClient(nil)
	project := config.Project{Name: "No Nginx"}

	err := client.ConfigureDomains(context.Background(), project, []string{"example.com"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nginx configuration missing")
}
