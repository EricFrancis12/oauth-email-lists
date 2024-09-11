package main

import (
	"log"
)

const (
	listenAddr      string = ":6009"
	postgresConnStr string = "user=postgres dbname=postgres password=CHANGE_ME sslmode=disable"
)

var storage = NewStorage(postgresConnStr)

func main() {
	if err := storage.Init(); err != nil {
		log.Fatal(err)
	}

	server := NewServer(listenAddr)
	log.Fatal(server.Run())
}
