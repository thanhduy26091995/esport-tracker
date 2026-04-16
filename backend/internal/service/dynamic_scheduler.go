package service

import (
	"math"
	"sort"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
)

// RoundSlot represents one scheduled 2v2 round: 4 active players, rest sit out.
type RoundSlot struct {
	Team1   [2]uuid.UUID
	Team2   [2]uuid.UUID
	SitOut  []uuid.UUID
}

// pairKey is a canonical (sorted) unordered pair of player IDs.
type pairKey struct{ A, B uuid.UUID }

func makePairKey(a, b uuid.UUID) pairKey {
	if a.String() < b.String() {
		return pairKey{a, b}
	}
	return pairKey{b, a}
}

// schedulerState tracks coverage and diversity metrics across rounds.
type schedulerState struct {
	allPairs        map[pairKey]bool
	opponentCovered map[pairKey]bool
	teammateCounts  map[pairKey]int
	opponentCounts  map[pairKey]int
	sitOutCounts    map[uuid.UUID]int
}

func newSchedulerState(players []*model.User) *schedulerState {
	s := &schedulerState{
		allPairs:        make(map[pairKey]bool),
		opponentCovered: make(map[pairKey]bool),
		teammateCounts:  make(map[pairKey]int),
		opponentCounts:  make(map[pairKey]int),
		sitOutCounts:    make(map[uuid.UUID]int),
	}
	for i := 0; i < len(players); i++ {
		for j := i + 1; j < len(players); j++ {
			s.allPairs[makePairKey(players[i].ID, players[j].ID)] = false
		}
		s.sitOutCounts[players[i].ID] = 0
	}
	return s
}

func (s *schedulerState) allCovered() bool {
	for k := range s.allPairs {
		if !s.opponentCovered[k] {
			return false
		}
	}
	return true
}

func (s *schedulerState) apply(slot RoundSlot) {
	// Update teammate counts
	s.teammateCounts[makePairKey(slot.Team1[0], slot.Team1[1])]++
	s.teammateCounts[makePairKey(slot.Team2[0], slot.Team2[1])]++
	// Update opponent counts and coverage
	for _, p1 := range slot.Team1 {
		for _, p2 := range slot.Team2 {
			k := makePairKey(p1, p2)
			s.opponentCovered[k] = true
			s.opponentCounts[k]++
		}
	}
	// Update sit-out counts
	for _, p := range slot.SitOut {
		s.sitOutCounts[p]++
	}
}

// GenerateSchedule2v2 produces a schedule of 2v2 rounds for the given players.
// It runs until every pair of players has been opponents at least once (covers C(N,2) pairs).
// Tier-balance is enforced per round: Pro players are paired with NonPro when possible.
func GenerateSchedule2v2(players []*model.User) []RoundSlot {
	n := len(players)
	if n < 4 {
		return nil
	}

	state := newSchedulerState(players)
	// Safety cap: theoretically need ceil(C(n,2)/4) rounds minimum.
	// Use n*n to give ample room for fairness constraints.
	maxRounds := n * n

	var slots []RoundSlot
	for round := 0; round < maxRounds && !state.allCovered(); round++ {
		slot := pickBestRound(players, state)
		state.apply(slot)
		slots = append(slots, slot)
	}
	return slots
}

// pickBestRound selects the best RoundSlot from all candidate 4-player subsets.
// Coverage (new opponent pairs) strictly dominates — a round covering more new pairs
// always beats a round covering fewer, regardless of fairness. Fairness is tie-breaker.
// Tier-balanced candidates (no Pro-Pro team) are always preferred over unbalanced ones
// within the same coverage tier.
func pickBestRound(players []*model.User, state *schedulerState) RoundSlot {
	n := len(players)

	playerIDs := make([]uuid.UUID, n)
	for i, p := range players {
		playerIDs[i] = p.ID
	}

	bestNewPairs := -1
	bestFairScore := math.MaxFloat64
	bestBalanced := false
	var bestSlot RoundSlot

	splits := [3][2][2]int{
		{{0, 1}, {2, 3}},
		{{0, 2}, {1, 3}},
		{{0, 3}, {1, 2}},
	}

	fourIdxCombinations(n, func(i, j, k, l int) {
		active := [4]*model.User{players[i], players[j], players[k], players[l]}
		sitOut := makeSitOut(playerIDs, []uuid.UUID{active[0].ID, active[1].ID, active[2].ID, active[3].ID})

		// Evaluate all 3 splits for this 4-player subset
		for _, split := range splits {
			t1 := [2]*model.User{active[split[0][0]], active[split[0][1]]}
			t2 := [2]*model.User{active[split[1][0]], active[split[1][1]]}

			balanced := isTierBalanced(t1[:], t2[:])
			newPairs, fairScore := computeScoreParts(t1[:], t2[:], sitOut, state)

			slot := RoundSlot{
				Team1:  [2]uuid.UUID{t1[0].ID, t1[1].ID},
				Team2:  [2]uuid.UUID{t2[0].ID, t2[1].ID},
				SitOut: sitOut,
			}

			if bestNewPairs == -1 {
				// First candidate ever
				bestNewPairs = newPairs
				bestFairScore = fairScore
				bestBalanced = balanced
				bestSlot = slot
				continue
			}

			// Coverage strictly dominates
			if newPairs > bestNewPairs {
				bestNewPairs = newPairs
				bestFairScore = fairScore
				bestBalanced = balanced
				bestSlot = slot
				continue
			}
			if newPairs < bestNewPairs {
				continue
			}

			// Same coverage: prefer tier-balanced
			if balanced && !bestBalanced {
				bestFairScore = fairScore
				bestBalanced = balanced
				bestSlot = slot
				continue
			}
			if !balanced && bestBalanced {
				continue
			}

			// Same tier-balance: pick better fairness score
			if fairScore < bestFairScore {
				bestFairScore = fairScore
				bestSlot = slot
			}
		}
	})

	return bestSlot
}

// computeScoreParts returns (newPairs, fairnessScore) for a candidate round.
// newPairs: how many new opponent pairs this round covers (higher = better).
// fairnessScore: combined penalty for repeated teammates, repeated opponents,
// and sit-out imbalance (lower = better), used as tie-breaker only.
func computeScoreParts(t1, t2 []*model.User, sitOut []uuid.UUID, state *schedulerState) (int, float64) {
	newPairs := 0
	maxRepeatOpp := 0
	for _, p1 := range t1 {
		for _, p2 := range t2 {
			k := makePairKey(p1.ID, p2.ID)
			if !state.opponentCovered[k] {
				newPairs++
			}
			if c := state.opponentCounts[k]; c > maxRepeatOpp {
				maxRepeatOpp = c
			}
		}
	}

	tm1 := state.teammateCounts[makePairKey(t1[0].ID, t1[1].ID)]
	tm2 := state.teammateCounts[makePairKey(t2[0].ID, t2[1].ID)]
	maxRepeatTm := int(math.Max(float64(tm1), float64(tm2)))

	sitImbalance := 0.0
	if len(sitOut) > 0 {
		minSit, maxSit := math.MaxInt32, 0
		for _, p := range sitOut {
			c := state.sitOutCounts[p]
			if c < minSit {
				minSit = c
			}
			if c > maxSit {
				maxSit = c
			}
		}
		if minSit == math.MaxInt32 {
			minSit = 0
		}
		sitImbalance = float64(maxSit - minSit)
	}

	fairScore := float64(maxRepeatTm) + float64(maxRepeatOpp)*0.5 + sitImbalance*2
	return newPairs, fairScore
}

// isTierBalanced returns true if neither team is all-Pro.
// A team is "unbalanced" if both members are Pro (tier == "pro").
func isTierBalanced(t1, t2 []*model.User) bool {
	return !allPro(t1) && !allPro(t2)
}

func allPro(team []*model.User) bool {
	for _, u := range team {
		if u.Tier != "pro" {
			return false
		}
	}
	return true
}

// fourIdxCombinations calls fn for every combination of 4 indices from [0, n).
func fourIdxCombinations(n int, fn func(i, j, k, l int)) {
	for i := 0; i < n-3; i++ {
		for j := i + 1; j < n-2; j++ {
			for k := j + 1; k < n-1; k++ {
				for l := k + 1; l < n; l++ {
					fn(i, j, k, l)
				}
			}
		}
	}
}

// makeSitOut returns the player IDs not in the active set.
func makeSitOut(all []uuid.UUID, active []uuid.UUID) []uuid.UUID {
	activeSet := make(map[uuid.UUID]bool, len(active))
	for _, id := range active {
		activeSet[id] = true
	}
	var sitOut []uuid.UUID
	for _, id := range all {
		if !activeSet[id] {
			sitOut = append(sitOut, id)
		}
	}
	// Sort for determinism in tests
	sort.Slice(sitOut, func(i, j int) bool {
		return sitOut[i].String() < sitOut[j].String()
	})
	return sitOut
}
