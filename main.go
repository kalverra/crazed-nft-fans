package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/kalverra/crazed-nft-fans/client"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	err := client.NewTransactionTracker()
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to initialize transaction tracker")
	}

}
