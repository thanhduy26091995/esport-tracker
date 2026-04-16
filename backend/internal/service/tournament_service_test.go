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
	assert.True(t, req.affectsScore(), "nil AffectsScore should default to true")
}

func TestCreateTournamentRequest_AffectsScoreExplicitFalse(t *testing.T) {
	f := false
	req := &CreateTournamentRequest{AffectsScore: &f}
	assert.False(t, req.affectsScore(), "explicit false should be false")
}

func TestCreateTournamentRequest_AffectsScoreExplicitTrue(t *testing.T) {
	tr := true
	req := &CreateTournamentRequest{AffectsScore: &tr}
	assert.True(t, req.affectsScore(), "explicit true should be true")
}

// ─── RecordMatchResultRequest — teamHandicap helper ───────────────────────────

func TestTeamHandicap_MaxOfTeam(t *testing.T) {
	u1 := makeUser("pro", 0.5)
	u2 := makeUser("normal", 1.0)
	userMap := map[uuid.UUID]*model.User{u1.ID: u1, u2.ID: u2}

	result := teamHandicap([]uuid.UUID{u1.ID, u2.ID}, userMap)
	assert.Equal(t, 1.0, result, "team handicap = max of individual handicaps")
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
	// 4 players → 2 teams → C(2,2)=1 match
	players := []*model.User{
		makeUser("pro", 0.5),
		makeUser("noop", 0),
		makeUser("normal", 0),
		makeUser("normal", 0),
	}
	matches, err := svc.generate2v2Schedule(players)
	require.NoError(t, err)
	assert.Len(t, matches, 1, "4 players → 2 teams → 1 match")
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

func TestGenerate2v2Schedule_TeamHandicapIsMax(t *testing.T) {
	svc := &TournamentService{}
	// Pro(0.5) + Normal(0) vs Normal(0) + Normal(0)
	pro := makeUser("pro", 0.5)
	noop := makeUser("noop", 0)
	normal1 := makeUser("normal", 0)
	normal2 := makeUser("normal", 0)
	players := []*model.User{pro, noop, normal1, normal2}

	matches, err := svc.generate2v2Schedule(players)
	require.NoError(t, err)
	require.Len(t, matches, 1)

	m := matches[0]
	// One team has the pro (handicap=0.5), other has no handicap
	maxHandicap := m.HandicapTeam1
	if m.HandicapTeam2 > maxHandicap {
		maxHandicap = m.HandicapTeam2
	}
	assert.Equal(t, 0.5, maxHandicap, "team containing pro should have handicap 0.5")
	minHandicap := m.HandicapTeam1
	if m.HandicapTeam2 < minHandicap {
		minHandicap = m.HandicapTeam2
	}
	assert.Equal(t, 0.0, minHandicap, "team without pro should have handicap 0")
}

func TestGenerate2v2Schedule_OddPlayers_ReturnsError(t *testing.T) {
	svc := &TournamentService{}
	players := makePlayers("normal", "normal", "normal") // odd
	_, err := svc.generate2v2Schedule(players)
	assert.Error(t, err)
}
