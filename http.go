package main

import (
	"encoding/json"
	"fmt"
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
	w.Header().Set(HTTPHeaderContentType, ContentTypeApplicationJson)
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func WriteUnauthorized(w http.ResponseWriter) error {
	return WriteJSON(w, http.StatusUnauthorized, NewJsonResponse(false, nil, unauthorized()))
}

func TelegramAPIMessageUrl(botID string) string {
	return fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botID)
}

func BearerHeader(value string) string {
	return "Bearer " + value
}
