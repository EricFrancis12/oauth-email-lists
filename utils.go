package main

import (
	"net/http"
	"os"
)

func setCookie(w http.ResponseWriter, name string, value string) {
	cookie := &http.Cookie{
		Name:     string(name),
		Value:    value,
		Path:     "/",
		MaxAge:   0, // Zero means no max age
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
