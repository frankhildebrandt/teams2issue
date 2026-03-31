package teams

import "time"

// Minimal types to parse Microsoft Graph change notifications payloads.
// Reference: https://learn.microsoft.com/en-us/graph/change-notifications-delivery-webhooks

type changeNotificationCollection struct {
	Value []changeNotification `json:"value"`
}

type changeNotification struct {
	SubscriptionID string `json:"subscriptionId"`
	ClientState    string `json:"clientState"`
	ChangeType     string `json:"changeType"`
	TenantID       string `json:"tenantId"`
	Resource       string `json:"resource"`

	SubscriptionExpirationDateTime time.Time         `json:"subscriptionExpirationDateTime"`
	ResourceData                  *changeResourceData `json:"resourceData"`

	LifecycleEvent string `json:"lifecycleEvent"`
}

type changeResourceData struct {
	ODataType string `json:"@odata.type"`
	ODataID   string `json:"@odata.id"`
	ODataETag string `json:"@odata.etag"`
	ID        string `json:"id"`
}

