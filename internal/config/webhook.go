package config

// WebhookConfigType holds webhook bus configuration
type WebhookConfigType struct {
	Provider string
}

// Webhook is the loaded webhook configuration
var Webhook *WebhookConfigType

func loadWebhook() *WebhookConfigType {
	return &WebhookConfigType{
		Provider: GetEnv("WEBHOOK_PROVIDER", "memory"),
	}
}
