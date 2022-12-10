package main

import (
	"immi/internal/funnel"
	"log"
	"net/http"
)

func main() {
	server, err := funnel.NewServer()
	if err != nil {
		log.Fatal(err)
	}

	err = http.ListenAndServe(":8080", server.Handler())
	if err != nil {
		log.Fatal(err)
	}
}
