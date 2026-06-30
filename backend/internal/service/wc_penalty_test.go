package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeCancelPenalty(t *testing.T) {
	tests := []struct {
		name           string
		stake          int
		penaltyPercent int
		enabled        bool
		want           int
	}{
		{"feature disabled", 100, 20, false, 0},
		{"zero percent", 100, 0, true, 0},
		{"zero stake", 0, 20, true, 0},
		{"20% of 100", 100, 20, true, 20},
		{"20% of 10", 10, 20, true, 2},
		{"floor: 20% of 7 = 1.4 → 1", 7, 20, true, 1},
		{"floor: 20% of 4 = 0.8 → 0", 4, 20, true, 0},
		{"floor: 50% of 1 = 0.5 → 0", 1, 50, true, 0},
		{"100% of 100", 100, 100, true, 100},
		{"10% of 100", 100, 10, true, 10},
		{"5% of 100", 100, 5, true, 5},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := computeCancelPenalty(tc.stake, tc.penaltyPercent, tc.enabled)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestComputeReducePenalty(t *testing.T) {
	tests := []struct {
		name           string
		original       int
		newStake       int
		maxPercent     int
		penaltyPercent int
		wantPenalty    int
		wantExcess     int
		wantAllowedMin int
	}{
		{
			// maxPercent=0 means "no reduction limit" — any reduction is free
			name: "maxPercent=0 means no limit",
			original: 100, newStake: 10, maxPercent: 0, penaltyPercent: 20,
			wantPenalty: 0, wantExcess: 0, wantAllowedMin: 0,
		},
		{
			name: "increase stake — no penalty",
			original: 100, newStake: 120, maxPercent: 50, penaltyPercent: 20,
			wantPenalty: 0, wantExcess: 0, wantAllowedMin: 0,
		},
		{
			name: "same stake — no penalty",
			original: 100, newStake: 100, maxPercent: 50, penaltyPercent: 20,
			wantPenalty: 0, wantExcess: 0, wantAllowedMin: 0,
		},
		{
			// allowedMin = 100 - floor(100*50/100) = 50; 60 >= 50 → free
			name: "reduce within 50% allowed — no penalty",
			original: 100, newStake: 60, maxPercent: 50, penaltyPercent: 20,
			wantPenalty: 0, wantExcess: 0, wantAllowedMin: 50,
		},
		{
			name: "reduce exactly to allowed min — no penalty",
			original: 100, newStake: 50, maxPercent: 50, penaltyPercent: 20,
			wantPenalty: 0, wantExcess: 0, wantAllowedMin: 50,
		},
		{
			// excess = 50-40 = 10; penalty = floor(10*20/100) = 2
			name: "reduce below allowed min — excess 10, penalty 2",
			original: 100, newStake: 40, maxPercent: 50, penaltyPercent: 20,
			wantPenalty: 2, wantExcess: 10, wantAllowedMin: 50,
		},
		{
			// excess = 50-49 = 1; penalty = floor(1*20/100) = floor(0.2) = 0
			name: "excess too small to yield penalty (floor rounds to 0)",
			original: 100, newStake: 49, maxPercent: 50, penaltyPercent: 20,
			wantPenalty: 0, wantExcess: 1, wantAllowedMin: 50,
		},
		{
			// allowedMin = 100-30=70; excess = 70-60=10; penalty = floor(10*10/100) = 1
			name: "30% max, 10% penalty: excess 10, penalty 1",
			original: 100, newStake: 60, maxPercent: 30, penaltyPercent: 10,
			wantPenalty: 1, wantExcess: 10, wantAllowedMin: 70,
		},
		{
			// allowedMin = 10-5=5; newStake=5 at allowedMin → free
			name: "small stake: reduce exactly to allowedMin",
			original: 10, newStake: 5, maxPercent: 50, penaltyPercent: 20,
			wantPenalty: 0, wantExcess: 0, wantAllowedMin: 5,
		},
		{
			// excess = 5-4 = 1; penalty = floor(1*20/100) = 0
			name: "small stake: 1-pt excess, penalty still 0 due to floor",
			original: 10, newStake: 4, maxPercent: 50, penaltyPercent: 20,
			wantPenalty: 0, wantExcess: 1, wantAllowedMin: 5,
		},
		{
			// allowedMin = 100-50=50; excess = 50-10=40; penalty = floor(40*50/100) = 20
			name: "deep reduction: 50% penalty percent",
			original: 100, newStake: 10, maxPercent: 50, penaltyPercent: 50,
			wantPenalty: 20, wantExcess: 40, wantAllowedMin: 50,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			penalty, excess, allowedMin := computeReducePenalty(tc.original, tc.newStake, tc.maxPercent, tc.penaltyPercent)
			assert.Equal(t, tc.wantPenalty, penalty, "penalty")
			assert.Equal(t, tc.wantExcess, excess, "excessReduction")
			assert.Equal(t, tc.wantAllowedMin, allowedMin, "allowedMinStake")
		})
	}
}
