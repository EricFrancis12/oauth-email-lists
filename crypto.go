package main

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

const oauthDecEncDelim string = "%&%&%&"

type OAuthDecEncoder struct {
	secret string
	delim  string
}

func NewOAuthDecEncoder(secret string, delim string) *OAuthDecEncoder {
	return &OAuthDecEncoder{
		secret: secret,
		delim:  delim,
	}
}

func (o OAuthDecEncoder) Encode(
	emailListID string,
	providerName ProviderName,
	outputIDs []string,
	redirectUrl string,
) (oauthID string, err error) {
	var parts = []string{
		encodePart(emailListID),
		encodePart(string(providerName)),
		encodePart(strings.Join(outputIDs, outputCookieDelim)),
		encodePart(redirectUrl),
	}
	return Encrypt(
		o.secret,
		strings.Join(parts, o.delim),
	)
}

func (o OAuthDecEncoder) Decode(oauthID string) (
	emailListID string,
	provider OAuthProvider,
	outputIDs []string,
	redirectUrl string,
	err error,
) {
	str, err := Decrypt(o.secret, oauthID)
	if err != nil {
		return "", nil, []string{}, "", err
	}

	parts := strings.Split(str, o.delim)
	if len(parts) < 4 {
		return "", nil, []string{}, "", invalidOauthID()
	}

	if emailListID, err = decodePart(parts[0]); err != nil {
		return "", nil, []string{}, "", invalidOauthID()
	}

	s, err := decodePart(parts[1])
	if err != nil {
		return "", nil, []string{}, "", invalidOauthID()
	}
	pn, err := ToProviderName(s)
	if err != nil {
		return "", nil, []string{}, "", invalidOauthID()
	}
	provider = NewOAuthProvider(pn)

	outputIDStr, err := decodePart(parts[2])
	if err != nil {
		return "", nil, []string{}, "", invalidOauthID()
	}
	outputIDs = strings.Split(outputIDStr, outputCookieDelim)

	if redirectUrl, err = decodePart(parts[3]); err != nil {
		return "", nil, []string{}, "", invalidOauthID()
	}

	return emailListID, provider, outputIDs, redirectUrl, nil
}

func Encrypt(secret, value string) (string, error) {
	claims := jwt.MapClaims{
		"data": value,
		"exp":  time.Now().Add(time.Hour * 1).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return signedToken, nil
}

func Decrypt(secret, tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return "", fmt.Errorf("parsing token: %w", err)
	}

	// Extract claims from the token
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if value, ok := claims["data"].(string); ok {
			return value, nil
		}
		return "", fmt.Errorf("value not found in claims")
	}

	return "", fmt.Errorf("invalid token")
}

func encodePart(part string) string {
	return url.QueryEscape(part)
}

func decodePart(part string) (string, error) {
	return url.QueryUnescape(part)
}

func validSecret(secret string) bool {
	return len(secret) == 32
}

func validDelim(delim string) bool {
	if len(delim) < MinDelimLength {
		return false
	}

	for _, char := range delim {
		equalBeforeAndAfterEncode := encodePart(string(char)) == string(char)
		if equalBeforeAndAfterEncode {
			return false
		}
	}

	return true
}

func NewUUID() string {
	return uuid.NewString()
}

func invalidOauthID() error {
	return fmt.Errorf("invalid oauthID")
}
