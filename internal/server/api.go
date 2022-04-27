package server

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/apachejuice/chomp/internal/server/auth"
	"github.com/gin-gonic/gin"
)

type API struct {
	eng *gin.Engine
	db  Database
}

// request json type
type requestJson map[string]string

func loadJson(c *gin.Context) (requestJson, error) {
	var data requestJson
	err := c.BindJSON(&data)
	return data, err
}

func json(c *gin.Context, status int, str string, args ...any) {
	d := []byte(fmt.Sprintf(str, args...))
	c.Data(status, "text/json", d)
}

func errJson(c *gin.Context, err error) {
	slog.Printf("API returned error response at endpoint %s (status %d) to %s: %s\n",
		c.Request.URL.Path, http.StatusBadRequest, c.ClientIP(), err.Error())
	json(c, http.StatusBadRequest, `{"error": "%s"}`, err.Error())
}

func params(c *gin.Context, names ...string) (map[string]string, error) {
	if len(names) == 0 {
		slog.Fatal("names must have a list of names")
	}

	res := make(map[string]string)
	for _, n := range names {
		text, ok := c.GetQuery(n)
		if !ok || text == "" {
			return nil, fmt.Errorf("required param '%s' missing", n)
		}

		res[n] = text
	}

	return res, nil
}

func NewApi() (*API, error) {
	db, err := NewDatabase()
	if err != nil {
		return nil, err
	}

	engine := gin.Default()
	engine.SetTrustedProxies(nil)
	return &API{
		eng: engine,
		db:  db,
	}, nil
}

func (a *API) Run(addr string) error {
	slog.Printf("Starting API %s with options: %s\n", config.APIConfig.Version, configStr)
	tlsConf := config.APIConfig.TLSConfig
	if tlsConf != nil {
		return a.eng.RunTLS(addr, tlsConf.CertFile, tlsConf.KeyFile)
	}

	return a.eng.Run(addr)
}

func (a *API) SetEndpoints() {
	br := config.APIConfig.BaseRoute
	a.eng.GET(filepath.Join(br, "/version"), func(c *gin.Context) {
		json(c, http.StatusOK, `{"version": "%s"}`, config.APIConfig.Version)
	})

	a.eng.POST(filepath.Join(br, "/login"), a.apiLogin)
	a.eng.POST(filepath.Join(br, "/logout"), a.apiLogout)
	a.eng.POST(filepath.Join(br, "/register"), a.apiRegister)
	a.eng.GET(filepath.Join(br, "/loggedIn"), a.apiLoggedIn)
}

func (a *API) apiLogin(c *gin.Context) {
	params, err := loadJson(c)
	if err != nil {
		errJson(c, err)
		return
	}

	session, err := a.db.Login(params["user"], params["pw"])
	if err != nil {
		errJson(c, err)
		return
	}

	json(c, http.StatusOK, `{"token": "%s"}`, session.Token)
	slog.Printf("New login from %s user '%s'\n", c.ClientIP(), session.Account.Username)
}

func (a *API) apiLogout(c *gin.Context) {
	params, err := loadJson(c)
	if err != nil {
		errJson(c, err)
		return
	}

	session, err := a.db.LogoutToken(params["token"])
	if err != nil {
		errJson(c, err)
		return
	}

	slog.Printf("Logging out from %s user '%s'\n", c.ClientIP(), session.Account.Username)
}

func (a *API) apiRegister(c *gin.Context) {
	params, err := loadJson(c)
	if err != nil {
		errJson(c, err)
		return
	}

	acc, err := auth.NewAccount(params["user"], params["pw"])
	if err != nil {
		errJson(c, err)
		return
	}

	err = a.db.AddAccount(acc)
	if err != nil {
		errJson(c, err)
		return
	}

	slog.Printf("User added: %s\n", acc.Username)
}

func (a *API) apiLoggedIn(c *gin.Context) {
	params, err := loadJson(c)
	if err != nil {
		errJson(c, err)
		return
	}

	session, err := a.db.GetSessionByToken(params["token"])
	if err != nil {
		errJson(c, err)
		return
	}

	loggedIn, err := a.db.IsLoggedIn(session.Account.Username)
	if err != nil {
		errJson(c, err)
		return
	}

	if !loggedIn {
		json(c, http.StatusOK, `{"account": "<none>"}`)
		return
	}

	json(c, http.StatusOK, `{"account": "%s"}`, session.Account.Username)
}
