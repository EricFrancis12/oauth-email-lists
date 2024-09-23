package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	decenc  *OAuthDecEncoder
	server  *Server
	storage *Storage
)

func init() {
	if runningFromServerless() {
		log.Printf("Cold start")
	}

	if err := safeLoadEnvs(); err != nil {
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

	listenAddr := fmtPort(fallbackIfEmpty(os.Getenv(EnvPort), defaultListenAddr))
	server = NewServer(listenAddr)
}

func main() {
	err := server.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func envFilePathsInLoadOrder() []string {
	return []string{
		filePathEnvLocal,
		filePathEnv,
	}
}

func safeLoadEnvs() error {
	validFilenames := []string{}
	for _, filePath := range envFilePathsInLoadOrder() {
		if fileExists(filePath) {
			validFilenames = append(validFilenames, filePath)
		}
	}
	if len(validFilenames) == 0 {
		return nil
	}
	return godotenv.Load(validFilenames...)
}

func fmtPort(port string) string {
	if string(port[0]) == ":" {
		return port
	}
	return ":" + port
}
