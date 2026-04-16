package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestEffectiveWinner_Team1HandicapCausesLoss(t *testing.T) {
	// Score 1-1 but team1 has 0.5 handicap: eff1=0.5, eff2=1.0 → team2 wins
	result := EffectiveWinner(1, 1, 0.5, 0)
	assert.Equal(t, 2, result, "tie score with team1 handicap 0.5 → team2 wins")
}

func TestEffectiveWinner_Team2HandicapCausesLoss(t *testing.T) {
	// Score 2-2 but team2 has 1.0 handicap: eff1=2.0, eff2=1.0 → team1 wins
	result := EffectiveWinner(2, 2, 0, 1.0)
	assert.Equal(t, 1, result, "tie score with team2 handicap 1.0 → team1 wins")
}

func TestEffectiveWinner_HandicapCausesDraw(t *testing.T) {
	// Score 3-2, team1 has 1.0 handicap: eff1=2.0, eff2=2.0 → draw
	result := EffectiveWinner(3, 2, 1.0, 0)
	assert.Equal(t, 0, result, "handicap exactly bridges the gap → draw")
}

func TestEffectiveWinner_BothHandicapTeam1StillWins(t *testing.T) {
	// Score 4-2, both have 0.5: eff1=3.5, eff2=1.5 → team1 wins
	result := EffectiveWinner(4, 2, 0.5, 0.5)
	assert.Equal(t, 1, result, "both teams have handicap, team1 still ahead")
}

func TestEffectiveWinner_ZeroZeroScore(t *testing.T) {
	result := EffectiveWinner(0, 0, 0, 0)
	assert.Equal(t, 0, result, "0-0 with no handicap → draw")
}

func TestEffectiveWinner_HighScores(t *testing.T) {
	result := EffectiveWinner(10, 9, 0, 0)
	assert.Equal(t, 1, result, "10-9 → team1 wins")
}

func TestEffectiveWinner_HandicapExceedsScore(t *testing.T) {
	// Team2 scored 0, team1 scored 1, team1 has handicap 2.0: eff1=-1, eff2=0 → team2 wins
	result := EffectiveWinner(1, 0, 2.0, 0)
	assert.Equal(t, 2, result, "team1 handicap > team1 score advantage → team2 wins")
}
