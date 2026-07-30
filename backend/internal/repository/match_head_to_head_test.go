package repository

import (
	"testing"
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// seedMatch inserts a match with the given participants (userID → teamNumber) and
// registers cleanup. winnerTeam sets matches.winner_team.
func seedMatch(t *testing.T, db *gorm.DB, matchType string, winnerTeam int, date time.Time, teams map[uuid.UUID]int) *model.Match {
	t.Helper()
	m := &model.Match{
		ID:         uuid.New(),
		MatchType:  matchType,
		WinnerTeam: winnerTeam,
		MatchDate:  date,
	}
	for uid, team := range teams {
		pc := -1
		if team == winnerTeam {
			pc = 1
		}
		m.Participants = append(m.Participants, model.MatchParticipant{
			ID:          uuid.New(),
			UserID:      uid,
			TeamNumber:  team,
			PointChange: pc,
		})
	}
	require.NoError(t, db.Create(m).Error)
	t.Cleanup(func() {
		db.Unscoped().Where("match_id = ?", m.ID).Delete(&model.MatchParticipant{})
		db.Unscoped().Delete(m)
	})
	return m
}

func TestGetHeadToHeadMatches_OpponentsOnly(t *testing.T) {
	db := openTestDB(t)
	repo := NewMatchRepository(db)

	a := seedUser(t, db, "H2H_A")
	b := seedUser(t, db, "H2H_B")
	c := seedUser(t, db, "H2H_C")

	// 1) A vs B opposing (1v1) — counts.
	seedMatch(t, db, "1v1", 1, time.Now().Add(-3*time.Hour), map[uuid.UUID]int{a.ID: 1, b.ID: 2})
	// 2) A and B on the SAME team (2v2 vs C+someone) — must NOT count.
	d := seedUser(t, db, "H2H_D")
	seedMatch(t, db, "2v2", 1, time.Now().Add(-2*time.Hour), map[uuid.UUID]int{a.ID: 1, b.ID: 1, c.ID: 2, d.ID: 2})
	// 3) A vs C (B absent) — must NOT count for A-vs-B.
	seedMatch(t, db, "1v1", 2, time.Now().Add(-1*time.Hour), map[uuid.UUID]int{a.ID: 1, c.ID: 2})
	// 4) A vs B opposing again (2v2), most recent — counts.
	seedMatch(t, db, "2v2", 2, time.Now(), map[uuid.UUID]int{a.ID: 1, c.ID: 1, b.ID: 2, d.ID: 2})

	rows, err := repo.GetHeadToHeadMatches(a.ID, b.ID)
	require.NoError(t, err)

	require.Len(t, rows, 2, "only the two opposing-side A-vs-B matches count")
	// Ordered most-recent first: match 4 then match 1.
	assert.True(t, rows[0].MatchDate.After(rows[1].MatchDate), "ordered by match_date DESC")
	// In match 4 A is team 1, in match 1 A is team 1.
	assert.Equal(t, 1, rows[0].P1Team)
	assert.Equal(t, 2, rows[0].WinnerTeam)
}

func TestGetHeadToHeadMatches_NeverMet(t *testing.T) {
	db := openTestDB(t)
	repo := NewMatchRepository(db)

	a := seedUser(t, db, "H2H_Lonely_A")
	b := seedUser(t, db, "H2H_Lonely_B")

	rows, err := repo.GetHeadToHeadMatches(a.ID, b.ID)
	require.NoError(t, err)
	assert.Empty(t, rows)
}
