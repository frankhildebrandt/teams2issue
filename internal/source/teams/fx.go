package teams

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewSource,
			fx.As(new(EventSource)),
		),
		NewBus,
		NewWebhookHandler,
		NewSubscriptionManager,
	),
	fx.Invoke(RegisterWebhookRoutes),
	fx.Invoke(RegisterLifecycle),
)
