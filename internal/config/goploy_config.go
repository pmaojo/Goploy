package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// GoployConfig represents the structure of the goploy.yaml configuration file.
// It contains a list of projects managed by the application.
type GoployConfig struct {
	Projects []Project `yaml:"projects"`
}

// Project represents a single project configuration within the goploy.yaml file.
// It includes connection details, repository information, and deployment settings.
type Project struct {
	// Name is the unique display name of the project.
	Name string `yaml:"name"`
	// Host is the SSH connection string (e.g., "user@host:port").
	Host string `yaml:"host"`
	// User is the SSH username (optional if included in Host).
	User string `yaml:"user"`
	// Port is the SSH port (optional if included in Host).
	Port string `yaml:"port"`
	// IdentityFile is the path to the SSH private key file.
	IdentityFile string `yaml:"identity_file"`
	// Path is the absolute path to the project directory on the remote server.
	Path string `yaml:"path"`
	// Repo is the Git repository URL.
	Repo string `yaml:"repo"`
	// NotifyEmails is a list of email addresses to notify upon deployment completion.
	NotifyEmails []string `yaml:"notify_emails"`
	// Caddy holds configuration for Caddy web server integration.
	Caddy *CaddyConfig `yaml:"caddy"`
}

// CaddyConfig defines the configuration for Caddy web server management for a project.
type CaddyConfig struct {
	// AdminURL is the URL of the Caddy admin API.
	AdminURL string `yaml:"admin_url"`
	// Server is the identifier for the Caddy server (if applicable).
	Server string `yaml:"server"`
	// Upstream is the upstream address that Caddy proxies to.
	Upstream string `yaml:"upstream"`
	// Email is the email address used for ACME registration.
	Email string `yaml:"email"`
	// Domains is a list of domains served by this project.
	Domains []string `yaml:"domains"`
}

// ParseGoployConfig parses the provided YAML byte data into a GoployConfig struct.
//
// Parameters:
//   - data: The raw YAML data as a byte slice.
//
// Returns:
//   - *GoployConfig: The parsed configuration object.
//   - error: An error if parsing fails.
func ParseGoployConfig(data []byte) (*GoployConfig, error) {
	var config GoployConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// LoadGoployConfig reads and parses the goploy configuration file from the given file path.
//
// Parameters:
//   - path: The file path to the YAML configuration file.
//
// Returns:
//   - *GoployConfig: The parsed configuration object.
//   - error: An error if the file cannot be read or parsed.
func LoadGoployConfig(path string) (*GoployConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseGoployConfig(data)
}
