package main

import (
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const defaultGoogleOAuthStateStr string = "5dYo3mAPGIp8OxZepJgs62YKoz0SDjatFhBVgw5JEg7KvucAjh8qkPovuBteJhPF"

func googleOAuthStateStr() string {
	ss := os.Getenv(EnvGoogleOAuthStateStr)
	if ss == "" {
		return defaultGoogleOAuthStateStr
	}
	return ss
}

func GoogleConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv(EnvGoogleClientID),
		ClientSecret: os.Getenv(EnvGoogleClientSecret),
		RedirectURL:  fmt.Sprintf("%s//%s/callback/google", os.Getenv(EnvProtocol), os.Getenv(EnvHostname)),
		Scopes:       []string{GoogleOauthScopeEmail, GoogleOauthScopeProfile},
		Endpoint:     google.Endpoint,
	}
}
