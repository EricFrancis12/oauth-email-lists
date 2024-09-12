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
	ID     string `json:"id"`
	UserID string `json:"userId"`
	Name   string `json:"name"`
}

func NewEmailList(userID string, name string) *EmailList {
	return &EmailList{
		ID:     NewUUID(),
		UserID: userID,
		Name:   name,
	}
}

type EmailListCreationReq struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
}

type EmailListUpdateReq struct {
	Name string `json:"name"`
}

type Subscriber struct {
	ID          string `json:"id"`
	EmailListID string `json:"emailListId"`
	Name        string `json:"name"`
	EmailAddr   string `json:"emailAddr"`
}

func NewSubscriber(emailListID string, name string, emailAddr string) *Subscriber {
	return &Subscriber{
		ID:          NewUUID(),
		EmailListID: emailListID,
		Name:        name,
		EmailAddr:   emailAddr,
	}
}

type SubscriberCreationReq struct {
	EmailListID string `json:"emailListId"`
	Name        string `json:"name"`
	EmailAddr   string `json:"emailAddr"`
}

type SubscriberUpdateReq struct {
	Name      string `json:"name"`
	EmailAddr string `json:"emailAddr"`
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
	EnvJWTSecret           string = "JWT_SECRET"
)

const (
	HTTPHeaderContentType string = "Content-Type"
	HTTPHeaderJWTToken    string = "X-JWT-Token"
)

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
