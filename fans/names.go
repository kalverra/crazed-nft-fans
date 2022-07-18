package fans

import (
	"fmt"
	"math/rand"
)

// newName generates a random new name
func newName() string {
	firstIndex, lastIndex := rand.Intn(len(firstNames)), rand.Intn(len(lastNames))
	return fmt.Sprintf("%s %s", firstNames[firstIndex], lastNames[lastIndex])
}

var firstNames = []string{
	"Paul",
	"Rick",
	"Morty",
	"Sergey",
	"Steve",
	"Ari",
	"Strider",
	"Aragorn",
	"Legolas",
	"Gimli",
	"Stilgar",
	"Kendrick",
	"Childish",
	"Connor",
	"Adam",
	"Meg",
}

var lastNames = []string{
	"Atreides",
	"Smith",
	"Nazarov",
	"Ellis",
	"Juels",
	"Lamar",
	"Gambino",
	"Hite",
	"Hamrick",
}
