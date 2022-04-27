package server

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var slog *log.Logger = nil

const logPrefix = "[chomp:server] "

func InitLog(file string) {
	f, err := os.Create(file)
	if err != nil {
		slog.Fatal(err)
	}

	mw := io.MultiWriter(f)
	gin.DefaultWriter = mw
	slog = log.New(mw, logPrefix, 0)

	slog.Printf("Starting chomp at %s\n", time.Now().Local().Format("Mon Jan 2 15:04:05"))
	slog.Printf("Using gin %s\n", gin.Version)
}
