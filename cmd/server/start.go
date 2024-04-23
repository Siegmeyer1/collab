package main

import (
	"diploma/src/app"
	"diploma/src/postgres"
	"fmt"
)

func main() {
	var err error

	fmt.Println("Starting server")

	postgres.Pool, err = postgres.MakePgxPool("postgres://postgres:postgres@localhost:5433/editor")
	if err != nil {
		panic(fmt.Sprintf("Connect to postgres failed: %v", err))
	}

	srv, err := app.BuildServer()
	if err != nil {
		panic(fmt.Sprintf("Server build failed: %v", err))
	}

	if err = srv.Start("localhost:1234"); err != nil {
		panic(fmt.Sprintf("Server start failed: %v", err))
	}
}
