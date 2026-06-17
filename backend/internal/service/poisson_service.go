package service

import (
	"math"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
)

// PoissonInput defines parameters for generating exact-score odds via Poisson model.
type PoissonInput struct {
	MatchID     uuid.UUID
	HomeLambda  float64
	AwayLambda  float64
	HouseMargin float64
	MinProb     float64
}

// PoissonScoreline is one entry in the Poisson output.
type PoissonScoreline struct {
	HomeScore   int     `json:"home_score"`
	AwayScore   int     `json:"away_score"`
	Probability float64 `json:"probability"`
	Odds        float64 `json:"odds"`
}

type PoissonService struct{}

func NewPoissonService() *PoissonService { return &PoissonService{} }

// PoissonProb returns P(X=k) for a Poisson distribution with mean lambda.
func (p *PoissonService) PoissonProb(lambda float64, k int) float64 {
	if lambda <= 0 || k < 0 {
		return 0
	}
	return math.Exp(-lambda) * math.Pow(lambda, float64(k)) / factorial(k)
}

// GenerateScoreOdds computes fair odds for all scorelines with probability >= MinProb,
// then applies the house margin (vig).
func (p *PoissonService) GenerateScoreOdds(input PoissonInput) ([]PoissonScoreline, []model.WcScoreOdds) {
	type candidate struct {
		h, a int
		prob float64
	}
	var candidates []candidate
	probSum := 0.0

	for h := 0; h <= 10; h++ {
		for a := 0; a <= 10; a++ {
			prob := p.PoissonProb(input.HomeLambda, h) * p.PoissonProb(input.AwayLambda, a)
			if prob >= input.MinProb {
				candidates = append(candidates, candidate{h, a, prob})
				probSum += prob
			}
		}
	}
	if len(candidates) == 0 || probSum == 0 {
		return nil, nil
	}

	margin := 1.0 - input.HouseMargin
	lines := make([]PoissonScoreline, 0, len(candidates))
	odds := make([]model.WcScoreOdds, 0, len(candidates))

	for _, c := range candidates {
		// Fair probability (relative to included pool) + house margin applied
		fairProb := c.prob / probSum
		vigProb := fairProb * margin
		var oddsVal float64
		if vigProb > 0 {
			oddsVal = math.Round((1.0/vigProb)*100) / 100
		}
		lines = append(lines, PoissonScoreline{
			HomeScore:   c.h,
			AwayScore:   c.a,
			Probability: math.Round(fairProb*10000) / 10000,
			Odds:        oddsVal,
		})
		odds = append(odds, model.WcScoreOdds{
			ID:        uuid.New(),
			MatchID:   input.MatchID,
			HomeScore: c.h,
			AwayScore: c.a,
			Odds:      oddsVal,
		})
	}
	return lines, odds
}

func factorial(n int) float64 {
	if n <= 1 {
		return 1
	}
	r := 1.0
	for i := 2; i <= n; i++ {
		r *= float64(i)
	}
	return r
}
