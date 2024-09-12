package main

import (
	"encoding/json"
	"net/http"
	"os"
)

const defaultCatchAllRedirectUrl = "https://bing.com"

func RedirectVisitor(w http.ResponseWriter, r *http.Request, url string) {
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func RedirectToCatchAllUrl(w http.ResponseWriter, r *http.Request) {
	RedirectVisitor(w, r, CatchAllUrl())
}

func CatchAllUrl() string {
	catchAllUrl := os.Getenv(EnvCatchAllRedirectUrl)
	if catchAllUrl == "" {
		return defaultCatchAllRedirectUrl
	}
	return catchAllUrl
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
