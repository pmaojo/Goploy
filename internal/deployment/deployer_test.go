package deployment

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"

	"allaboutapps.dev/aw/go-starter/internal/config"
	"github.com/stretchr/testify/assert"
)

// TestHelperProcess is a magic function that runs the test command helper.
// This is used to mock exec.Command.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	// This code runs inside the "exec" command during the test.
	// We read arguments to know what to simulate.
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command provided\n")
		os.Exit(2)
	}

	cmd, cmdArgs := args[0], args[1:]

	if cmd == "ssh" {
		// Simulate SSH output
		fmt.Printf("Mock SSH Output: %s\n", cmdArgs)
		// Check command args to ensure they are what we expect
		// ssh <host> <remote_command>
		if len(cmdArgs) >= 2 {
			fmt.Printf("Executing on host: %s\n", cmdArgs[0])
			fmt.Printf("Command: %s\n", cmdArgs[1])
		}
		os.Exit(0)
	}

	os.Exit(1)
}


func TestDeployer_Deploy(t *testing.T) {
	// Custom runner to mock exec.Command
	mockRunner := func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	d := NewSSHDeployer()
	d.CmdRunner = mockRunner

	p := config.Project{
		Name: "Test Project",
		Host: "localhost",
		Path: "/var/www/test",
		Repo: "git@github.com:test/test.git",
	}

	var buf bytes.Buffer
	err := d.Deploy(p, &buf)

	assert.NoError(t, err)
	output := buf.String()

	// Verify the deployer logic set up the correct command structure
	assert.Contains(t, output, "Connecting to localhost...")
	assert.Contains(t, output, "Running: cd /var/www/test && git pull && docker compose pull && docker compose up -d")

	// Verify the mock execution happened
	assert.Contains(t, output, "Mock SSH Output")
	assert.Contains(t, output, "Executing on host: localhost")
	assert.Contains(t, output, "Command: cd /var/www/test && git pull && docker compose pull && docker compose up -d")
}

func TestDeployer_Deploy_Real_Execution_Fails_Gracefully(t *testing.T) {
	// This test ensures that if exec fails (exit code != 0), Deploy returns an error
	mockRunner := func(command string, args ...string) *exec.Cmd {
		// Create a command that fails
		cmd := exec.Command("false")
		return cmd
	}

	d := NewSSHDeployer()
	d.CmdRunner = mockRunner

	p := config.Project{
		Name: "Test Project",
		Host: "localhost",
		Path: "/path",
	}

	err := d.Deploy(p, io.Discard)
	assert.Error(t, err)
}
