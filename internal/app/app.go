package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	authpkg "github.com/frankhildebrandt/teams2issue/internal/auth"
	"github.com/frankhildebrandt/teams2issue/internal/config"
	daemonhttp "github.com/frankhildebrandt/teams2issue/internal/http"
	"github.com/frankhildebrandt/teams2issue/internal/observability"
	jirapkg "github.com/frankhildebrandt/teams2issue/internal/provider/jira"
	teamspkg "github.com/frankhildebrandt/teams2issue/internal/source/teams"
	"go.uber.org/fx"
)

func New(cfg config.Config) *fx.App {
	return fx.New(
		fx.Supply(cfg),
		observability.Module,
		daemonhttp.Module,
		authpkg.Module,
		jirapkg.Module,
		teamspkg.Module,
		fx.Invoke(registerLifecycleLogs),
	)
}

func Run(ctx context.Context, cfg config.Config) error {
	app := New(cfg)

	startCtx, cancelStart := context.WithTimeout(ctx, cfg.App.StartupTimeout)
	defer cancelStart()

	if err := app.Start(startCtx); err != nil {
		return err
	}

	var runErr error

	select {
	case sig := <-app.Wait():
		if sig.ExitCode != 0 {
			runErr = fmt.Errorf("application exited with code %d", sig.ExitCode)
		}
	case <-ctx.Done():
		if !errors.Is(ctx.Err(), context.Canceled) {
			runErr = ctx.Err()
		}
	}

	stopCtx, cancelStop := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancelStop()

	if err := app.Stop(stopCtx); err != nil {
		return errors.Join(runErr, err)
	}

	return runErr
}

type lifecycleParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	Logger    *slog.Logger
	Config    config.Config
	Teams     teamspkg.EventSource
	Jira      jirapkg.IssueService
	OAuth2    authpkg.Service
}

func registerLifecycleLogs(params lifecycleParams) {
	params.Lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			params.Logger.Info("daemon bootstrap completed",
				"app", params.Config.App.Name,
				"environment", params.Config.App.Environment,
				"teams_source", params.Teams.Name(),
				"jira_provider", params.Jira.Name(),
				"oauth2_provider", params.OAuth2.Name(),
			)
			return nil
		},
		OnStop: func(context.Context) error {
			params.Logger.Info("daemon stopping", "app", params.Config.App.Name)
			return nil
		},
	})
}
