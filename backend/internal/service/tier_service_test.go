package service

import (
	"errors"
	"testing"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── EvaluateTier ─────────────────────────────────────────────────────────────

func TestEvaluateTier_ProAboveThreshold(t *testing.T) {
	assert.Equal(t, TierPro, EvaluateTier(0.65, 20, 10))
}

func TestEvaluateTier_ProExactlyAtThreshold(t *testing.T) {
	assert.Equal(t, TierPro, EvaluateTier(0.60, 10, 10), "60%% at exactly 10 matches = pro")
}

func TestEvaluateTier_NormalJustBelowProThreshold(t *testing.T) {
	assert.Equal(t, TierNormal, EvaluateTier(0.59, 10, 10))
}

func TestEvaluateTier_NormalExactlyAtLowerThreshold(t *testing.T) {
	assert.Equal(t, TierNormal, EvaluateTier(0.40, 10, 10), "40%% at 10 matches = normal")
}

func TestEvaluateTier_NoobJustBelowNormalThreshold(t *testing.T) {
	assert.Equal(t, TierNoob, EvaluateTier(0.39, 10, 10))
}

func TestEvaluateTier_NoobZeroWinRate(t *testing.T) {
	assert.Equal(t, TierNoob, EvaluateTier(0.00, 10, 10))
}

func TestEvaluateTier_ProPerfectWinRate(t *testing.T) {
	assert.Equal(t, TierPro, EvaluateTier(1.00, 10, 10))
}

func TestEvaluateTier_InsufficientMatches_NineGames(t *testing.T) {
	// 80%% win rate but only 9 matches → still default normal (insufficient sample)
	assert.Equal(t, TierNormal, EvaluateTier(0.80, 9, 10))
}

func TestEvaluateTier_InsufficientMatches_ZeroGames(t *testing.T) {
	assert.Equal(t, TierNormal, EvaluateTier(0.00, 0, 10))
}

func TestEvaluateTier_InsufficientMatches_ZeroWinRateNineGames(t *testing.T) {
	// Would be noob if enough games, but 9 < 10 → normal
	assert.Equal(t, TierNormal, EvaluateTier(0.00, 9, 10))
}

func TestEvaluateTier_NormalMidRange(t *testing.T) {
	assert.Equal(t, TierNormal, EvaluateTier(0.50, 20, 10))
}

// ─── TierService.RecalculateForUsers ─────────────────────────────────────────

// fakeUserStatsRepo records calls for inspection.
type fakeUserStatsRepo struct {
	// winRates controls what GetWinRatesBatch returns for each user ID.
	winRates map[uuid.UUID]model.UserWithStats
	// batchErr, if set, is returned by GetWinRatesBatch.
	batchErr error
	// updateErr, if set, is returned by UpdateTier (for every call).
	updateErr error
	// updatedTiers records each (id → tier) pair passed to UpdateTier.
	updatedTiers map[uuid.UUID]string
	// allIDs is returned by GetAllIDs.
	allIDs []uuid.UUID
}

func newFakeRepo() *fakeUserStatsRepo {
	return &fakeUserStatsRepo{
		winRates:     make(map[uuid.UUID]model.UserWithStats),
		updatedTiers: make(map[uuid.UUID]string),
	}
}

func (f *fakeUserStatsRepo) GetWinRatesBatch(ids []uuid.UUID) (map[uuid.UUID]model.UserWithStats, error) {
	if f.batchErr != nil {
		return nil, f.batchErr
	}
	result := make(map[uuid.UUID]model.UserWithStats)
	for _, id := range ids {
		if row, ok := f.winRates[id]; ok {
			result[id] = row
		}
	}
	return result, nil
}

func (f *fakeUserStatsRepo) UpdateTier(id uuid.UUID, tier string) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.updatedTiers[id] = tier
	return nil
}

func (f *fakeUserStatsRepo) GetAllIDs() ([]uuid.UUID, error) {
	return f.allIDs, nil
}

func (f *fakeUserStatsRepo) setUser(id uuid.UUID, winRate float64, totalMatches int) {
	u := model.UserWithStats{}
	u.ID = id
	u.WinRate = winRate
	u.TotalMatches = totalMatches
	u.WonMatches = int(float64(totalMatches) * winRate)
	f.winRates[id] = u
}

// ─────────────────────────────────────────────────────────────────────────────

func TestRecalculateForUsers_EmptyList(t *testing.T) {
	repo := newFakeRepo()
	svc := NewTierService(repo, nil)

	err := svc.RecalculateForUsers([]uuid.UUID{})

	require.NoError(t, err)
	assert.Empty(t, repo.updatedTiers, "no UpdateTier calls for empty list")
}

func TestRecalculateForUsers_SetsTierPro(t *testing.T) {
	repo := newFakeRepo()
	svc := NewTierService(repo, nil)
	id := uuid.New()
	repo.setUser(id, 0.65, 20)

	require.NoError(t, svc.RecalculateForUsers([]uuid.UUID{id}))
	assert.Equal(t, TierPro, repo.updatedTiers[id])
}

func TestRecalculateForUsers_SetsTierNormal(t *testing.T) {
	repo := newFakeRepo()
	svc := NewTierService(repo, nil)
	id := uuid.New()
	repo.setUser(id, 0.50, 15)

	require.NoError(t, svc.RecalculateForUsers([]uuid.UUID{id}))
	assert.Equal(t, TierNormal, repo.updatedTiers[id])
}

func TestRecalculateForUsers_SetsTierNoob(t *testing.T) {
	repo := newFakeRepo()
	svc := NewTierService(repo, nil)
	id := uuid.New()
	repo.setUser(id, 0.30, 10)

	require.NoError(t, svc.RecalculateForUsers([]uuid.UUID{id}))
	assert.Equal(t, TierNoob, repo.updatedTiers[id])
}

func TestRecalculateForUsers_InsufficientMatchesStaysNormal(t *testing.T) {
	repo := newFakeRepo()
	svc := NewTierService(repo, nil)
	id := uuid.New()
	repo.setUser(id, 0.90, 5) // 90% win rate, but only 5 matches

	require.NoError(t, svc.RecalculateForUsers([]uuid.UUID{id}))
	assert.Equal(t, TierNormal, repo.updatedTiers[id])
}

func TestRecalculateForUsers_MultipleUsers(t *testing.T) {
	repo := newFakeRepo()
	svc := NewTierService(repo, nil)

	proID := uuid.New()
	normalID := uuid.New()
	noobID := uuid.New()
	newbieID := uuid.New()

	repo.setUser(proID, 0.70, 20)
	repo.setUser(normalID, 0.50, 10)
	repo.setUser(noobID, 0.20, 10)
	repo.setUser(newbieID, 0.80, 3) // insufficient matches

	ids := []uuid.UUID{proID, normalID, noobID, newbieID}
	require.NoError(t, svc.RecalculateForUsers(ids))

	assert.Equal(t, TierPro, repo.updatedTiers[proID])
	assert.Equal(t, TierNormal, repo.updatedTiers[normalID])
	assert.Equal(t, TierNoob, repo.updatedTiers[noobID])
	assert.Equal(t, TierNormal, repo.updatedTiers[newbieID])
}

func TestRecalculateForUsers_BatchFetchError_ReturnsError(t *testing.T) {
	repo := newFakeRepo()
	repo.batchErr = errors.New("db connection lost")
	svc := NewTierService(repo, nil)

	err := svc.RecalculateForUsers([]uuid.UUID{uuid.New()})

	assert.Error(t, err)
	assert.Empty(t, repo.updatedTiers, "no UpdateTier calls when batch fetch fails")
}

func TestRecalculateForUsers_UpdateTierError_ContinuesOtherUsers(t *testing.T) {
	// UpdateTier fails, but the function should not abort — it logs and continues.
	// We verify the service doesn't return an error for individual tier update failures.
	repo := newFakeRepo()
	repo.updateErr = errors.New("write conflict")
	svc := NewTierService(repo, nil)

	id1 := uuid.New()
	id2 := uuid.New()
	repo.setUser(id1, 0.70, 20)
	repo.setUser(id2, 0.30, 10)

	// RecalculateForUsers returns nil even when UpdateTier fails for individual users
	err := svc.RecalculateForUsers([]uuid.UUID{id1, id2})
	assert.NoError(t, err)
}

func TestRecalculateForUsers_UserNotInBatchResult_Skipped(t *testing.T) {
	// If a user ID is not in the batch result (e.g. deleted mid-flight), skip it.
	repo := newFakeRepo()
	svc := NewTierService(repo, nil)
	id := uuid.New()
	// intentionally do NOT add id to repo.winRates

	require.NoError(t, svc.RecalculateForUsers([]uuid.UUID{id}))
	assert.Empty(t, repo.updatedTiers, "no update when user absent from batch result")
}

// ─── TierService.RecalculateAllTiers ─────────────────────────────────────────

func TestRecalculateAllTiers_DelegatesToRecalculateForUsers(t *testing.T) {
	repo := newFakeRepo()
	svc := NewTierService(repo, nil)

	id1, id2 := uuid.New(), uuid.New()
	repo.allIDs = []uuid.UUID{id1, id2}
	repo.setUser(id1, 0.65, 12)
	repo.setUser(id2, 0.30, 10)

	require.NoError(t, svc.RecalculateAllTiers())

	assert.Equal(t, TierPro, repo.updatedTiers[id1])
	assert.Equal(t, TierNoob, repo.updatedTiers[id2])
}

func TestRecalculateAllTiers_EmptyDatabase(t *testing.T) {
	repo := newFakeRepo()
	repo.allIDs = []uuid.UUID{} // no users
	svc := NewTierService(repo, nil)

	err := svc.RecalculateAllTiers()

	require.NoError(t, err)
	assert.Empty(t, repo.updatedTiers)
}
