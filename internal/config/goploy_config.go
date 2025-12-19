package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// GoployConfig represents the structure of the goploy.yaml configuration file.
type GoployConfig struct {
	Projects []Project `yaml:"projects"`
}

// Project represents a single project configuration.
type Project struct {
	Name         string       `yaml:"name"`
	Host         string       `yaml:"host"`
	User         string       `yaml:"user"`
	Port         string       `yaml:"port"`
	IdentityFile string       `yaml:"identity_file"`
	Path         string       `yaml:"path"`
	Repo         string       `yaml:"repo"`
	NotifyEmails []string     `yaml:"notify_emails"`
	Caddy        *CaddyConfig `yaml:"caddy"`
	Nginx        *NginxConfig `yaml:"nginx"`
}

type CaddyConfig struct {
	AdminURL string   `yaml:"admin_url"`
	Server   string   `yaml:"server"`
	Upstream string   `yaml:"upstream"`
	Email    string   `yaml:"email"`
	Domains  []string `yaml:"domains"`
}

type NginxConfig struct {
	ConfigPath       string   `yaml:"config_path"`        // e.g. /etc/nginx/sites-available
	SitesEnabledPath string   `yaml:"sites_enabled_path"` // e.g. /etc/nginx/sites-enabled
	ReloadCmd        string   `yaml:"reload_cmd"`         // e.g. "sudo systemctl reload nginx"
	Upstream         string   `yaml:"upstream"`
	Domains          []string `yaml:"domains"`
}

// ParseGoployConfig parses the provided YAML data into a GoployConfig struct.
func ParseGoployConfig(data []byte) (*GoployConfig, error) {
	var config GoployConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// LoadGoployConfig reads and parses the configuration file from the given path.
func LoadGoployConfig(path string) (*GoployConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseGoployConfig(data)
}
