package auth

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	Username string
	PwHash   []byte
}

func NewAccount(username, password string) (Account, error) {
	if len(username) < 5 || len(username) > 99 {
		return Account{}, fmt.Errorf("username too long: must be 5-100 characters")
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

	if len(pw) < 6 {
		return fmt.Errorf("password must be longer than 6 characters")
	}

	hasUpper := false
	hasLower := false
	hasNumberOrCntrl := false
	for _, c := range pw {
		if unicode.IsUpper(c) {
			hasUpper = true
		}

		if unicode.IsLower(c) {
			hasLower = true
		}

		if unicode.IsNumber(c) || unicode.IsControl(c) {
			hasNumberOrCntrl = true
		}
	}

	if !(hasUpper && hasLower && hasNumberOrCntrl) {
		return fmt.Errorf("password must contain all of: upper case letter, lower case letter, number or control characters")
	}

	return nil
}

func hashPw(pw string) ([]byte, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return bytes, err
}
