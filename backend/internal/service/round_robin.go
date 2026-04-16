package service

// MatchPair represents two participants (by index) that play each other
type MatchPair struct {
	A int // index into participants slice
	B int // index into participants slice
}

// GenerateRoundRobin generates a round-robin schedule using polygon rotation.
// Returns rounds, each round is a list of MatchPair (participant indices).
// Bye matches (where A or B == -1) are excluded.
func GenerateRoundRobin(n int) [][]MatchPair {
	players := make([]int, n)
	for i := range players {
		players[i] = i
	}

	hasBye := n%2 == 1
	if hasBye {
		players = append(players, -1) // -1 = bye
		n++
	}

	numRounds := n - 1
	rounds := make([][]MatchPair, 0, numRounds)

	for round := 0; round < numRounds; round++ {
		pairs := make([]MatchPair, 0, n/2)
		for i := 0; i < n/2; i++ {
			a := players[i]
			b := players[n-1-i]
			if a != -1 && b != -1 {
				pairs = append(pairs, MatchPair{A: a, B: b})
			}
		}
		if len(pairs) > 0 {
			rounds = append(rounds, pairs)
		}
		// Rotate: fix players[0], rotate players[1:]
		last := players[n-1]
		copy(players[2:], players[1:n-1])
		players[1] = last
	}

	return rounds
}
