package main

import (
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func GoogleConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv(EnvGoogleClientID),
		ClientSecret: os.Getenv(EnvGoogleClientSecret),
		RedirectURL:  fmt.Sprintf("%s//%s/callback/google", os.Getenv(EnvProtocol), os.Getenv(EnvHostname)),
		Scopes:       []string{GoogleOauthScopeEmail, GoogleOauthScopeProfile},
		Endpoint:     google.Endpoint,
	}
}
