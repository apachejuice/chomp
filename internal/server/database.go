package server

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/apachejuice/chomp/internal/server/auth"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const (
	dbPath = "chess.db"
)

var (
	errTokenExpired = fmt.Errorf("token expired")
)

// The database holds information about the accounts, games and logins of the server
type Database struct {
	db *sql.DB
}

func NewDatabase() (Database, error) {
	_, err := os.Stat(dbPath)
	empty := errors.Is(err, os.ErrNotExist)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return Database{}, err
	}

	// we do wanna make sure the database is actually fine; Open() may not check anything
	err = db.Ping()
	if err != nil {
		return Database{}, err
	}

	slog.Printf("Opened database %s\n", dbPath)
	if empty {
		dbInit(db, config.APIConfig.AllowGuestLogin)
	}

	return Database{db: db}, nil
}

func (d *Database) LogoutToken(token string) (auth.Session, error) {
	session, err := d.GetSessionByToken(token)
	if err != nil {
		return auth.Session{}, err
	}

	return session, d.Logout(session)
}

func (d *Database) GetSessionByToken(token string) (auth.Session, error) {
	if !d.checkToken(token) {
		return auth.Session{}, errTokenExpired
	}

	// the token is always bound to an account, kind of like a username and password combined
	statement := "SELECT Token, Username FROM Accounts"
	rows, err := d.db.Query(statement)
	if err != nil {
		return auth.Session{}, err
	}

	defer rows.Close()
	for rows.Next() {
		var dbToken string
		var user string
		rows.Scan(&dbToken, &user)
		if dbToken == token {
			// get the account
			acc, err := d.GetAccount(user)
			if err != nil {
				return auth.Session{}, err
			}

			return auth.Session{Account: acc, Token: token}, nil
		}
	}

	err = rows.Err()
	if err != nil {
		return auth.Session{}, err
	}

	return auth.Session{}, fmt.Errorf("invalid token")
}

func (d *Database) GetAccount(username string) (auth.Account, error) {
	if !d.hasAccount(username) {
		return auth.Account{}, fmt.Errorf("no such account: %s", username)
	}

	statement := "SELECT Username, PwHash FROM Accounts"
	rows, err := d.db.Query(statement)
	if err != nil {
		return auth.Account{}, err
	}

	defer rows.Close()
	for rows.Next() {
		var user string
		var pwhash []byte
		rows.Scan(&user, &pwhash)
		if user == username {
			return auth.Account{Username: user, PwHash: pwhash}, nil
		}
	}

	err = rows.Err()
	if err != nil {
		return auth.Account{}, err
	}

	return auth.Account{}, fmt.Errorf("something went wrong")
}

func (d *Database) AddAccount(account auth.Account) error {
	if d.hasAccount(account.Username) {
		return fmt.Errorf("cannot add account: username '%s' is taken", account.Username)
	}

	statement := fmt.Sprintf(`
	INSERT INTO Accounts (Username, PwHash)
	VALUES ('%s', '%s');
	`, account.Username, string(account.PwHash))

	_, err := d.db.Exec(statement)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) Login(username, password string) (auth.Session, error) {
	if !d.hasAccount(username) {
		return auth.Session{}, fmt.Errorf("no such account: %s", username)
	}

	acc, err := d.GetAccount(username)
	if err != nil {
		return auth.Session{}, err
	}

	loggedIn, err := d.IsLoggedIn(username)
	if err != nil {
		return auth.Session{}, err
	}

	if loggedIn {
		return auth.Session{}, fmt.Errorf("already logged in")
	}

	// ok, not logged in and account exists: set the login bit
	_, err = d.db.Exec(fmt.Sprintf("UPDATE Accounts SET LoggedIn=1 WHERE Username = '%s'", username))
	if err != nil {
		return auth.Session{}, err
	}

	if checkPw(password, acc.PwHash) {
		token := auth.GetToken()
		d.addSessionToken(acc, token)
		return auth.Session{Account: acc, Token: token}, nil
	}

	return auth.Session{}, fmt.Errorf("invalid password")
}

func (d *Database) Logout(session auth.Session) error {
	_, err := d.db.Exec(fmt.Sprintf("UPDATE Accounts SET LoggedIn = 0, TokenExpiration = NULL WHERE Username = '%s'", session.Account.Username))
	if err != nil {
		return err
	}

	return d.removeSessionToken(session.Account)
}

func checkPw(password string, hash []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, []byte(password))
	return err == nil
}

func (d *Database) addSessionToken(acc auth.Account, token string) {
	_, err := d.db.Exec(fmt.Sprintf("UPDATE Accounts SET Token = '%s', TokenExpiration = datetime('now') WHERE Username = '%s';", token, acc.Username))
	if err != nil {
		slog.Fatal(err)
	}

	slog.Printf("Created a new session token for user '%s' - expires at %s\n", acc.Username, time.Now().Local().Add(time.Hour))
}

func (d *Database) IsLoggedIn(username string) (bool, error) {
	rows, err := d.db.Query("SELECT LoggedIn, Username FROM Accounts;")
	if err != nil {
		return false, err
	}

	defer rows.Close()
	for rows.Next() {
		var loggedIn int
		var user string
		err = rows.Scan(&loggedIn, &user)
		if err != nil {
			return false, err
		}

		if user == username {
			return loggedIn == 1, nil
		}
	}

	err = rows.Err()
	if err != nil {
		return false, err
	}

	return false, fmt.Errorf("no such user")
}

func (d *Database) getSessionToken(acc auth.Account) (string, error) {
	rows, err := d.db.Query("SELECT Username, Token FROM Accounts;")
	if err != nil {
		return "", err
	}

	defer rows.Close()
	for rows.Next() {
		var user, token string
		rows.Scan(&user, &token)

		if user == acc.Username {
			return token, nil
		}
	}

	err = rows.Err()
	if err != nil {
		return "", err
	}

	return "", fmt.Errorf("not logged in")
}

func (d *Database) removeSessionToken(acc auth.Account) error {
	token, err := d.getSessionToken(acc)
	if err != nil {
		return err
	} else if err == nil && token == "" {
		return fmt.Errorf("no session token to begin with")
	}

	_, err = d.db.Exec(fmt.Sprintf("UPDATE Accounts SET Token = '' WHERE Username = '%s';", acc.Username))
	if err != nil {
		return err
	}

	slog.Printf("Removed session token for user '%s'\n", acc.Username)
	return nil
}

func (d *Database) hasAccount(username string) bool {
	statement := "SELECT Username FROM Accounts;"
	rows, err := d.db.Query(statement)
	if err != nil {
		return false
	}

	defer rows.Close()
	for rows.Next() {
		var user string
		err = rows.Scan(&user)
		if err != nil {
			return false
		}

		if username == user {
			return true
		}
	}

	err = rows.Err()
	if err != nil {
		return false
	}

	return false
}

func (d *Database) checkToken(token string) bool {
	rows, err := d.db.Query(fmt.Sprintf("SELECT Token, TokenExpiration FROM Accounts WHERE Token = '%s'", token))
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()
	for rows.Next() {
		var dbToken string
		var tokenExpires time.Time
		err = rows.Scan(&dbToken, &tokenExpires)
		if err != nil {
			log.Fatal(err)
		}

		if dbToken == token {
			time := tokenExpires.Add(time.Hour).In(time.Local)
			if time.After(time.Local()) {
				d.db.Exec(
					fmt.Sprintf("UPDATE Accounts SET LoggedIn = 0, TokenExpiration = NULL, Token = '' WHERE TOKEN = '%s'", token),
				)
				return false
			}
		}
	}

	return true
}

func dbInit(db *sql.DB, guestsTable bool) {
	// we probably wouldn't need the LoggedIn value there, but i find it cleaner
	// than checking for the existence of an authentication token.
	statement := `
	CREATE TABLE Accounts (
		Username VARCHAR(100),
		PwHash BINARY(60),
		Token CHAR(24) DEFAULT '',
		TokenExpiration DATETIME,
		LoggedIn INT DEFAULT 0
	);`

	_, err := db.Exec(statement)
	if err != nil {
		slog.Fatal(err)
	}

	if guestsTable {
		_, err = db.Exec(`
		CREATE TABLE Guests (
			Nick VARCHAR(100),
			Token CHAR(24) DEFAULT '',
			TokenExpiration DATETIME
		)
		`)

		if err != nil {
			slog.Fatal(err)
		}
	}

	slog.Printf("Populating database %s\n", dbPath)
}
