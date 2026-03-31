package teams

import (
	"github.com/frankhildebrandt/teams2issue/internal/config"
	"github.com/labstack/echo/v4"
)

func RegisterWebhookRoutes(e *echo.Echo, _ config.Config, handler *WebhookHandler) {
	// Change notifications webhook endpoint (notificationUrl)
	e.POST("/graph/notifications", handler.HandleNotifications)

	// Lifecycle notifications webhook endpoint (lifecycleNotificationUrl)
	e.POST("/graph/lifecycle", handler.HandleLifecycle)
}

