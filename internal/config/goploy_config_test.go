package config_test

import (
	"testing"

	"allaboutapps.dev/aw/go-starter/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGoployConfig_Valid(t *testing.T) {
	yamlData := []byte(`
projects:
  - name: "Project Alpha"
    host: "ssh://user@192.168.1.100:22"
    path: "/var/www/alpha"
    repo: "https://github.com/user/alpha.git"
  - name: "Project Beta"
    host: "ssh://admin@10.0.0.5"
    path: "/opt/beta"
`)

	cfg, err := config.ParseGoployConfig(yamlData)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Len(t, cfg.Projects, 2)

	assert.Equal(t, "Project Alpha", cfg.Projects[0].Name)
	assert.Equal(t, "ssh://user@192.168.1.100:22", cfg.Projects[0].Host)
	assert.Equal(t, "/var/www/alpha", cfg.Projects[0].Path)
	assert.Equal(t, "https://github.com/user/alpha.git", cfg.Projects[0].Repo)

	assert.Equal(t, "Project Beta", cfg.Projects[1].Name)
	assert.Equal(t, "ssh://admin@10.0.0.5", cfg.Projects[1].Host)
	assert.Equal(t, "/opt/beta", cfg.Projects[1].Path)
	assert.Empty(t, cfg.Projects[1].Repo) // Optional field check if omitted in yaml (though I didn't omit it in struct definition yet, assuming flexible)
}

func TestParseGoployConfig_InvalidYAML(t *testing.T) {
	yamlData := []byte(`
projects:
  - name: "Project Alpha"
    host: [this is broken
`)

	_, err := config.ParseGoployConfig(yamlData)
	assert.Error(t, err)
}
