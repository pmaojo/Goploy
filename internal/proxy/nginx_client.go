package proxy

import (
	"context"
	"fmt"
	"strings"

	"github.com/pmaojo/goploy/internal/config"
	"github.com/pmaojo/goploy/internal/deployment"
)

// NginxClient handles Nginx configuration via SSH.
type NginxClient struct {
	controller deployment.Controller
}

// NewNginxClient creates a new NginxClient.
func NewNginxClient(controller deployment.Controller) *NginxClient {
	return &NginxClient{
		controller: controller,
	}
}

// ConfigureDomains configures Nginx for the given project and domains.
func (n *NginxClient) ConfigureDomains(ctx context.Context, project config.Project, domains []string) error {
	if project.Nginx == nil {
		return fmt.Errorf("nginx configuration missing on project")
	}

	if len(domains) == 0 {
		return fmt.Errorf("at least one domain must be provided")
	}

	confName := project.Name
	// Sanitize name for filename
	confName = strings.ReplaceAll(confName, " ", "_")
	confName = strings.ToLower(confName)

	configContent := n.generateConfig(project.Nginx, domains)

	// Assume sites-available and sites-enabled structure by default, but allow override
	configPath := project.Nginx.ConfigPath
	if configPath == "" {
		configPath = "/etc/nginx/sites-available"
	}

	sitesEnabledPath := project.Nginx.SitesEnabledPath
	if sitesEnabledPath == "" {
		sitesEnabledPath = "/etc/nginx/sites-enabled"
	}

	// We need to upload to a temp location first because we likely don't have permission to write to /etc/nginx directly.
	// If the user is not root, we need sudo.
	// Our runCommand doesn't handle interactive sudo, so we assume passwordless sudo.
	remoteTempPath := fmt.Sprintf("/tmp/%s.nginx.conf", confName)
	remoteFinalPath := fmt.Sprintf("%s/%s", configPath, confName)

	// 1. Upload config to temp path
	if err := n.controller.UploadFile(project, []byte(configContent), remoteTempPath); err != nil {
		return fmt.Errorf("failed to upload nginx config: %w", err)
	}

	// 2. Move to final path (using sudo if needed)
	moveCmd := fmt.Sprintf("sudo mv %s %s", remoteTempPath, remoteFinalPath)
	if err := n.controller.RunCommand(project, moveCmd); err != nil {
		return fmt.Errorf("failed to move config file (ensure passwordless sudo is configured for the deploy user): %w", err)
	}

	// 3. Symlink if sites-enabled path is not set to "-" (explicit disable)
	if sitesEnabledPath != "-" {
		linkCmd := fmt.Sprintf("sudo ln -sf %s %s/%s", remoteFinalPath, sitesEnabledPath, confName)
		if err := n.controller.RunCommand(project, linkCmd); err != nil {
			return fmt.Errorf("failed to symlink config to %s: %w", sitesEnabledPath, err)
		}
	}

	// 4. Test config
	if err := n.controller.RunCommand(project, "sudo nginx -t"); err != nil {
		return fmt.Errorf("nginx config test failed: %w", err)
	}

	// 5. Reload Nginx
	reloadCmd := project.Nginx.ReloadCmd
	if reloadCmd == "" {
		reloadCmd = "sudo systemctl reload nginx"
	}
	if err := n.controller.RunCommand(project, reloadCmd); err != nil {
		return fmt.Errorf("failed to reload nginx: %w", err)
	}

	return nil
}

func (n *NginxClient) generateConfig(cfg *config.NginxConfig, domains []string) string {
	domainList := strings.Join(domains, " ")

	return fmt.Sprintf(`server {
    listen 80;
    server_name %s;

    location / {
        proxy_pass http://%s;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
`, domainList, cfg.Upstream)
}
