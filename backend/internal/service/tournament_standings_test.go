package service

import (
	"testing"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── Fixtures ────────────────────────────────────────────────────────────────

func makeTournamentTeam(p1Name, p2Name string) model.TournamentTeam {
	return model.TournamentTeam{
		ID:        uuid.New(),
		Player1ID: uuid.New(),
		Player2ID: uuid.New(),
		Player1:   &model.User{ID: uuid.New(), Name: p1Name},
		Player2:   &model.User{ID: uuid.New(), Name: p2Name},
	}
}

// completedGroupMatch builds a completed group match with a given score.
// effectiveWinner: 0=draw, 1=team1, 2=team2.
func completedGroupMatch(team1, team2 model.TournamentTeam, score1, score2, winner int) model.TournamentMatch {
	t1id := team1.ID
	t2id := team2.ID
	return model.TournamentMatch{
		Stage:           "group",
		Status:          "completed",
		Team1TeamID:     &t1id,
		Team2TeamID:     &t2id,
		ActualScore1:    &score1,
		ActualScore2:    &score2,
		EffectiveWinner: winner,
	}
}

func pendingGroupMatch(team1, team2 model.TournamentTeam) model.TournamentMatch {
	t1id := team1.ID
	t2id := team2.ID
	return model.TournamentMatch{
		Stage:       "group",
		Status:      "pending",
		Team1TeamID: &t1id,
		Team2TeamID: &t2id,
	}
}

// ─── ComputeStandings ────────────────────────────────────────────────────────

func TestComputeStandings_EmptyTeamsAndMatches(t *testing.T) {
	result := ComputeStandings(nil, nil, 4)
	assert.Empty(t, result)
}

func TestComputeStandings_OneTeam_NoMatches(t *testing.T) {
	team := makeTournamentTeam("A1", "A2")
	result := ComputeStandings([]model.TournamentTeam{team}, nil, 4)
	require.Len(t, result, 1)
	assert.Equal(t, 0, result[0].Points)
	assert.Equal(t, 0, result[0].Played)
}

func TestComputeStandings_Team1WinsAllMatches_12Points(t *testing.T) {
	teams := []model.TournamentTeam{
		makeTournamentTeam("A1", "A2"),
		makeTournamentTeam("B1", "B2"),
		makeTournamentTeam("C1", "C2"),
		makeTournamentTeam("D1", "D2"),
		makeTournamentTeam("E1", "E2"),
	}
	// Team[0] beats everyone
	var matches []model.TournamentMatch
	for i := 1; i < len(teams); i++ {
		matches = append(matches, completedGroupMatch(teams[0], teams[i], 2, 0, 1))
	}
	result := ComputeStandings(teams, matches, 4)
	// Winner is first
	assert.Equal(t, teams[0].ID, result[0].TeamID)
	assert.Equal(t, 12, result[0].Points, "4 wins × 3 pts = 12")
	assert.Equal(t, 4, result[0].Won)
	assert.Equal(t, 4, result[0].Played)
	assert.Equal(t, 1, result[0].Seed, "top team gets seed 1")
}

func TestComputeStandings_DrawGivesOnePointEach(t *testing.T) {
	a := makeTournamentTeam("A1", "A2")
	b := makeTournamentTeam("B1", "B2")
	match := completedGroupMatch(a, b, 1, 1, 0) // draw

	result := ComputeStandings([]model.TournamentTeam{a, b}, []model.TournamentMatch{match}, 4)
	require.Len(t, result, 2)
	for _, s := range result {
		assert.Equal(t, 1, s.Points, "draw gives 1 pt each")
		assert.Equal(t, 1, s.Drawn)
		assert.Equal(t, 0, s.Won)
		assert.Equal(t, 0, s.Lost)
	}
}

func TestComputeStandings_GoalDifference(t *testing.T) {
	a := makeTournamentTeam("A1", "A2")
	b := makeTournamentTeam("B1", "B2")
	match := completedGroupMatch(a, b, 3, 1, 1) // a wins 3-1

	result := ComputeStandings([]model.TournamentTeam{a, b}, []model.TournamentMatch{match}, 4)
	require.Len(t, result, 2)

	byID := map[uuid.UUID]TeamStanding{}
	for _, s := range result {
		byID[s.TeamID] = s
	}
	assert.Equal(t, 2, byID[a.ID].GD, "winner GD = +2")
	assert.Equal(t, -2, byID[b.ID].GD, "loser GD = -2")
	assert.Equal(t, 3, byID[a.ID].GF)
	assert.Equal(t, 1, byID[a.ID].GA)
}

func TestComputeStandings_SortByPointsPrimary(t *testing.T) {
	a := makeTournamentTeam("A1", "A2")
	b := makeTournamentTeam("B1", "B2")
	c := makeTournamentTeam("C1", "C2")

	matches := []model.TournamentMatch{
		completedGroupMatch(a, b, 2, 0, 1), // a wins
		completedGroupMatch(a, c, 2, 0, 1), // a wins
		completedGroupMatch(b, c, 1, 1, 0), // draw
	}
	result := ComputeStandings([]model.TournamentTeam{a, b, c}, matches, 4)
	require.Len(t, result, 3)
	assert.Equal(t, a.ID, result[0].TeamID, "a has most points (6), should be first")
}

func TestComputeStandings_SortByGDSecondary(t *testing.T) {
	a := makeTournamentTeam("A1", "A2")
	b := makeTournamentTeam("B1", "B2")
	c := makeTournamentTeam("C1", "C2")

	// a beats b 3-0 (3pts, GD+3); c beats b 1-0 (3pts, GD+1); a vs c not played
	matches := []model.TournamentMatch{
		completedGroupMatch(a, b, 3, 0, 1),
		completedGroupMatch(c, b, 1, 0, 1),
	}
	result := ComputeStandings([]model.TournamentTeam{a, b, c}, matches, 4)
	require.Len(t, result, 3)
	// a and c both have 3pts; a has GD+3, c has GD+1 → a first
	assert.Equal(t, a.ID, result[0].TeamID, "higher GD should rank first when points tied")
}

func TestComputeStandings_SortByGFTertiary(t *testing.T) {
	a := makeTournamentTeam("A1", "A2")
	b := makeTournamentTeam("B1", "B2")
	c := makeTournamentTeam("C1", "C2")

	// Both a and c win by same GD (+1) but a scored more
	matches := []model.TournamentMatch{
		completedGroupMatch(a, b, 2, 1, 1), // a wins, GF=2
		completedGroupMatch(c, b, 1, 0, 1), // c wins, GF=1
	}
	result := ComputeStandings([]model.TournamentTeam{a, b, c}, matches, 4)
	require.Len(t, result, 3)
	// Points tied (3), GD tied (+1), GF: a=2 > c=1 → a first
	assert.Equal(t, a.ID, result[0].TeamID, "higher GF should rank first when points and GD tied")
}

// ─── Seed assignment ──────────────────────────────────────────────────────────

func TestComputeStandings_Top4Seeds(t *testing.T) {
	teams := make([]model.TournamentTeam, 5)
	for i := range teams {
		teams[i] = makeTournamentTeam("P1", "P2")
	}
	// No matches — all tied at 0; seeds still assigned to first 4 in result
	result := ComputeStandings(teams, nil, 4)
	require.Len(t, result, 5)
	seeded := 0
	for _, s := range result {
		if s.Seed > 0 {
			seeded++
		}
	}
	assert.Equal(t, 4, seeded, "knockoutSize=4 → exactly 4 teams get seeds")
	assert.Equal(t, 0, result[4].Seed, "5th team should have seed=0")
}

func TestComputeStandings_Top2Seeds_KnockoutSize2(t *testing.T) {
	teams := make([]model.TournamentTeam, 4)
	for i := range teams {
		teams[i] = makeTournamentTeam("P1", "P2")
	}
	result := ComputeStandings(teams, nil, 2)
	require.Len(t, result, 4)
	assert.Equal(t, 1, result[0].Seed, "1st team gets seed 1")
	assert.Equal(t, 2, result[1].Seed, "2nd team gets seed 2")
	assert.Equal(t, 0, result[2].Seed, "3rd team gets seed 0 with knockoutSize=2")
	assert.Equal(t, 0, result[3].Seed, "4th team gets seed 0 with knockoutSize=2")
}

func TestComputeStandings_SeedOrderMatchesSortOrder(t *testing.T) {
	teams := []model.TournamentTeam{
		makeTournamentTeam("A1", "A2"),
		makeTournamentTeam("B1", "B2"),
		makeTournamentTeam("C1", "C2"),
		makeTournamentTeam("D1", "D2"),
		makeTournamentTeam("E1", "E2"),
	}
	// Engineer distinct point totals: A=12, B=9, C=6, D=3, E=0
	matches := []model.TournamentMatch{
		completedGroupMatch(teams[0], teams[1], 2, 0, 1),
		completedGroupMatch(teams[0], teams[2], 2, 0, 1),
		completedGroupMatch(teams[0], teams[3], 2, 0, 1),
		completedGroupMatch(teams[0], teams[4], 2, 0, 1),
		completedGroupMatch(teams[1], teams[2], 2, 0, 1),
		completedGroupMatch(teams[1], teams[3], 2, 0, 1),
		completedGroupMatch(teams[1], teams[4], 2, 0, 1),
		completedGroupMatch(teams[2], teams[3], 2, 0, 1),
		completedGroupMatch(teams[2], teams[4], 2, 0, 1),
		completedGroupMatch(teams[3], teams[4], 2, 0, 1),
	}
	result := ComputeStandings(teams, matches, 4)
	require.Len(t, result, 5)
	for i, s := range result {
		if i < 4 {
			assert.Equal(t, i+1, s.Seed, "position %d should have seed %d", i+1, i+1)
		} else {
			assert.Equal(t, 0, s.Seed, "position %d should have seed 0", i+1)
		}
	}
}

// ─── Match filtering ──────────────────────────────────────────────────────────

func TestComputeStandings_PendingMatchExcluded(t *testing.T) {
	a := makeTournamentTeam("A1", "A2")
	b := makeTournamentTeam("B1", "B2")
	pending := pendingGroupMatch(a, b)

	result := ComputeStandings([]model.TournamentTeam{a, b}, []model.TournamentMatch{pending}, 4)
	for _, s := range result {
		assert.Equal(t, 0, s.Played, "pending match must not be counted")
		assert.Equal(t, 0, s.Points)
	}
}

func TestComputeStandings_NonGroupStageMatchExcluded(t *testing.T) {
	a := makeTournamentTeam("A1", "A2")
	b := makeTournamentTeam("B1", "B2")

	t1id := a.ID
	t2id := b.ID
	s1, s2, w := 2, 0, 1
	semiMatch := model.TournamentMatch{
		Stage:           "semi",
		Status:          "completed",
		Team1TeamID:     &t1id,
		Team2TeamID:     &t2id,
		ActualScore1:    &s1,
		ActualScore2:    &s2,
		EffectiveWinner: w,
	}
	result := ComputeStandings([]model.TournamentTeam{a, b}, []model.TournamentMatch{semiMatch}, 4)
	for _, s := range result {
		assert.Equal(t, 0, s.Played, "semi match must not affect group standings")
	}
}

func TestComputeStandings_ZeroUUIDTeamIDSkipped(t *testing.T) {
	a := makeTournamentTeam("A1", "A2")
	b := makeTournamentTeam("B1", "B2")

	// Simulate the old bug: team IDs were zero UUID
	nilID := uuid.Nil
	s1, s2, w := 2, 0, 1
	badMatch := model.TournamentMatch{
		Stage:           "group",
		Status:          "completed",
		Team1TeamID:     &nilID, // zero UUID → no team in map
		Team2TeamID:     &nilID,
		ActualScore1:    &s1,
		ActualScore2:    &s2,
		EffectiveWinner: w,
	}
	// Should not panic, and should not affect any team
	result := ComputeStandings([]model.TournamentTeam{a, b}, []model.TournamentMatch{badMatch}, 4)
	for _, s := range result {
		assert.Equal(t, 0, s.Played, "zero-UUID team match must be skipped")
	}
}

func TestComputeStandings_NilTeamIDSkipped(t *testing.T) {
	a := makeTournamentTeam("A1", "A2")
	b := makeTournamentTeam("B1", "B2")

	s1, s2 := 2, 0
	badMatch := model.TournamentMatch{
		Stage:        "group",
		Status:       "completed",
		Team1TeamID:  nil, // nil pointer
		Team2TeamID:  nil,
		ActualScore1: &s1,
		ActualScore2: &s2,
	}
	result := ComputeStandings([]model.TournamentTeam{a, b}, []model.TournamentMatch{badMatch}, 4)
	for _, s := range result {
		assert.Equal(t, 0, s.Played, "nil team ID match must be skipped")
	}
}

// ─── Played/Won/Drawn/Lost accumulation ──────────────────────────────────────

func TestComputeStandings_PlayedCountAccumulates(t *testing.T) {
	a := makeTournamentTeam("A1", "A2")
	b := makeTournamentTeam("B1", "B2")
	c := makeTournamentTeam("C1", "C2")

	matches := []model.TournamentMatch{
		completedGroupMatch(a, b, 2, 1, 1),
		completedGroupMatch(a, c, 1, 1, 0),
	}
	result := ComputeStandings([]model.TournamentTeam{a, b, c}, matches, 4)
	byID := map[uuid.UUID]TeamStanding{}
	for _, s := range result {
		byID[s.TeamID] = s
	}
	assert.Equal(t, 2, byID[a.ID].Played, "a played 2 matches")
	assert.Equal(t, 1, byID[b.ID].Played, "b played 1 match")
	assert.Equal(t, 1, byID[c.ID].Played, "c played 1 match")

	assert.Equal(t, 1, byID[a.ID].Won)
	assert.Equal(t, 1, byID[a.ID].Drawn)
	assert.Equal(t, 0, byID[a.ID].Lost)
	assert.Equal(t, 4, byID[a.ID].Points, "3 for win + 1 for draw")
}
