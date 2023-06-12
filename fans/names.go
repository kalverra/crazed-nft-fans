package fans

import "math/rand"

var (
	firstNames = []string{
		"Adam",
		"Meg",
		"Pere",
		"Paul",
		"Turanga",
		"Bender",
		"Philip",
		"Professor Hubert",
		"Satoshi",
		"Vitalik",
	}

	lastNames = []string{
		"Hamrick",
		"Atreides",
		"Zoidberg",
		"Leela",
		"Fry",
		"Bending Rodriguez",
		"Farnsworth",
		"Nakamoto",
		"Buterin",
	}
)

// generateName gets a random name from our lists
func generateName() string {
	return firstNames[rand.Intn(len(firstNames))] + " " + lastNames[rand.Intn(len(lastNames))]
}
