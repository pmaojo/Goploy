package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode"

	"github.com/pmaojo/goploy/internal/config"
)

// Configurator defines the contract for updating HTTP domain routing.
type Configurator interface {
	ConfigureDomains(ctx context.Context, project config.Project, domains []string) error
}

// CaddyClient talks to the Caddy admin API.
type CaddyClient struct {
	httpClient *http.Client
}

// NewCaddyClient builds a Configurator backed by the Caddy admin API.
func NewCaddyClient(httpClient *http.Client) *CaddyClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &CaddyClient{httpClient: httpClient}
}

var errNotFound = errors.New("resource not found")

const defaultServerName = "goploy"

// ConfigureDomains ensures the given domains route to the project's upstream via Caddy.
func (c *CaddyClient) ConfigureDomains(ctx context.Context, project config.Project, domains []string) error {
	if project.Caddy == nil {
		return errors.New("caddy configuration missing on project")
	}

	baseURL := strings.TrimSuffix(project.Caddy.AdminURL, "/")
	if baseURL == "" {
		return errors.New("caddy admin_url is required")
	}

	upstream := strings.TrimSpace(project.Caddy.Upstream)
	if upstream == "" {
		return errors.New("caddy upstream is required")
	}

	if len(domains) == 0 {
		return errors.New("at least one domain must be provided")
	}

	serverName := project.Caddy.Server
	if serverName == "" {
		serverName = defaultServerName
	}

	routeID := routeIDFromProject(project.Name)
	route := buildRoutePayload(routeID, domains, upstream)
	routeURL := fmt.Sprintf("%s/config/apps/http/servers/%s/routes/%s", baseURL, serverName, routeID)

	if err := c.putJSON(ctx, routeURL, route); err != nil {
		if !errors.Is(err, errNotFound) {
			return err
		}

		serverURL := fmt.Sprintf("%s/config/apps/http/servers/%s", baseURL, serverName)
		serverPayload := map[string]any{
			"listen": []string{":443", ":80"},
		}

		if err := c.putJSON(ctx, serverURL, serverPayload); err != nil {
			return err
		}

		return c.putJSON(ctx, routeURL, route)
	}

	return nil
}

type caddyRoute struct {
	ID       string        `json:"@id"`
	Match    []caddyMatch  `json:"match"`
	Handle   []caddyHandle `json:"handle"`
	Terminal bool          `json:"terminal"`
}

type caddyMatch struct {
	Host []string `json:"host"`
}

type caddyHandle struct {
	Handler   string          `json:"handler"`
	Upstreams []caddyUpstream `json:"upstreams"`
}

type caddyUpstream struct {
	Dial string `json:"dial"`
}

func buildRoutePayload(routeID string, domains []string, upstream string) caddyRoute {
	return caddyRoute{
		ID: routeID,
		Match: []caddyMatch{
			{Host: domains},
		},
		Handle: []caddyHandle{
			{
				Handler: "reverse_proxy",
				Upstreams: []caddyUpstream{
					{Dial: upstream},
				},
			},
		},
		Terminal: true,
	}
}

func (c *CaddyClient) putJSON(ctx context.Context, url string, payload any) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return errNotFound
	}

	if resp.StatusCode >= http.StatusBadRequest {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("caddy admin request failed (status %d): %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}

	return nil
}

func routeIDFromProject(name string) string {
	lower := strings.ToLower(name)
	var b strings.Builder
	for _, r := range lower {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		case r == ' ' || r == '_' || r == '-':
			b.WriteRune('-')
		}
	}

	cleaned := strings.Trim(b.String(), "-")
	if cleaned == "" {
		cleaned = "project"
	}
	return "goploy-" + cleaned
}
