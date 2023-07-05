package main

import (
	"encoding/json"
	"math/big"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
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
	r := chi.NewRouter()
	// r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	err := president.WatchChain()
	if err != nil {
		log.Fatal().Err(err).Msg("Error watching chain")
	}
	err = president.RecruitFans(10)
	if err != nil {
		log.Fatal().Err(err).Msg("Error recruiting fans")
	}
	err = president.FundFans(convert.EtherToWei(big.NewFloat(100)))
	if err != nil {
		log.Fatal().Err(err).Msg("Error funding fans")
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "dash.html")
	})
	r.Get("/blockData", blockData)

	log.Info().Msg("Starting at http://localhost:3000")
	log.Fatal().Err(http.ListenAndServe(":3000", r)).Msg("Error running router")
}

func blockData(w http.ResponseWriter, r *http.Request) {
	var blocks []*president.TrackedBlock

	blockNumber := r.URL.Query().Get("blockNumber")
	if blockNumber != "" {
		blockNum, err := strconv.ParseUint(blockNumber, 10, 64)
		if err != nil {
			log.Error().Err(err).Msg("Error parsing block number")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		blocks = president.BlocksSinceNumber(blockNum)
	} else {
		blocks = president.AllBlocks()
	}

	ret, err := json.Marshal(blocks)
	if err != nil {
		log.Error().Err(err).Msg("Error marshaling blocks")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(ret)
	if err != nil {
		log.Error().Err(err).Msg("Error writing response")
		w.WriteHeader(http.StatusInternalServerError)
	}
}
