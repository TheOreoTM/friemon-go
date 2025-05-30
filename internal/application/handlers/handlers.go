package handlers

import (
	"math/rand"
)

func randomInt(min, max int) int {
	return rand.Intn(max-min) + min
}
