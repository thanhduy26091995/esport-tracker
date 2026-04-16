package service

import (
	"testing"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── CreateTournamentRequest validation ────────────────────────────────────────

func TestCreateTournamentRequest_AffectsScoreDefaultsTrue(t *testing.T) {
	req := &CreateTournamentRequest{}
	assert.True(t, *req.resolvedAffectsScore(), "nil AffectsScore should default to true")
}

func TestCreateTournamentRequest_AffectsScoreExplicitFalse(t *testing.T) {
	f := false
	req := &CreateTournamentRequest{AffectsScore: &f}
	assert.False(t, *req.resolvedAffectsScore(), "explicit false should be false")
}

func TestCreateTournamentRequest_AffectsScoreExplicitTrue(t *testing.T) {
	tr := true
	req := &CreateTournamentRequest{AffectsScore: &tr}
	assert.True(t, *req.resolvedAffectsScore(), "explicit true should be true")
}

// ─── RecordMatchResultRequest — teamHandicap helper ───────────────────────────

func TestTeamHandicap_MaxOfTeam(t *testing.T) {
	// Pro with 0.5 handicap paired with Normal(0): team handicap = max = 0.5
	u1 := makeUser("pro", 0.5)
	u2 := makeUser("normal", 0)
	userMap := map[uuid.UUID]*model.User{u1.ID: u1, u2.ID: u2}

	result := teamHandicap([]uuid.UUID{u1.ID, u2.ID}, userMap)
	assert.Equal(t, 0.5, result, "team handicap = max (most penalizing) of members")
}

func TestTeamHandicap_SinglePlayer(t *testing.T) {
	u := makeUser("pro", 0.5)
	userMap := map[uuid.UUID]*model.User{u.ID: u}

	result := teamHandicap([]uuid.UUID{u.ID}, userMap)
	assert.Equal(t, 0.5, result)
}

func TestTeamHandicap_ZeroHandicaps(t *testing.T) {
	u1 := makeUser("normal", 0)
	u2 := makeUser("normal", 0)
	userMap := map[uuid.UUID]*model.User{u1.ID: u1, u2.ID: u2}

	result := teamHandicap([]uuid.UUID{u1.ID, u2.ID}, userMap)
	assert.Equal(t, 0.0, result)
}

func TestTeamHandicap_MissingPlayerInMap(t *testing.T) {
	// unknown ID → treated as handicap 0, should not panic
	unknownID := uuid.New()
	userMap := map[uuid.UUID]*model.User{}
	result := teamHandicap([]uuid.UUID{unknownID}, userMap)
	assert.Equal(t, 0.0, result, "unknown player defaults to 0 handicap")
}

// ─── generate1v1Schedule ──────────────────────────────────────────────────────

func TestGenerate1v1Schedule_CorrectMatchCount(t *testing.T) {
	svc := &TournamentService{}
	players := makePlayers("normal", "normal", "normal", "normal")
	matches, err := svc.generate1v1Schedule(players)
	require.NoError(t, err)
	assert.Len(t, matches, 6, "4 players → C(4,2)=6 matches")
}

func TestGenerate1v1Schedule_HandicapSnapshotted(t *testing.T) {
	svc := &TournamentService{}
	players := []*model.User{
		makeUser("pro", 0.5),
		makeUser("normal", 0),
		makeUser("normal", 0),
	}
	matches, err := svc.generate1v1Schedule(players)
	require.NoError(t, err)

	// At least one match should have a non-zero handicap (the pro player)
	hasHandicap := false
	for _, m := range matches {
		if m.HandicapTeam1 > 0 || m.HandicapTeam2 > 0 {
			hasHandicap = true
			break
		}
	}
	assert.True(t, hasHandicap, "pro player's handicap should be snapshotted in at least one match")
}

func TestGenerate1v1Schedule_StatusPending(t *testing.T) {
	svc := &TournamentService{}
	players := makePlayers("normal", "normal", "normal")
	matches, err := svc.generate1v1Schedule(players)
	require.NoError(t, err)
	for _, m := range matches {
		assert.Equal(t, "pending", m.Status)
	}
}

func TestGenerate1v1Schedule_RoundsAndOrderSet(t *testing.T) {
	svc := &TournamentService{}
	players := makePlayers("normal", "normal", "normal", "normal")
	matches, err := svc.generate1v1Schedule(players)
	require.NoError(t, err)
	for _, m := range matches {
		assert.Greater(t, m.Round, 0, "round must be ≥ 1")
		assert.Greater(t, m.MatchOrder, 0, "match_order must be ≥ 1")
	}
}

// ─── generate2v2Schedule ─────────────────────────────────────────────────────

func TestGenerate2v2Schedule_CorrectMatchCount(t *testing.T) {
	svc := &TournamentService{}
	// 4 players: dynamic scheduler covers all C(4,2)=6 opponent pairs in ~2 rounds
	players := []*model.User{
		makeUser("pro", 0.5),
		makeUser("noop", 0),
		makeUser("normal", 0),
		makeUser("normal", 0),
	}
	matches, err := svc.generate2v2Schedule(players)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(matches), 2, "4 players need at least 2 rounds to cover all opponent pairs")
	covered := make(map[pairKey]bool)
	for _, m := range matches {
		for _, p1 := range []uuid.UUID{m.Team1Player1ID, *m.Team1Player2ID} {
			for _, p2 := range []uuid.UUID{m.Team2Player1ID, *m.Team2Player2ID} {
				covered[makePairKey(p1, p2)] = true
			}
		}
	}
	for i := 0; i < len(players); i++ {
		for j := i + 1; j < len(players); j++ {
			k := makePairKey(players[i].ID, players[j].ID)
			assert.True(t, covered[k], "pair not covered")
		}
	}
}

func TestGenerate2v2Schedule_EachMatchHasTwoPlayersPerTeam(t *testing.T) {
	svc := &TournamentService{}
	players := makePlayers("normal", "normal", "normal", "normal")
	matches, err := svc.generate2v2Schedule(players)
	require.NoError(t, err)
	for _, m := range matches {
		assert.NotEqual(t, uuid.Nil, m.Team1Player1ID)
		require.NotNil(t, m.Team1Player2ID, "2v2 match must have 2 players per team")
		assert.NotEqual(t, uuid.Nil, m.Team2Player1ID)
		require.NotNil(t, m.Team2Player2ID)
	}
}

func TestGenerate2v2Schedule_TeamHandicapIsMin(t *testing.T) {
	svc := &TournamentService{}
	pro := makeUser("pro", 0.5)
	noop := makeUser("noop", 0)
	normal1 := makeUser("normal", 0)
	normal2 := makeUser("normal", 0)
	players := []*model.User{pro, noop, normal1, normal2}

	matches, err := svc.generate2v2Schedule(players)
	require.NoError(t, err)
	require.NotEmpty(t, matches)

	// Every round where pro is active should have their team's handicap = 0.5
	proID := pro.ID
	for _, m := range matches {
		inT1 := m.Team1Player1ID == proID || (m.Team1Player2ID != nil && *m.Team1Player2ID == proID)
		inT2 := m.Team2Player1ID == proID || (m.Team2Player2ID != nil && *m.Team2Player2ID == proID)
		if inT1 {
			assert.Equal(t, 0.5, m.HandicapTeam1, "team1 with pro should have handicap 0.5")
		}
		if inT2 {
			assert.Equal(t, 0.5, m.HandicapTeam2, "team2 with pro should have handicap 0.5")
		}
	}
}

func TestGenerate2v2Schedule_OddPlayers_ReturnsError(t *testing.T) {
	svc := &TournamentService{}
	players := makePlayers("normal", "normal", "normal") // odd
	_, err := svc.generate2v2Schedule(players)
	assert.Error(t, err)
}
