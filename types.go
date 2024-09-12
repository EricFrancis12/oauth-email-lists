package main

import (
	"fmt"
	"time"
)

type User struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	HashedPassword string    `json:"hashedPassword"`
	CreatedAt      time.Time `json:"createdAt"`
}

func NewUser(name string, password string) (*User, error) {
	hpw, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:             NewUUID(),
		Name:           name,
		HashedPassword: string(hpw),
		CreatedAt:      time.Now(),
	}, nil
}

type UserCreationReq struct {
	Name     string
	Password string
}

type UserUpdateReq struct {
	Name     string
	Password string
}

type EmailList struct {
	ID   string
	Name string
}

type Subscriber struct {
	ID          string
	EmailListID string
	Name        string
	EmailAddr   string
}

type CookieName string

const (
	CookieNameEmailListID  CookieName = "emailListId"
	CookieNameProviderName CookieName = "providerName"
	CookieNameCreatedAt    CookieName = "createdAt"
)

const (
	EnvCatchAllRedirectUrl string = "CATCH_ALL_REDIRECT_URL"
	EnvCryptoSecret        string = "CRYPTO_SECRET"
	EnvGoogleClientID      string = "GOOGLE_CLIENT_ID"
	EnvGoogleClientSecret  string = "GOOGLE_CLIENT_Secret"
)

const HTTPHeaderContentType string = "Content-Type"

type ProviderName string

const (
	ProviderNameDiscord ProviderName = "Discord"
	ProviderNameGoogle  ProviderName = "Google"
)

var providerNames = []ProviderName{
	ProviderNameDiscord,
	ProviderNameGoogle,
}

func ToProviderName(str string) (ProviderName, error) {
	for _, pn := range providerNames {
		if string(pn) == str {
			return pn, nil
		}
	}
	return "", fmt.Errorf("invalid ProviderName %s", str)
}
