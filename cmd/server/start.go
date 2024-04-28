package main

import (
	"context"
	"diploma/src/app"
	"diploma/src/config"
	"diploma/src/postgres"
	"diploma/src/redis"
	"fmt"
)

func main() {
	var err error

	fmt.Println("Starting server")

	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("loading config: %v", err))
	}

	postgres.Pool, err = postgres.MakePgxPool(&cfg.Postgres)
	if err != nil {
		panic(fmt.Sprintf("Connect to postgres failed: %v", err))
	}

	redisClient, err := redis.NewClient(&cfg.Redis)
	if err != nil {
		panic(fmt.Sprintf("Build redis client failed: %v", err))
	}

	redis.DefaultQueue = redis.NewQueue(redisClient)
	go redis.DefaultQueue.Serve(context.Background())

	srv, err := app.BuildServer()
	if err != nil {
		panic(fmt.Sprintf("Server build failed: %v", err))
	}

	if err = srv.Start(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)); err != nil {
		panic(fmt.Sprintf("Server start failed: %v", err))
	}
}
