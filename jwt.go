package main

import (
	"log"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/getsentry/raven-go"
)

func createToken() string {
	// create a signer for rsa 256
	t := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"router": os.Args[2],
		"exp":    time.Now().Add(time.Minute * 1).Unix(),
		"nbf":    time.Now().Unix(),
	})

	tokenString, err := t.SignedString(signKey)
	if err != nil {
		raven.CaptureError(err, ravenContext)
		log.Println(err)
		return ""
	}

	return tokenString
}
