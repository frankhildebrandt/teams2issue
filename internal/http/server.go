package daemonhttp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	nethttp "net/http"

	"github.com/frankhildebrandt/teams2issue/internal/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

type Server struct {
	echo       *echo.Echo
	server     *nethttp.Server
	readiness  *Readiness
	logger     *slog.Logger
	shutdowner fx.Shutdowner
	address    string
	listener   net.Listener
}

func NewEcho() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	return e
}

func NewServer(e *echo.Echo, cfg config.Config, readiness *Readiness, registry *prometheus.Registry, logger *slog.Logger, shutdowner fx.Shutdowner) *Server {
	RegisterRoutes(e, cfg, readiness, registry)

	server := &nethttp.Server{
		Addr:         cfg.HTTP.Address,
		Handler:      e,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	return &Server{
		echo:       e,
		server:     server,
		readiness:  readiness,
		logger:     logger,
		shutdowner: shutdowner,
		address:    cfg.HTTP.Address,
	}
}

func RegisterLifecycle(lifecycle fx.Lifecycle, server *Server, cfg config.Config) {
	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			listener, err := net.Listen("tcp", server.address)
			if err != nil {
				return fmt.Errorf("listen on %s: %w", server.address, err)
			}

			server.listener = listener
			server.readiness.SetReady(true)
			server.logger.Info("http server listening", "address", server.address)

			go func() {
				if err := server.server.Serve(listener); err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
					server.logger.Error("http server stopped unexpectedly", "error", err)
					if shutdownErr := server.shutdowner.Shutdown(fx.ExitCode(1)); shutdownErr != nil {
						server.logger.Error("failed to trigger application shutdown", "error", shutdownErr)
					}
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			server.readiness.SetReady(false)

			shutdownCtx, cancel := context.WithTimeout(ctx, cfg.HTTP.ShutdownTimeout)
			defer cancel()

			if err := server.server.Shutdown(shutdownCtx); err != nil {
				return fmt.Errorf("shutdown http server: %w", err)
			}

			if err := server.echo.Close(); err != nil {
				return fmt.Errorf("close echo server: %w", err)
			}

			server.logger.Info("http server stopped", "address", server.address)
			return nil
		},
	})
}
