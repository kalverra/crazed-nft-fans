package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/kalverra/crazed-nft-fans/convert"
	"github.com/kalverra/crazed-nft-fans/president"
)

func buildRoutes() *http.Server {
	r := chi.NewRouter()
	// r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "dash.html")
	})
	r.Get("/blockData", blockData)
	r.Put("/increaseIntensity", increaseIntensity)
	r.Put("/decreaseIntensity", decreaseIntensity)
	r.Put("/spike", spike)

	return &http.Server{
		Addr:         ":3333",
		Handler:      r,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}
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

// TODO: could clean this up
func increaseIntensity(w http.ResponseWriter, r *http.Request) {
	newTarget := president.IncreaseGasTarget()
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte(convert.WeiToGwei(newTarget).String()))
	if err != nil {
		log.Error().Err(err).Msg("Error writing response")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func decreaseIntensity(w http.ResponseWriter, r *http.Request) {
	newTarget := president.DecreaseGasTarget()
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte(convert.WeiToGwei(newTarget).String()))
	if err != nil {
		log.Error().Err(err).Msg("Error writing response")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func spike(w http.ResponseWriter, r *http.Request) {
	newTarget := president.Spike()
	_, err := w.Write([]byte(convert.WeiToGwei(newTarget).String()))
	if err != nil {
		log.Error().Err(err).Msg("Error writing response")
		w.WriteHeader(http.StatusInternalServerError)
	}
}
