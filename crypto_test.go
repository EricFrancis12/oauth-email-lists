package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDelim(t *testing.T) {
	assert.True(t, validDelim(oauthDecEncDelim))
}

func TestCrypto(t *testing.T) {
	var (
		message = "Hello World"
		secret  = "123456789_123456789_123456789_12"
		delim   = "%&%&%&"
	)

	assert.True(t, validSecret(secret))
	assert.True(t, validDelim(delim))

	t.Run("Correct usage", func(t *testing.T) {
		encrypted, err := Encrypt(secret, message)
		assert.Nil(t, err)
		assert.NotEqual(t, message, encrypted)

		decrypted, err := Decrypt(secret, encrypted)
		assert.Nil(t, err)
		assert.Equal(t, decrypted, message)
	})

	t.Run("Decrypt with wrong key", func(t *testing.T) {
		encrypted, err := Encrypt(secret, message)
		assert.Nil(t, err)
		assert.NotEqual(t, message, encrypted)

		decrypted, err := Decrypt("wrong-secret-key", encrypted)
		assert.NotNil(t, err)
		assert.NotEqual(t, decrypted, message)
	})

	t.Run("Test OAuthDecEncoder", func(t *testing.T) {
		var (
			emailListID  string       = "abcdefgh"
			providerName ProviderName = ProviderNameGoogle
			redirectUrls []string     = testingUrls()
			outputIDs    []string     = []string{
				"1234",
				"5678",
			}
		)

		de := NewOAuthDecEncoder(secret, delim)

		for _, redirectUrl := range redirectUrls {
			encrypted, err := de.Encode(emailListID, providerName, outputIDs, redirectUrl)
			assert.Nil(t, err)
			assert.NotEqual(t, emailListID, encrypted)
			assert.NotEqual(t, providerName, encrypted)
			assert.NotEqual(t, string(providerName), encrypted)
			for _, outputID := range outputIDs {
				assert.NotEqual(t, outputID, encrypted)
			}
			assert.NotEqual(t, redirectUrl, encrypted)
			assert.NotEqual(t, emailListID+de.delim+string(providerName), encrypted)

			decEmailListID, decProvider, decOutputIDs, decRedirectUrl, err := de.Decode(encrypted)
			assert.Nil(t, err)
			assert.Equal(t, emailListID, decEmailListID)
			assert.Equal(t, providerName, decProvider.Name())
			assert.Equal(t, outputIDs, decOutputIDs)
			assert.Equal(t, redirectUrl, decRedirectUrl)
		}
	})
}

func TestUrls(t *testing.T) {
	var (
		secret = "123456789_123456789_123456789_12"
		urls   = testingUrls()
	)

	assert.True(t, validSecret(secret))

	t.Run("Test encodePart() and decodePart()", func(t *testing.T) {
		for _, url := range urls {
			encodedUrl := encodePart(url)
			assert.NotEqual(t, url, encodedUrl)

			decodedUrl, err := decodePart(encodedUrl)
			assert.Nil(t, err)
			assert.Equal(t, url, decodedUrl)
		}
	})

	t.Run("Test Encrpyt() and Decrpyt()", func(t *testing.T) {
		for _, url := range urls {
			encryptedUrl, err := Encrypt(secret, url)
			assert.Nil(t, err)
			assert.NotEqual(t, secret, encryptedUrl)
			assert.NotEqual(t, url, encryptedUrl)

			decryptedUrl, err := Decrypt(secret, encryptedUrl)
			assert.Nil(t, err)
			assert.NotEqual(t, secret, decryptedUrl)
			assert.Equal(t, url, decryptedUrl)
		}
	})
}

func TestUUID(t *testing.T) {
	uuid := NewUUID()
	assert.NotContains(t, uuid, oauthDecEncDelim)
	assert.NotContains(t, uuid, outputCookieDelim)
}

func testingUrls() []string {
	return []string{
		"http://bing.com",
		"http://bing.com/1/2/3",
		"https://bing.com",
		"https://bing.com/1/2/3",
		"http://subdoamin.bing.com",
		"http://subdoamin.bing.com/1/2/3",
		"https://subdoamin.bing.com",
		"https://subdoamin.bing.com/1/2/3",
		"http://bing.com?one=1&hello=true",
		"http://bing.com/1/2/3?one=1&hello=true",
		"https://bing.com?one=1&hello=true",
		"https://bing.com/1/2/3?one=1&hello=true",
		"http://subdoamin.bing.com?one=1&hello=true",
		"http://subdoamin.bing.com/1/2/3?one=1&hello=true",
		"https://subdoamin.bing.com?one=1&hello=true",
		"https://subdoamin.bing.com/1/2/3?one=1&hello=true",
	}
}
