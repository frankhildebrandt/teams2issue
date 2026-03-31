package jira

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewNoopService,
			fx.As(new(IssueService)),
		),
	),
)
