package teams

import (
	"context"
	"errors"
)

type Event struct {
	ID      string
	Channel string
	Body    string
}

type EventSource interface {
	Name() string
	Start(context.Context) error
}

type NoopSource struct{}

func NewNoopSource() *NoopSource {
	return &NoopSource{}
}

func (s *NoopSource) Name() string {
	return "teams-noop"
}

func (s *NoopSource) Start(context.Context) error {
	return errors.New("teams integration is scaffolded but not implemented")
}
