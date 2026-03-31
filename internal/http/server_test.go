package daemonhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/frankhildebrandt/teams2issue/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
)

func TestRegisterRoutes(t *testing.T) {
	e := echo.New()
	readiness := NewReadiness()
	readiness.SetReady(true)
	registry := prometheus.NewRegistry()
	cfg := config.Config{
		Metrics: config.MetricsConfig{
			Enabled: true,
			Path:    "/metrics",
		},
	}

	RegisterRoutes(e, cfg, readiness, registry)

	for _, tc := range []struct {
		path string
		code int
	}{
		{path: "/healthz", code: http.StatusOK},
		{path: "/readyz", code: http.StatusOK},
		{path: "/metrics", code: http.StatusOK},
	} {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		if rec.Code != tc.code {
			t.Fatalf("expected %s to return %d, got %d", tc.path, tc.code, rec.Code)
		}
	}
}
