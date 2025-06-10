package game

import "math"

// CalculateELO calculates the new ELO ratings for two players after a match.
// player1Rating: ELO of player 1
// player2Rating: ELO of player 2
// kFactor: The K-factor, determines how much the ELO changes. Common values are 16, 32.
// result: 1.0 for player 1 win, 0.5 for draw, 0.0 for player 2 win.
func CalculateELO(player1Rating, player2Rating, kFactor int, result float64) (newPlayer1Rating, newPlayer2Rating int) {
	// Expected score for player 1
	expected1 := 1.0 / (1.0 + math.Pow(10, float64(player2Rating-player1Rating)/400.0))

	// Expected score for player 2
	expected2 := 1.0 / (1.0 + math.Pow(10, float64(player1Rating-player2Rating)/400.0))

	// New rating for player 1
	newPlayer1Rating = player1Rating + int(float64(kFactor)*(result-expected1))

	// New rating for player 2
	newPlayer2Rating = player2Rating + int(float64(kFactor)*((1.0-result)-expected2))

	return
}
