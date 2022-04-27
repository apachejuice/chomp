package auth

import (
	"encoding/base64"
	"log"

	uuid "github.com/satori/go.uuid"
)

func GetToken() string {
	id, err := uuid.NewV4()
	if err != nil {
		log.Fatal(err)
	}

	return base64.StdEncoding.EncodeToString(id.Bytes())
}
