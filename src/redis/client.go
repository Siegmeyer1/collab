package redis

import (
	"fmt"

	"collab/src/config"
	"github.com/go-redis/redis/v8"
)

func NewClient(cfg *config.RedisConfig) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       cfg.DB,
		Password: cfg.Password,
	})

	return client, nil
}
