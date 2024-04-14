package main

import (
	"diploma/src/app"
	"fmt"
)

func main() {
	fmt.Println("Starting server")

	srv, err := app.BuildServer()
	if err != nil {
		panic(fmt.Sprintf("Server build failed: %v", err))
	}

	if err := srv.Start("localhost:1234"); err != nil {
		panic(fmt.Sprintf("Server start failed: %v", err))
	}
}
