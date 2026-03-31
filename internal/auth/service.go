package auth

import (
	"github.com/frankhildebrandt/teams2issue/internal/config"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewNoopService,
			fx.As(new(Service)),
		),
	),
)

type Service interface {
	Name() string
	Settings() config.OAuth2Config
}

type NoopService struct {
	cfg config.Config
}

func NewNoopService(cfg config.Config) *NoopService {
	return &NoopService{cfg: cfg}
}

func (s *NoopService) Name() string {
	return "oauth2-noop"
}

func (s *NoopService) Settings() config.OAuth2Config {
	return s.cfg.OAuth2
}
