package config

type QueueConfigType struct {
	Provider string
}

var Queue *QueueConfigType

func loadQueue() *QueueConfigType {
	return &QueueConfigType{
		Provider: GetEnv("QUEUE_PROVIDER", ""),
	}
}
