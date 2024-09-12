package main

import (
	"net/http"
	"strings"
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

type ProviderCookie struct {
	EmailListID  string
	ProviderName ProviderName
	OutputIDs    []string
	CreatedAt    time.Time
}

func NewProviderCookie(emailListID string, providerName ProviderName, outputIDs []string) *ProviderCookie {
	return &ProviderCookie{
		EmailListID:  emailListID,
		ProviderName: providerName,
		OutputIDs:    outputIDs,
		CreatedAt:    time.Now(),
	}
}

func ProviderCookieFrom(r *http.Request) (*ProviderCookie, error) {
	elidc, err := r.Cookie(string(CookieNameEmailListID))
	if err != nil {
		return nil, err
	}

	pnc, err := r.Cookie(string(CookieNameProviderName))
	if err != nil {
		return nil, err
	}
	providerName, err := ToProviderName(pnc.Value)
	if err != nil {
		return nil, err
	}

	oidc, err := r.Cookie(string(CookieNameEmailListID))
	if err != nil {
		return nil, err
	}
	outputIDs := strings.Split(oidc.Value, outputCookieDelim)

	cac, err := r.Cookie(string(CookieNameCreatedAt))
	if err != nil {
		return nil, err
	}
	createdAt, err := time.Parse(timestampLayout, cac.Value)
	if err != nil {
		return nil, err
	}

	return &ProviderCookie{
		EmailListID:  elidc.Value,
		ProviderName: providerName,
		OutputIDs:    outputIDs,
		CreatedAt:    createdAt,
	}, nil
}

func (pc ProviderCookie) Set(w http.ResponseWriter) {
	setCookie(w, CookieNameEmailListID, pc.EmailListID)
	setCookie(w, CookieNameProviderName, string(pc.ProviderName))
	setCookie(w, CookieNameOutputIDs, strings.Join(pc.OutputIDs, outputCookieDelim))
	setCookie(w, CookieNameCreatedAt, pc.CreatedAt.Format(timestampFormat))
}
