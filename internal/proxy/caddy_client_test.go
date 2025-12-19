package proxy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/pmaojo/goploy/internal/config"
	"github.com/stretchr/testify/require"
)

type routePayload struct {
	ID    string `json:"@id"`
	Match []struct {
		Host []string `json:"host"`
	} `json:"match"`
	Handle []struct {
		Handler   string `json:"handler"`
		Upstreams []struct {
			Dial string `json:"dial"`
		} `json:"upstreams"`
	} `json:"handle"`
	Terminal bool `json:"terminal"`
}

func TestCaddyClient_ConfiguresDomains(t *testing.T) {
	var requestCount int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)

		require.Equal(t, http.MethodPut, r.Method)
		require.Equal(t, "/config/apps/http/servers/srv0/routes/goploy-project-alpha", r.URL.Path)

		var payload routePayload
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))

		require.Equal(t, "goploy-project-alpha", payload.ID)
		require.Len(t, payload.Match, 1)
		require.Equal(t, []string{"alpha.example.com", "www.alpha.test"}, payload.Match[0].Host)
		require.Len(t, payload.Handle, 1)
		require.Equal(t, "reverse_proxy", payload.Handle[0].Handler)
		require.Len(t, payload.Handle[0].Upstreams, 1)
		require.Equal(t, "localhost:3000", payload.Handle[0].Upstreams[0].Dial)

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewCaddyClient(nil)
	project := config.Project{
		Name: "Project Alpha",
		Caddy: &config.CaddyConfig{
			AdminURL: srv.URL,
			Server:   "srv0",
			Upstream: "localhost:3000",
		},
	}

	err := client.ConfigureDomains(context.Background(), project, []string{"alpha.example.com", "www.alpha.test"})
	require.NoError(t, err)
	require.Equal(t, int32(1), atomic.LoadInt32(&requestCount))
}

func TestCaddyClient_CreatesServerWhenMissing(t *testing.T) {
	var requestCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&requestCount, 1)

		switch current {
		case 1:
			require.Equal(t, "/config/apps/http/servers/goploy/routes/goploy-project-beta", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		case 2:
			require.Equal(t, http.MethodPut, r.Method)
			require.Equal(t, "/config/apps/http/servers/goploy", r.URL.Path)

			var serverCfg struct {
				Listen []string `json:"listen"`
			}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&serverCfg))
			require.ElementsMatch(t, []string{":443", ":80"}, serverCfg.Listen)
			w.WriteHeader(http.StatusCreated)
		case 3:
			require.Equal(t, http.MethodPut, r.Method)
			require.Equal(t, "/config/apps/http/servers/goploy/routes/goploy-project-beta", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected request %d", current)
		}
	}))
	defer srv.Close()

	client := NewCaddyClient(nil)
	project := config.Project{
		Name: "Project Beta",
		Caddy: &config.CaddyConfig{
			AdminURL: srv.URL + "/", // include trailing slash to test trimming
			Upstream: "localhost:4000",
		},
	}

	err := client.ConfigureDomains(context.Background(), project, []string{"beta.example.com"})
	require.NoError(t, err)
	require.Equal(t, int32(3), atomic.LoadInt32(&requestCount))
}
