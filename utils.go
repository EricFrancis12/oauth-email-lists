package main

import (
	"net/http"
	"os"

	"github.com/google/uuid"
)

func NewUUID() string {
	return uuid.NewString()
}

func setCookie(w http.ResponseWriter, name CookieName, value string) {
	cookie := &http.Cookie{
		Name:     string(name),
		Value:    value,
		Path:     "/",
		MaxAge:   0, // No max age
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func FallbackIfEmpty(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
