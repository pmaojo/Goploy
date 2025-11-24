package deployment

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

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
		if len(cmdArgs) >= 2 {
			fmt.Printf("Executing on host: %s\n", cmdArgs[0])
			fmt.Printf("Command: %s\n", cmdArgs[1])
		}

		// Simulate long running process if log streaming
		if strings.Contains(cmdArgs[1], "logs -f") {
			// keep running until signal or timeout
			// We can simulate some output
			fmt.Println("Log line 1")
			time.Sleep(100 * time.Millisecond) // wait a bit
		}

		os.Exit(0)
	}

	os.Exit(1)
}

func mockRunner(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestSSHClient_Deploy(t *testing.T) {
	c := NewSSHClient()
	c.CmdRunner = mockRunner

	p := config.Project{
		Name: "Test Project",
		Host: "localhost",
		Path: "/var/www/test",
	}

	var buf bytes.Buffer
	err := c.Deploy(p, &buf)

	assert.NoError(t, err)
	output := buf.String()

	assert.Contains(t, output, "Connecting to localhost...")
	assert.Contains(t, output, "Running: cd /var/www/test && git pull && docker compose pull && docker compose up -d --build")
	assert.Contains(t, output, "Mock SSH Output")
}

func TestSSHClient_StreamLogs(t *testing.T) {
	c := NewSSHClient()
	c.CmdRunner = mockRunner

	p := config.Project{
		Name: "Test Project",
		Host: "localhost",
		Path: "/var/www/test",
	}

	var buf bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := c.StreamLogs(ctx, p, &buf)

	assert.NoError(t, err)
	output := buf.String()

	assert.Contains(t, output, "Streaming logs from localhost...")
	// Expected command: cd /var/www/test && docker compose logs -f
	assert.Contains(t, output, "Running: cd /var/www/test && docker compose logs -f")
	assert.Contains(t, output, "Log line 1")
}

func TestSSHClient_Restart(t *testing.T) {
	c := NewSSHClient()
	c.CmdRunner = mockRunner

	p := config.Project{
		Name: "Test Project",
		Host: "localhost",
		Path: "/var/www/test",
	}

	var buf bytes.Buffer
	err := c.Restart(p, &buf)

	assert.NoError(t, err)
	output := buf.String()

	assert.Contains(t, output, "Restarting project on localhost...")
	assert.Contains(t, output, "Running: cd /var/www/test && docker compose restart")
}

func TestSSHClient_Stop(t *testing.T) {
	c := NewSSHClient()
	c.CmdRunner = mockRunner

	p := config.Project{
		Name: "Test Project",
		Host: "localhost",
		Path: "/var/www/test",
	}

	var buf bytes.Buffer
	err := c.Stop(p, &buf)

	assert.NoError(t, err)
	output := buf.String()

	assert.Contains(t, output, "Stopping project on localhost...")
	assert.Contains(t, output, "Running: cd /var/www/test && docker compose stop")
}
