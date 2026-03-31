package jira

import (
	"context"
	"errors"

	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewNoopService,
			fx.As(new(IssueService)),
		),
	),
)

type CreateIssueInput struct {
	Summary     string
	Description string
}

type Issue struct {
	Key string
}

type IssueService interface {
	Name() string
	CreateIssue(context.Context, CreateIssueInput) (Issue, error)
}

type NoopService struct{}

func NewNoopService() *NoopService {
	return &NoopService{}
}

func (s *NoopService) Name() string {
	return "jira-noop"
}

func (s *NoopService) CreateIssue(context.Context, CreateIssueInput) (Issue, error) {
	return Issue{}, errors.New("jira integration is scaffolded but not implemented")
}
