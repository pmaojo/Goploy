## 🌟 About Goploy

Goploy is a powerful yet minimalist self-hosted deployment manager designed for developers who value control and simplicity. It offers a dual interface: a rich Terminal User Interface (TUI) for interactive management and a robust HTTP API for programmatic control. Seamlessly manage your remote Docker Compose projects, trigger deployments, monitor logs, and control containers with ease.

## ✨ Key Features

### 💻 Terminal User Interface (TUI)
*   **Intuitive Project Definition**: Configure all your projects using a simple `goploy.yaml` file.
*   **Interactive Navigation**: Effortlessly browse and select projects with keyboard shortcuts.
*   **One-Click Deployment**: Trigger `git pull` and `docker compose up -d` on remote hosts with a single key press.
*   **Real-time Logging**: Stream live deployment output and container logs directly in the terminal.
*   **Container Control**: Perform essential actions like Restart, Stop, and Shell into your containers.
*   **Status Monitoring**: Get a live view of container health, status, and metadata.

### 🌐 HTTP API
*   **Programmatic Control**: Integrate Goploy into your CI/CD pipelines or custom tools via HTTP endpoints.
*   **Trigger Deployments**: Deploy specific git references (branches, tags, or commits) remotely.
*   **Stream Logs**: Consume live deployment and container logs over HTTP for real-time feedback.
*   **Secure Access**: All API interactions are protected by API Key authentication.

### 📧 Notifications
*   **Email Alerts**: Automatically receive deployment status notifications (Success/Failure) for critical updates.

## 🚀 Architecture & Tech Stack

Goploy is built with performance and simplicity in mind, leveraging the following technologies:

*   **Go (Golang)**: The core language for its concurrency, performance, and ability to compile to a single static binary.
*   **Tview**: Powers the interactive and responsive Terminal User Interface.
*   **Echo**: A high-performance, minimalist Go web framework for the HTTP API.
*   **Docker Compose**: The standard for defining and running multi-container Docker applications, managed remotely by Goploy.
*   **SSH**: Securely executes commands on remote servers, enabling seamless deployments and container management.

## ⚡ Performance Highlights

Goploy is engineered for efficiency, making it ideal for resource-constrained environments:

*   **Blazing Fast Startup**: Initializes in less than **500ms**, getting you up and running instantly.
*   **Minimal Resource Utilization**: Maintains a low memory footprint of less than **30MB idle**, ensuring your server resources are free for your applications.

## 🛠️ Installation & Usage

### Prerequisites

Before you begin, ensure you have:
*   [Go (1.21+)](https://golang.org/doc/install) installed.
*   Docker and Docker Compose installed on your remote deployment targets.
*   SSH access to your remote servers with appropriate keys configured.

### Install Goploy

You can install Goploy directly using `go install`:

```bash
go install github.com/your-org/goploy@latest # Replace 'your-org' with the actual GitHub organization/user
```

This will place the `goploy` executable in your `$GOPATH/bin` directory.

### Configuration

Create a `goploy.yaml` file in your working directory (or specify its path with `GOPLOY_CONFIG_PATH` environment variable). See the [Configuration](#-configuration) section for details.

### Running the TUI

To launch the interactive Terminal User Interface:

```bash
goploy tui
```

### Running the HTTP API Server

To start the HTTP API server, you must provide the `GOPLOY_API_KEY` environment variable. Configure mailer settings if you want email notifications.

```bash
export GOPLOY_API_KEY="your-super-secret-api-key-here"
export SERVER_ECHO_LISTEN_ADDRESS=":8080" # Optional, default is :8080

# --- Optional: Configure SMTP for email notifications ---
export SERVER_MAILER_TRANSPORTER="smtp" # Use "mock" for testing
export SERVER_SMTP_HOST="smtp.mailtrap.io"
export SERVER_SMTP_PORT="2525"
export SERVER_SMTP_USERNAME="your-smtp-username"
export SERVER_SMTP_PASSWORD="your-smtp-password"
# --------------------------------------------------------

goploy server
```

## ⚙️ Configuration

### `goploy.yaml`

Define your projects, their remote hosts, repositories, and notification preferences in a `goploy.yaml` file:

```yaml
projects:
  - name: "Marketing Site"
    host: "deploy@192.168.1.10:22"
    path: "/var/www/marketing"
    repo: "git@github.com:company/marketing.git"
    identity_file: "~/.ssh/id_rsa" # Optional: specify SSH key
    notify_emails:
      - "devops@company.com"
      - "lead@company.com"
  - name: "Backend API"
    host: "admin@api.production.com"
    path: "/opt/services/backend"
    repo: "https://github.com/company/backend.git"
    # identity_file is optional; if omitted, SSH agent or default keys are used.
```

### Environment Variables

Configure the server, API authentication, and email settings using environment variables:

| Variable                          | Description                                                                                                                              | Default Value |
| :-------------------------------- | :--------------------------------------------------------------------------------------------------------------------------------------- | :------------ |
| `GOPLOY_API_KEY`                  | **Required**. The Bearer token used for authenticating API requests.                                                                     |               |
| `GOPLOY_CONFIG_PATH`              | Path to the `goploy.yaml` configuration file.                                                                                            | `./goploy.yaml` |
| `SERVER_ECHO_LISTEN_ADDRESS`      | The address and port for the HTTP API server to listen on.                                                                               | `:8080`       |
| `SERVER_MAILER_TRANSPORTER`       | Mail transport to use (`smtp` for real emails, `mock` for development/testing without sending).                                          | `mock`        |
| `SERVER_SMTP_HOST`                | SMTP host for sending emails (e.g., `smtp.gmail.com`). Required if `SERVER_MAILER_TRANSPORTER` is `smtp`.                                |               |
| `SERVER_SMTP_PORT`                | SMTP port (e.g., `587` for TLS, `465` for SSL). Required if `SERVER_MAILER_TRANSPORTER` is `smtp`.                                      |               |
| `SERVER_SMTP_USERNAME`            | SMTP username for authentication. Required if `SERVER_MAILER_TRANSPORTER` is `smtp`.                                                     |               |
| `SERVER_SMTP_PASSWORD`            | SMTP password for authentication. Required if `SERVER_MAILER_TRANSPORTER` is `smtp`.                                                     |               |

## 🌐 HTTP API Usage

All API requests must include the `Authorization` header with your `GOPLOY_API_KEY` as a Bearer token.

`Authorization: Bearer <GOPLOY_API_KEY>`

Assume the server is running on `http://localhost:8080` and `GOPLOY_API_KEY` is set to `$GOPLOY_API_KEY`.

### List Projects

`GET /api/v1/projects`
Returns a list of all configured projects.

```bash
curl -H "Authorization: Bearer $GOPLOY_API_KEY" http://localhost:8080/api/v1/projects
```

### Get Project Status

`GET /api/v1/projects/:name/status`
Returns the current status, active branch, and container health of a specific project.
*(Note: Project names with spaces should be URL-encoded)*

```bash
curl -H "Authorization: Bearer $GOPLOY_API_KEY" http://localhost:8080/api/v1/projects/Marketing%20Site/status
```

### Trigger Deployment

`POST /api/v1/projects/:name/deploy`
Triggers a deployment for the specified project. You can optionally provide a `ref` (branch, tag, or commit hash) in the request body. The response is streamed as plain text logs.

```bash
# Deploy 'main' branch
curl -X POST \
     -H "Authorization: Bearer $GOPLOY_API_KEY" \
     -H "Content-Type: application/json" \
     -d '{"ref": "main"}' \
     http://localhost:8080/api/v1/projects/Marketing%20Site/deploy

# Deploy default branch (as configured in goploy.yaml or repo)
curl -X POST \
     -H "Authorization: Bearer $GOPLOY_API_KEY" \
     http://localhost:8080/api/v1/projects/Backend%20API/deploy
```

### Stream Logs

`GET /api/v1/projects/:name/logs`
Streams the live `docker compose logs -f` output for the project's containers.

```bash
curl -H "Authorization: Bearer $GOPLOY_API_KEY" http://localhost:8080/api/v1/projects/Marketing%20Site/logs
```

## 🗺️ Roadmap

Goploy is continuously evolving. Here's a look at the current and planned features:

### Functional Requirements (FR)

*   [x] **FR1: Project Definition (YAML)**: Parse user-defined configuration (e.g., `goploy.yaml`) specifying projects.
*   [x] **FR2: Main TUI Navigation**: Interactive list of projects with keyboard navigation.
*   [x] **FR3: Interactive Deployment Workflow**: Trigger deployment via key press.
*   [x] **FR4: Real-time Logging (Deployment)**: Stream output of remote commands to a log panel.
*   [x] **FR5: Real-time Logging (Monitoring)**: Stream application logs (`docker compose logs -f`) to a log panel.
*   [x] **FR6: Basic Container Control**: Restart, Stop, and Shell Access via shortcuts.
*   [x] **FR7: Status and Metadata Display**: Monitor container status and metadata.
*   [x] **FR8: Remote Secure Execution**: Execute commands via SSH.
*   [ ] **FR9: Error Reporting**: Report failures in the TUI.
*   [x] **FR10: HTTP API**: Control deployments via REST API.
*   [x] **FR11: Notifications**: Email alerts for deployment status.

### Non-Functional Requirements (NFR)

*   [x] **NFR1: Performance (Startup)**: Fast initialization (< 500ms).
*   [x] **NFR2: Resource Utilization**: Low memory footprint (< 30MB idle).
*   [ ] **NFR3: Concurrency and Responsiveness**: Responsive UI during background tasks.
*   [ ] **NFR4: Distribution**: Single statically compiled binary.
*   [ ] **NFR5: Keyboard Usability**: Full keyboard control.

---
