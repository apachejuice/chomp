package auth

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	Username string
	PwHash   []byte
}

func NewAccount(username, password string) (Account, error) {
	if len(username) < 5 || len(username) > 99 {
		return Account{}, fmt.Errorf("username must be 5-100 characters")
	}

	err := pwCheckRequirements(password)
	if err != nil {
		return Account{}, err
	}

	hash, err := hashPw(password)
	if err != nil {
		return Account{}, err
	}

	return Account{Username: username, PwHash: hash}, nil
}

func pwCheckRequirements(pw string) error {
	if len(pw) == 0 || len(strings.ReplaceAll(pw, " ", "")) == 0 {
		return fmt.Errorf("password cannot be empty or consist entirely of whitespace")
	}

	if len(pw) < 8 {
		return fmt.Errorf("password must be longer than 8 characters")
	}

	return nil
}

func hashPw(pw string) ([]byte, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return bytes, err
}
