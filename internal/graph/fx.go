package graph

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(
		NewHTTPClient,
		NewTokenProvider,
		NewClient,
	),
)

