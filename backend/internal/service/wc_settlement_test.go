package service

import (
	"testing"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// helpers to build minimal WcBet for tests
func handicapBet(choice string, stake int, odds float64, handicap float64, handicapTeam string) *model.WcBet {
	c := choice
	h := handicap
	ht := handicapTeam
	return &model.WcBet{
		ID:                   uuid.New(),
		WcUserID:             uuid.New(),
		BetType:              model.WcBetTypeHandicap,
		BetChoice:            &c,
		Stake:                stake,
		OddsSnapshot:         odds,
		HandicapSnapshot:     &h,
		HandicapTeamSnapshot: &ht,
	}
}

func exactScoreBet(homeGuess, awayGuess, stake int, odds float64) *model.WcBet {
	return &model.WcBet{
		ID:                 uuid.New(),
		WcUserID:           uuid.New(),
		BetType:            model.WcBetTypeExactScore,
		Stake:              stake,
		OddsSnapshot:       odds,
		PredictedHomeScore: &homeGuess,
		PredictedAwayScore: &awayGuess,
	}
}

// ─── evaluateHandicapBet ──────────────────────────────────────────────────────

func TestHandicap_HomeGivesHalfBall_HomeWins(t *testing.T) {
	// Home gives 0.5 → adjusted_home = 2 - 0.5 = 1.5 > 1 → home wins handicap
	bet := handicapBet("home", 100, 1.90, 0.5, "home")
	result, payout := evaluateHandicapBet(bet, 2, 1)
	assert.Equal(t, model.WcResultWin, result)
	assert.InDelta(t, 190.0, payout, 0.001)
}

func TestHandicap_HomeGivesHalfBall_HomeLoses(t *testing.T) {
	// Home gives 0.5 → adjusted_home = 1 - 0.5 = 0.5 < 1 → away wins handicap
	bet := handicapBet("home", 100, 1.90, 0.5, "home")
	result, payout := evaluateHandicapBet(bet, 1, 1)
	assert.Equal(t, model.WcResultLose, result)
	assert.InDelta(t, 0.0, payout, 0.001)
}

func TestHandicap_AwayGivesHalfBall_AwayBetWins(t *testing.T) {
	// Away gives 0.5 → adjusted_home = 1 + 0.5 = 1.5 > 0 → home wins; bettor picked away → lose
	bet := handicapBet("away", 100, 1.95, 0.5, "away")
	result, payout := evaluateHandicapBet(bet, 1, 0)
	assert.Equal(t, model.WcResultLose, result)
	assert.InDelta(t, 0.0, payout, 0.001)
}

func TestHandicap_AwayGivesHalfBall_AwayBetLoses(t *testing.T) {
	// Away gives 0.5 → adjusted_home = 0 + 0.5 = 0.5 < 1 → away wins; bettor picked away → win
	bet := handicapBet("away", 100, 1.95, 0.5, "away")
	result, payout := evaluateHandicapBet(bet, 0, 1)
	assert.Equal(t, model.WcResultWin, result)
	assert.InDelta(t, 195.0, payout, 0.001)
}

func TestHandicap_WholeNumber_Push(t *testing.T) {
	// Home gives 1.0 → adjusted_home = 2 - 1 = 1 == away 1 → push, stake returned
	bet := handicapBet("home", 200, 1.85, 1.0, "home")
	result, payout := evaluateHandicapBet(bet, 2, 1)
	assert.Equal(t, model.WcResultPush, result)
	assert.InDelta(t, 200.0, payout, 0.001) // stake returned
}

func TestHandicap_WholeNumber_HomeWinsHandicap(t *testing.T) {
	// Home gives 1.0 → adjusted_home = 3 - 1 = 2 > 1 → home wins; bettor picked home
	bet := handicapBet("home", 100, 1.85, 1.0, "home")
	result, payout := evaluateHandicapBet(bet, 3, 1)
	assert.Equal(t, model.WcResultWin, result)
	assert.InDelta(t, 185.0, payout, 0.001)
}

func TestHandicap_HomeGivesOneAndHalf_LargeWin(t *testing.T) {
	// Home gives 1.5 → adjusted_home = 3 - 1.5 = 1.5 > 0 → home wins
	bet := handicapBet("home", 1000, 1.90, 1.5, "home")
	result, payout := evaluateHandicapBet(bet, 3, 0)
	assert.Equal(t, model.WcResultWin, result)
	assert.InDelta(t, 1900.0, payout, 0.001)
}

func TestHandicap_HomeGivesOneAndHalf_NotEnough(t *testing.T) {
	// Home gives 1.5 → adjusted_home = 1 - 1.5 = -0.5 < 0 → away wins; bettor picked home → lose
	bet := handicapBet("home", 100, 1.90, 1.5, "home")
	result, payout := evaluateHandicapBet(bet, 1, 0)
	assert.Equal(t, model.WcResultLose, result)
	assert.InDelta(t, 0.0, payout, 0.001)
}

func TestHandicap_ZeroZeroScoreAwayGivesHalf_HomeWins(t *testing.T) {
	// Away gives 0.5 → adjusted_home = 0 + 0.5 = 0.5 > 0 → home wins
	bet := handicapBet("home", 50, 1.95, 0.5, "away")
	result, payout := evaluateHandicapBet(bet, 0, 0)
	assert.Equal(t, model.WcResultWin, result)
	assert.InDelta(t, 97.5, payout, 0.001) // round(50 * 1.95, 2) = 97.5
}

func TestHandicap_PayoutFloorRounding(t *testing.T) {
	// stake=100, odds=1.879 → round(187.9, 2) = 187.9
	bet := handicapBet("home", 100, 1.879, 0.5, "home")
	result, payout := evaluateHandicapBet(bet, 2, 1)
	assert.Equal(t, model.WcResultWin, result)
	assert.InDelta(t, 187.9, payout, 0.001)
}

// ─── Quarter handicap (split bet) ────────────────────────────────────────────

func TestHandicap_QuarterBall_HomeGives125_WinByTwo(t *testing.T) {
	// Home gives 1.25 (split: -1.0 / -1.5). France wins 2-0:
	// -1.0: adjusted=1 > 0 → win; -1.5: adjusted=0.5 > 0 → win → WIN FULL
	bet := handicapBet("home", 3, 2.10, 1.25, "home")
	result, payout := evaluateHandicapBet(bet, 2, 0)
	assert.Equal(t, model.WcResultWin, result)
	assert.InDelta(t, 6.3, payout, 0.001) // round(3 * 2.10, 2) = 6.3
}

func TestHandicap_QuarterBall_HomeGives125_WinByOne(t *testing.T) {
	// Home gives 1.25 (split: -1.0 / -1.5). France wins 1-0:
	// -1.0: adjusted=0 == 0 → push; -1.5: adjusted=-0.5 < 0 → lose → LOSE HALF
	bet := handicapBet("home", 3, 2.10, 1.25, "home")
	result, payout := evaluateHandicapBet(bet, 1, 0)
	assert.Equal(t, model.WcResultLoseHalf, result)
	assert.InDelta(t, 1.5, payout, 0.001) // round(1.5, 2) = 1.5
}

func TestHandicap_QuarterBall_HomeGives125_Draw(t *testing.T) {
	// Home gives 1.25. Draw 0-0: both sub-handicaps lose → LOSE FULL
	bet := handicapBet("home", 3, 2.10, 1.25, "home")
	result, payout := evaluateHandicapBet(bet, 0, 0)
	assert.Equal(t, model.WcResultLose, result)
	assert.InDelta(t, 0.0, payout, 0.001)
}

func TestHandicap_QuarterBall_HomeGives075_WinByOne(t *testing.T) {
	// Home gives 0.75 (split: -0.5 / -1.0). Home wins 1-0:
	// -0.5: adjusted=0.5 > 0 → win; -1.0: adjusted=0 == 0 → push → WIN HALF
	bet := handicapBet("home", 4, 1.90, 0.75, "home")
	result, payout := evaluateHandicapBet(bet, 1, 0)
	assert.Equal(t, model.WcResultWinHalf, result)
	assert.InDelta(t, 5.8, payout, 0.001) // round(2 * 1.90 + 2, 2) = 5.8
}

func TestHandicap_QuarterBall_HomeGives075_Draw(t *testing.T) {
	// Home gives 0.75. Draw 0-0:
	// -0.5: adjusted=-0.5 < 0 → lose; -1.0: adjusted=-1 < 0 → lose → LOSE FULL
	bet := handicapBet("home", 4, 1.90, 0.75, "home")
	result, payout := evaluateHandicapBet(bet, 0, 0)
	assert.Equal(t, model.WcResultLose, result)
	assert.InDelta(t, 0.0, payout, 0.001)
}

func TestHandicap_QuarterBall_AwayGets125_WinByOne(t *testing.T) {
	// Home gives 1.25 (split: -1.0 / -1.5). Bet on AWAY, home wins 1-0:
	// -1.0: adjusted=0 == 0 → push; -1.5: adjusted=-0.5 < 0 → away wins → WIN HALF for away bettor
	bet := handicapBet("away", 3, 1.78, 1.25, "home")
	result, payout := evaluateHandicapBet(bet, 1, 0)
	assert.Equal(t, model.WcResultWinHalf, result)
	assert.InDelta(t, 4.17, payout, 0.01) // round(1.5 * 1.78 + 1.5, 2) = 4.17
}

func TestHandicap_QuarterBall_EvenStake(t *testing.T) {
	// Even stake=2, home gives 1.25, wins 1-0 (lose half)
	bet := handicapBet("home", 2, 2.10, 1.25, "home")
	result, payout := evaluateHandicapBet(bet, 1, 0)
	assert.Equal(t, model.WcResultLoseHalf, result)
	assert.InDelta(t, 1.0, payout, 0.001) // round(1.0, 2) = 1.0
}

// ─── evaluateOverUnderBet ─────────────────────────────────────────────────────

func ouBet(choice string, stake int, odds float64, line float64) *model.WcBet {
	c := choice
	l := line
	return &model.WcBet{
		ID:               uuid.New(),
		WcUserID:         uuid.New(),
		BetType:          model.WcBetTypeOverUnder,
		BetChoice:        &c,
		Stake:            stake,
		OddsSnapshot:     odds,
		HandicapSnapshot: &l,
	}
}

func TestOverUnder_QuarterLine275_Total3_OverWinsHalf(t *testing.T) {
	// Line 2.75 (split: 2.5 / 3.0), total = 3: over wins 2.5, pushes 3.0 → WIN HALF
	bet := ouBet("over", 4, 1.90, 2.75)
	result, payout := evaluateOverUnderBet(bet, 2, 1)
	assert.Equal(t, model.WcResultWinHalf, result)
	assert.InDelta(t, 5.8, payout, 0.001) // round(2*1.90 + 2, 2) = 5.8
}

func TestOverUnder_QuarterLine275_Total3_UnderLosesHalf(t *testing.T) {
	// Line 2.75 (split: 2.5 / 3.0), total = 3: under loses 2.5, pushes 3.0 → LOSE HALF
	bet := ouBet("under", 20, 1.91, 2.75)
	result, payout := evaluateOverUnderBet(bet, 3, 0)
	assert.Equal(t, model.WcResultLoseHalf, result)
	assert.InDelta(t, 10.0, payout, 0.001) // half stake refunded
}

func TestOverUnder_QuarterLine275_Total4_OverWinsFull(t *testing.T) {
	// Line 2.75, total = 4: over wins both sub-lines → WIN FULL
	bet := ouBet("over", 3, 2.00, 2.75)
	result, payout := evaluateOverUnderBet(bet, 3, 1)
	assert.Equal(t, model.WcResultWin, result)
	assert.InDelta(t, 6.0, payout, 0.001)
}

func TestOverUnder_QuarterLine275_Total2_UnderWinsFull(t *testing.T) {
	// Line 2.75, total = 2: under wins both sub-lines → WIN FULL
	bet := ouBet("under", 3, 1.85, 2.75)
	result, payout := evaluateOverUnderBet(bet, 1, 1)
	assert.Equal(t, model.WcResultWin, result)
	assert.InDelta(t, 5.55, payout, 0.001)
}

func TestOverUnder_QuarterLine325_Total3_OverLosesHalf(t *testing.T) {
	// Line 3.25 (split: 3.0 / 3.5), total = 3: over pushes 3.0, loses 3.5 → LOSE HALF
	bet := ouBet("over", 4, 1.90, 3.25)
	result, payout := evaluateOverUnderBet(bet, 2, 1)
	assert.Equal(t, model.WcResultLoseHalf, result)
	assert.InDelta(t, 2.0, payout, 0.001)
}

func TestOverUnder_QuarterLine325_Total3_UnderWinsHalf(t *testing.T) {
	// Line 3.25 (split: 3.0 / 3.5), total = 3: under pushes 3.0, wins 3.5 → WIN HALF
	bet := ouBet("under", 4, 1.90, 3.25)
	result, payout := evaluateOverUnderBet(bet, 2, 1)
	assert.Equal(t, model.WcResultWinHalf, result)
	assert.InDelta(t, 5.8, payout, 0.001) // round(2*1.90 + 2, 2) = 5.8
}

func TestOverUnder_FullLine3_Total3_Push(t *testing.T) {
	// Whole line 3.0, total = 3 → PUSH, stake returned (regression)
	bet := ouBet("over", 5, 1.95, 3.0)
	result, payout := evaluateOverUnderBet(bet, 2, 1)
	assert.Equal(t, model.WcResultPush, result)
	assert.InDelta(t, 5.0, payout, 0.001)
}

func TestOverUnder_HalfLine25_Total3_OverWinsFull(t *testing.T) {
	// Half line 2.5, total = 3 → over WIN FULL (regression)
	bet := ouBet("over", 2, 1.97, 2.5)
	result, payout := evaluateOverUnderBet(bet, 2, 1)
	assert.Equal(t, model.WcResultWin, result)
	assert.InDelta(t, 3.94, payout, 0.001)
}

// ─── evaluateOverUnderPrediction ──────────────────────────────────────────────

func ouPrediction(choice string, points int, mult float64, line float64) *model.WcPrediction {
	c := choice
	l := line
	return &model.WcPrediction{
		ID:                 uuid.New(),
		WcUserID:           uuid.New(),
		PredictionType:     model.WcPredictionTypeOverUnder,
		PredictionChoice:   &c,
		Points:             points,
		MultiplierSnapshot: mult,
		HandicapSnapshot:   &l,
	}
}

func TestOUPrediction_QuarterLine275_Total3_OverWinHalf(t *testing.T) {
	// Production bug: Spain 3-0 Austria, line 2.75, over → must be WIN HALF, not full win
	bet := ouPrediction("over", 4, 1.97, 2.75)
	result, earned := evaluateOverUnderPrediction(bet, 3, 0)
	assert.Equal(t, model.WcResultWinHalf, result)
	assert.InDelta(t, 5.94, earned, 0.001) // round(2*1.97 + 2, 2) = 5.94
}

func TestOUPrediction_QuarterLine275_Total3_UnderLoseHalf(t *testing.T) {
	// Production bug: Spain 3-0 Austria, line 2.75, under → must be LOSE HALF (half refund), not full loss
	bet := ouPrediction("under", 20, 1.91, 2.75)
	result, earned := evaluateOverUnderPrediction(bet, 3, 0)
	assert.Equal(t, model.WcResultLoseHalf, result)
	assert.InDelta(t, 10.0, earned, 0.001)
}

func TestOUPrediction_QuarterLine325_Total3_UnderWinHalf(t *testing.T) {
	// Line 3.25 (split: 3.0 / 3.5), total = 3: under pushes 3.0, wins 3.5 → WIN HALF
	bet := ouPrediction("under", 5, 1.90, 3.25)
	result, earned := evaluateOverUnderPrediction(bet, 2, 1)
	assert.Equal(t, model.WcResultWinHalf, result)
	assert.InDelta(t, 7.25, earned, 0.001) // round(2.5*1.90 + 2.5, 2) = 7.25
}

func TestOUPrediction_QuarterLine275_Total4_OverCorrectFull(t *testing.T) {
	// Line 2.75, total = 4 → over CORRECT full (regression)
	bet := ouPrediction("over", 3, 2.00, 2.75)
	result, earned := evaluateOverUnderPrediction(bet, 3, 1)
	assert.Equal(t, model.WcResultCorrect, result)
	assert.InDelta(t, 6.0, earned, 0.001)
}

func TestOUPrediction_FullLine3_Total3_Void(t *testing.T) {
	// Whole line 3.0, total = 3 → VOID, points returned (regression)
	bet := ouPrediction("over", 5, 1.95, 3.0)
	result, earned := evaluateOverUnderPrediction(bet, 2, 1)
	assert.Equal(t, model.WcResultVoid, result)
	assert.InDelta(t, 5.0, earned, 0.001)
}

// ─── evaluateExactScoreBet ────────────────────────────────────────────────────

func TestExactScore_CorrectPrediction(t *testing.T) {
	bet := exactScoreBet(2, 1, 100, 6.00)
	result, payout := evaluateExactScoreBet(bet, 2, 1)
	assert.Equal(t, model.WcResultWin, result)
	assert.InDelta(t, 600.0, payout, 0.001)
}

func TestExactScore_WrongHomeScore(t *testing.T) {
	bet := exactScoreBet(2, 1, 100, 6.00)
	result, payout := evaluateExactScoreBet(bet, 3, 1)
	assert.Equal(t, model.WcResultLose, result)
	assert.InDelta(t, 0.0, payout, 0.001)
}

func TestExactScore_WrongAwayScore(t *testing.T) {
	bet := exactScoreBet(1, 0, 100, 5.00)
	result, payout := evaluateExactScoreBet(bet, 1, 1)
	assert.Equal(t, model.WcResultLose, result)
	assert.InDelta(t, 0.0, payout, 0.001)
}

func TestExactScore_ZeroZeroCorrect(t *testing.T) {
	bet := exactScoreBet(0, 0, 200, 3.50)
	result, payout := evaluateExactScoreBet(bet, 0, 0)
	assert.Equal(t, model.WcResultWin, result)
	assert.InDelta(t, 700.0, payout, 0.001)
}

func TestExactScore_ZeroZeroWrong(t *testing.T) {
	bet := exactScoreBet(0, 0, 100, 3.50)
	result, payout := evaluateExactScoreBet(bet, 1, 0)
	assert.Equal(t, model.WcResultLose, result)
	assert.InDelta(t, 0.0, payout, 0.001)
}

func TestExactScore_HighOdds(t *testing.T) {
	// stake=33, odds=7.00 → round(231, 2) = 231.0
	bet := exactScoreBet(0, 3, 33, 7.00)
	result, payout := evaluateExactScoreBet(bet, 0, 3)
	assert.Equal(t, model.WcResultWin, result)
	assert.InDelta(t, 231.0, payout, 0.001)
}

func TestExactScore_OffByOne_HomeHigher(t *testing.T) {
	bet := exactScoreBet(2, 0, 100, 8.00)
	result, payout := evaluateExactScoreBet(bet, 3, 0)
	assert.Equal(t, model.WcResultLose, result)
	assert.InDelta(t, 0.0, payout, 0.001)
}

func TestExactScore_NoPush(t *testing.T) {
	// Exact score bet has no push — only win or lose
	bet := exactScoreBet(1, 1, 100, 4.00)
	result, _ := evaluateExactScoreBet(bet, 2, 1)
	assert.NotEqual(t, model.WcResultPush, result)
}
