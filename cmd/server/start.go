package main

import (
	"context"
	"diploma/src/app"
	"diploma/src/postgres"
	"diploma/src/redis"
	"flag"
	"fmt"
)

func main() {
	var err error
	var port int

	fmt.Println("Starting server")

	flag.IntVar(&port, "p", 1234, "Provide a port number")

	flag.Parse()

	postgres.Pool, err = postgres.MakePgxPool("postgres://postgres:postgres@localhost:5433/editor")
	if err != nil {
		panic(fmt.Sprintf("Connect to postgres failed: %v", err))
	}

	redisClient, err := redis.NewClient("localhost:6380")
	if err != nil {
		panic(fmt.Sprintf("Build redis client failed: %v", err))
	}
	redis.DefaultQueue = redis.NewQueue(redisClient)
	go redis.DefaultQueue.Serve(context.Background())

	srv, err := app.BuildServer()
	if err != nil {
		panic(fmt.Sprintf("Server build failed: %v", err))
	}

	if err = srv.Start(fmt.Sprintf("localhost:%d", port)); err != nil {
		panic(fmt.Sprintf("Server start failed: %v", err))
	}
}
