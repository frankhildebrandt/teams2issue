package daemonhttp

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(
		NewReadiness,
		NewEcho,
		NewServer,
	),
	fx.Invoke(RegisterLifecycle),
)
