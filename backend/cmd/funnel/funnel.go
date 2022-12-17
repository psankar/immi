package main

import (
	"immi/internal/funnel"
	"immi/internal/idb"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func main() {
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()
	db, err := idb.NewPGDB()
	if err != nil {
		log.Fatal().Err(err)
		return
	}

	server, err := funnel.NewServer(funnel.FunnelConfig{
		BatchSize:     1024,
		BatchDuration: time.Second * 5,
		DB:            db,
		Logger:        &log,
	})
	if err != nil {
		log.Fatal().Err(err)
		return
	}

	err = http.ListenAndServe(":8080", server.Handler())
	if err != nil {
		log.Fatal().Err(err)
		return
	}
}
