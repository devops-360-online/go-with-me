package repositories

import (
    "github.com/go-redis/redis/v8"
    "github.com/devops-360-online/go-with-me/config"
    // "context"
)

func NewRedisClient(cfg *config.Config) *redis.Client {
    rdb := redis.NewClient(&redis.Options{
        Addr: cfg.RedisAddress,
    })
    return rdb
}
