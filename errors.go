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

func missingEnv(envVars ...string) error {
	if len(envVars) == 0 {
		return fmt.Errorf("unknown missingEnv error")
	}

	var err = fmt.Errorf("missing required environment variable: %s", envVars[0])
	if len(envVars) > 1 {
		err = fmt.Errorf("missing required environment variables: %s", strings.Join(envVars, ", "))
	}
	return err
}
