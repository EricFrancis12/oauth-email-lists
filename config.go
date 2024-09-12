package main

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Config struct {
}

func NewConfig() *Config {
	return &Config{}
}

func (c Config) Google() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv(EnvGoogleClientID),
		ClientSecret: os.Getenv(EnvGoogleClientSecret),
		RedirectURL:  "http://localhost:6009/callback/google",
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}
}
