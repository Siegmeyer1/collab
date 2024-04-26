package redis

import (
	"github.com/go-redis/redis/v8"
)

func NewClient(addr string) (*redis.Client, error) {
	return redis.NewClient(&redis.Options{Addr: addr}), nil
}
