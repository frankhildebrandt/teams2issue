package teams

import (
	"net/http"
	"net/url"

	"github.com/frankhildebrandt/teams2issue/internal/config"
	"github.com/labstack/echo/v4"
)

type WebhookHandler struct {
	cfg config.Config

	notifications chan changeNotificationCollection
	lifecycle     chan changeNotificationCollection
}

func NewWebhookHandler(cfg config.Config) *WebhookHandler {
	return &WebhookHandler{
		cfg:           cfg,
		notifications: make(chan changeNotificationCollection, 512),
		lifecycle:     make(chan changeNotificationCollection, 128),
	}
}

func (h *WebhookHandler) Notifications() <-chan changeNotificationCollection { return h.notifications }
func (h *WebhookHandler) Lifecycle() <-chan changeNotificationCollection     { return h.lifecycle }

func (h *WebhookHandler) HandleNotifications(c echo.Context) error {
	if token := c.QueryParam("validationToken"); token != "" {
		decoded, err := url.QueryUnescape(token)
		if err != nil {
			return c.String(http.StatusBadRequest, "invalid validationToken")
		}
		return c.Blob(http.StatusOK, "text/plain; charset=utf-8", []byte(decoded))
	}

	var payload changeNotificationCollection
	if err := c.Bind(&payload); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	// Validate clientState on a best-effort basis. If absent, ignore (some lifecycle events).
	if expected := h.cfg.Teams.ClientState; expected != "" {
		for _, n := range payload.Value {
			if n.ClientState != "" && n.ClientState != expected {
				return c.NoContent(http.StatusUnauthorized)
			}
		}
	}

	select {
	case h.notifications <- payload:
	default:
		// Avoid blocking the webhook. Downstream will rely on Graph retries.
	}

	return c.NoContent(http.StatusAccepted)
}

func (h *WebhookHandler) HandleLifecycle(c echo.Context) error {
	if token := c.QueryParam("validationToken"); token != "" {
		decoded, err := url.QueryUnescape(token)
		if err != nil {
			return c.String(http.StatusBadRequest, "invalid validationToken")
		}
		return c.Blob(http.StatusOK, "text/plain; charset=utf-8", []byte(decoded))
	}

	var payload changeNotificationCollection
	if err := c.Bind(&payload); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	if expected := h.cfg.Teams.ClientState; expected != "" {
		for _, n := range payload.Value {
			if n.ClientState != "" && n.ClientState != expected {
				return c.NoContent(http.StatusUnauthorized)
			}
		}
	}

	select {
	case h.lifecycle <- payload:
	default:
	}

	return c.NoContent(http.StatusAccepted)
}

