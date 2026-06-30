package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func intPtr(v int) *int { return &v }

// ─── selectBettingScore ───────────────────────────────────────────────────────

func TestSelectBettingScore_Regular_UsesFullTime(t *testing.T) {
	home, away := selectBettingScore("REGULAR", intPtr(2), intPtr(1), nil, nil)
	assert.Equal(t, 2, *home)
	assert.Equal(t, 1, *away)
}

func TestSelectBettingScore_ExtraTime_UsesRegularTime(t *testing.T) {
	// fullTime=3-1 (after ET), regularTime=1-1 (90 min) → must use regularTime
	home, away := selectBettingScore("EXTRA_TIME", intPtr(3), intPtr(1), intPtr(1), intPtr(1))
	assert.Equal(t, 1, *home, "should use 90-min score, not ET result")
	assert.Equal(t, 1, *away)
}

func TestSelectBettingScore_PenaltyShootout_UsesRegularTime(t *testing.T) {
	// fullTime=2-2 (after ET), regularTime=2-2, penalties=5-4
	home, away := selectBettingScore("PENALTY_SHOOTOUT", intPtr(2), intPtr(2), intPtr(2), intPtr(2))
	assert.Equal(t, 2, *home)
	assert.Equal(t, 2, *away)
}

func TestSelectBettingScore_ExtraTime_MissingRegularTime_FallsBackToFullTime(t *testing.T) {
	// API bug or old data: duration=EXTRA_TIME but no regularTime field → fall back to fullTime
	home, away := selectBettingScore("EXTRA_TIME", intPtr(3), intPtr(2), nil, nil)
	assert.Equal(t, 3, *home, "fallback to fullTime when regularTime absent")
	assert.Equal(t, 2, *away)
}

func TestSelectBettingScore_ExtraTime_OnlyOneRegularField_FallsBackToFullTime(t *testing.T) {
	// Only rtHome present (malformed response) → fall back to fullTime
	home, away := selectBettingScore("EXTRA_TIME", intPtr(3), intPtr(2), intPtr(1), nil)
	assert.Equal(t, 3, *home, "partial regularTime must not be used")
	assert.Equal(t, 2, *away)
}

func TestSelectBettingScore_EmptyDuration_UsesFullTime(t *testing.T) {
	home, away := selectBettingScore("", intPtr(1), intPtr(0), intPtr(0), intPtr(0))
	assert.Equal(t, 1, *home)
	assert.Equal(t, 0, *away)
}

func TestSelectBettingScore_NoScoreYet_ReturnsNil(t *testing.T) {
	// Match not finished — all fields nil
	home, away := selectBettingScore("REGULAR", nil, nil, nil, nil)
	assert.Nil(t, home)
	assert.Nil(t, away)
}
