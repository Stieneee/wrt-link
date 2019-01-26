package main

import (
	"os"

	"github.com/getsentry/raven-go"
)

var ravenContext = map[string]string{
	"version": "0.0.0",
	"API":     os.Args[1],
	"router":  os.Args[2],
}

func ravenInit() {
	raven.SetDSN(os.Getenv("SENTRY"))
}
