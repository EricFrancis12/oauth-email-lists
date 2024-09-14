package main

import (
	"fmt"
	"net/url"
	"os"
	"time"
)

type Campaign struct {
	EmailListID  string       `json:"emailListId"`
	ProviderName ProviderName `json:"providerName"`
	OutputIDs    []string     `json:"outputIds"`
	RedirectUrl  string       `json:"redirectUrl"`
}

func (c Campaign) Link() (string, error) {
	var (
		protocol = os.Getenv(EnvProtocol)
		hostname = os.Getenv(EnvHostname)
	)

	if protocol == "" || hostname == "" {
		return "", missingEnv(EnvProtocol, EnvHostname)
	}

	oauthID, err := decenc.Encode(c.EmailListID, c.ProviderName, c.OutputIDs, c.RedirectUrl)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s//%s/c?c=%s", protocol, hostname, url.QueryEscape(oauthID)), nil
}

type EmailList struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewEmailList(userID string, name string) *EmailList {
	now := time.Now()
	return &EmailList{
		ID:        NewUUID(),
		UserID:    userID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type EmailListCreationReq struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
}

type EmailListUpdateReq struct {
	Name string `json:"name"`
}

type Output interface {
	OutputName() OutputName
	GetUserID() string
	Handle(emailAddr string, name string) error
}

type OutputsData map[OutputName][]Output

type OutputCreationReq struct {
	UserID     string     `json:"userId"`
	OutputName OutputName `json:"outputName"`
	ListID     string     `json:"listId"`
}

type OutputUpdateReq struct {
	OutputName OutputName `json:"outputName"`
	ListID     string     `json:"listId"`
}

type AWeberOutput struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	ListID    string    `json:"listId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ResendOutput struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId"`
	AudienceID string    `json:"audienceId"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type Subscriber struct {
	ID                 string       `json:"id"`
	EmailListID        string       `json:"emailListId"`
	UserID             string       `json:"userId"`
	SourceProviderName ProviderName `json:"sourceProviderName"`
	Name               string       `json:"name"`
	EmailAddr          string       `json:"emailAddr"`
	CreatedAt          time.Time    `json:"createdAt"`
	UpdatedAt          time.Time    `json:"updatedAt"`
}

func NewSubscriber(
	emailListID string,
	userID string,
	sourceProviderName ProviderName,
	name string,
	emailAddr string,
) *Subscriber {
	now := time.Now()
	return &Subscriber{
		ID:                 NewUUID(),
		EmailListID:        emailListID,
		UserID:             userID,
		SourceProviderName: sourceProviderName,
		Name:               name,
		EmailAddr:          emailAddr,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

type SubscriberCreationReq struct {
	EmailListID        string       `json:"emailListId"`
	UserID             string       `json:"userId"`
	SourceProviderName ProviderName `json:"sourceProviderName"`
	Name               string       `json:"name"`
	EmailAddr          string       `json:"emailAddr"`
}

type SubscriberUpdateReq struct {
	Name      string `json:"name"`
	EmailAddr string `json:"emailAddr"`
}

type User struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	HashedPassword string    `json:"hashedPassword"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func NewUser(name string, password string) (*User, error) {
	hpw, err := hashPassword(password)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &User{
		ID:             NewUUID(),
		Name:           name,
		HashedPassword: string(hpw),
		CreatedAt:      now,
		UpdatedAt:      now,
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

type CookieName string

const (
	CookieNameCreatedAt    CookieName = "createdAt"
	CookieNameEmailListID  CookieName = "emailListId"
	CookieNameOutputIDs    CookieName = "outputIds"
	CookieNameProviderName CookieName = "providerName"
	CookieNameRedirectURL  CookieName = "redirectUrl"
)

const (
	EnvCatchAllRedirectUrl string = "CATCH_ALL_REDIRECT_URL"
	EnvCookieSecret        string = "COOKIE_SECRET"
	EnvCryptoSecret        string = "CRYPTO_SECRET"
	EnvGoogleClientID      string = "GOOGLE_CLIENT_ID"
	EnvGoogleClientSecret  string = "GOOGLE_CLIENT_Secret"
	EnvHostname            string = "HOST_NAME"
	EnvJWTSecret           string = "JWT_SECRET"
	EnvPort                string = "PORT"
	EnvProtocol            string = "PROTOCOL"
	EnvPostgresConnStr     string = "POSTGRES_CONN_STR"
	EnvResendApiKey        string = "RESEND_API_KEY"
	EnvRootPassword        string = "ROOT_PASSWORD"
	EnvRootUsername        string = "ROOT_USERNAME"
)

const (
	HTTPHeaderContentType string = "Content-Type"
	HTTPHeaderJWTToken    string = "X-JWT-Token"
)

type OutputName string

const (
	OutputNameAWeber OutputName = "aweber"
	OutputNameResend OutputName = "resend"
)

var outputNames = []OutputName{}

func ToOutputName(str string) (OutputName, error) {
	for _, on := range outputNames {
		if string(on) == str {
			return on, nil
		}
	}
	return "", fmt.Errorf("invalid OutputName %s", str)
}

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

const QueryParamC string = "c"
