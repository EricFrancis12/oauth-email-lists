package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	config  = NewConfig()
	decenc  *OAuthDecEncoder
	storage *Storage
)

func main() {
	if err := SafeLoadEnvs(filePathEnv); err != nil {
		log.Fatal(err)
	}

	secret := os.Getenv(EnvCryptoSecret)
	if !validSecret(secret) {
		log.Fatal(missingEnv(EnvCryptoSecret))
	}

	decenc = NewOAuthDecEncoder(secret, oauthDecEncDelim)

	postgresConnStr := os.Getenv(EnvPostgresConnStr)
	if postgresConnStr == "" {
		log.Fatal(missingEnv(EnvPostgresConnStr))
	}

	storage = NewStorage(postgresConnStr)
	if err := storage.Init(); err != nil {
		log.Fatal(err)
	}

	server := NewServer(mustFmtPort(os.Getenv(EnvPort)))
	log.Fatal(server.Run())
}

func SafeLoadEnvs(filenames ...string) error {
	validFilenames := []string{}
	for _, fn := range filenames {
		if fileExists(fn) {
			validFilenames = append(validFilenames, fn)
		}
	}
	if len(validFilenames) == 0 {
		return nil
	}
	return godotenv.Load(validFilenames...)
}

func mustFmtPort(port string) string {
	if port == "" {
		panic("argument port is missing")
	}
	if string(port[0]) == ":" {
		return port
	}
	return ":" + port
}
