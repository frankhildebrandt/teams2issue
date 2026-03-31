package teams

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/frankhildebrandt/teams2issue/internal/config"
	"github.com/frankhildebrandt/teams2issue/internal/graph"
)

type SubscriptionManager struct {
	cfg   config.Config
	graph *graph.Client
}

func NewSubscriptionManager(cfg config.Config, graphClient *graph.Client) *SubscriptionManager {
	return &SubscriptionManager{cfg: cfg, graph: graphClient}
}

type subscription struct {
	ID               string    `json:"id"`
	Resource         string    `json:"resource"`
	ChangeType       string    `json:"changeType"`
	NotificationURL  string    `json:"notificationUrl"`
	LifecycleURL     string    `json:"lifecycleNotificationUrl,omitempty"`
	ExpirationDateTime time.Time `json:"expirationDateTime"`
	ClientState      string    `json:"clientState"`
}

type createSubscriptionRequest struct {
	ChangeType              string `json:"changeType"`
	NotificationURL         string `json:"notificationUrl"`
	LifecycleNotificationURL string `json:"lifecycleNotificationUrl,omitempty"`
	Resource                string `json:"resource"`
	ExpirationDateTime      string `json:"expirationDateTime"`
	ClientState             string `json:"clientState"`
}

func (m *SubscriptionManager) EnsureChannelMessagesSubscription(ctx context.Context) (subscription, error) {
	if m.cfg.Teams.TeamID == "" || m.cfg.Teams.ChannelID == "" {
		return subscription{}, fmt.Errorf("teams.team_id and teams.channel_id must be set")
	}
	if m.cfg.Teams.WebhookBaseURL == "" {
		return subscription{}, fmt.Errorf("teams.webhook_base_url must be set (public https)")
	}
	if m.cfg.Teams.ClientState == "" {
		return subscription{}, fmt.Errorf("teams.client_state must be set")
	}

	base := strings.TrimRight(m.cfg.Teams.WebhookBaseURL, "/")
	notificationURL := base + "/graph/notifications"
	lifecycleURL := base + "/graph/lifecycle"

	resource := fmt.Sprintf("/teams/%s/channels/%s/messages", m.cfg.Teams.TeamID, m.cfg.Teams.ChannelID)
	exp := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339Nano)

	req := createSubscriptionRequest{
		ChangeType:              "created",
		NotificationURL:         notificationURL,
		LifecycleNotificationURL: lifecycleURL,
		Resource:                resource,
		ExpirationDateTime:      exp,
		ClientState:             m.cfg.Teams.ClientState,
	}

	var out subscription
	if _, err := m.graph.DoJSON(ctx, "POST", "/subscriptions", req, &out); err != nil {
		// Duplicate subscription requests return 409 Conflict. Try to locate and reuse the existing subscription.
		if strings.Contains(err.Error(), "status=409") {
			existing, findErr := m.findExistingSubscription(ctx, resource, notificationURL)
			if findErr != nil {
				return subscription{}, err
			}
			return existing, nil
		}
		return subscription{}, err
	}
	return out, nil
}

func (m *SubscriptionManager) Renew(ctx context.Context, subscriptionID string, expiration time.Time) (subscription, error) {
	if subscriptionID == "" {
		return subscription{}, fmt.Errorf("subscriptionID must not be empty")
	}

	body := map[string]string{
		"expirationDateTime": expiration.UTC().Format(time.RFC3339Nano),
	}

	var out subscription
	if _, err := m.graph.DoJSON(ctx, "PATCH", "/subscriptions/"+subscriptionID, body, &out); err != nil {
		return subscription{}, err
	}
	return out, nil
}

type subscriptionsList struct {
	Value []subscription `json:"value"`
}

func (m *SubscriptionManager) findExistingSubscription(ctx context.Context, resource string, notificationURL string) (subscription, error) {
	// Best-effort: list subscriptions and match by resource + notificationUrl.
	// NOTE: Tenants can have many subscriptions; if this becomes an issue, persist subscriptionId.
	var out subscriptionsList
	if _, err := m.graph.DoJSON(ctx, "GET", "/subscriptions?$top=999", nil, &out); err != nil {
		return subscription{}, err
	}

	for _, s := range out.Value {
		if s.Resource == resource && s.NotificationURL == notificationURL {
			return s, nil
		}
	}

	return subscription{}, fmt.Errorf("existing subscription not found for resource %q", resource)
}

