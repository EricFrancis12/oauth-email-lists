package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

func IsRootUser(user *User) bool {
	if user == nil {
		return false
	}
	return user.Name == os.Getenv(EnvRootUsername) && user.HashedPassword == os.Getenv(EnvRootPassword)
}

func NewRootUser(username string, password string) (*User, error) {
	user, err := NewUser(username, password)
	if err != nil {
		return nil, err
	}
	user.HashedPassword = password
	if !IsRootUser(user) {
		return nil, fmt.Errorf("invalid root user credentials")
	}
	user.ID = rootUserID
	return user, nil
}

func useProtectedRoute(w http.ResponseWriter, r *http.Request) (*User, error) {
	tokenStr := r.Header.Get(HTTPHeaderJWTToken)
	token, err := validateJWT(tokenStr)
	if err != nil || !token.Valid {
		username, password, ok := r.BasicAuth()
		if !ok || username == "" || password == "" {
			return nil, unauthorizedError()
		}

		// Check if credentials match root user credentials
		rootUser, err := NewRootUser(username, password)
		if err == nil {
			return rootUser, nil
		}

		user, err := storage.GetUserByUsernameAndPassword(username, password)
		if err != nil {
			return nil, err
		}

		if jwtStr, err := newJWTStr(user); err == nil {
			w.Header().Add(HTTPHeaderJWTToken, jwtStr)
		}

		return user, nil
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := claims["userID"].(string)

	user, err := storage.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func Auth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := useProtectedRoute(w, r)
		if err != nil && user != nil {
			WriteUnauthorized(w)
			return
		}

		h(w, r)
	}
}

func RootAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := useProtectedRoute(w, r)
		if err != nil || !IsRootUser(user) {
			WriteUnauthorized(w)
			return
		}

		h(w, r)
	}
}

func newJWTStr(user *User) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt": jwtExpiry,
		"userID":    user.ID,
	}

	secret := os.Getenv(EnvJWTSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func validateJWT(tokenStr string) (*jwt.Token, error) {
	jwtSecret := os.Getenv(EnvJWTSecret)
	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(jwtSecret), nil
	})
}

// TODO: add more constraints to password (length, etc...)
func hashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func unauthorizedError() error {
	return fmt.Errorf("unauthorized")
}

func WriteUnauthorized(w http.ResponseWriter) error {
	return WriteJSON(w, http.StatusUnauthorized, unauthorizedError())
}
