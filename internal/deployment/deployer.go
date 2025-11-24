package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"allaboutapps.dev/aw/go-starter/internal/config"
	"allaboutapps.dev/aw/go-starter/internal/mailer"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"
)

// Controller defines the interface for controlling a project.
type Controller interface {
	Deploy(project config.Project, output io.Writer, ref string) error
	StreamLogs(ctx context.Context, project config.Project, output io.Writer) error
	Restart(project config.Project, output io.Writer) error
	Stop(project config.Project, output io.Writer) error
	ListServices(project config.Project) ([]string, error)
	RunShell(project config.Project, service string) error
	GetStatus(ctx context.Context, project config.Project) (ProjectStatus, error)
}

// SSHClient implements Controller using golang.org/x/crypto/ssh.
type SSHClient struct {
	Mailer *mailer.Mailer
}

var _ Controller = (*SSHClient)(nil)

// NewSSHClient creates a new SSHClient.
func NewSSHClient(mailer *mailer.Mailer) *SSHClient {
	return &SSHClient{
		Mailer: mailer,
	}
}

// connect establishes an SSH connection to the project host.
func (c *SSHClient) connect(project config.Project) (*ssh.Client, error) {
	// 1. Determine Host, User, Port
	host := project.Host
	user := project.User
	port := project.Port

	// If host contains user@ or :port, parse it
	if strings.Contains(host, "@") {
		parts := strings.SplitN(host, "@", 2)
		if user == "" {
			user = parts[0]
		}
		host = parts[1]
	}
	if strings.Contains(host, ":") {
		parts := strings.SplitN(host, ":", 2)
		host = parts[0]
		if port == "" {
			port = parts[1]
		}
	}

	// Defaults
	if user == "" {
		user = os.Getenv("USER") // fallback to current user
	}
	if port == "" {
		port = "22"
	}

	// 2. Prepare Auth Methods
	authMethods := []ssh.AuthMethod{}

	// Identity File
	identityFile := project.IdentityFile
	if identityFile == "" {
		// Default to ~/.ssh/id_rsa
		home, err := os.UserHomeDir()
		if err == nil {
			identityFile = home + "/.ssh/id_rsa"
		}
	} else {
		// Expand ~ if present
		if strings.HasPrefix(identityFile, "~/") {
			home, err := os.UserHomeDir()
			if err == nil {
				identityFile = home + identityFile[1:]
			}
		}
	}

	key, err := os.ReadFile(identityFile)
	if err == nil {
		signer, err := ssh.ParsePrivateKey(key)
		if err == nil {
			authMethods = append(authMethods, ssh.PublicKeys(signer))
		}
	}

	// TODO: Add Agent support if needed (requires golang.org/x/crypto/ssh/agent)

	// 3. Host Key Verification
	// We use ~/.ssh/known_hosts
	home, err := os.UserHomeDir()
	var hostKeyCallback ssh.HostKeyCallback
	if err == nil {
		knownHostsFile := home + "/.ssh/known_hosts"
		hostKeyCallback, err = knownhosts.New(knownHostsFile)
		if err != nil {
			// Fallback or error? User asked for "safely", but if known_hosts doesn't exist or is unreadable...
			// We might want to allow InsecureIgnoreHostKey ONLY if explicitly configured, but for now let's be strict
			// or just warn.
			// However, knownhosts.New returns error if file doesn't exist usually.
			// Let's assume strict check for "safely".
			// If file doesn't exist, create an empty one? No, knownhosts.New handles non-existent file? No.
			// If it fails, we return error.
			return nil, fmt.Errorf("failed to load known_hosts: %w", err)
		}
	} else {
		return nil, fmt.Errorf("failed to get user home dir: %w", err)
	}

	clientConfig := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         10 * time.Second,
	}

	addr := net.JoinHostPort(host, port)
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial ssh %s: %w", addr, err)
	}

	return client, nil
}

// Deploy connects to the project host and runs the deployment commands.
func (c *SSHClient) Deploy(project config.Project, output io.Writer, ref string) error {
	fmt.Fprintf(output, "Connecting to %s...\n", project.Host)

	// Buffer output for email notification
	var logBuffer strings.Builder
	multiOutput := io.MultiWriter(output, &logBuffer)

	client, err := c.connect(project)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer client.Close()

	commands := []string{
		fmt.Sprintf("cd %q", project.Path),
		"git fetch --all",
	}

	if ref != "" {
		// Checkout specific ref
		commands = append(commands, fmt.Sprintf("git checkout %s", ref))
	}

	commands = append(commands,
		"git pull",
		"docker compose pull",
		"docker compose up -d --build",
	)
	remoteCommand := strings.Join(commands, " && ")

	fmt.Fprintf(multiOutput, "Running: %s\n", remoteCommand)

	err = c.runSession(client, remoteCommand, multiOutput, multiOutput, nil)

	// Send notification if configured
	if c.Mailer != nil && len(project.NotifyEmails) > 0 {
		status := "SUCCESS"
		if err != nil {
			status = "FAILURE"
		}
		// Run in background to not block deployment response?
		// User said "Stream response", so maybe blocking here is fine as it's the last step.
		// Or we can just log it.
		// Since we want to notify "depending on yaml config", we do it here.
		notifErr := c.Mailer.SendDeploymentNotification(context.Background(), project.NotifyEmails, project.Name, status, logBuffer.String())
		if notifErr != nil {
			fmt.Fprintf(output, "Failed to send notification: %v\n", notifErr)
		} else {
			fmt.Fprintf(output, "Notification sent to %v\n", project.NotifyEmails)
		}
	}

	return err
}

// StreamLogs streams the logs from the remote project.
func (c *SSHClient) StreamLogs(ctx context.Context, project config.Project, output io.Writer) error {
	fmt.Fprintf(output, "Streaming logs from %s...\n", project.Host)

	client, err := c.connect(project)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer client.Close()

	commands := []string{
		fmt.Sprintf("cd %q", project.Path),
		"docker compose logs -f",
	}
	remoteCommand := strings.Join(commands, " && ")

	fmt.Fprintf(output, "Running: %s\n", remoteCommand)

	return c.runSession(client, remoteCommand, output, output, ctx)
}

// Restart restarts the project containers.
func (c *SSHClient) Restart(project config.Project, output io.Writer) error {
	fmt.Fprintf(output, "Restarting project on %s...\n", project.Host)

	client, err := c.connect(project)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer client.Close()

	commands := []string{
		fmt.Sprintf("cd %q", project.Path),
		"docker compose restart",
	}
	remoteCommand := strings.Join(commands, " && ")

	fmt.Fprintf(output, "Running: %s\n", remoteCommand)

	return c.runSession(client, remoteCommand, output, output, nil)
}

// Stop stops the project containers.
func (c *SSHClient) Stop(project config.Project, output io.Writer) error {
	fmt.Fprintf(output, "Stopping project on %s...\n", project.Host)

	client, err := c.connect(project)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer client.Close()

	commands := []string{
		fmt.Sprintf("cd %q", project.Path),
		"docker compose stop",
	}
	remoteCommand := strings.Join(commands, " && ")

	fmt.Fprintf(output, "Running: %s\n", remoteCommand)

	return c.runSession(client, remoteCommand, output, output, nil)
}

// ListServices fetches the list of services for the project.
func (c *SSHClient) ListServices(project config.Project) ([]string, error) {
	client, err := c.connect(project)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer client.Close()

	commands := []string{
		fmt.Sprintf("cd %q", project.Path),
		"docker compose config --services",
	}
	remoteCommand := strings.Join(commands, " && ")

	var b strings.Builder
	if err := c.runSession(client, remoteCommand, &b, &b, nil); err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	services := strings.Split(strings.TrimSpace(b.String()), "\n")
	var validServices []string
	for _, s := range services {
		s = strings.TrimSpace(s)
		if s != "" {
			validServices = append(validServices, s)
		}
	}

	return validServices, nil
}

// RunShell starts an interactive shell session for the service.
func (c *SSHClient) RunShell(project config.Project, service string) error {
	client, err := c.connect(project)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	commands := []string{
		fmt.Sprintf("cd %q", project.Path),
		fmt.Sprintf("docker compose exec -it %s /bin/sh", service),
	}
	remoteCommand := strings.Join(commands, " && ")

	// Request PTY
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		state, err := term.MakeRaw(fd)
		if err != nil {
			return fmt.Errorf("failed to make raw terminal: %w", err)
		}
		defer term.Restore(fd, state)

		w, h, err := term.GetSize(fd)
		if err == nil {
			if err := session.RequestPty("xterm", h, w, ssh.TerminalModes{
				ssh.ECHO:          1,
				ssh.TTY_OP_ISPEED: 14400,
				ssh.TTY_OP_OSPEED: 14400,
			}); err != nil {
				return fmt.Errorf("failed to request pty: %w", err)
			}

			// Handle window resize? (Advanced, skipped for now)
		}
	}

	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if err := session.Run(remoteCommand); err != nil {
		if _, ok := err.(*ssh.ExitError); ok {
			// Process exited with non-zero status
			return nil // It's expected when user exits shell
		}
		return fmt.Errorf("remote shell error: %w", err)
	}

	return nil
}

// GetStatus returns the status of the project.
func (c *SSHClient) GetStatus(ctx context.Context, project config.Project) (ProjectStatus, error) {
	client, err := c.connect(project)
	if err != nil {
		return ProjectStatus{}, fmt.Errorf("connection failed: %w", err)
	}
	defer client.Close()

	commands := []string{
		fmt.Sprintf("cd %q", project.Path),
		"(git rev-parse --abbrev-ref HEAD || echo '')",
		"echo '---SPLIT---'",
		"docker compose ps -a --format json",
	}
	remoteCommand := strings.Join(commands, " && ")

	var b strings.Builder
	if err := c.runSession(client, remoteCommand, &b, &b, ctx); err != nil {
		return ProjectStatus{}, fmt.Errorf("failed to get status: %w", err)
	}

	output := b.String()
	parts := strings.Split(output, "---SPLIT---")
	if len(parts) < 2 {
		return ProjectStatus{}, fmt.Errorf("unexpected output format: %s", output)
	}

	branch := strings.TrimSpace(parts[0])
	jsonOutput := strings.TrimSpace(parts[1])

	var containers []ContainerStatus
	if err := json.Unmarshal([]byte(jsonOutput), &containers); err != nil {
		lines := strings.Split(jsonOutput, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			var container ContainerStatus
			if err := json.Unmarshal([]byte(line), &container); err == nil {
				containers = append(containers, container)
			}
		}
	}

	status := "Down"
	runningCount := 0
	if len(containers) > 0 {
		for _, c := range containers {
			if strings.ToLower(c.State) == "running" {
				runningCount++
			}
		}
		if runningCount == len(containers) {
			status = "Healthy"
		} else if runningCount > 0 {
			status = "Partial"
		}
	}

	var lastDeployed time.Time
	for _, c := range containers {
		t, err := parseDockerTime(c.CreatedAt)
		if err == nil {
			if t.After(lastDeployed) {
				lastDeployed = t
			}
		}
	}

	return ProjectStatus{
		Name:           project.Name,
		Branch:         branch,
		LastDeployedAt: lastDeployed,
		Status:         status,
		Containers:     containers,
	}, nil
}

func (c *SSHClient) runSession(client *ssh.Client, cmd string, stdout, stderr io.Writer, ctx context.Context) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	session.Stdout = stdout
	session.Stderr = stderr

	if err := session.Start(cmd); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- session.Wait()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		// Attempt to send signal or close session
		// session.Signal(ssh.SIGINT) or just close
		// Closing session usually kills the remote command if pty is not allocated,
		// but here we didn't allocate pty for batch commands.
		// We'll return nil if cancelled intentionally? Or ctx error.
		// The caller (StreamLogs) expects to return when ctx is done.
		return ctx.Err()
	}
}

func parseDockerTime(s string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05 +0000 UTC",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}
