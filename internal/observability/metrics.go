package observability

import (
	"fmt"

	"github.com/frankhildebrandt/teams2issue/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

func NewRegistry(cfg config.Config) (*prometheus.Registry, error) {
	registry := prometheus.NewRegistry()

	if err := registry.Register(collectors.NewGoCollector()); err != nil {
		return nil, fmt.Errorf("register go collector: %w", err)
	}
	if err := registry.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
		return nil, fmt.Errorf("register process collector: %w", err)
	}

	appInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: cfg.Metrics.Namespace,
			Name:      "app_info",
			Help:      "Static application metadata.",
		},
		[]string{"name", "environment"},
	)
	appInfo.WithLabelValues(cfg.App.Name, cfg.App.Environment).Set(1)

	if err := registry.Register(appInfo); err != nil {
		return nil, fmt.Errorf("register app info gauge: %w", err)
	}

	return registry, nil
}
