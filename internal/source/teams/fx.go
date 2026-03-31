package teams

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewNoopSource,
			fx.As(new(EventSource)),
		),
	),
)
