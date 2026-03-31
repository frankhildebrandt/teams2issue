package teams

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/frankhildebrandt/teams2issue/internal/config"
	"github.com/frankhildebrandt/teams2issue/internal/graph"
	"go.uber.org/fx"
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

type Source struct {
	cfg    config.Config
	logger *slog.Logger

	graph *graph.Client
	bus   *Bus

	subscriptions *SubscriptionManager
	webhooks      *WebhookHandler

	dedupeMu sync.Mutex
	dedupe   map[string]time.Time
}

func NewSource(cfg config.Config, logger *slog.Logger, graphClient *graph.Client, bus *Bus, subscriptions *SubscriptionManager, webhooks *WebhookHandler) *Source {
	return &Source{
		cfg:           cfg,
		logger:        logger,
		graph:         graphClient,
		bus:           bus,
		subscriptions: subscriptions,
		webhooks:      webhooks,
		dedupe:        map[string]time.Time{},
	}
}

func (s *Source) Name() string {
	return "teams-graph"
}

func (s *Source) Start(context.Context) error {
	return nil
}

type lifecycleParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	Source    *Source
	Logger    *slog.Logger
}

func RegisterLifecycle(p lifecycleParams) {
	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Ensure subscription exists at startup.
			sub, err := p.Source.subscriptions.EnsureChannelMessagesSubscription(ctx)
			if err != nil {
				return fmt.Errorf("ensure teams subscription: %w", err)
			}
			p.Logger.Info("teams subscription created",
				"subscription_id", sub.ID,
				"resource", sub.Resource,
				"expires_at", sub.ExpirationDateTime,
			)

			// Process notifications and lifecycle events in background.
			go p.Source.runNotificationLoop(ctx)
			go p.Source.runLifecycleLoop(ctx)

			// Renew subscription periodically.
			go func() {
				ticker := time.NewTicker(30 * time.Minute)
				defer ticker.Stop()
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						newExp := time.Now().Add(2 * time.Hour)
						renewed, renewErr := p.Source.subscriptions.Renew(context.Background(), sub.ID, newExp)
						if renewErr != nil {
							p.Logger.Error("failed to renew teams subscription", "error", renewErr, "subscription_id", sub.ID)
							continue
						}
						sub = renewed
						p.Logger.Info("teams subscription renewed", "subscription_id", sub.ID, "expires_at", sub.ExpirationDateTime)
					}
				}
			}()

			return nil
		},
	})
}

type graphChatMessage struct {
	ID        string `json:"id"`
	ReplyToID string `json:"replyToId"`

	CreatedDateTime time.Time `json:"createdDateTime"`

	Body struct {
		Content string `json:"content"`
	} `json:"body"`

	From struct {
		User struct {
			DisplayName string `json:"displayName"`
		} `json:"user"`
	} `json:"from"`
}

func (s *Source) runNotificationLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case payload := <-s.webhooks.Notifications():
			s.handleNotificationPayload(ctx, payload)
		}
	}
}

func (s *Source) runLifecycleLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case payload := <-s.webhooks.Lifecycle():
			for _, n := range payload.Value {
				if n.LifecycleEvent == "reauthorizationRequired" && n.SubscriptionID != "" {
					// Best-effort reauthorize. If it fails, Graph will keep retrying lifecycle events.
					_, err := s.graph.DoJSON(context.Background(), "POST", "/subscriptions/"+n.SubscriptionID+"/reauthorize", nil, nil)
					if err != nil {
						s.logger.Error("failed to reauthorize subscription", "error", err, "subscription_id", n.SubscriptionID)
					} else {
						s.logger.Info("subscription reauthorized", "subscription_id", n.SubscriptionID)
					}
				}
			}
		}
	}
}

func (s *Source) handleNotificationPayload(ctx context.Context, payload changeNotificationCollection) {
	for _, n := range payload.Value {
		if n.ChangeType != "created" || n.ResourceData == nil || n.ResourceData.ID == "" {
			continue
		}

		key := n.SubscriptionID + ":" + n.ChangeType + ":" + n.ResourceData.ID
		if s.isDuplicate(key) {
			continue
		}

		msg, err := s.fetchMessage(ctx, n.ResourceData.ID)
		if err != nil {
			s.logger.Error("failed to fetch message", "error", err, "message_id", n.ResourceData.ID)
			continue
		}

		threadID := msg.ID
		isReply := false
		if msg.ReplyToID != "" {
			threadID = msg.ReplyToID
			isReply = true
		}

		created := ChannelMessageCreated{
			TeamID:    s.cfg.Teams.TeamID,
			ChannelID: s.cfg.Teams.ChannelID,
			ThreadID:  threadID,
			MessageID: msg.ID,
			Body:      msg.Body.Content,
			From:      msg.From.User.DisplayName,
			CreatedAt: msg.CreatedDateTime,
			IsReply:   isReply,
		}
		s.bus.Publish(ctx, created)

		if !isReply {
			s.bus.Publish(ctx, ThreadCreated{
				TeamID:    created.TeamID,
				ChannelID: created.ChannelID,
				ThreadID:  created.ThreadID,
				MessageID: created.MessageID,
				Body:      created.Body,
				From:      created.From,
				CreatedAt: created.CreatedAt,
			})
		}
	}
}

func (s *Source) fetchMessage(ctx context.Context, messageID string) (graphChatMessage, error) {
	path := fmt.Sprintf("/teams/%s/channels/%s/messages/%s", s.cfg.Teams.TeamID, s.cfg.Teams.ChannelID, messageID)
	var out graphChatMessage
	_, err := s.graph.DoJSON(ctx, "GET", path, nil, &out)
	return out, err
}

func (s *Source) isDuplicate(key string) bool {
	now := time.Now()

	s.dedupeMu.Lock()
	defer s.dedupeMu.Unlock()

	// Best-effort pruning.
	for k, ts := range s.dedupe {
		if now.Sub(ts) > 6*time.Hour {
			delete(s.dedupe, k)
		}
	}

	if _, ok := s.dedupe[key]; ok {
		return true
	}
	s.dedupe[key] = now
	return false
}
