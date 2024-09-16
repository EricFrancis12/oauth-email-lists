package main

import (
	"fmt"
	"strings"
)

func unauthorized() error {
	return fmt.Errorf("unauthorized")
}

func userIDNotProvided() error {
	return fmt.Errorf("user ID not provided")
}

func outputIDNotProvided() error {
	return fmt.Errorf("output ID not provided")
}

func invalidOauthID() error {
	return fmt.Errorf("invalid oauthID")
}

func missingEnv(envVars ...string) error {
	if len(envVars) == 0 {
		return fmt.Errorf("unknown missingEnv error")
	}
	if len(envVars) == 1 {
		return fmt.Errorf("missing required environment variable: %s", envVars[0])
	}
	return fmt.Errorf("missing required environment variables: %s", strings.Join(envVars, ", "))
}
