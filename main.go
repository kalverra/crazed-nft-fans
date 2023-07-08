package main

import (
	"math/big"

	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"

	"github.com/kalverra/crazed-nft-fans/config"
	"github.com/kalverra/crazed-nft-fans/convert"
	"github.com/kalverra/crazed-nft-fans/president"
)

func init() {
	err := config.ReadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading config")
	}
	if err := config.InitLogging(config.Current.LogLevel); err != nil {
		log.Fatal().Err(err).Msg("Error initializing logging")
	}
}

func main() {
	router := buildRoutes()
	err := president.WatchChain()
	if err != nil {
		log.Fatal().Err(err).Msg("Error watching chain")
	}
	err = president.RecruitFans(100)
	if err != nil {
		log.Fatal().Err(err).Msg("Error recruiting fans")
	}
	err = president.FundFans(convert.EtherToWei(big.NewFloat(100)))
	if err != nil {
		log.Fatal().Err(err).Msg("Error funding fans")
	}

	log.Info().Msg("Starting at http://localhost:3333")

	log.Fatal().Err(router.ListenAndServe()).Msg("Error running router")
}
