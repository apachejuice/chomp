package main

import (
	"fmt"
	"os"

	"github.com/apachejuice/chomp/internal/server"
	uuid "github.com/satori/go.uuid"
)

const (
	serverLogFile = "server.log"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s [addresss]\n", os.Args[0])
		os.Exit(2)
	}

	fmt.Printf("Starting Chomp, logging to: %s\n", serverLogFile)
	initialize()
	api, err := server.NewApi()
	if err != nil {
		panic(err)
	}

	api.SetEndpoints()
	if err = api.Run(os.Args[1]); err != nil {
		panic(err)
	}
}

func initialize() {
	// we wanna keep all logs, so copy them to a folder.
	if _, err := os.Stat(serverLogFile + ".old"); err == nil {
		dir, err := os.Stat("logs")
		if err != nil {
			os.Mkdir("logs", os.FileMode(0755))
			dir, _ = os.Stat("logs")
		}

		if !dir.Mode().IsDir() {
			panic("must have a 'logs' folder, it is a file right now")
		}

		// get a unique name for the file
		id, err := uuid.NewV4()
		if err != nil {
			panic(err)
		}

		err = os.Rename(serverLogFile+".old", fmt.Sprintf("logs/server-%s.log", id))
		if err != nil {
			panic(err)
		}
	}

	if info, err := os.Stat(serverLogFile); err == nil {
		// server.log exists; move it to server.log.old and rewrite the log file
		if info.Mode().IsRegular() {
			os.Rename(serverLogFile, serverLogFile+".old")
		}
	}

	server.InitLog(serverLogFile)
	server.LoadConfig()
}
