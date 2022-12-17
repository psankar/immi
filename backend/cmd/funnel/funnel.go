package main

import (
	"immi/internal/funnel"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	server, err := funnel.NewServer(funnel.FunnelConfig{
		BatchSize:     1024,
		BatchDuration: time.Second * 5,
		Logger:        &logger,
	})
	if err != nil {
		log.Fatal(err)
	}

	err = http.ListenAndServe(":8080", server.Handler())
	if err != nil {
		log.Fatal(err)
	}
}
