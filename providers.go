package main

import (
	"net/http"
	"time"
)

const (
	timestampFormat   string = time.UnixDate
	timestampLayout   string = "Mon Jan 2 15:04:05 CDT 2006"
	outputCookieDelim string = "---"
)

type OAuthProvider interface {
	Name() ProviderName
	Redirect(w http.ResponseWriter, r *http.Request)
}

func NewOAuthProvider(providerName ProviderName) OAuthProvider {
	switch providerName {
	case ProviderNameDiscord:
		return DiscordProvider{}
	case ProviderNameGoogle:
		return GoogleProvider{}
	}
	return nil
}

type DiscordProvider struct{}

func (dp DiscordProvider) Name() ProviderName {
	return ProviderNameDiscord
}

func (dp DiscordProvider) Redirect(w http.ResponseWriter, r *http.Request) {
	// TODO: ...
}

type GoogleProvider struct{}

func (gp GoogleProvider) Name() ProviderName {
	return ProviderNameGoogle
}

func (gp GoogleProvider) Redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, config.Google().AuthCodeURL(googleOAuthStateString), http.StatusTemporaryRedirect)
}

type GoogleProviderResult struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

func (gpr GoogleProviderResult) ToSubscriber(emailListID string) Subscriber {
	return Subscriber{
		ID:          NewUUID(),
		EmailListID: emailListID,
		Name:        gpr.Name,
		EmailAddr:   gpr.Email,
	}
}
