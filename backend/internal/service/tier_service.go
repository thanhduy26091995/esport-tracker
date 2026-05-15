package service

import (
	"log"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
)

// UserStatsRepo is the minimal repository surface needed by TierService.
// *repository.UserRepository satisfies this interface automatically.
type UserStatsRepo interface {
	GetWinRatesBatch(ids []uuid.UUID) (map[uuid.UUID]model.UserWithStats, error)
	UpdateTier(id uuid.UUID, tier string) error
	GetAllIDs() ([]uuid.UUID, error)
}

const (
	TierPro    = "pro"
	TierNormal = "normal"
	TierNoob   = "noob"

	proThreshold     = 0.60
	normalThreshold  = 0.40
	defaultMinMatches = 5
)

type TierService struct {
	userRepo  UserStatsRepo
	configSvc *ConfigService
}

func NewTierService(userRepo UserStatsRepo, configSvc *ConfigService) *TierService {
	return &TierService{userRepo: userRepo, configSvc: configSvc}
}

// EvaluateTier returns the tier string for a given win rate and match count.
// Players with fewer than minMatches matches remain at the default "normal" tier.
// Draws (point_change=0) are excluded from both numerator and denominator before calling this.
func EvaluateTier(winRate float64, totalMatches int, minMatches int) string {
	if totalMatches < minMatches {
		return TierNormal
	}
	switch {
	case winRate >= proThreshold:
		return TierPro
	case winRate >= normalThreshold:
		return TierNormal
	default:
		return TierNoob
	}
}

// RecalculateForUsers fetches win rates for the given user IDs in one batch query,
// evaluates the tier for each, and persists the result. Errors are logged but not
// returned as fatal — match mutations must not be blocked by a tier update failure.
func (s *TierService) RecalculateForUsers(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	minMatches := defaultMinMatches
	if s.configSvc != nil {
		if v, err := s.configSvc.GetMinMatchesForTier(); err == nil {
			minMatches = v
		}
	}

	stats, err := s.userRepo.GetWinRatesBatch(ids)
	if err != nil {
		log.Printf("tier: failed to fetch win rates for batch %v: %v", ids, err)
		return err
	}

	for _, id := range ids {
		row, ok := stats[id]
		if !ok {
			continue
		}
		tier := EvaluateTier(row.WinRate, row.TotalMatches, minMatches)
		if updateErr := s.userRepo.UpdateTier(id, tier); updateErr != nil {
			log.Printf("tier: failed to update tier for user %s: %v", id, updateErr)
		}
	}
	return nil
}

// RecalculateAllTiers recalculates and persists tier for every active user.
// Called once at startup to backfill tiers from existing match history.
func (s *TierService) RecalculateAllTiers() error {
	ids, err := s.userRepo.GetAllIDs()
	if err != nil {
		return err
	}
	return s.RecalculateForUsers(ids)
}
