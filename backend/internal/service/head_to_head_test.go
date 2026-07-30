package service

import (
	"testing"
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// row is a small helper to build an H2HRow. `date` orders matches; higher = more recent.
func h2hRow(p1Team, winnerTeam, date int) model.H2HRow {
	return model.H2HRow{
		MatchID:    uuid.New(),
		MatchType:  "1v1",
		WinnerTeam: winnerTeam,
		MatchDate:  time.Unix(int64(date), 0),
		P1Team:     p1Team,
	}
}

func TestAggregateHeadToHead_NoHistory(t *testing.T) {
	p1 := &model.User{ID: uuid.New(), Name: "A", IsActive: true}
	p2 := &model.User{ID: uuid.New(), Name: "B", IsActive: true}

	resp := aggregateHeadToHead(p1, p2, nil, nil)

	assert.Equal(t, 0, resp.TotalMatches)
	assert.Equal(t, 0, resp.Player1Wins)
	assert.Equal(t, 0, resp.Player2Wins)
	assert.Equal(t, 0.0, resp.Player1WinRate)
	assert.Equal(t, 0.0, resp.Player2WinRate)
	assert.Empty(t, resp.Form)
	assert.Empty(t, resp.RecentMatches)
	assert.Nil(t, resp.CurrentStreak.PlayerID, "no streak owner when there are no matches")
	assert.Equal(t, 0, resp.CurrentStreak.Count)
}

func TestAggregateHeadToHead_WinsAndComplement(t *testing.T) {
	p1 := &model.User{ID: uuid.New(), Name: "A"}
	p2 := &model.User{ID: uuid.New(), Name: "B"}

	// Most-recent first. P1 on team 1 throughout; winner alternates.
	rows := []model.H2HRow{
		h2hRow(1, 1, 5), // P1 win
		h2hRow(1, 2, 4), // P1 loss
		h2hRow(1, 1, 3), // P1 win
		h2hRow(1, 1, 2), // P1 win
		h2hRow(1, 2, 1), // P1 loss
	}

	resp := aggregateHeadToHead(p1, p2, rows, nil)

	assert.Equal(t, 5, resp.TotalMatches)
	assert.Equal(t, 3, resp.Player1Wins)
	assert.Equal(t, 2, resp.Player2Wins)
	assert.Equal(t, resp.TotalMatches, resp.Player1Wins+resp.Player2Wins, "wins must sum to total (no draws)")
	assert.InDelta(t, 0.6, resp.Player1WinRate, 1e-9)
	assert.InDelta(t, 0.4, resp.Player2WinRate, 1e-9)
	assert.InDelta(t, 1.0, resp.Player1WinRate+resp.Player2WinRate, 1e-9)
	// Form is most-recent first from P1 POV.
	assert.Equal(t, []string{"W", "L", "W", "W", "L"}, resp.Form)
}

func TestAggregateHeadToHead_MixedMatchTypesAggregated(t *testing.T) {
	p1 := &model.User{ID: uuid.New(), Name: "A"}
	p2 := &model.User{ID: uuid.New(), Name: "B"}

	rows := []model.H2HRow{
		{MatchID: uuid.New(), MatchType: "2v2", WinnerTeam: 1, MatchDate: time.Unix(3, 0), P1Team: 1}, // P1 win
		{MatchID: uuid.New(), MatchType: "1v2", WinnerTeam: 1, MatchDate: time.Unix(2, 0), P1Team: 2}, // P1 loss
		{MatchID: uuid.New(), MatchType: "1v1", WinnerTeam: 2, MatchDate: time.Unix(1, 0), P1Team: 2}, // P1 win
	}

	resp := aggregateHeadToHead(p1, p2, rows, nil)

	assert.Equal(t, 3, resp.TotalMatches, "all match types aggregated together")
	assert.Equal(t, 2, resp.Player1Wins)
	assert.Equal(t, 1, resp.Player2Wins)
}

func TestAggregateHeadToHead_OrientationFlips(t *testing.T) {
	a := &model.User{ID: uuid.New(), Name: "A"}
	b := &model.User{ID: uuid.New(), Name: "B"}

	// From A's POV: team 1, wins 2 of 3.
	aRows := []model.H2HRow{
		h2hRow(1, 1, 3), // A win
		h2hRow(1, 1, 2), // A win
		h2hRow(1, 2, 1), // A loss
	}
	// Same matches from B's POV: B is team 2 (opposing), so results invert.
	bRows := []model.H2HRow{
		h2hRow(2, 1, 3), // B loss
		h2hRow(2, 1, 2), // B loss
		h2hRow(2, 2, 1), // B win
	}

	fromA := aggregateHeadToHead(a, b, aRows, nil)
	fromB := aggregateHeadToHead(b, a, bRows, nil)

	assert.Equal(t, fromA.Player1Wins, fromB.Player2Wins, "A's wins == B-view's player2 wins")
	assert.Equal(t, fromA.Player2Wins, fromB.Player1Wins)
	assert.Equal(t, fromA.TotalMatches, fromB.TotalMatches)
	assert.Equal(t, []string{"W", "W", "L"}, fromA.Form)
	assert.Equal(t, []string{"L", "L", "W"}, fromB.Form, "form flips with orientation")
}

func TestAggregateHeadToHead_FormCappedAtTen(t *testing.T) {
	p1 := &model.User{ID: uuid.New()}
	p2 := &model.User{ID: uuid.New()}

	rows := make([]model.H2HRow, 15)
	for i := range rows {
		rows[i] = h2hRow(1, 1, 15-i) // all P1 wins, most-recent first
	}

	resp := aggregateHeadToHead(p1, p2, rows, nil)

	assert.Equal(t, 15, resp.TotalMatches, "totals count the full history")
	assert.Equal(t, 15, resp.Player1Wins)
	assert.Len(t, resp.Form, h2hListLimit, "form is capped at 10")
	assert.Len(t, resp.RecentMatches, h2hListLimit, "recent list is capped at 10")
}

func TestAggregateHeadToHead_StreakOwnerAndCount(t *testing.T) {
	p1 := &model.User{ID: uuid.New(), Name: "A"}
	p2 := &model.User{ID: uuid.New(), Name: "B"}

	// Most-recent first: P1 lost the last two, then won earlier.
	rows := []model.H2HRow{
		h2hRow(1, 2, 5), // P1 loss (most recent)
		h2hRow(1, 2, 4), // P1 loss
		h2hRow(1, 1, 3), // P1 win
	}

	resp := aggregateHeadToHead(p1, p2, rows, nil)

	require.NotNil(t, resp.CurrentStreak.PlayerID)
	assert.Equal(t, p2.ID, *resp.CurrentStreak.PlayerID, "P2 owns the current streak (won the last two)")
	assert.Equal(t, 2, resp.CurrentStreak.Count)
}

func TestAggregateHeadToHead_RecentMatchLineups(t *testing.T) {
	p1 := &model.User{ID: uuid.New(), Name: "A"}
	p2 := &model.User{ID: uuid.New(), Name: "B"}

	row := model.H2HRow{MatchID: uuid.New(), MatchType: "2v2", WinnerTeam: 1, MatchDate: time.Unix(1, 0), P1Team: 1}
	lineups := map[uuid.UUID][]model.H2HParticipant{
		row.MatchID: {
			{UserID: p1.ID, Name: "A", Team: 1},
			{UserID: uuid.New(), Name: "C", Team: 1},
			{UserID: p2.ID, Name: "B", Team: 2},
			{UserID: uuid.New(), Name: "D", Team: 2},
		},
	}

	resp := aggregateHeadToHead(p1, p2, []model.H2HRow{row}, lineups)

	require.Len(t, resp.RecentMatches, 1)
	m := resp.RecentMatches[0]
	assert.True(t, m.Player1Won)
	assert.Len(t, m.Participants, 4, "full lineup for a 2v2 match")
	team1, team2 := 0, 0
	for _, part := range m.Participants {
		switch part.Team {
		case 1:
			team1++
		case 2:
			team2++
		}
	}
	assert.Equal(t, 2, team1)
	assert.Equal(t, 2, team2)
}

func TestComputeH2HStreak_P1LeadingWins(t *testing.T) {
	p1 := uuid.New()
	p2 := uuid.New()
	rows := []model.H2HRow{
		h2hRow(1, 1, 3), // P1 win (most recent)
		h2hRow(1, 1, 2), // P1 win
		h2hRow(1, 2, 1), // P1 loss
	}

	streak := computeH2HStreak(rows, p1, p2)

	require.NotNil(t, streak.PlayerID)
	assert.Equal(t, p1, *streak.PlayerID)
	assert.Equal(t, 2, streak.Count)
}
