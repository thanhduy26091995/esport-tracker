package service

// EffectiveWinner determines the match winner after applying handicap.
// handicap_rate is stored as a positive number (e.g. 0.5) and represents
// a PENALTY subtracted from the team's score.
// Pro players with higher handicap_rate must win by more than the handicap to count.
// Example: Pro with handicap=0.5, draws 2-2 → eff = 2 - 0.5 = 1.5 < 2.0 → opponent wins.
// Returns: 1 = team1 wins, 2 = team2 wins, 0 = draw
func EffectiveWinner(score1, score2 int, handicap1, handicap2 float64) int {
	eff1 := float64(score1) - handicap1
	eff2 := float64(score2) - handicap2
	if eff1 > eff2 {
		return 1
	}
	if eff2 > eff1 {
		return 2
	}
	return 0
}
