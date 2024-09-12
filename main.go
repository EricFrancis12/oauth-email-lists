package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

const (
	listenAddr      string = ":6009"
	postgresConnStr string = "user=postgres dbname=postgres password=CHANGE_ME sslmode=disable"
)

var (
	config  = NewConfig()
	decenc  *OAuthDecEncoder
	storage = NewStorage(postgresConnStr)
)

func main() {
	if err := SafeLoadEnvs(filePathEnv); err != nil {
		log.Fatal(err)
	}

	secret := os.Getenv(EnvCryptoSecret)
	if !validSecret(secret) {
		log.Fatalf("invalid environment variable %s", EnvCryptoSecret)
	}

	decenc = NewOAuthDecEncoder(secret, oauthDecEncDelim)

	if err := storage.Init(); err != nil {
		log.Fatal(err)
	}

	server := NewServer(listenAddr)
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
