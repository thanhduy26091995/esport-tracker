package service

// computeCancelPenalty returns the wallet deduction for cancelling a pending bet.
// Returns 0 when enabled is false or either numeric argument is ≤ 0.
func computeCancelPenalty(stake, penaltyPercent int, enabled bool) float64 {
	if !enabled || penaltyPercent <= 0 || stake <= 0 {
		return 0
	}
	return float64(stake) * float64(penaltyPercent) / 100.0
}

// computeReducePenalty returns the wallet penalty, excess reduction amount, and allowed
// minimum stake for a stake-reduction operation.
//
// maxPercent=0 means "no reduction limit" — all reductions are free and penalty is 0.
// When newStake >= originalStake (no reduction), all return values are 0.
func computeReducePenalty(originalStake, newStake, maxPercent, penaltyPercent int) (penalty float64, excessReduction, allowedMinStake int) {
	if maxPercent == 0 || newStake >= originalStake {
		return 0, 0, 0
	}
	maxReduction := int(float64(originalStake) * float64(maxPercent) / 100.0)
	allowedMinStake = originalStake - maxReduction
	if newStake >= allowedMinStake {
		return 0, 0, allowedMinStake
	}
	excessReduction = allowedMinStake - newStake
	if penaltyPercent > 0 {
		penalty = float64(excessReduction) * float64(penaltyPercent) / 100.0
	}
	return penalty, excessReduction, allowedMinStake
}
