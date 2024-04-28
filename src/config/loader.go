package config

import (
	"context"
	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/env"
	"github.com/heetch/confita/backend/file"
	"github.com/heetch/confita/backend/flags"
)

func LoadConfig() (*Config, error) {
	loader := confita.NewLoader(
		env.NewBackend(),
		file.NewBackend("./config.json"),
		flags.NewBackend(),
	)

	// default values go here
	cfg := &Config{
		Host: "127.0.0.1",
		Port: 1234,

		StorageBackend: StoragePostgres,
		QueueBackend:   RedisPubSub,

		Postgres: PostgresConfig{
			Login:    "postgres",
			Password: "postgres",
			DBName:   "editor",
			ConnURL:  "postgres://localhost:5433",
		},

		Redis: RedisConfig{
			Host: "localhost",
			Port: 6380,
			DB:   1,
		},
	}

	if err := loader.Load(context.Background(), cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
