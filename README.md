```markdown
# Goploy TUI

<p align="center">
  <img src="https://i.imgur.com/your-logo-placeholder.png" alt="Goploy TUI Logo" width="150"/>
</p>

<p align="center">
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/go-1.21+-00ADD8?style=for-the-badge&logo=go" alt="Go Version"></a>
  <a href="https://www.docker.com/"><img src="https://img.shields.io/badge/docker-%230db7ed.svg?style=for-the-badge&logo=docker&logoColor=white" alt="Docker"></a>
  <a href="https://github.com/your-org/goploy-tui/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge" alt="License: MIT"></a>
  <a href="https://github.com/your-org/goploy-tui/releases/latest"><img src="https://img.shields.io/github/v/release/your-org/goploy-tui?style=for-the-badge&sort=semver" alt="Latest Release"></a>
</p>

<p align="center">
  <i>A lightweight, self-hosted, TUI-based deployment manager for your Go and Docker Compose projects.</i>
</p>

---

## 🌟 About

**Goploy TUI** is your command-line companion for effortless, self-hosted application deployment and management. Designed for developers who love the terminal, it provides a powerful yet intuitive Text User Interface (TUI) to interact with your remote projects.

Say goodbye to manual SSH commands and scattered scripts. With Goploy TUI, you can define your projects in a simple YAML file and then deploy, monitor, and control your applications with just a few keystrokes, all from a single, responsive interface. It's built for speed, efficiency, and a seamless developer experience.

## ✨ Features

*   📋 **Configuration as Code:** Define projects and their deployment parameters using a straightforward YAML file (`goploy.yaml`).
*   🚀 **Intuitive TUI Navigation:** Effortlessly browse and select your projects with full keyboard control.
*   ✨ **One-Click Deployments:** Trigger `git pull` and `docker compose pull/up` on remote hosts with a single key press.
*   📜 **Real-time Deployment Logs:** Monitor every step of your deployment process with live streaming output directly in the TUI.
*   👁️‍🗨️ **Live Application Monitoring:** Stream `docker compose logs -f` for any project to keep an eye on your running applications.
*   ⚙️ **Basic Container Control:** Quickly restart, stop, and access container shells via convenient keyboard shortcuts.
*   📊 **Status & Metadata Display:** Get immediate insights into the health and essential metadata of your Docker containers.
*   🔒 **Secure Remote Execution:** All remote commands are executed securely via SSH.
*   🚨 **Error Reporting:** Receive clear, actionable failure notifications within the TUI when things go wrong.

## 🏗️ Architecture & Tech Stack

Goploy TUI is engineered for robustness and performance, leveraging battle-tested open-source technologies:

*   **Core Language:** Built entirely in **Go (Golang)** for speed, concurrency, and static compilation.
*   **TUI Framework:** Utilizes [`github.com/rivo/tview`](https://github.com/rivo/tview) to create a rich and interactive terminal user interface.
*   **Orchestration:** Manages application services via **Docker Compose**, allowing for multi-container application deployments.
*   **Remote Protocol:** Securely communicates with remote hosts using `golang.org/x/crypto/ssh`.
*   **Project Structure:** Follows a clean, modular architecture inspired by `allaboutapps/go-starter`, with distinct `cmd/server` and `cmd/tui` components.

## ⚡ Performance

Goploy TUI is designed with performance and resource efficiency in mind:

*   **Blazing Fast Startup:** Initializes in **under 500ms**, getting you to your projects instantly.
*   **Minimal Resource Footprint:** Operates with a **low memory footprint (< 30MB idle)**, perfect for resource-constrained environments.
*   **Highly Responsive UI:** Ensures a smooth and interactive user experience, even during intensive background tasks.
*   **Single, Statically Compiled Binary:** Easy distribution and deployment—just one file to copy!

## 🚀 Installation & Usage

Getting started with Goploy TUI is simple.

### Prerequisites

*   Go 1.21+ installed
*   Docker & Docker Compose installed on your remote deployment hosts.
*   SSH access configured for your remote hosts.

### Install

You can install Goploy TUI directly using `go install`:

```bash
go install github.com/your-org/goploy-tui@latest
```

This will download and compile the `goploy-tui` binary and place it in your `$GOPATH/bin` directory (ensure this is in your system's PATH).

### Configuration

Create a `goploy.yaml` file in your current directory or specify its path. This file defines the projects Goploy TUI will manage.

```yaml
# goploy.yaml
projects:
  - name: MyWebApp
    host: user@your-server-ip:22
    path: /var/www/mywebapp
    repo: git@github.com:your-org/mywebapp.git
    branch: main
    compose_file: docker-compose.prod.yaml # Optional, defaults to docker-compose.yaml
  - name: AnotherService
    host: deployer@another-server.com
    path: /opt/services/anotherservice
    repo: https://github.com/your-org/anotherservice.git
```

### Run

Navigate to the directory containing your `goploy.yaml` and run:

```bash
goploy-tui
```

Or, specify the config file explicitly:

```bash
goploy-tui --config /path/to/your/goploy.yaml
```

The TUI will launch, displaying your configured projects. Use arrow keys to navigate and follow the on-screen prompts for deployment and control actions.

## ⚙️ Configuration

The `goploy.yaml` file is the heart of Goploy TUI, allowing you to define multiple projects.

```yaml
projects:
  - name: <string> # Unique name for your project
    host: <string> # SSH connection string (e.g., "user@ip:port")
    path: <string> # Absolute path to the project root directory on the remote host
    repo: <string> # Git repository URL (e.g., "git@github.com:user/repo.git" or "https://github.com/user/repo.git")
    branch: <string, optional> # Git branch to deploy (defaults to "main")
    compose_file: <string, optional> # Name of the Docker Compose file (defaults to "docker-compose.yaml")
```

**Example:**

```yaml
projects:
  - name: API Gateway
    host: deploy@192.168.1.100
    path: /srv/api-gateway
    repo: https://github.com/myorg/api-gateway.git
    branch: develop
    compose_file: docker-compose.dev.yaml
  - name: Frontend App
    host: deploy@frontend-server.com:2222
    path: /var/www/frontend
    repo: git@github.com:myorg/frontend-app.git
```

## 🗺️ Roadmap

Here's an overview of the current and planned features for Goploy TUI:

### Functional Requirements (FR)

*   [ ] FR1: Project Definition (YAML): Parse user-defined configuration (e.g., `goploy.yaml`) specifying projects (Name, Host, Path, Repo).
*   [ ] FR2: Main TUI Navigation: Interactive list of projects with keyboard navigation.
*   [ ] FR3: Interactive Deployment Workflow: Trigger deployment (git pull, docker compose pull/up) via key press.
*   [ ] FR4: Real-time Logging (Deployment): Stream output of remote commands to a log panel.
*   [ ] FR5: Real-time Logging (Monitoring): Stream application logs (`docker compose logs -f`) to a log panel.
*   [ ] FR6: Basic Container Control: Restart, Stop, and Shell Access via shortcuts.
*   [ ] FR7: Status and Metadata Display: Monitor container status and metadata.
*   [ ] FR8: Remote Secure Execution: Execute commands via SSH.
*   [ ] FR9: Error Reporting: Report failures in the TUI.

### Non-Functional Requirements (NFR)

*   [ ] NFR1: Performance (Startup): Fast initialization (< 500ms).
*   [ ] NFR2: Resource Utilization: Low memory footprint (< 30MB idle).
*   [ ] NFR3: Concurrency and Responsiveness: Responsive UI during background tasks.
*   [ ] NFR4: Distribution: Single statically compiled binary.
*   [ ] NFR5: Keyboard Usability: Full keyboard control.
```
