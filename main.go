package main

import (
	"time"

	"github.com/kalverra/crazed-nft-fans/config"
	"github.com/kalverra/crazed-nft-fans/fans"
	"github.com/rs/zerolog/log"
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
	president, err := fans.NewPresident()
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating president")
	}
	if err = president.RecruitFans(5); err != nil {
		log.Fatal().Err(err).Msg("Error recruiting fans")
	}
	president.ActivateFans()
	time.Sleep(time.Second * 5)
}
