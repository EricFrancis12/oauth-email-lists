package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootUser(t *testing.T) {
	assert.Nil(t, safeLoadEnvs(filePathEnv))

	var (
		rootUsername = os.Getenv(EnvRootUsername)
		rootPassword = os.Getenv(EnvRootPassword)
	)

	assert.NotEqual(t, "", rootUsername)
	assert.NotEqual(t, "", rootPassword)

	rootUser, err := NewRootUser(rootUsername, rootPassword)
	assert.Nil(t, err)
	assert.Equal(t, rootUsername, rootUser.Name)
	assert.Equal(t, rootPassword, rootUser.HashedPassword)
	assert.Equal(t, rootUserID, rootUser.ID)
	assert.Less(t, len(rootUser.ID), len(NewUUID()))
}
