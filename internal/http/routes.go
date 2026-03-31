package daemonhttp

import (
	nethttp "net/http"

	"github.com/frankhildebrandt/teams2issue/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRoutes(e *echo.Echo, cfg config.Config, readiness *Readiness, registry *prometheus.Registry) {
	e.GET("/healthz", func(c echo.Context) error {
		return c.JSON(nethttp.StatusOK, map[string]string{"status": "ok"})
	})

	e.GET("/readyz", func(c echo.Context) error {
		if !readiness.IsReady() {
			return c.JSON(nethttp.StatusServiceUnavailable, map[string]string{"status": "starting"})
		}
		return c.JSON(nethttp.StatusOK, map[string]string{"status": "ready"})
	})

	if cfg.Metrics.Enabled {
		e.GET(cfg.Metrics.Path, echo.WrapHandler(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))
	}
}
