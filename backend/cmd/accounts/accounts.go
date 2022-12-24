package main

import (
	"immi/internal/accounts"
	"immi/internal/idb"
	"net/http"
	"os"

	"github.com/rs/zerolog"
)

func main() {
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()
	db, err := idb.NewPGDB(&log)
	if err != nil {
		log.Fatal().Err(err).Msg("Immi Funnel cannot talk to DB")
		return
	}

	server, err := accounts.NewServer(accounts.AccountsConfig{
		DB:     db,
		Logger: &log,
	})
	if err != nil {
		log.Fatal().Err(err)
		return
	}

	log.Error().Msg("Starting Accounts service")
	err = http.ListenAndServe(":8080", server.Handler())
	if err != nil {
		log.Fatal().Err(err)
		return
	}
}
