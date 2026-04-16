package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// handicap_rate semantics: negative = penalty, positive = bonus
// eff_score = actual_score + handicap_rate
// Pro players have negative handicap_rate (e.g. -0.5): a draw means they lose effectively

func TestEffectiveWinner_Team1WinsByScore(t *testing.T) {
	result := EffectiveWinner(3, 1, 0, 0)
	assert.Equal(t, 1, result, "team1 scores more → team1 wins")
}

func TestEffectiveWinner_Team2WinsByScore(t *testing.T) {
	result := EffectiveWinner(1, 4, 0, 0)
	assert.Equal(t, 2, result, "team2 scores more → team2 wins")
}

func TestEffectiveWinner_DrawNoHandicap(t *testing.T) {
	result := EffectiveWinner(2, 2, 0, 0)
	assert.Equal(t, 0, result, "equal scores, no handicap → draw")
}

func TestEffectiveWinner_ProPlayerDrawsLoses(t *testing.T) {
	// Cuban (Pro, handicap=-0.5) draws 2-2: eff_Cuban=1.5, eff_opp=2.0 → opponent wins
	result := EffectiveWinner(2, 2, -0.5, 0)
	assert.Equal(t, 2, result, "Pro with -0.5 handicap draws → opponent wins")
}

func TestEffectiveWinner_ProPlayerWinsByEnough(t *testing.T) {
	// Pro (handicap=-0.5) wins 3-2: eff1=2.5, eff2=2.0 → Pro team wins
	result := EffectiveWinner(3, 2, -0.5, 0)
	assert.Equal(t, 1, result, "Pro wins by enough to overcome handicap → team1 wins")
}

func TestEffectiveWinner_HandicapCausesDraw(t *testing.T) {
	// Score 3-4, team2 has -1.0 handicap: eff1=3.0, eff2=3.0 → draw
	result := EffectiveWinner(3, 4, 0, -1.0)
	assert.Equal(t, 0, result, "handicap exactly bridges the gap → draw")
}

func TestEffectiveWinner_BothHandicapTeam1StillWins(t *testing.T) {
	// Score 4-2, both have -0.5: eff1=3.5, eff2=1.5 → team1 wins
	result := EffectiveWinner(4, 2, -0.5, -0.5)
	assert.Equal(t, 1, result, "both teams have same handicap, higher score wins")
}

func TestEffectiveWinner_ZeroZeroScore(t *testing.T) {
	result := EffectiveWinner(0, 0, 0, 0)
	assert.Equal(t, 0, result, "0-0 with no handicap → draw")
}

func TestEffectiveWinner_HighScores(t *testing.T) {
	result := EffectiveWinner(10, 9, 0, 0)
	assert.Equal(t, 1, result, "10-9 → team1 wins")
}

func TestEffectiveWinner_LargeNegativeHandicapCausesLoss(t *testing.T) {
	// Team1 won 3-1 but has -3.0 handicap: eff1=0.0, eff2=1.0 → team2 wins
	result := EffectiveWinner(3, 1, -3.0, 0)
	assert.Equal(t, 2, result, "large negative handicap overrides score advantage → team2 wins")
}
