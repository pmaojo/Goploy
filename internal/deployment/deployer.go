package deployment

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"allaboutapps.dev/aw/go-starter/internal/config"
)

// Controller defines the interface for controlling a project.
type Controller interface {
	Deploy(project config.Project, output io.Writer) error
	StreamLogs(ctx context.Context, project config.Project, output io.Writer) error
	Restart(project config.Project, output io.Writer) error
	Stop(project config.Project, output io.Writer) error
}

// SSHClient implements Controller using the system's `ssh` binary.
type SSHClient struct {
	// CmdRunner allows mocking the command execution for testing.
	// If nil, it uses the real os/exec.Command.
	CmdRunner func(name string, arg ...string) *exec.Cmd
}

// NewSSHClient creates a new SSHClient.
func NewSSHClient() *SSHClient {
	return &SSHClient{}
}

// Deploy connects to the project host and runs the deployment commands.
func (c *SSHClient) Deploy(project config.Project, output io.Writer) error {
	fmt.Fprintf(output, "Connecting to %s...\n", project.Host)

	commands := []string{
		fmt.Sprintf("cd %s", project.Path),
		"git pull",
		"docker compose pull",
		"docker compose up -d --build",
	}
	// Join commands with && so subsequent commands only run if previous ones succeed
	remoteCommand := strings.Join(commands, " && ")

	fmt.Fprintf(output, "Running: %s\n", remoteCommand)

	return c.runSSH(project.Host, remoteCommand, output, nil)
}

// StreamLogs streams the logs from the remote project.
func (c *SSHClient) StreamLogs(ctx context.Context, project config.Project, output io.Writer) error {
	fmt.Fprintf(output, "Streaming logs from %s...\n", project.Host)

	// Command: cd path && docker compose logs -f
	commands := []string{
		fmt.Sprintf("cd %s", project.Path),
		"docker compose logs -f",
	}
	remoteCommand := strings.Join(commands, " && ")

	fmt.Fprintf(output, "Running: %s\n", remoteCommand)

	return c.runSSH(project.Host, remoteCommand, output, ctx)
}

// Restart restarts the project containers.
func (c *SSHClient) Restart(project config.Project, output io.Writer) error {
	fmt.Fprintf(output, "Restarting project on %s...\n", project.Host)

	commands := []string{
		fmt.Sprintf("cd %s", project.Path),
		"docker compose restart",
	}
	remoteCommand := strings.Join(commands, " && ")

	fmt.Fprintf(output, "Running: %s\n", remoteCommand)

	return c.runSSH(project.Host, remoteCommand, output, nil)
}

// Stop stops the project containers.
func (c *SSHClient) Stop(project config.Project, output io.Writer) error {
	fmt.Fprintf(output, "Stopping project on %s...\n", project.Host)

	commands := []string{
		fmt.Sprintf("cd %s", project.Path),
		"docker compose stop",
	}
	remoteCommand := strings.Join(commands, " && ")

	fmt.Fprintf(output, "Running: %s\n", remoteCommand)

	return c.runSSH(project.Host, remoteCommand, output, nil)
}

// runSSH executes a remote command via SSH.
func (c *SSHClient) runSSH(host, remoteCommand string, output io.Writer, ctx context.Context) error {
	args := []string{host, remoteCommand}

	var cmd *exec.Cmd
	if c.CmdRunner != nil {
		// Mock path
		cmd = c.CmdRunner("ssh", args...)

		// If context is provided, we need to handle it.
		// Since we can't easily attach context to a mock command created by CmdRunner (which returns *Cmd),
		// we rely on the caller to not care about cancellation in tests OR we wrap it.
		// For now, in real usage, we use exec.CommandContext.
	} else {
		if ctx != nil {
			cmd = exec.CommandContext(ctx, "ssh", args...)
		} else {
			cmd = exec.Command("ssh", args...)
		}
	}

	cmd.Stdout = output
	cmd.Stderr = output

	if err := cmd.Run(); err != nil {
		// If cancelled, it might return error
		if ctx != nil && ctx.Err() == context.Canceled {
			return nil // Stopped explicitly
		}
		return fmt.Errorf("ssh command failed: %w", err)
	}

	return nil
}
