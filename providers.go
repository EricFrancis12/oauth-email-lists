package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
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
	var (
		clientID    = os.Getenv(EnvDiscordClientID)
		protocol    = os.Getenv(EnvProtocol)
		hostname    = os.Getenv(EnvHostname)
		redirectUri = url.QueryEscape(fmt.Sprintf("%s//%s/callback/discord", protocol, hostname))
	)

	url := fmt.Sprintf(
		"https://discord.com/oauth2/authorize?client_id=%s&response_type=code&redirect_uri=%s&scope=email+identify",
		clientID,
		redirectUri,
	)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (dpr DiscordProviderResp) ToSubscriber(emailListID string) Subscriber {
	return Subscriber{
		ID:          NewUUID(),
		EmailListID: emailListID,
		Name:        dpr.Username,
		EmailAddr:   dpr.Email,
	}
}

func (dpr DiscordProviderResp) Result() ProviderResult {
	return ProviderResult{
		Name:      dpr.Username,
		EmailAddr: dpr.Email,
	}
}

type GoogleProvider struct{}

func (gp GoogleProvider) Name() ProviderName {
	return ProviderNameGoogle
}

func (gp GoogleProvider) Redirect(w http.ResponseWriter, r *http.Request) {
	url := GoogleConfig().AuthCodeURL(googleOAuthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (gpr GoogleProviderResp) ToSubscriber(emailListID string) Subscriber {
	return Subscriber{
		ID:          NewUUID(),
		EmailListID: emailListID,
		Name:        gpr.Name,
		EmailAddr:   gpr.Email,
	}
}

func (gpr GoogleProviderResp) Result() ProviderResult {
	return ProviderResult{
		Name:      gpr.Name,
		EmailAddr: gpr.Email,
	}
}
