# teams2issue

`teams2issue` now contains a runnable Go daemon bootstrap with Cobra/Viper CLI wiring, Fx-based dependency injection, structured `slog` logging, Echo HTTP endpoints, Prometheus metrics, and stubbed integration seams for Teams, Jira, and OAuth2.

## Project layout

- `cmd/teams2issue`: CLI entrypoint.
- `internal/app`: Fx composition root and lifecycle management.
- `internal/config`: typed configuration loading, defaults, and validation.
- `internal/http`: Echo server, health checks, readiness, and metrics route.
- `internal/observability`: `slog` logger and Prometheus registry setup.
- `internal/provider/jira`: Jira scaffolding interface and noop implementation.
- `internal/source/teams`: Teams scaffolding interface and noop implementation.
- `internal/auth`: OAuth2 scaffolding service.
- `configs/`: local development config and example config.

## Local development

Install dependencies and run the daemon:

```bash
go mod tidy
go run ./cmd/teams2issue daemon
```

Use a different configuration file when needed:

```bash
go run ./cmd/teams2issue daemon --config ./configs/config.example.yaml
```

Environment variables override file values. Nested keys use the `TEAMS2ISSUE_` prefix and `_` separators, for example:

```bash
TEAMS2ISSUE_LOGGING_LEVEL=debug TEAMS2ISSUE_HTTP_ADDRESS=127.0.0.1:9090 go run ./cmd/teams2issue daemon
```

## Endpoints

After startup, the daemon exposes:

- `GET /healthz` for process health.
- `GET /readyz` for bootstrap readiness.
- `GET /metrics` for Prometheus metrics.

Quick smoke test:

```bash
curl -fsS http://127.0.0.1:8080/healthz
curl -fsS http://127.0.0.1:8080/readyz
curl -fsS http://127.0.0.1:8080/metrics | head
```

## Current integration status

The Teams source, Jira provider, and OAuth2 handling are scaffolded only. They are registered as interfaces with noop/stub implementations so the daemon can start without any external systems.
