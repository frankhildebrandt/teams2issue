package teams

import (
	"context"
	"log/slog"
	"sync"
)

type Bus struct {
	logger *slog.Logger

	mu   sync.RWMutex
	subs map[chan DomainEvent]struct{}
}

func NewBus(logger *slog.Logger) *Bus {
	return &Bus{
		logger: logger,
		subs:   map[chan DomainEvent]struct{}{},
	}
}

func (b *Bus) Subscribe(buffer int) (<-chan DomainEvent, func()) {
	if buffer <= 0 {
		buffer = 64
	}

	ch := make(chan DomainEvent, buffer)

	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()

	unsubscribe := func() {
		b.mu.Lock()
		if _, ok := b.subs[ch]; ok {
			delete(b.subs, ch)
			close(ch)
		}
		b.mu.Unlock()
	}

	return ch, unsubscribe
}

func (b *Bus) Publish(ctx context.Context, evt DomainEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subs {
		select {
		case <-ctx.Done():
			return
		case ch <- evt:
		default:
			b.logger.Warn("dropping domain event (subscriber slow)",
				"event", evt.EventName(),
			)
		}
	}
}

