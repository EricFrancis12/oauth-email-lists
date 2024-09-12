package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

func useProtectedRoute(w http.ResponseWriter, r *http.Request) (*User, error) {
	// - check for jwt
	// - if valid jwt, then fetch user
	// 		- if user found, return user, nil
	// 		- else return nil, err
	// - else, then check if request has basic auth
	// 		- if has basic auth, fetch user
	//			- if user found, return user, nil
	// 			- else return nil, err
	// 		- else return nil, err

	tokenStr := r.Header.Get(HTTPHeaderJWTToken)
	token, err := validateJWT(tokenStr)
	if err != nil || !token.Valid {
		username, password, ok := r.BasicAuth()
		if !ok || username == "" || password == "" {
			return nil, unauthorizedError()
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

func newJWTStr(user *User) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt": 15000,
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
