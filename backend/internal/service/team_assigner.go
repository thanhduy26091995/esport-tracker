package service

import (
	"errors"
	"math/rand"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
)

// Team represents 1 or 2 player IDs forming a team
type Team struct {
	Players []uuid.UUID
}

// AssignTeams1v1 shuffles players and returns them as individual teams for round-robin
func AssignTeams1v1(users []*model.User) []Team {
	shuffled := make([]*model.User, len(users))
	copy(shuffled, users)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

	teams := make([]Team, len(shuffled))
	for i, u := range shuffled {
		teams[i] = Team{Players: []uuid.UUID{u.ID}}
	}
	return teams
}

// AssignTeams2v2 creates balanced 2-player teams.
// Rules: Pro players must be paired with Noop (preferred) or Normal players.
// If impossible (too many Pros), allow Pro-Pro pairing.
func AssignTeams2v2(users []*model.User) ([]Team, error) {
	if len(users)%2 != 0 {
		return nil, errors.New("2v2 requires an even number of players")
	}

	// Separate by tier
	var pros, noops, normals []*model.User
	for _, u := range users {
		switch u.Tier {
		case "pro":
			pros = append(pros, u)
		case "noop":
			noops = append(noops, u)
		default: // "normal" or anything else
			normals = append(normals, u)
		}
	}

	// Shuffle each tier group
	rand.Shuffle(len(pros), func(i, j int) { pros[i], pros[j] = pros[j], pros[i] })
	rand.Shuffle(len(noops), func(i, j int) { noops[i], noops[j] = noops[j], noops[i] })
	rand.Shuffle(len(normals), func(i, j int) { normals[i], normals[j] = normals[j], normals[i] })

	// Pool of non-pro players (noops preferred, then normals)
	pool := append(noops, normals...)

	var pairs [][2]*model.User

	// Pair each pro with a non-pro
	for _, pro := range pros {
		if len(pool) > 0 {
			partner := pool[0]
			pool = pool[1:]
			pairs = append(pairs, [2]*model.User{pro, partner})
		} else {
			// No non-pro left — will pair pros together below
			pool = append(pool, pro) // put back in pool
		}
	}

	// Pair remaining in pool (could include excess pros)
	rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	for i := 0; i+1 < len(pool); i += 2 {
		pairs = append(pairs, [2]*model.User{pool[i], pool[i+1]})
	}

	// Convert pairs to teams
	teams := make([]Team, len(pairs))
	for i, pair := range pairs {
		teams[i] = Team{Players: []uuid.UUID{pair[0].ID, pair[1].ID}}
	}
	return teams, nil
}
