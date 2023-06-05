package config

import (
	"math/rand"
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
		Name:                    "Indifferent",
		MaxPendingTransactions:  1,
		TransactionBlockTimeout: 15,
		GasPriceMultiplier:      1.1,
	}
	Curious = &CrazedLevel{
		Name:                    "Curious",
		MaxPendingTransactions:  2,
		TransactionBlockTimeout: 13,
		GasPriceMultiplier:      1.25,
	}
	Interested = &CrazedLevel{
		Name:                    "Interested",
		MaxPendingTransactions:  3,
		TransactionBlockTimeout: 10,
		GasPriceMultiplier:      1.5,
	}
	Obsessed = &CrazedLevel{
		Name:                    "Obsessed",
		MaxPendingTransactions:  5,
		TransactionBlockTimeout: 5,
		GasPriceMultiplier:      1.75,
	}
	Manic = &CrazedLevel{
		Name:                    "Manic",
		MaxPendingTransactions:  10,
		TransactionBlockTimeout: 2,
		GasPriceMultiplier:      2.0,
	}
)

type CrazedLevel struct {
	Name                    string
	MaxPendingTransactions  int
	TransactionBlockTimeout int
	GasPriceMultiplier      float64
}

// GetCrazedLevel retrieves the current crazed level, processing Mixed if necessary
func (c *Config) GetCrazedLevel() *CrazedLevel {
	crazedLevel := c.CrazedLevel
	if crazedLevel == "Mixed" {
		crazedLevel = CrazedLevelsMapping[rand.Intn(len(CrazedLevels))]
	}
	return CrazedLevels[crazedLevel]
}
