package config

import (
	"math/rand"
	"time"
)

var (
	// CrazedLevelsMapping maps numbers to the adjectives for easier random selection
	CrazedLevelsMapping = []string{"Indifferent", "Curious", "Interested", "Obsessed", "Manic"}
	// CrazedLevels holds the crazed level configs
	CrazedLevels = map[string]*CrazedLevel{
		"Indifferent": Indifferent,
		"Curious":     Curious,
		"Interested":  Interested,
		"Obsessed":    Obsessed,
		"Manic":       Manic,
	}

	Indifferent = &CrazedLevel{
		Name:                   "Indifferent",
		MaxPendingTransactions: 1,
		TransactionTimeout:     time.Minute * 5,
		GasPriceMultiplier:     1.1,
	}
	Curious = &CrazedLevel{
		Name:                   "Curious",
		MaxPendingTransactions: 2,
		TransactionTimeout:     time.Minute * 2,
		GasPriceMultiplier:     1.25,
	}
	Interested = &CrazedLevel{
		Name:                   "Interested",
		MaxPendingTransactions: 3,
		TransactionTimeout:     time.Minute,
		GasPriceMultiplier:     1.5,
	}
	Obsessed = &CrazedLevel{
		Name:                   "Obsessed",
		MaxPendingTransactions: 5,
		TransactionTimeout:     time.Second * 30,
		GasPriceMultiplier:     1.75,
	}
	Manic = &CrazedLevel{
		Name:                   "Manic",
		MaxPendingTransactions: 10,
		TransactionTimeout:     time.Second * 10,
		GasPriceMultiplier:     2.0,
	}
	President = &CrazedLevel{
		Name:                   "President",
		MaxPendingTransactions: 1000,
		TransactionTimeout:     time.Second * 10,
		GasPriceMultiplier:     1.1,
	}
)

type CrazedLevel struct {
	Name string
	// MaxPendingTransactions is the maximum number of pending transactions a fan will have at any given time
	MaxPendingTransactions int
	// TransactionBlockTimeout is the number of blocks a fan will wait for a transaction to be mined before resending
	TransactionTimeout time.Duration
	// GasPriceMultiplier is the multiplier applied when increasing the gas price for a transaction
	GasPriceMultiplier float64
}

// GetCrazedLevel retrieves the current crazed level, processing Mixed if necessary
func (c *Config) GetCrazedLevel() *CrazedLevel {
	crazedLevel := c.CrazedLevel
	if crazedLevel == "Mixed" {
		crazedLevel = CrazedLevelsMapping[rand.Intn(len(CrazedLevels))]
	}
	return CrazedLevels[crazedLevel]
}
