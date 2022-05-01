package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/apachejuice/chomp/internal/server"
	"github.com/gin-gonic/gin"
)

const defaultConfig = `
{
    "version": "v1.0",
    "apiConfig": {
        "apiVersion": "v1.0",
        "allowGuestLogin": false,
        "bannedIPs": [],
        "baseRoute": "/api/v1",
        "serveAddress": "<YOUR IP HERE>",
        "tlsConfig": null
    },
    "dbConfig": {
        "accountDatabase": "accounts.db"
    }
}
`

const usageTemplate = `Usage: %s [verb] [arguments]

[arguments] are ones specified after the verb below OR one in this list:
	--logfile=FILE		Sets the log file (default chomp.log)
				The old log file will be FILE.old.
	--help			Show this help text

[verb] is one of:
	init			Creates a new chomp.json with the given configuration
		--batch		Don't prompt the user for configurations, instead create a default one
	run			Runs chomp with the default configuration
		--debug		Sets debug mode (default release)

Bugreport address: <https://github.com/apachejuice/chomp/issues>
`

func showUsage() {
	fmt.Printf(usageTemplate, os.Args[0])
	os.Exit(2)
}

func ask(q string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s ", q)
	text, _ := reader.ReadString('\n')
	return strings.TrimSuffix(text, "\n")
}

func yesNo(q string) bool {
	ans := ask(fmt.Sprintf("%s [Y/N]", q))

	for {
		ans = strings.ToLower(ans)
		if ans == "n" || ans == "no" {
			return false
		} else if ans == "y" || ans == "yes" {
			return true
		}

		continue
	}
}

func cmdErrorf(base string, args ...any) {
	fmt.Printf(base, args...)
	os.Exit(2)
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		showUsage()
	}

	for _, a := range args {
		if a == "--help" {
			showUsage()
		}
	}

	verb := strings.ToLower(args[0])
	switch verb {
	case "init":
		runInit(args[1:])
		return
	case "run":
		startChomp(args[1:])
		return
	default:
		fmt.Printf("Unknown command verb: %s\n", verb)
		os.Exit(2)
	}
}

func runInit(args []string) {
	batch := false
	for _, e := range args {
		if e == "--batch" {
			batch = true
		} else {
			cmdErrorf("init: unknown argument: %s", e)
		}
	}

	if batch {
		err := os.WriteFile("chomp.json", []byte(defaultConfig), os.FileMode(0644))
		if err != nil {
			cmdErrorf("internal error: %s", err.Error())
		}
	} else {
		ver := ask("What is the chomp version you're using?")
		gl := yesNo("Will you allow guest logins?")
		banned := ask("Enter a comma-separated list of banned IP addresses:")
		base := ask("Enter the base route of your API (default '/'):")
		if base == "" {
			base = "/"
		}

		addr := ask("Enter the address (host:port) you wish to serve at:")
		accDb := ask("Enter the name for the accounts database file:")
		c := server.ChompConfig{
			Version: ver,
			APIConfig: server.APIConfig{
				Version:         ver,
				AllowGuestLogin: gl,
				BannedIPs:       strings.Split(strings.ReplaceAll(banned, " ", ""), ","),
				BaseRoute:       base,
				ServeAddress:    addr,
			},
			DBConfig: server.DatabaseConfig{
				AccountDatabase: accDb,
			},
		}

		useTLS := yesNo("Do you wish to enable TLS? (you need a certificate and key)")
		if useTLS {
			whitelist := ask("Enter a comma-separated list of hosts to whitelist:")
			dirCache := ask("Enter the cache directory:")

			c.APIConfig.TLSConfig = &server.TLSConfig{}
			c.APIConfig.TLSConfig.HostWhitelist = strings.Split(strings.ReplaceAll(whitelist, " ", ""), ",")
			c.APIConfig.TLSConfig.DirCache = dirCache
		}

		data, err := json.Marshal(c)
		if err != nil {
			cmdErrorf("internal error: %s", err.Error())
		}

		os.WriteFile("chomp.json", data, os.FileMode(0644))
	}

	fmt.Println("wrote chomp.json")
}

func startChomp(args []string) {
	gin.SetMode(gin.ReleaseMode)
	for _, e := range args {
		if e == "--debug" {
			gin.SetMode(gin.DebugMode)
		} else {
			cmdErrorf("init: unknown argument: %s", e)
		}
	}

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
