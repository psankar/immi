package main

import (
	"immi/internal/funnel"
	"log"
	"net/http"
	"time"
)

func main() {
	server, err := funnel.NewServer(funnel.FunnelConfig{
		BatchSize:     1024,
		BatchDuration: time.Second * 5,
	})
	if err != nil {
		log.Fatal(err)
	}

	err = http.ListenAndServe(":8080", server.Handler())
	if err != nil {
		log.Fatal(err)
	}
}
