package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/valentinesamuel/activelog/internal/config"
)

type Provider struct {
	client *redis.Client
}

func Connect() (*Provider, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Cache.Redis.Address,
		Password: config.Cache.Redis.Password,
		DB:       config.Cache.Redis.DB,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		panic(fmt.Sprintf("failed to connect to redis: %v", err))
	}

	return &Provider{client: client}, nil
}

func (rc *Provider) Get(key string) (string, error) {
	value, err := rc.client.Get(context.Background(), key).Result()
	if err != nil {
		return "", err
	}
	return value, nil
}

func (rc *Provider) Set(key string, value string, ttl time.Duration) error {
	err := rc.client.Set(context.Background(), key, value, ttl).Err()
	if err != nil {
		return err
	}
	return nil
}
