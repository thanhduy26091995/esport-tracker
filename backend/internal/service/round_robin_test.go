package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRoundRobin_TwoPlayers(t *testing.T) {
	rounds := GenerateRoundRobin(2)
	require.Len(t, rounds, 1, "2 players → 1 round")
	assert.Len(t, rounds[0], 1, "1 match in the round")
	assert.Equal(t, MatchPair{A: 0, B: 1}, rounds[0][0])
}

func TestGenerateRoundRobin_ThreePlayers(t *testing.T) {
	// N=3 (odd) → bye added → N=4, 3 rounds, 1 real match each
	rounds := GenerateRoundRobin(3)
	require.Len(t, rounds, 3, "3 players → 3 rounds")
	totalMatches := countMatches(rounds)
	assert.Equal(t, 3, totalMatches, "3 players → 3 total matches (C(3,2))")
}

func TestGenerateRoundRobin_FourPlayers(t *testing.T) {
	rounds := GenerateRoundRobin(4)
	require.Len(t, rounds, 3, "4 players → 3 rounds")
	for i, r := range rounds {
		assert.Len(t, r, 2, "round %d should have 2 matches", i+1)
	}
	assert.Equal(t, 6, countMatches(rounds), "4 players → 6 total matches (C(4,2))")
}

func TestGenerateRoundRobin_FivePlayers(t *testing.T) {
	rounds := GenerateRoundRobin(5)
	assert.Equal(t, 10, countMatches(rounds), "5 players → 10 total matches (C(5,2))")
}

func TestGenerateRoundRobin_SixPlayers(t *testing.T) {
	rounds := GenerateRoundRobin(6)
	require.Len(t, rounds, 5, "6 players → 5 rounds")
	assert.Equal(t, 15, countMatches(rounds), "6 players → 15 total matches (C(6,2))")
}

func TestGenerateRoundRobin_EachPairAppearsOnce(t *testing.T) {
	for n := 2; n <= 8; n++ {
		rounds := GenerateRoundRobin(n)
		seen := make(map[[2]int]bool)
		for _, round := range rounds {
			for _, pair := range round {
				key := [2]int{pair.A, pair.B}
				if pair.A > pair.B {
					key = [2]int{pair.B, pair.A}
				}
				assert.False(t, seen[key], "n=%d: pair (%d,%d) appears more than once", n, pair.A, pair.B)
				seen[key] = true
			}
		}
		expected := n * (n - 1) / 2
		assert.Equal(t, expected, len(seen), "n=%d: expected %d unique pairs", n, expected)
	}
}

func TestGenerateRoundRobin_NoBye(t *testing.T) {
	// All pairs must be valid player indices (no -1 leaks)
	for n := 2; n <= 8; n++ {
		rounds := GenerateRoundRobin(n)
		for _, round := range rounds {
			for _, pair := range round {
				assert.GreaterOrEqual(t, pair.A, 0, "n=%d: pair.A should not be -1 (bye)", n)
				assert.GreaterOrEqual(t, pair.B, 0, "n=%d: pair.B should not be -1 (bye)", n)
				assert.Less(t, pair.A, n, "n=%d: pair.A out of range", n)
				assert.Less(t, pair.B, n, "n=%d: pair.B out of range", n)
			}
		}
	}
}

func TestGenerateRoundRobin_EachPlayerOncePerRound(t *testing.T) {
	// In each round, each player appears at most once
	for n := 2; n <= 8; n++ {
		rounds := GenerateRoundRobin(n)
		for ri, round := range rounds {
			playersInRound := make(map[int]bool)
			for _, pair := range round {
				assert.False(t, playersInRound[pair.A], "n=%d round %d: player %d appears twice", n, ri, pair.A)
				assert.False(t, playersInRound[pair.B], "n=%d round %d: player %d appears twice", n, ri, pair.B)
				playersInRound[pair.A] = true
				playersInRound[pair.B] = true
			}
		}
	}
}

func TestGenerateRoundRobin_TotalMatchFormula(t *testing.T) {
	for n := 2; n <= 16; n++ {
		rounds := GenerateRoundRobin(n)
		expected := n * (n - 1) / 2
		actual := countMatches(rounds)
		assert.Equal(t, expected, actual, "n=%d: C(n,2) formula must hold", n)
	}
}

// countMatches sums all matches across all rounds
func countMatches(rounds [][]MatchPair) int {
	total := 0
	for _, r := range rounds {
		total += len(r)
	}
	return total
}
