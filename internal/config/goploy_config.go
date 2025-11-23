package config

import (
	"gopkg.in/yaml.v3"
)

// GoployConfig represents the structure of the goploy.yaml configuration file.
type GoployConfig struct {
	Projects []Project `yaml:"projects"`
}

// Project represents a single project configuration.
type Project struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Path string `yaml:"path"`
	Repo string `yaml:"repo"`
}

// ParseGoployConfig parses the provided YAML data into a GoployConfig struct.
func ParseGoployConfig(data []byte) (*GoployConfig, error) {
	var config GoployConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
