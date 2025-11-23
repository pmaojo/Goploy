# Goploy TUI

Goploy TUI is a lightweight, self-hosted, TUI-based deployment manager.

## Roadmap

### Functional Requirements (FR)

- [x] **FR1: Project Definition (YAML)**: Parse user-defined configuration (e.g., `goploy.yaml`) specifying projects (Name, Host, Path, Repo).
- [x] **FR2: Main TUI Navigation**: Interactive list of projects with keyboard navigation.
- [ ] **FR3: Interactive Deployment Workflow**: Trigger deployment (git pull, docker compose pull/up) via key press.
- [ ] **FR4: Real-time Logging (Deployment)**: Stream output of remote commands to a log panel.
- [ ] **FR5: Real-time Logging (Monitoring)**: Stream application logs (`docker compose logs -f`) to a log panel.
- [ ] **FR6: Basic Container Control**: Restart, Stop, and Shell Access via shortcuts.
- [ ] **FR7: Status and Metadata Display**: Monitor container status and metadata.
- [ ] **FR8: Remote Secure Execution**: Execute commands via SSH.
- [ ] **FR9: Error Reporting**: Report failures in the TUI.

### Non-Functional Requirements (NFR)

- [ ] **NFR1: Performance (Startup)**: Fast initialization (< 500ms).
- [ ] **NFR2: Resource Utilization**: Low memory footprint (< 30MB idle).
- [ ] **NFR3: Concurrency and Responsiveness**: Responsive UI during background tasks.
- [ ] **NFR4: Distribution**: Single statically compiled binary.
- [ ] **NFR5: Keyboard Usability**: Full keyboard control.

### Technical Requirements (TR)

- [ ] **TR1: Core Language**: Go (Golang).
- [ ] **TR2: TUI Framework**: `github.com/rivo/tview`.
- [ ] **TR3: Orchestration**: Docker Compose.
- [ ] **TR4: Remote Protocol**: `golang.org/x/crypto/ssh`.
- [ ] **TR5: Project Structure**: Based on `allaboutapps/go-starter`, keeping `cmd/server` and adding `cmd/tui`.

## Project Requirements

These requirements ensure the tool meets the goals of being lightweight, self-hosted, and TUI-driven, using the Go ecosystem and Docker Compose orchestration.

### I. Functional Requirements (FR)
These define what the system must do to facilitate deployment and management.

* **FR1: Project Definition (YAML)**: The application must read and parse a user-defined configuration file (e.g., `goploy.yaml`) specifying projects. This configuration must include:
  * A unique project name.
  * Target host connection details (SSH URL/credentials).
  * The remote directory path containing the `docker-compose.yml` file.
  * The Git repository URL (for source-based deployments).
* **FR2: Main TUI Navigation**: The primary view must display an interactive, scrollable list of all defined projects (FR1) and allow navigation via keyboard input (arrows, tabs, etc.).
* **FR3: Interactive Deployment Workflow**: The user must be able to select a project and trigger the full, multi-step deployment sequence with a single key press (e.g., `[D] Deploy`). The sequence must execute on the target host:
  * `git pull` (to update source code).
  * `docker compose pull` (for pre-built images).
  * `docker compose up -d --build` (to build if necessary and run containers).
* **FR4: Real-time Logging (Deployment)**: During the deployment process (FR3), the TUI must display the live, streaming output (stdout/stderr) of all remote commands (Git, Docker build, Docker compose) in a dedicated log panel.
* **FR5: Real-time Logging (Monitoring)**: The user must be able to select a running project and view its current application logs using the `docker compose logs -f` command, streamed directly into a dedicated TUI panel (FR4).
* **FR6: Basic Container Control**: The user must be able to trigger essential container actions on a selected deployed project via TUI shortcuts:
  * Restart (`docker compose restart`).
  * Stop (`docker compose stop`).
  * Shell Access (`docker compose exec service-name /bin/sh`).
* **FR7: Status and Metadata Display**: The TUI must continuously monitor and display the operational status of the project's containers (Running, Stopped, Exited, etc.) and key metadata (e.g., Git branch, last deployed time).
* **FR8: Remote Secure Execution**: All deployment and control commands must be executed securely on the target host via SSH.
* **FR9: Error Reporting**: Any failure during the workflow (e.g., SSH connection failed, Git pull error, Docker build error) must immediately be reported to the user via a clear TUI alert/message in the log panel (FR4).

### II. Non-Functional Requirements (NFR)
These define how the system performs, focusing on the goals of being lightweight and robust.

* **NFR1: Performance (Startup)**: The application must initialize and render the main TUI screen quickly (e.g., under 500 milliseconds).
* **NFR2: Resource Utilization**: The application must maintain a very low memory footprint (e.g., ideally under 30 MB of RAM during idle monitoring).
* **NFR3: Concurrency and Responsiveness**: The TUI must remain fully responsive and navigable even while one or more deployments (FR3) or log streaming sessions (FR5) are running concurrently in the background using Go concurrency primitives.
* **NFR4: Distribution and Self-Contained Nature**: The final application must be distributed as a single, statically compiled binary file, minimizing dependencies on the execution host.
* **NFR5: Keyboard Usability**: The entire application must be fully usable and navigable using only the keyboard (shortcuts and directional keys).

### III. Technical Requirements (TR)
These define the required technology stack, leveraging your chosen Go starter.

* **TR1: Core Language**: The application must be developed using Go (Golang).
* **TR2: TUI Framework**: The application must utilize a robust Go TUI library, such as `github.com/rivo/tview`, for building the interactive console interface.
* **TR3: Orchestration**: The deployment core must rely on sending standard Docker Compose commands to the target host.
* **TR4: Remote Protocol**: Remote host connectivity and command execution must be handled using the Go SSH library (`golang.org/x/crypto/ssh`).
* **TR5: Project Structure**: The project must be built upon the existing structure of the `allaboutapps/go-starter`, adapting the configuration and logging components and replacing the HTTP server entrypoint (`cmd/server`) with a TUI entrypoint (`cmd/tui`).
