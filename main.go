package main

import (
	"fmt"

	"github.com/apachejuice/chomp/internal/server"
)

func main() {
	fmt.Println("Starting Chomp...")
	initialize()
	api, err := server.NewApi()
	if err != nil {
		panic(err)
	}

	api.SetEndpoints()
	if err = api.Run(); err != nil {
		panic(err)
	}
}

func initialize() {
	server.InitLog()
	server.LoadConfig()
}
