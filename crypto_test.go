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
		delim   = "%"
	)

	t.Run("Valid test vars", func(t *testing.T) {
		assert.True(t, validSecret(secret))
		assert.True(t, validDelim(delim))
	})

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
		assert.Nil(t, err)
		assert.NotEqual(t, decrypted, message)
	})

	t.Run("Test OAuthDecEncoder", func(t *testing.T) {
		var (
			emailListID  string       = "abcdefgh"
			providerName ProviderName = ProviderNameGoogle
			outputIDs                 = []string{
				"1234",
				"5678",
			}
		)

		de := NewOAuthDecEncoder(secret, delim)

		encrypted, err := de.Encode(emailListID, providerName, outputIDs)
		assert.Nil(t, err)
		assert.NotEqual(t, emailListID, encrypted)
		assert.NotEqual(t, providerName, encrypted)
		assert.NotEqual(t, string(providerName), encrypted)
		assert.NotEqual(t, emailListID+de.delim+string(providerName), encrypted)

		decEmailListID, decProvider, decOutputIDs, err := de.Decode(encrypted)
		assert.Nil(t, err)
		assert.Equal(t, emailListID, decEmailListID)
		assert.Equal(t, providerName, decProvider.Name())
		assert.Equal(t, outputIDs, decOutputIDs)
	})
}

func TestUUID(t *testing.T) {
	uuid := NewUUID()
	assert.NotContains(t, uuid, oauthDecEncDelim)
	assert.NotContains(t, uuid, outputCookieDelim)
}
