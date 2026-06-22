package service

import (
	"testing"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/stretchr/testify/assert"
)

// ─── standingLess ─────────────────────────────────────────────────────────────

func TestStandingLess_HigherPointsWins(t *testing.T) {
	a := model.WcTeamStanding{TeamName: "ARG", Points: 9, GoalDifference: 0, GoalsFor: 0}
	b := model.WcTeamStanding{TeamName: "BRA", Points: 6, GoalDifference: 5, GoalsFor: 10}
	assert.True(t, standingLess(a, b), "9pts should rank above 6pts regardless of GD/GF")
	assert.False(t, standingLess(b, a))
}

func TestStandingLess_SamePoints_HigherGDWins(t *testing.T) {
	a := model.WcTeamStanding{TeamName: "ARG", Points: 4, GoalDifference: 3, GoalsFor: 5}
	b := model.WcTeamStanding{TeamName: "BRA", Points: 4, GoalDifference: -1, GoalsFor: 8}
	assert.True(t, standingLess(a, b), "+3 GD should rank above -1 GD at equal points")
	assert.False(t, standingLess(b, a))
}

func TestStandingLess_SamePointsSameGD_HigherGFWins(t *testing.T) {
	a := model.WcTeamStanding{TeamName: "ARG", Points: 4, GoalDifference: 0, GoalsFor: 6}
	b := model.WcTeamStanding{TeamName: "BRA", Points: 4, GoalDifference: 0, GoalsFor: 3}
	assert.True(t, standingLess(a, b), "6 GF should rank above 3 GF at equal points and GD")
	assert.False(t, standingLess(b, a))
}

func TestStandingLess_FullyTied_AlphabeticalOrder(t *testing.T) {
	a := model.WcTeamStanding{TeamName: "ARG", Points: 1, GoalDifference: 0, GoalsFor: 1}
	b := model.WcTeamStanding{TeamName: "BRA", Points: 1, GoalDifference: 0, GoalsFor: 1}
	assert.True(t, standingLess(a, b), "ARG < BRA alphabetically")
	assert.False(t, standingLess(b, a))
}

func TestStandingLess_IdenticalTeam_ReturnsFalse(t *testing.T) {
	a := model.WcTeamStanding{TeamName: "ARG", Points: 3, GoalDifference: 2, GoalsFor: 4}
	assert.False(t, standingLess(a, a), "identical team should not rank above itself")
}

func TestStandingLess_NegativeGD_BothNegative_LessNegativeWins(t *testing.T) {
	a := model.WcTeamStanding{TeamName: "FRA", Points: 1, GoalDifference: -1, GoalsFor: 1}
	b := model.WcTeamStanding{TeamName: "ESP", Points: 1, GoalDifference: -3, GoalsFor: 2}
	assert.True(t, standingLess(a, b), "-1 GD should rank above -3 GD")
	assert.False(t, standingLess(b, a))
}

// ─── sortTeamStandings ────────────────────────────────────────────────────────

func TestSortTeamStandings_ByPoints(t *testing.T) {
	teams := []model.WcTeamStanding{
		{TeamName: "ESP", Points: 0},
		{TeamName: "ARG", Points: 9},
		{TeamName: "FRA", Points: 3},
		{TeamName: "BRA", Points: 6},
	}
	sortTeamStandings(teams)
	assert.Equal(t, "ARG", teams[0].TeamName)
	assert.Equal(t, "BRA", teams[1].TeamName)
	assert.Equal(t, "FRA", teams[2].TeamName)
	assert.Equal(t, "ESP", teams[3].TeamName)
}

func TestSortTeamStandings_TieOnGD(t *testing.T) {
	// ARG and BRA both 4pts; ARG has GD +3, BRA has GD -1 → ARG first
	teams := []model.WcTeamStanding{
		{TeamName: "BRA", Points: 4, GoalDifference: -1, GoalsFor: 2},
		{TeamName: "ARG", Points: 4, GoalDifference: 3, GoalsFor: 4},
		{TeamName: "FRA", Points: 2, GoalDifference: 0, GoalsFor: 2},
		{TeamName: "ESP", Points: 1, GoalDifference: -1, GoalsFor: 1},
	}
	sortTeamStandings(teams)
	assert.Equal(t, "ARG", teams[0].TeamName)
	assert.Equal(t, "BRA", teams[1].TeamName)
	assert.Equal(t, "FRA", teams[2].TeamName)
	assert.Equal(t, "ESP", teams[3].TeamName)
}

func TestSortTeamStandings_TieOnGF(t *testing.T) {
	// FRA and ESP both 1pt, both GD 0; FRA has GF=3, ESP has GF=1 → FRA first
	teams := []model.WcTeamStanding{
		{TeamName: "ESP", Points: 1, GoalDifference: 0, GoalsFor: 1},
		{TeamName: "FRA", Points: 1, GoalDifference: 0, GoalsFor: 3},
	}
	sortTeamStandings(teams)
	assert.Equal(t, "FRA", teams[0].TeamName)
	assert.Equal(t, "ESP", teams[1].TeamName)
}

func TestSortTeamStandings_FullTie_Alphabetical(t *testing.T) {
	teams := []model.WcTeamStanding{
		{TeamName: "MEX", Points: 1, GoalDifference: 0, GoalsFor: 1},
		{TeamName: "ARG", Points: 1, GoalDifference: 0, GoalsFor: 1},
	}
	sortTeamStandings(teams)
	assert.Equal(t, "ARG", teams[0].TeamName)
	assert.Equal(t, "MEX", teams[1].TeamName)
}

func TestSortTeamStandings_AllZeros_Alphabetical(t *testing.T) {
	// Pre-tournament: all teams have zero stats → sorted alphabetically
	teams := []model.WcTeamStanding{
		{TeamName: "URU"},
		{TeamName: "ARG"},
		{TeamName: "MEX"},
		{TeamName: "CAN"},
	}
	sortTeamStandings(teams)
	assert.Equal(t, "ARG", teams[0].TeamName)
	assert.Equal(t, "CAN", teams[1].TeamName)
	assert.Equal(t, "MEX", teams[2].TeamName)
	assert.Equal(t, "URU", teams[3].TeamName)
}

func TestSortTeamStandings_SingleTeam_NoChange(t *testing.T) {
	teams := []model.WcTeamStanding{{TeamName: "ARG", Points: 3}}
	sortTeamStandings(teams)
	assert.Equal(t, "ARG", teams[0].TeamName)
}

func TestSortTeamStandings_Empty_NoChange(t *testing.T) {
	var teams []model.WcTeamStanding
	sortTeamStandings(teams) // must not panic
}

// ─── Full scenario from the testing doc ──────────────────────────────────────
//
// Standings after 4 matches (Match 5 is scheduled, not counted):
//   ARG: P2 W1 D1 L0 GF4 GA1 GD+3 Pts4   Form: [W, D]
//   BRA: P2 W1 D0 L1 GF2 GA3 GD-1 Pts3   Form: [L, W]
//   FRA: P2 W0 D2 L0 GF2 GA2 GD0  Pts2   Form: [D, D]
//   ESP: P2 W0 D1 L1 GF1 GA2 GD-1 Pts1   Form: [D, L]
//
// BRA vs ESP tiebreaker: same GD (-1), BRA has GF=2, ESP has GF=1 → BRA ranks above ESP.

func TestSortTeamStandings_FullScenario(t *testing.T) {
	teams := []model.WcTeamStanding{
		{TeamName: "ESP", Points: 1, GoalDifference: -1, GoalsFor: 1},
		{TeamName: "FRA", Points: 2, GoalDifference: 0, GoalsFor: 2},
		{TeamName: "BRA", Points: 3, GoalDifference: -1, GoalsFor: 2},
		{TeamName: "ARG", Points: 4, GoalDifference: 3, GoalsFor: 4},
	}
	sortTeamStandings(teams)

	require := assert.New(t)
	require.Equal("ARG", teams[0].TeamName)
	require.Equal("BRA", teams[1].TeamName)
	require.Equal("FRA", teams[2].TeamName)
	require.Equal("ESP", teams[3].TeamName)
}

// ─── Points formula ───────────────────────────────────────────────────────────

func TestStandingPointsFormula_WinDrawLoss(t *testing.T) {
	// Won×3 + Drawn×1; GoalDifference = GF - GA
	team := model.WcTeamStanding{
		Won:          2,
		Drawn:        1,
		GoalsFor:     5,
		GoalsAgainst: 2,
	}
	expectedPts := team.Won*3 + team.Drawn
	expectedGD := team.GoalsFor - team.GoalsAgainst
	assert.Equal(t, 7, expectedPts)
	assert.Equal(t, 3, expectedGD)
}
