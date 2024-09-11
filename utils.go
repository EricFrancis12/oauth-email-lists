package main

func FallbackIfEmpty(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
