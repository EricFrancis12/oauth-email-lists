package main

import (
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

type OAuthProvider interface {
	GetResult() (OAuthResult, error)
}

type OAuthResult struct {
	Name      string
	EmailAddr string
}

func (o OAuthResult) ToNewSubscriber(emailListID string) *Subscriber {
	return &Subscriber{
		ID:          NewUUID(),
		EmailListID: emailListID,
		Name:        o.Name,
		EmailAddr:   o.EmailAddr,
	}
}

const (
	HTTPHeaderContentType string = "Content-Type"
)
