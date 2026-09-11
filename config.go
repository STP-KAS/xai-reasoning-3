package main

import (
	"os"
	"strings"
)

type Config struct {
	Bind       string
	Name       string
	IBAN       string
	BIC        string
	Kaspa      string
	PrepaidKey string
	DataDir    string
}

func loadConfig() Config {
	c := Config{
		Bind:       env("DESK_BIND", "127.0.0.1:8091"),
		Name:       env("DESK_NAME", "Desk"),
		IBAN:       compactIBAN(os.Getenv("DESK_IBAN")),
		BIC:        strings.ToUpper(strings.TrimSpace(os.Getenv("DESK_BIC"))),
		Kaspa:      strings.TrimSpace(os.Getenv("DESK_KASPA")),
		PrepaidKey: strings.TrimSpace(os.Getenv("DESK_PREPAID_KEY")),
		DataDir:    env("DESK_DATA", "data"),
	}
	c.Kaspa = strings.TrimPrefix(c.Kaspa, "kaspa:")
	c.Kaspa = strings.TrimPrefix(c.Kaspa, "kaspatest:")
	return c
}

func env(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}
