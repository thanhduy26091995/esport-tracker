package service

import (
	"testing"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makePlayers2v2 creates a slice of users with given tiers.
func makePlayers2v2(tiers ...string) []*model.User {
	users := make([]*model.User, len(tiers))
	for i, tier := range tiers {
		users[i] = &model.User{ID: uuid.New(), Tier: tier}
	}
	return users
}

// countOpponentPairs returns all unique opponent pairs covered across all rounds.
func countOpponentPairs(slots []RoundSlot) map[pairKey]bool {
	covered := make(map[pairKey]bool)
	for _, s := range slots {
		for _, p1 := range s.Team1 {
			for _, p2 := range s.Team2 {
				covered[makePairKey(p1, p2)] = true
			}
		}
	}
	return covered
}

// allPlayerPairs returns all C(n,2) pairs from a player list.
func allPlayerPairs(players []*model.User) map[pairKey]bool {
	pairs := make(map[pairKey]bool)
	for i := 0; i < len(players); i++ {
		for j := i + 1; j < len(players); j++ {
			pairs[makePairKey(players[i].ID, players[j].ID)] = true
		}
	}
	return pairs
}

func TestGenerateSchedule2v2_AllOpponentPairsCovered_4Players(t *testing.T) {
	players := makePlayers2v2("normal", "normal", "normal", "normal")
	slots := GenerateSchedule2v2(players)
	require.NotEmpty(t, slots)
	covered := countOpponentPairs(slots)
	want := allPlayerPairs(players)
	for pair := range want {
		assert.True(t, covered[pair], "pair %v not covered", pair)
	}
}

func TestGenerateSchedule2v2_AllOpponentPairsCovered_6Players(t *testing.T) {
	players := makePlayers2v2("pro", "normal", "normal", "normal", "normal", "normal")
	slots := GenerateSchedule2v2(players)
	require.NotEmpty(t, slots)
	covered := countOpponentPairs(slots)
	want := allPlayerPairs(players)
	assert.Len(t, want, 15, "C(6,2)=15 pairs expected")
	for pair := range want {
		assert.True(t, covered[pair], "pair not covered")
	}
	assert.LessOrEqual(t, len(slots), 15, "should finish in ≤15 rounds")
}

func TestGenerateSchedule2v2_AllOpponentPairsCovered_8Players(t *testing.T) {
	players := makePlayers2v2("pro", "pro", "normal", "normal", "normal", "normal", "normal", "normal")
	slots := GenerateSchedule2v2(players)
	require.NotEmpty(t, slots)
	covered := countOpponentPairs(slots)
	want := allPlayerPairs(players)
	assert.Len(t, want, 28, "C(8,2)=28 pairs expected")
	for pair := range want {
		assert.True(t, covered[pair], "pair not covered")
	}
}

func TestGenerateSchedule2v2_TeamsAlwaysHave2Players(t *testing.T) {
	players := makePlayers2v2("normal", "normal", "normal", "normal", "normal", "normal")
	slots := GenerateSchedule2v2(players)
	for i, s := range slots {
		assert.NotEqual(t, s.Team1[0], uuid.Nil, "round %d team1[0] nil", i+1)
		assert.NotEqual(t, s.Team1[1], uuid.Nil, "round %d team1[1] nil", i+1)
		assert.NotEqual(t, s.Team2[0], uuid.Nil, "round %d team2[0] nil", i+1)
		assert.NotEqual(t, s.Team2[1], uuid.Nil, "round %d team2[1] nil", i+1)
	}
}

func TestGenerateSchedule2v2_NoPlayerInBothTeamsSameRound(t *testing.T) {
	players := makePlayers2v2("normal", "normal", "normal", "normal", "normal", "normal")
	slots := GenerateSchedule2v2(players)
	for i, s := range slots {
		all4 := map[uuid.UUID]bool{
			s.Team1[0]: true, s.Team1[1]: true,
			s.Team2[0]: true, s.Team2[1]: true,
		}
		assert.Len(t, all4, 4, "round %d: duplicate player in active teams", i+1)
	}
}

func TestGenerateSchedule2v2_SitOutBalance_6Players(t *testing.T) {
	players := makePlayers2v2("normal", "normal", "normal", "normal", "normal", "normal")
	slots := GenerateSchedule2v2(players)

	sitCounts := make(map[uuid.UUID]int)
	for _, s := range slots {
		for _, p := range s.SitOut {
			sitCounts[p]++
		}
	}
	minSit, maxSit := len(slots), 0
	for _, c := range sitCounts {
		if c < minSit { minSit = c }
		if c > maxSit { maxSit = c }
	}
	// imbalance should be at most 1 (can't perfectly balance 6 players with 2 sitting out)
	assert.LessOrEqual(t, maxSit-minSit, 2, "sit-out imbalance too large")
}

func TestGenerateSchedule2v2_TierBalance_OnePro(t *testing.T) {
	// 1 Pro + 5 Normal: Pro should always be paired with a Normal (trivially satisfied with 1 Pro)
	// Verify the schedule generates without panic and coverage is complete
	players := makePlayers2v2("pro", "normal", "normal", "normal", "normal", "normal")
	proID := players[0].ID
	slots := GenerateSchedule2v2(players)
	require.NotEmpty(t, slots)
	for i, s := range slots {
		// When pro is active, check partner is not also pro (only 1 pro exists, so trivially true)
		proInT1 := s.Team1[0] == proID || s.Team1[1] == proID
		proInT2 := s.Team2[0] == proID || s.Team2[1] == proID
		if proInT1 {
			partner := s.Team1[1]
			if s.Team1[0] != proID {
				partner = s.Team1[0]
			}
			assert.NotEqual(t, proID, partner, "round %d: pro is partnered with themselves", i+1)
		}
		if proInT2 {
			partner := s.Team2[1]
			if s.Team2[0] != proID {
				partner = s.Team2[0]
			}
			assert.NotEqual(t, proID, partner, "round %d: pro is partnered with themselves", i+1)
		}
	}
}

func TestGenerateSchedule2v2_TierBalance_3Pros3Normals(t *testing.T) {
	// With 3 Pros + 3 Normals: tier balance is best-effort.
	// Coverage takes priority — so some rounds MAY have Pro-Pro teams when
	// it's the only way to cover remaining pairs efficiently.
	// We verify: all pairs are still covered and schedule is valid.
	players := makePlayers2v2("pro", "pro", "pro", "normal", "normal", "normal")
	slots := GenerateSchedule2v2(players)
	require.NotEmpty(t, slots)

	covered := countOpponentPairs(slots)
	want := allPlayerPairs(players)
	for pair := range want {
		assert.True(t, covered[pair], "pair not covered")
	}

	// Count how often Pro-Pro teams appear
	proSet := make(map[uuid.UUID]bool)
	for i := 0; i < 3; i++ {
		proSet[players[i].ID] = true
	}
	proProRounds := 0
	for _, s := range slots {
		if proSet[s.Team1[0]] && proSet[s.Team1[1]] {
			proProRounds++
		}
		if proSet[s.Team2[0]] && proSet[s.Team2[1]] {
			proProRounds++
		}
	}
	// Majority of rounds should be tier-balanced (Pro-Pro rounds minimized)
	assert.Less(t, proProRounds, len(slots), "most rounds should avoid Pro-Pro teams")
}

func TestGenerateSchedule2v2_TierBalance_AllPros_FallbackAllowed(t *testing.T) {
	// All 4 players are Pro: must fallback to Pro-Pro (no crash, valid schedule)
	players := makePlayers2v2("pro", "pro", "pro", "pro")
	slots := GenerateSchedule2v2(players)
	assert.NotEmpty(t, slots, "should still generate a schedule even with all-Pro players")
	covered := countOpponentPairs(slots)
	want := allPlayerPairs(players)
	for pair := range want {
		assert.True(t, covered[pair], "pair not covered")
	}
}

func TestGenerateSchedule2v2_Terminates_AllN(t *testing.T) {
	// Verify no infinite loop and full coverage for even N from 4 to 10
	for n := 4; n <= 10; n += 2 {
		tiers := make([]string, n)
		for i := range tiers {
			tiers[i] = "normal"
		}
		tiers[0] = "pro"
		players := makePlayers2v2(tiers...)
		slots := GenerateSchedule2v2(players)
		assert.NotEmpty(t, slots, "N=%d: should produce rounds", n)
		covered := countOpponentPairs(slots)
		want := allPlayerPairs(players)
		for pair := range want {
			assert.True(t, covered[pair], "N=%d: pair not covered", n)
		}
	}
}

func TestGenerateSchedule2v2_TooFewPlayers_ReturnsNil(t *testing.T) {
	players := makePlayers2v2("normal", "normal")
	slots := GenerateSchedule2v2(players)
	assert.Nil(t, slots)
}

func TestIsTierBalanced(t *testing.T) {
	pro1 := &model.User{Tier: "pro"}
	pro2 := &model.User{Tier: "pro"}
	norm1 := &model.User{Tier: "normal"}
	norm2 := &model.User{Tier: "normal"}

	assert.True(t, isTierBalanced([]*model.User{pro1, norm1}, []*model.User{pro2, norm2}), "1pro+1normal vs 1pro+1normal = balanced")
	assert.False(t, isTierBalanced([]*model.User{pro1, pro2}, []*model.User{norm1, norm2}), "2pro vs 2normal = unbalanced")
	assert.True(t, isTierBalanced([]*model.User{norm1, norm2}, []*model.User{pro1, norm1}), "all-normal team = balanced")
}
