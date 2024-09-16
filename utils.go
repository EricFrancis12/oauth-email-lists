package main

import (
	"os"
)

func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func fallbackIfEmpty(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
