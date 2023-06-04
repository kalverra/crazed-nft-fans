package main

import (
	"log"

	"github.com/kalverra/crazed-nft-fans/config"
)

func main() {
	err := config.ReadConfig()
	if err != nil {
		log.Fatal(err)
	}

}
