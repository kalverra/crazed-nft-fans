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
		"Dr. John",
	}

	lastNames = []string{
		"Hamrick",
		"Atreides",
		"Zoidberg",
		"Leela",
		"Fry",
		"Bending Rodriguez",
	}
)

// generateName gets a random name from our lists
func generateName() string {
	return firstNames[rand.Intn(len(firstNames))] + " " + lastNames[rand.Intn(len(lastNames))]
}
