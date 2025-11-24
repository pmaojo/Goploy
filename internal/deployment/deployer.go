package deployment

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"allaboutapps.dev/aw/go-starter/internal/config"
)

// Deployer defines the interface for deploying a project.
type Deployer interface {
	Deploy(project config.Project, output io.Writer) error
}

// SSHDeployer implements Deployer using the system's `ssh` binary.
type SSHDeployer struct {
	// CmdRunner allows mocking the command execution for testing.
	// If nil, it uses the real os/exec.Command.
	CmdRunner func(name string, arg ...string) *exec.Cmd
}

// NewSSHDeployer creates a new SSHDeployer.
func NewSSHDeployer() *SSHDeployer {
	return &SSHDeployer{}
}

// Deploy connects to the project host and runs the deployment commands.
func (d *SSHDeployer) Deploy(project config.Project, output io.Writer) error {
	fmt.Fprintf(output, "Connecting to %s...\n", project.Host)

	// Construct the remote command chain
	// We want to force a pseudo-terminal allocation (-t) if we want colorful output or interactive behavior,
	// but for automation, we usually don't.
	// However, docker compose output often looks better with a PTY.
	// For this log streaming FR, standard pipe is better to capture all output reliably.

	commands := []string{
		fmt.Sprintf("cd %s", project.Path),
		"git pull",
		"docker compose pull",
		"docker compose up -d",
	}
	// Join commands with && so subsequent commands only run if previous ones succeed
	remoteCommand := strings.Join(commands, " && ")

	fmt.Fprintf(output, "Running: %s\n", remoteCommand)

	// Prepare the SSH command: ssh <host> <remote_command>
	// We rely on the system's ssh config for authentication.
	args := []string{project.Host, remoteCommand}

	var cmd *exec.Cmd
	if d.CmdRunner != nil {
		cmd = d.CmdRunner("ssh", args...)
	} else {
		cmd = exec.Command("ssh", args...)
	}

	// Wire up output streaming
	cmd.Stdout = output
	cmd.Stderr = output

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ssh command failed: %w", err)
	}

	return nil
}

// Helper to check if we are in a test environment (optional, but good for safety)
func isTestEnv() bool {
	return os.Getenv("GO_TEST_ENV") == "1"
}
