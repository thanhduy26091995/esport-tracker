package service

import (
	"math"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── PoissonProb ──────────────────────────────────────────────────────────────

func TestPoissonProb_ZeroGoals(t *testing.T) {
	// P(X=0 | λ) = e^(-λ)
	svc := NewPoissonService()
	lambda := 1.5
	got := svc.PoissonProb(lambda, 0)
	assert.InDelta(t, math.Exp(-lambda), got, 1e-9)
}

func TestPoissonProb_OneGoal(t *testing.T) {
	// P(X=1 | λ) = λ * e^(-λ)
	svc := NewPoissonService()
	lambda := 2.0
	got := svc.PoissonProb(lambda, 1)
	assert.InDelta(t, lambda*math.Exp(-lambda), got, 1e-9)
}

func TestPoissonProb_NegativeLambda_ReturnsZero(t *testing.T) {
	svc := NewPoissonService()
	assert.Equal(t, 0.0, svc.PoissonProb(-1.0, 2))
}

func TestPoissonProb_ZeroLambda_ReturnsZero(t *testing.T) {
	svc := NewPoissonService()
	assert.Equal(t, 0.0, svc.PoissonProb(0.0, 0))
}

func TestPoissonProb_NegativeK_ReturnsZero(t *testing.T) {
	svc := NewPoissonService()
	assert.Equal(t, 0.0, svc.PoissonProb(1.5, -1))
}

func TestPoissonProb_SumApproachesOne(t *testing.T) {
	// Sum of P(X=k) for k=0..20 should be very close to 1 for moderate λ.
	svc := NewPoissonService()
	lambda := 1.8
	sum := 0.0
	for k := 0; k <= 20; k++ {
		sum += svc.PoissonProb(lambda, k)
	}
	assert.InDelta(t, 1.0, sum, 0.001)
}

// ─── GenerateScoreOdds ────────────────────────────────────────────────────────

func TestGenerateScoreOdds_ReturnsNonEmptyForTypicalLambdas(t *testing.T) {
	svc := NewPoissonService()
	input := PoissonInput{
		MatchID:     uuid.New(),
		HomeLambda:  1.5,
		AwayLambda:  1.2,
		HouseMargin: 0.05,
		MinProb:     0.001,
	}
	lines, odds := svc.GenerateScoreOdds(input)
	require.NotEmpty(t, lines, "expected scorelines for typical lambdas")
	assert.Equal(t, len(lines), len(odds), "lines and odds slices must match in length")
}

func TestGenerateScoreOdds_AllOddsPositive(t *testing.T) {
	svc := NewPoissonService()
	input := PoissonInput{
		MatchID:     uuid.New(),
		HomeLambda:  1.5,
		AwayLambda:  1.2,
		HouseMargin: 0.05,
		MinProb:     0.001,
	}
	_, odds := svc.GenerateScoreOdds(input)
	for _, o := range odds {
		assert.Greater(t, o.Odds, 0.0, "odds must be positive for %d-%d", o.HomeScore, o.AwayScore)
	}
}

func TestGenerateScoreOdds_MatchIDPropagated(t *testing.T) {
	svc := NewPoissonService()
	matchID := uuid.New()
	input := PoissonInput{
		MatchID:     matchID,
		HomeLambda:  1.5,
		AwayLambda:  1.2,
		HouseMargin: 0.05,
		MinProb:     0.001,
	}
	_, odds := svc.GenerateScoreOdds(input)
	require.NotEmpty(t, odds)
	for _, o := range odds {
		assert.Equal(t, matchID, o.MatchID, "all odds must carry the correct MatchID")
	}
}

func TestGenerateScoreOdds_HighMinProb_FewerScorelines(t *testing.T) {
	svc := NewPoissonService()
	base := PoissonInput{MatchID: uuid.New(), HomeLambda: 1.5, AwayLambda: 1.2, HouseMargin: 0.05}

	base.MinProb = 0.001
	linesLow, _ := svc.GenerateScoreOdds(base)

	base.MinProb = 0.05
	linesHigh, _ := svc.GenerateScoreOdds(base)

	assert.Greater(t, len(linesLow), len(linesHigh), "stricter MinProb must produce fewer scorelines")
}

func TestGenerateScoreOdds_HigherMargin_HigherOdds(t *testing.T) {
	// odds = 1 / (fairProb * (1 - HouseMargin)), so a larger HouseMargin
	// reduces the denominator and raises the posted odds.
	svc := NewPoissonService()
	input := PoissonInput{MatchID: uuid.New(), HomeLambda: 1.5, AwayLambda: 1.2, MinProb: 0.01}

	input.HouseMargin = 0.02
	_, oddsSmallMargin := svc.GenerateScoreOdds(input)

	input.HouseMargin = 0.10
	_, oddsLargeMargin := svc.GenerateScoreOdds(input)

	require.NotEmpty(t, oddsSmallMargin)
	require.NotEmpty(t, oddsLargeMargin)
	assert.Greater(t, oddsLargeMargin[0].Odds, oddsSmallMargin[0].Odds, "larger HouseMargin must produce higher posted odds")
}

func TestGenerateScoreOdds_ZeroZeroIncluded(t *testing.T) {
	// 0-0 is always a likely scoreline for balanced lambdas.
	svc := NewPoissonService()
	_, odds := svc.GenerateScoreOdds(PoissonInput{
		MatchID:     uuid.New(),
		HomeLambda:  1.2,
		AwayLambda:  1.2,
		HouseMargin: 0.05,
		MinProb:     0.001,
	})
	found := false
	for _, o := range odds {
		if o.HomeScore == 0 && o.AwayScore == 0 {
			found = true
			break
		}
	}
	assert.True(t, found, "0-0 scoreline must be included for typical lambdas")
}

func TestGenerateScoreOdds_ImpossibleMinProb_ReturnsNil(t *testing.T) {
	// MinProb > any single P(h,a) → nothing qualifies.
	svc := NewPoissonService()
	lines, odds := svc.GenerateScoreOdds(PoissonInput{
		MatchID:     uuid.New(),
		HomeLambda:  1.5,
		AwayLambda:  1.2,
		HouseMargin: 0.05,
		MinProb:     1.0, // no scoreline can have prob ≥ 1
	})
	assert.Nil(t, lines)
	assert.Nil(t, odds)
}

func TestGenerateScoreOdds_ProbabilityRoundedToFourDecimals(t *testing.T) {
	svc := NewPoissonService()
	lines, _ := svc.GenerateScoreOdds(PoissonInput{
		MatchID:     uuid.New(),
		HomeLambda:  1.5,
		AwayLambda:  1.2,
		HouseMargin: 0.05,
		MinProb:     0.001,
	})
	require.NotEmpty(t, lines)
	for _, l := range lines {
		rounded := math.Round(l.Probability*10000) / 10000
		assert.InDelta(t, rounded, l.Probability, 1e-9, "probability must be rounded to 4 decimals")
	}
}
