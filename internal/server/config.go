package server

import (
	encjson "encoding/json"
	"log"
	"os"
)

type ChompConfig struct {
	Version   string         `json:"version"`
	APIConfig APIConfig      `json:"apiConfig"`
	DBConfig  DatabaseConfig `json:"dbConfig"`
}

type APIConfig struct {
	Version         string     `json:"apiVersion"`
	AllowGuestLogin bool       `json:"allowGuestLogin"`
	BannedIPs       []string   `json:"bannedIPs"`
	BaseRoute       string     `json:"baseRoute"`
	ServeAddress    string     `json:"serveAddress"`
	TLSConfig       *TLSConfig `json:"tlsConfig"`
}

type DatabaseConfig struct {
	AccountDatabase string `json:"accountDatabase"`
}

type TLSConfig struct {
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
}

var config *ChompConfig = nil
var configStr = ""

const configFile = "chomp.json"

func LoadConfig() {
	data, err := os.ReadFile(configFile)
	if err != nil {
		slog.Fatal(err)
	}

	var c *ChompConfig = new(ChompConfig)
	err = encjson.Unmarshal(data, c)
	if err != nil {
		slog.Fatal(err)
	}

	slog.Printf("Loaded configuration from %s\n", configFile)
	slog.Println("====================== CHOMP ======================")
	slog.Printf("Welome to Chomp %s!\n", c.Version)
	slog.Println("===================================================")

	configStr = string(data)
	config = c

	if c.APIConfig.TLSConfig == nil {
		log.Println("WARNING: TLSConfig not set, serving HTTP")
		log.Println("DO NOT SEND SENSITIVE DATA OVER THIS CONNECTION!")
		slog.Println("Running on http; no encryption enabled!")
	}
}
