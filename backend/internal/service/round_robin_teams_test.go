package service

import (
	"testing"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeTeam builds a TournamentTeam with a real UUID and two real player IDs.
func makeTeam() model.TournamentTeam {
	return model.TournamentTeam{
		ID:        uuid.New(),
		Player1ID: uuid.New(),
		Player2ID: uuid.New(),
	}
}

func makeTeams(n int) []model.TournamentTeam {
	teams := make([]model.TournamentTeam, n)
	for i := range teams {
		teams[i] = makeTeam()
	}
	return teams
}

// ─── GenerateTeamSchedule ─────────────────────────────────────────────────────

func TestGenerateTeamSchedule_5Teams_10Matches(t *testing.T) {
	teams := makeTeams(5)
	slots := GenerateTeamSchedule(teams)
	assert.Len(t, slots, 10, "5 teams → C(5,2) = 10 matches")
}

func TestGenerateTeamSchedule_4Teams_6Matches(t *testing.T) {
	teams := makeTeams(4)
	slots := GenerateTeamSchedule(teams)
	assert.Len(t, slots, 6, "4 teams → C(4,2) = 6 matches")
}

func TestGenerateTeamSchedule_6Teams_15Matches(t *testing.T) {
	teams := makeTeams(6)
	slots := GenerateTeamSchedule(teams)
	assert.Len(t, slots, 15, "6 teams → C(6,2) = 15 matches")
}

func TestGenerateTeamSchedule_NoPairRepeated(t *testing.T) {
	for n := 4; n <= 8; n++ {
		teams := makeTeams(n)
		slots := GenerateTeamSchedule(teams)

		type teamPair struct{ A, B uuid.UUID }
		seen := make(map[teamPair]bool)
		for _, s := range slots {
			a, b := s.Team1.ID, s.Team2.ID
			if a.String() > b.String() {
				a, b = b, a
			}
			key := teamPair{a, b}
			assert.False(t, seen[key], "n=%d: team pair (%v, %v) appears more than once", n, a, b)
			seen[key] = true
		}
		expected := n * (n - 1) / 2
		assert.Equal(t, expected, len(seen), "n=%d: should have exactly C(n,2) unique pairs", n)
	}
}

func TestGenerateTeamSchedule_EachTeamPlaysNMinus1Times(t *testing.T) {
	for n := 4; n <= 7; n++ {
		teams := makeTeams(n)
		slots := GenerateTeamSchedule(teams)

		appearances := make(map[uuid.UUID]int)
		for _, s := range slots {
			appearances[s.Team1.ID]++
			appearances[s.Team2.ID]++
		}
		for _, team := range teams {
			assert.Equal(t, n-1, appearances[team.ID],
				"n=%d: each team should play exactly %d times", n, n-1)
		}
	}
}

func TestGenerateTeamSchedule_RoundAndOrderSet(t *testing.T) {
	teams := makeTeams(5)
	slots := GenerateTeamSchedule(teams)
	for _, s := range slots {
		assert.Greater(t, s.Round, 0, "round must be ≥ 1")
		assert.Greater(t, s.Order, 0, "order must be ≥ 1")
	}
}

func TestGenerateTeamSchedule_5Teams_5Rounds2MatchesEach(t *testing.T) {
	teams := makeTeams(5)
	slots := GenerateTeamSchedule(teams)

	byRound := make(map[int]int)
	for _, s := range slots {
		byRound[s.Round]++
	}
	require.Len(t, byRound, 5, "5 teams → 5 rounds")
	for r, count := range byRound {
		assert.Equal(t, 2, count, "round %d should have exactly 2 matches", r)
	}
}

func TestGenerateTeamSchedule_NoZeroUUIDTeamIDs(t *testing.T) {
	teams := makeTeams(5)
	slots := GenerateTeamSchedule(teams)
	for _, s := range slots {
		assert.NotEqual(t, uuid.Nil, s.Team1.ID, "Team1.ID must not be zero UUID")
		assert.NotEqual(t, uuid.Nil, s.Team2.ID, "Team2.ID must not be zero UUID")
	}
}

func TestGenerateTeamSchedule_TeamDataPreserved(t *testing.T) {
	teams := makeTeams(4)
	slots := GenerateTeamSchedule(teams)

	// Collect all team IDs seen in output
	seen := make(map[uuid.UUID]bool)
	for _, s := range slots {
		seen[s.Team1.ID] = true
		seen[s.Team2.ID] = true
	}
	for _, team := range teams {
		assert.True(t, seen[team.ID], "team %v must appear in the schedule", team.ID)
	}
}
