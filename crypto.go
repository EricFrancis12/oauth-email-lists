package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
)

const oauthDecEncDelim string = "%"

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

func (o OAuthDecEncoder) Encode(emailListID string, providerName ProviderName) (oauthID string, err error) {
	return Encrypt(
		o.secret,
		encodePart(emailListID)+o.delim+encodePart(string(providerName)),
	)

}

func (o OAuthDecEncoder) Decode(oauthID string) (emailListID string, provider OAuthProvider, err error) {
	str, err := Decrypt(o.secret, oauthID)
	if err != nil {
		return "", nil, err
	}

	parts := strings.Split(str, o.delim)
	if len(parts) != 2 {
		return "", nil, invalidOauthID()
	}

	if emailListID, err = decodePart(parts[0]); err != nil {
		return "", nil, invalidOauthID()
	}

	if s, err := decodePart(parts[1]); err != nil {
		return "", nil, invalidOauthID()
	} else {
		pn, err := ToProviderName(s)
		if err != nil {
			return "", nil, invalidOauthID()
		}
		provider = NewOAuthProvider(pn)
	}

	return emailListID, provider, nil
}

func Encrypt(secret, value string) (string, error) {
	block, err := aes.NewCipher([]byte(secret))
	if err != nil {
		return "", err
	}

	plainText := []byte(value)

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plainText))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plainText)

	return base64.RawStdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(secret, value string) (string, error) {
	ciphertext, err := base64.RawStdEncoding.DecodeString(value)
	if err != nil {
		return "", fmt.Errorf("decoding base64: %w", err)
	}

	block, err := aes.NewCipher([]byte(secret))
	if err != nil {
		return "", err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
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
	for _, char := range delim {
		if encodePart(string(char)) == string(char) {
			return false
		}
	}
	return true
}

func invalidOauthID() error {
	return fmt.Errorf("invalid oauthID")
}
