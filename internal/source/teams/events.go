package teams

import "time"

type DomainEvent interface {
	EventName() string
}

type ThreadCreated struct {
	TeamID    string
	ChannelID string

	ThreadID  string
	MessageID string

	Body      string
	From      string
	CreatedAt time.Time
}

func (ThreadCreated) EventName() string { return "teams.thread.created" }

type ChannelMessageCreated struct {
	TeamID    string
	ChannelID string

	ThreadID  string
	MessageID string

	Body      string
	From      string
	CreatedAt time.Time

	IsReply bool
}

func (ChannelMessageCreated) EventName() string { return "teams.channel_message.created" }
