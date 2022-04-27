package server

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

var slog *log.Logger = nil

const (
	logPrefix = "[chomp:server] "
)

var serverLogFile string = "chomp.log"

func InitLog() {
	logf, ok := os.LookupEnv("CHOMP_LOGFILE")
	if ok {
		serverLogFile = logf
	}

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

	fmt.Printf("Logging to %s\n", serverLogFile)

	f, err := os.Create(serverLogFile)
	if err != nil {
		panic(err)
	}

	mw := io.MultiWriter(f)
	gin.DefaultWriter = mw
	slog = log.New(mw, logPrefix, log.LstdFlags)

	slog.Printf("Starting chomp at %s\n", time.Now().Local().Format("Mon Jan 2 15:04:05"))
	slog.Printf("Using gin %s\n", gin.Version)
}
