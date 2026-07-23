package service

import (
	"testing"
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- isBetLocked ---

func TestIsBetLocked(t *testing.T) {
	future := time.Now().Add(1 * time.Hour)
	past := time.Now().Add(-1 * time.Hour)

	tests := []struct {
		name  string
		match model.WcMatch
		want  bool
	}{
		{
			name:  "scheduled, no lock time — not locked",
			match: model.WcMatch{Status: model.WcStatusScheduled, MatchDate: future},
			want:  false,
		},
		{
			name:  "live — locked",
			match: model.WcMatch{Status: model.WcStatusLive, MatchDate: future},
			want:  true,
		},
		{
			name:  "completed — locked",
			match: model.WcMatch{Status: model.WcStatusCompleted, MatchDate: future},
			want:  true,
		},
		{
			name:  "cancelled — locked",
			match: model.WcMatch{Status: model.WcStatusCancelled, MatchDate: future},
			want:  true,
		},
		{
			name:  "scheduled, match_date passed — locked",
			match: model.WcMatch{Status: model.WcStatusScheduled, MatchDate: past},
			want:  true,
		},
		{
			name:  "scheduled with future lock time — not yet locked",
			match: model.WcMatch{Status: model.WcStatusScheduled, MatchDate: future, BetsLockedAt: &future},
			want:  false,
		},
		{
			name:  "scheduled with past lock time — locked",
			match: model.WcMatch{Status: model.WcStatusScheduled, MatchDate: future, BetsLockedAt: &past},
			want:  true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, isBetLocked(&tc.match))
		})
	}
}

// --- validatePenaltyConfig ---

func TestValidatePenaltyConfig(t *testing.T) {
	tests := []struct {
		name                string
		cancelPercent       int
		reduceMaxPercent    int
		reducePenaltyPercent int
		wantErr             bool
		errContains         string
	}{
		{
			name: "all valid — boundary zeros",
			cancelPercent: 0, reduceMaxPercent: 0, reducePenaltyPercent: 0,
			wantErr: false,
		},
		{
			name: "all valid — boundary 100s",
			cancelPercent: 100, reduceMaxPercent: 100, reducePenaltyPercent: 100,
			wantErr: false,
		},
		{
			name: "typical valid config",
			cancelPercent: 20, reduceMaxPercent: 50, reducePenaltyPercent: 20,
			wantErr: false,
		},
		{
			name: "cancelPercent negative",
			cancelPercent: -1, reduceMaxPercent: 50, reducePenaltyPercent: 20,
			wantErr: true, errContains: "cancel_penalty_percent",
		},
		{
			name: "cancelPercent over 100",
			cancelPercent: 101, reduceMaxPercent: 50, reducePenaltyPercent: 20,
			wantErr: true, errContains: "cancel_penalty_percent",
		},
		{
			name: "reduceMaxPercent negative",
			cancelPercent: 20, reduceMaxPercent: -1, reducePenaltyPercent: 20,
			wantErr: true, errContains: "bet_reduce_max_percent",
		},
		{
			name: "reduceMaxPercent over 100",
			cancelPercent: 20, reduceMaxPercent: 101, reducePenaltyPercent: 20,
			wantErr: true, errContains: "bet_reduce_max_percent",
		},
		{
			name: "reducePenaltyPercent negative",
			cancelPercent: 20, reduceMaxPercent: 50, reducePenaltyPercent: -1,
			wantErr: true, errContains: "bet_reduce_penalty_percent",
		},
		{
			name: "reducePenaltyPercent over 100",
			cancelPercent: 20, reduceMaxPercent: 50, reducePenaltyPercent: 101,
			wantErr: true, errContains: "bet_reduce_penalty_percent",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePenaltyConfig(tc.cancelPercent, tc.reduceMaxPercent, tc.reducePenaltyPercent)
			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// --- sortBetHistoryNewestFirst ---

func TestSortBetHistoryNewestFirst(t *testing.T) {
	now := time.Now()
	oldest := now.Add(-2 * time.Hour)
	middle := now.Add(-1 * time.Hour)
	newest := now

	t.Run("three items sorted oldest-first → newest-first after sort", func(t *testing.T) {
		items := []model.BetHistoryItem{
			{ID: "a", CreatedAt: oldest},
			{ID: "b", CreatedAt: newest},
			{ID: "c", CreatedAt: middle},
		}
		sortBetHistoryNewestFirst(items)
		assert.Equal(t, "b", items[0].ID)
		assert.Equal(t, "c", items[1].ID)
		assert.Equal(t, "a", items[2].ID)
	})

	t.Run("already sorted — unchanged", func(t *testing.T) {
		items := []model.BetHistoryItem{
			{ID: "newest", CreatedAt: newest},
			{ID: "middle", CreatedAt: middle},
			{ID: "oldest", CreatedAt: oldest},
		}
		sortBetHistoryNewestFirst(items)
		assert.Equal(t, "newest", items[0].ID)
		assert.Equal(t, "middle", items[1].ID)
		assert.Equal(t, "oldest", items[2].ID)
	})

	t.Run("single item — unchanged", func(t *testing.T) {
		items := []model.BetHistoryItem{{ID: "only", CreatedAt: now}}
		sortBetHistoryNewestFirst(items)
		assert.Equal(t, "only", items[0].ID)
	})

	t.Run("empty slice — no panic", func(t *testing.T) {
		items := []model.BetHistoryItem{}
		sortBetHistoryNewestFirst(items)
		assert.Len(t, items, 0)
	})

	t.Run("mixed regular and custom kinds retain correct order", func(t *testing.T) {
		items := []model.BetHistoryItem{
			{ID: "r1", Kind: "regular", CreatedAt: oldest},
			{ID: "c1", Kind: "custom", CreatedAt: newest},
			{ID: "r2", Kind: "regular", CreatedAt: middle},
		}
		sortBetHistoryNewestFirst(items)
		assert.Equal(t, "c1", items[0].ID)
		assert.Equal(t, "r2", items[1].ID)
		assert.Equal(t, "r1", items[2].ID)
	})
}
