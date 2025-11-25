package test

import (
	"context"
	"testing"

	"github.com/pmaojo/goploy/internal/api"
	"github.com/pmaojo/goploy/internal/api/router"
	"github.com/pmaojo/goploy/internal/config"
	"github.com/pmaojo/goploy/internal/deployment"
)

// WithTestServer returns a fully configured server (using the default server config).
func WithTestServer(t *testing.T, closure func(s *api.Server)) {
	t.Helper()
	defaultConfig := config.DefaultServiceConfigFromEnv()
	WithTestServerConfigurable(t, defaultConfig, closure)
}

// WithTestServerConfigurable returns a fully configured server, allowing for configuration using the provided server config.
func WithTestServerConfigurable(t *testing.T, serverConfig config.Server, closure func(s *api.Server)) {
	t.Helper()
	ctx := t.Context()
	WithTestServerConfigurableContext(ctx, t, serverConfig, closure)
}

// WithTestServerConfigurableContext returns a fully configured server, allowing for configuration using the provided server config.
// The provided context will be used during setup (instead of the default background context).
func WithTestServerConfigurableContext(ctx context.Context, t *testing.T, serverConfig config.Server, closure func(s *api.Server)) {
	t.Helper()
	execClosureNewTestServer(ctx, t, serverConfig, closure)
}

// Executes closure on a new test server
func execClosureNewTestServer(ctx context.Context, t *testing.T, serverConfig config.Server, closure func(s *api.Server)) {
	t.Helper()

	// https://stackoverflow.com/questions/43424787/how-to-use-next-available-port-in-http-listenandserve
	// You may use port 0 to indicate you're not specifying an exact port but you want a free, available port selected by the system
	serverConfig.Echo.ListenAddress = ":0"

	// Mock dependencies
	mockGoployConfig := &config.GoployConfig{
		Projects: []config.Project{},
	}
	mockMailer := NewTestMailer(t)
	// mockController can be nil for basic server tests, or we can use a mock/SSH client if needed.
	// Since we are just testing the server setup here, nil might be risky if NewServer uses it immediately.
	// But looking at NewServer, it just assigns it.
	var mockController deployment.Controller = deployment.NewSSHClient(nil)

	s := api.NewServer(serverConfig, mockGoployConfig, mockMailer, mockController)

	err := router.Init(s)
	if err != nil {
		t.Fatalf("Failed to init router: %v", err)
	}

	closure(s)

	// echo is managed and should close automatically after running the test
	if err := s.Echo.Shutdown(ctx); err != nil {
		t.Fatalf("failed to shutdown server: %v", err)
	}

	// disallow any further refs to managed object after running the test
	//nolint: wastedassign
	s = nil
}
