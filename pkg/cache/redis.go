package cache

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/valentinesamuel/activelog/internal/config"
)

type RedisClient struct {
	client *redis.Client
}

func Connect() (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Address,
		Password: config.Redis.PASSWORD,
		DB:       config.Redis.DB,
	})

	pong, err := client.Ping(context.Background()).Result()
	if err != nil {
		panic(fmt.Sprintf("failed to connect to redis: %v", err))
		return nil, err
	}

	fmt.Println(pong)
	log.Println("⚡ ️🗑️  Connected to Redis")

	return &RedisClient{client: client}, nil
}

func (rc *RedisClient) Get(key string) (string, error) {
	value, err := rc.client.Get(context.Background(), key).Result()
	if err != nil {
		return "", err
	}
	return value, nil
}

func (rc *RedisClient) Set(key string, value string) error {
	err := rc.client.Set(context.Background(), key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}
