package config

// WebhookConfigType holds webhook bus configuration
type WebhookConfigType struct {
	Provider      string
	StreamMaxLen  int64
	RetryPollSecs int
	NATSUrl       string
}

// Webhook is the loaded webhook configuration
var Webhook *WebhookConfigType

func loadWebhook() *WebhookConfigType {
	return &WebhookConfigType{
		Provider:      GetEnv("WEBHOOK_PROVIDER", "memory"),
		StreamMaxLen:  int64(GetEnvInt("WEBHOOK_STREAM_MAX_LEN", 10000)),
		RetryPollSecs: GetEnvInt("WEBHOOK_RETRY_POLL_SECONDS", 30),
		NATSUrl:       GetEnv("NATS_URL", "nats://localhost:4222"),
	}
}
