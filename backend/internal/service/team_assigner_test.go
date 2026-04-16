package service

import (
	"testing"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helpers

func makeUser(tier string, handicap float64) *model.User {
	return &model.User{
		ID:           uuid.New(),
		Name:         "Player",
		Tier:         tier,
		HandicapRate: handicap,
		IsActive:     true,
	}
}

func makePlayers(tiers ...string) []*model.User {
	users := make([]*model.User, len(tiers))
	for i, tier := range tiers {
		users[i] = makeUser(tier, 0)
	}
	return users
}

// ─── AssignTeams1v1 ────────────────────────────────────────────────────────────

func TestAssignTeams1v1_ReturnsOneTeamPerPlayer(t *testing.T) {
	players := makePlayers("normal", "normal", "normal", "normal")
	teams := AssignTeams1v1(players)
	assert.Len(t, teams, 4)
	for i, team := range teams {
		assert.Len(t, team.Players, 1, "team %d should have exactly 1 player", i)
	}
}

func TestAssignTeams1v1_AllPlayersPresent(t *testing.T) {
	players := makePlayers("normal", "pro", "noop")
	teams := AssignTeams1v1(players)
	require.Len(t, teams, 3)

	seen := make(map[uuid.UUID]bool)
	for _, team := range teams {
		for _, id := range team.Players {
			seen[id] = true
		}
	}
	for _, p := range players {
		assert.True(t, seen[p.ID], "player %s missing from teams", p.ID)
	}
}

// ─── AssignTeams2v2 ────────────────────────────────────────────────────────────

func TestAssignTeams2v2_OddCount_ReturnsError(t *testing.T) {
	players := makePlayers("normal", "normal", "normal")
	_, err := AssignTeams2v2(players)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "even")
}

func TestAssignTeams2v2_AllNormal_ReturnsTwoTeams(t *testing.T) {
	players := makePlayers("normal", "normal", "normal", "normal")
	teams, err := AssignTeams2v2(players)
	require.NoError(t, err)
	assert.Len(t, teams, 2)
	for i, team := range teams {
		assert.Len(t, team.Players, 2, "team %d should have 2 players", i)
	}
}

func TestAssignTeams2v2_AllPlayersPresent(t *testing.T) {
	players := makePlayers("normal", "normal", "pro", "noop")
	teams, err := AssignTeams2v2(players)
	require.NoError(t, err)

	seen := make(map[uuid.UUID]bool)
	for _, team := range teams {
		for _, id := range team.Players {
			seen[id] = true
		}
	}
	for _, p := range players {
		assert.True(t, seen[p.ID], "player %s missing from assigned teams", p.ID)
	}
}

func TestAssignTeams2v2_ProNotPairedWithPro(t *testing.T) {
	// Run multiple times to account for randomness
	pro1 := makeUser("pro", 0.5)
	pro2 := makeUser("pro", 0.5)
	noop1 := makeUser("noop", 0)
	normal1 := makeUser("normal", 0)
	players := []*model.User{pro1, pro2, noop1, normal1}

	proIDs := map[uuid.UUID]bool{pro1.ID: true, pro2.ID: true}

	for i := 0; i < 20; i++ {
		teams, err := AssignTeams2v2(players)
		require.NoError(t, err)
		for _, team := range teams {
			allPro := true
			for _, id := range team.Players {
				if !proIDs[id] {
					allPro = false
					break
				}
			}
			assert.False(t, allPro, "iteration %d: a team is all-Pro (should be impossible with available non-Pros)", i)
		}
	}
}

func TestAssignTeams2v2_ProPreferNoop(t *testing.T) {
	// 1 Pro, 1 Noop, 2 Normals → Pro must be paired with Noop
	pro := makeUser("pro", 0.5)
	noop := makeUser("noop", 0)
	normal1 := makeUser("normal", 0)
	normal2 := makeUser("normal", 0)
	players := []*model.User{pro, noop, normal1, normal2}

	for i := 0; i < 20; i++ {
		teams, err := AssignTeams2v2(players)
		require.NoError(t, err)
		// Find team containing the pro
		for _, team := range teams {
			hasPro := false
			hasNoop := false
			for _, id := range team.Players {
				if id == pro.ID {
					hasPro = true
				}
				if id == noop.ID {
					hasNoop = true
				}
			}
			if hasPro {
				assert.True(t, hasNoop, "iteration %d: Pro should be paired with Noop when available", i)
			}
		}
	}
}

func TestAssignTeams2v2_AllPros_FallbackAllowed(t *testing.T) {
	// 4 pros — should succeed (fallback Pro-Pro pairing)
	players := makePlayers("pro", "pro", "pro", "pro")
	teams, err := AssignTeams2v2(players)
	require.NoError(t, err)
	assert.Len(t, teams, 2)
}

func TestAssignTeams2v2_NoProPlayers(t *testing.T) {
	players := makePlayers("noop", "noop", "normal", "normal")
	teams, err := AssignTeams2v2(players)
	require.NoError(t, err)
	assert.Len(t, teams, 2)
	for _, team := range teams {
		assert.Len(t, team.Players, 2)
	}
}

func TestAssignTeams2v2_SixPlayers_ThreeTeams(t *testing.T) {
	players := makePlayers("pro", "normal", "normal", "noop", "normal", "normal")
	teams, err := AssignTeams2v2(players)
	require.NoError(t, err)
	assert.Len(t, teams, 3)
}

func TestAssignTeams2v2_NoDuplicatePlayers(t *testing.T) {
	players := makePlayers("pro", "noop", "normal", "normal", "normal", "normal")
	teams, err := AssignTeams2v2(players)
	require.NoError(t, err)

	seen := make(map[uuid.UUID]bool)
	for _, team := range teams {
		for _, id := range team.Players {
			assert.False(t, seen[id], "player %s assigned to multiple teams", id)
			seen[id] = true
		}
	}
}
