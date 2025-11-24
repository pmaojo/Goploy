package deployment

import (
	"time"
)

// ContainerStatus represents the status of a single container.
type ContainerStatus struct {
	Name      string    `json:"Name"`
	State     string    `json:"State"` // e.g., "running", "exited"
	Status    string    `json:"Status"` // e.g., "Up 2 hours", "Exited (0) 5 seconds ago"
	CreatedAt string    `json:"CreatedAt"` // Raw timestamp string
    ExitCode  int       `json:"ExitCode"`
    Service   string    `json:"Service"`
}

// ProjectStatus represents the aggregated status of the project.
type ProjectStatus struct {
	Name            string
	Branch          string
	LastDeployedAt  time.Time
	Status          string // "Healthy", "Partial", "Down"
	Containers      []ContainerStatus
}
