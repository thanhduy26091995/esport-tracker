package service

import (
	"testing"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── Helpers ──────────────────────────────────────────────────────────────────

var (
	uid1 = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	uid2 = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	uid3 = uuid.MustParse("00000000-0000-0000-0000-000000000003")
)

func types(hs []model.Highlight) []string {
	out := make([]string, len(hs))
	for i, h := range hs {
		out[i] = h.Type
	}
	return out
}

func findType(hs []model.Highlight, typ string) *model.Highlight {
	for i := range hs {
		if hs[i].Type == typ {
			return &hs[i]
		}
	}
	return nil
}

// ─── winRateFromForm ──────────────────────────────────────────────────────────

func TestWinRateFromForm_Nil(t *testing.T) {
	wr, n := winRateFromForm(nil)
	assert.Equal(t, 0.0, wr)
	assert.Equal(t, 0, n)
}

func TestWinRateFromForm_EmptyResults(t *testing.T) {
	wr, n := winRateFromForm(&repository.RecentFormData{Results: []bool{}})
	assert.Equal(t, 0.0, wr)
	assert.Equal(t, 0, n)
}

func TestWinRateFromForm_AllWins(t *testing.T) {
	wr, n := winRateFromForm(&repository.RecentFormData{Results: []bool{true, true, true, true, true}})
	assert.InDelta(t, 1.0, wr, 0.001)
	assert.Equal(t, 5, n)
}

func TestWinRateFromForm_AllLosses(t *testing.T) {
	wr, n := winRateFromForm(&repository.RecentFormData{Results: []bool{false, false, false, false, false}})
	assert.InDelta(t, 0.0, wr, 0.001)
	assert.Equal(t, 5, n)
}

func TestWinRateFromForm_Mixed(t *testing.T) {
	// 8 wins, 2 losses → 0.80
	results := []bool{true, true, true, true, true, true, true, true, false, false}
	wr, n := winRateFromForm(&repository.RecentFormData{Results: results})
	assert.InDelta(t, 0.80, wr, 0.001)
	assert.Equal(t, 10, n)
}

func TestWinRateFromForm_FourOfTen(t *testing.T) {
	results := []bool{true, true, true, true, false, false, false, false, false, false}
	wr, n := winRateFromForm(&repository.RecentFormData{Results: results})
	assert.InDelta(t, 0.40, wr, 0.001)
	assert.Equal(t, 10, n)
}

// ─── priorityForStreak ────────────────────────────────────────────────────────

func TestPriorityForStreak_BelowMed(t *testing.T) {
	assert.Equal(t, 60, priorityForStreak(3))
	assert.Equal(t, 60, priorityForStreak(4))
}

func TestPriorityForStreak_AtMed(t *testing.T) {
	assert.Equal(t, 80, priorityForStreak(5))
	assert.Equal(t, 80, priorityForStreak(9))
}

func TestPriorityForStreak_AtHigh(t *testing.T) {
	assert.Equal(t, 95, priorityForStreak(10))
	assert.Equal(t, 95, priorityForStreak(15))
}

// ─── streakHighlights ─────────────────────────────────────────────────────────

func TestStreakHighlights_BelowThreshold(t *testing.T) {
	sd := &repository.StreakData{CurrentWin: 2, CurrentLose: 2, CurrentUnbeaten: 2}
	assert.Empty(t, streakHighlights(uid1, "Alice", sd))
}

func TestStreakHighlights_WinAtThreshold(t *testing.T) {
	sd := &repository.StreakData{CurrentWin: 3}
	hs := streakHighlights(uid1, "Alice", sd)
	require.Len(t, hs, 1)
	assert.Equal(t, model.TypeStreakWin, hs[0].Type)
	assert.Equal(t, model.SectionTrending, hs[0].Section)
	assert.Equal(t, 60, hs[0].Priority)
	assert.Equal(t, 3.0, hs[0].Value)
}

func TestStreakHighlights_WinHighPriority(t *testing.T) {
	sd := &repository.StreakData{CurrentWin: 10}
	hs := streakHighlights(uid1, "Alice", sd)
	h := findType(hs, model.TypeStreakWin)
	require.NotNil(t, h)
	assert.Equal(t, 95, h.Priority)
}

func TestStreakHighlights_LoseAtThreshold(t *testing.T) {
	sd := &repository.StreakData{CurrentLose: 3}
	hs := streakHighlights(uid1, "Alice", sd)
	h := findType(hs, model.TypeStreakLose)
	require.NotNil(t, h)
	assert.Equal(t, model.SectionDailyRecap, h.Section)
}

func TestStreakHighlights_UnbeatenOnlyWhenGreaterThanWin(t *testing.T) {
	// CurrentUnbeaten == CurrentWin → no unbeaten highlight
	sd := &repository.StreakData{CurrentWin: 5, CurrentUnbeaten: 5}
	hs := streakHighlights(uid1, "Alice", sd)
	assert.Nil(t, findType(hs, model.TypeStreakUnbeaten))
}

func TestStreakHighlights_UnbeatenWithDraw(t *testing.T) {
	// Unbeaten run includes draws → CurrentUnbeaten > CurrentWin
	sd := &repository.StreakData{CurrentWin: 3, CurrentUnbeaten: 5}
	hs := streakHighlights(uid1, "Alice", sd)
	u := findType(hs, model.TypeStreakUnbeaten)
	require.NotNil(t, u)
	assert.Equal(t, model.SectionTrending, u.Section)
	assert.Equal(t, 5.0, u.Value)
}

// ─── pointsHighlights ─────────────────────────────────────────────────────────

func TestPointsHighlights_ZeroGainedAndLost(t *testing.T) {
	pm := &repository.PointsMovement{}
	assert.Empty(t, pointsHighlights(uid1, "Alice", pm))
}

func TestPointsHighlights_SmallGain(t *testing.T) {
	pm := &repository.PointsMovement{PointsGained: 10}
	hs := pointsHighlights(uid1, "Alice", pm)
	h := findType(hs, model.TypePointsGainedToday)
	require.NotNil(t, h)
	assert.Equal(t, 50, h.Priority)
}

func TestPointsHighlights_MedGain(t *testing.T) {
	pm := &repository.PointsMovement{PointsGained: 50}
	hs := pointsHighlights(uid1, "Alice", pm)
	h := findType(hs, model.TypePointsGainedToday)
	require.NotNil(t, h)
	assert.Equal(t, 65, h.Priority)
}

func TestPointsHighlights_HighGain(t *testing.T) {
	pm := &repository.PointsMovement{PointsGained: 100}
	hs := pointsHighlights(uid1, "Alice", pm)
	h := findType(hs, model.TypePointsGainedToday)
	require.NotNil(t, h)
	assert.Equal(t, 80, h.Priority)
}

func TestPointsHighlights_Lost(t *testing.T) {
	pm := &repository.PointsMovement{PointsLost: 30}
	hs := pointsHighlights(uid1, "Alice", pm)
	h := findType(hs, model.TypePointsLostToday)
	require.NotNil(t, h)
	assert.Equal(t, 45, h.Priority)
	assert.Equal(t, 30.0, h.Value)
}

// ─── rankHighlights ───────────────────────────────────────────────────────────

func TestRankHighlights_NoChange(t *testing.T) {
	rd := &repository.RankData{CurrentRank: 5, YesterdayRank: 5}
	assert.Empty(t, rankHighlights(uid1, "Alice", rd))
}

func TestRankHighlights_ClimbedOne(t *testing.T) {
	rd := &repository.RankData{CurrentRank: 5, YesterdayRank: 6}
	hs := rankHighlights(uid1, "Alice", rd)
	h := findType(hs, model.TypeRankClimbed)
	require.NotNil(t, h)
	assert.Equal(t, 65, h.Priority)
	assert.Equal(t, 1.0, h.Value)
}

func TestRankHighlights_ClimbedToTop3(t *testing.T) {
	rd := &repository.RankData{CurrentRank: 2, YesterdayRank: 5}
	hs := rankHighlights(uid1, "Alice", rd)
	h := findType(hs, model.TypeRankClimbed)
	require.NotNil(t, h)
	assert.Equal(t, 90, h.Priority)
}

func TestRankHighlights_ClimbedThreePlaces(t *testing.T) {
	rd := &repository.RankData{CurrentRank: 6, YesterdayRank: 9}
	hs := rankHighlights(uid1, "Alice", rd)
	h := findType(hs, model.TypeRankClimbed)
	require.NotNil(t, h)
	assert.Equal(t, 75, h.Priority)
}

func TestRankHighlights_EnteredTop10(t *testing.T) {
	rd := &repository.RankData{CurrentRank: 10, YesterdayRank: 11}
	hs := rankHighlights(uid1, "Alice", rd)
	top10 := findType(hs, model.TypeMilestoneTop10)
	require.NotNil(t, top10)
	assert.Equal(t, model.SectionCompetitive, top10.Section)
	assert.Equal(t, 90, top10.Priority)
}

func TestRankHighlights_AlreadyInTop10_NoMilestone(t *testing.T) {
	rd := &repository.RankData{CurrentRank: 8, YesterdayRank: 9}
	hs := rankHighlights(uid1, "Alice", rd)
	assert.Nil(t, findType(hs, model.TypeMilestoneTop10))
}

func TestRankHighlights_Dropped(t *testing.T) {
	rd := &repository.RankData{CurrentRank: 8, YesterdayRank: 5}
	hs := rankHighlights(uid1, "Alice", rd)
	h := findType(hs, model.TypeRankDropped)
	require.NotNil(t, h)
	assert.Equal(t, model.SectionCompetitive, h.Section)
	assert.Equal(t, 3.0, h.Value)
}

// ─── fastestClimberToday / biggestCollapseToday ───────────────────────────────

func TestFastestClimberToday_Empty(t *testing.T) {
	assert.Nil(t, fastestClimberToday(nil, nil))
}

func TestFastestClimberToday_AllZero(t *testing.T) {
	today := map[uuid.UUID]*repository.PointsMovement{
		uid1: {PointsGained: 0},
	}
	assert.Nil(t, fastestClimberToday(today, map[uuid.UUID]string{uid1: "Alice"}))
}

func TestFastestClimberToday_PicksBest(t *testing.T) {
	today := map[uuid.UUID]*repository.PointsMovement{
		uid1: {PointsGained: 30},
		uid2: {PointsGained: 80},
		uid3: {PointsGained: 10},
	}
	names := map[uuid.UUID]string{uid1: "Alice", uid2: "Bob", uid3: "Charlie"}
	hs := fastestClimberToday(today, names)
	require.Len(t, hs, 1)
	assert.Equal(t, model.TypeFastestClimberToday, hs[0].Type)
	assert.Equal(t, uid2, hs[0].PlayerID)
	assert.Equal(t, 80.0, hs[0].Value)
}

func TestBiggestCollapseToday_Empty(t *testing.T) {
	assert.Nil(t, biggestCollapseToday(nil, nil))
}

func TestBiggestCollapseToday_AllZero(t *testing.T) {
	today := map[uuid.UUID]*repository.PointsMovement{uid1: {PointsLost: 0}}
	assert.Nil(t, biggestCollapseToday(today, map[uuid.UUID]string{uid1: "Alice"}))
}

func TestBiggestCollapseToday_PicksBest(t *testing.T) {
	today := map[uuid.UUID]*repository.PointsMovement{
		uid1: {PointsLost: 5},
		uid2: {PointsLost: 60},
	}
	names := map[uuid.UUID]string{uid1: "Alice", uid2: "Bob"}
	hs := biggestCollapseToday(today, names)
	require.Len(t, hs, 1)
	assert.Equal(t, model.TypeBiggestCollapse, hs[0].Type)
	assert.Equal(t, uid2, hs[0].PlayerID)
}

// ─── formHighlights ───────────────────────────────────────────────────────────

func TestFormHighlights_TooFewResults(t *testing.T) {
	fd := &repository.RecentFormData{Results: []bool{true, true, true, true}}
	assert.Empty(t, formHighlights(uid1, "Alice", fd))
}

func TestFormHighlights_Hot(t *testing.T) {
	// 8/10 = 0.80 → form_hot
	results := make([]bool, 10)
	for i := 0; i < 8; i++ {
		results[i] = true
	}
	fd := &repository.RecentFormData{Results: results}
	hs := formHighlights(uid1, "Alice", fd)
	h := findType(hs, model.TypeFormHot)
	require.NotNil(t, h)
	assert.Equal(t, model.SectionTrending, h.Section)
}

func TestFormHighlights_JustBelowHot(t *testing.T) {
	// 7/10 = 0.70 → not form_hot, not form_cold, not form_stable (0.70 > 0.65)
	results := make([]bool, 10)
	for i := 0; i < 7; i++ {
		results[i] = true
	}
	fd := &repository.RecentFormData{Results: results}
	assert.Empty(t, formHighlights(uid1, "Alice", fd))
}

func TestFormHighlights_Cold(t *testing.T) {
	// 4/10 = 0.40 → form_cold (exactly at FormColdMaxWinRate)
	results := []bool{true, true, true, true, false, false, false, false, false, false}
	fd := &repository.RecentFormData{Results: results}
	hs := formHighlights(uid1, "Alice", fd)
	h := findType(hs, model.TypeFormCold)
	require.NotNil(t, h)
	assert.Equal(t, model.SectionDailyRecap, h.Section)
}

func TestFormHighlights_Stable(t *testing.T) {
	// 5/10 = 0.50, count=10 >= 8 → form_stable
	results := []bool{true, true, true, true, true, false, false, false, false, false}
	fd := &repository.RecentFormData{Results: results}
	hs := formHighlights(uid1, "Alice", fd)
	h := findType(hs, model.TypeFormStable)
	require.NotNil(t, h)
	assert.Equal(t, model.SectionSocial, h.Section)
}

func TestFormHighlights_StableRequires8Results(t *testing.T) {
	// 4/7 = 0.57 (stable range) but count < 8 → no stable
	results := []bool{true, true, true, true, false, false, false}
	fd := &repository.RecentFormData{Results: results}
	assert.Empty(t, formHighlights(uid1, "Alice", fd))
}

// ─── activityHighlights ───────────────────────────────────────────────────────

func TestActivityHighlights_BelowThreshold(t *testing.T) {
	wa := &repository.WeeklyActivityData{ConsecutiveActiveDays: 4}
	assert.Empty(t, activityHighlights(uid1, "Alice", wa))
}

func TestActivityHighlights_AtThreshold(t *testing.T) {
	wa := &repository.WeeklyActivityData{ConsecutiveActiveDays: 5}
	hs := activityHighlights(uid1, "Alice", wa)
	h := findType(hs, model.TypeActiveStreakDays)
	require.NotNil(t, h)
	assert.Equal(t, model.SectionSocial, h.Section)
	assert.Equal(t, 5.0, h.Value)
}

// ─── mostActiveToday ──────────────────────────────────────────────────────────

func TestMostActiveToday_BelowThreshold(t *testing.T) {
	today := map[uuid.UUID]*repository.PointsMovement{uid1: {MatchCount: 4}}
	assert.Nil(t, mostActiveToday(today, map[uuid.UUID]string{uid1: "Alice"}))
}

func TestMostActiveToday_AtThreshold(t *testing.T) {
	today := map[uuid.UUID]*repository.PointsMovement{uid1: {MatchCount: 5}}
	names := map[uuid.UUID]string{uid1: "Alice"}
	hs := mostActiveToday(today, names)
	assert.NotNil(t, findType(hs, model.TypeMostActiveToday))
	assert.Nil(t, findType(hs, model.TypeMarathon))
}

func TestMostActiveToday_Marathon(t *testing.T) {
	today := map[uuid.UUID]*repository.PointsMovement{
		uid1: {MatchCount: 5},
		uid2: {MatchCount: 20},
	}
	names := map[uuid.UUID]string{uid1: "Alice", uid2: "Bob"}
	hs := mostActiveToday(today, names)
	require.NotNil(t, findType(hs, model.TypeMostActiveToday))
	require.NotNil(t, findType(hs, model.TypeMarathon))
	assert.Equal(t, uid2, findType(hs, model.TypeMarathon).PlayerID)
}

// ─── hotColdHighlights ────────────────────────────────────────────────────────

func TestHotColdHighlights_TooFewDecisiveMatches(t *testing.T) {
	fd := &repository.RecentFormData{Results: make([]bool, 9)} // 9 < 10
	sd := &repository.StreakData{}
	pm := &repository.PointsMovement{MatchCount: 10}
	assert.Empty(t, hotColdHighlights(uid1, "Alice", sd, fd, pm))
}

func TestHotColdHighlights_Hot(t *testing.T) {
	results := make([]bool, 10)
	for i := 0; i < 9; i++ {
		results[i] = true
	}
	fd := &repository.RecentFormData{Results: results}
	sd := &repository.StreakData{}
	pm := &repository.PointsMovement{MatchCount: 5}
	hs := hotColdHighlights(uid1, "Alice", sd, fd, pm)
	h := findType(hs, model.TypeHotPlayer)
	require.NotNil(t, h)
	assert.Equal(t, model.SectionTrending, h.Section)
	assert.Equal(t, 95, h.Priority)
}

func TestHotColdHighlights_HotRequiresMinMatchesToday(t *testing.T) {
	results := make([]bool, 10)
	for i := 0; i < 9; i++ {
		results[i] = true
	}
	fd := &repository.RecentFormData{Results: results}
	sd := &repository.StreakData{}
	pm := &repository.PointsMovement{MatchCount: 4} // below HotPlayerMinMatchesToday=5
	hs := hotColdHighlights(uid1, "Alice", sd, fd, pm)
	assert.Nil(t, findType(hs, model.TypeHotPlayer))
}

func TestHotColdHighlights_WinRateJustBelowHot(t *testing.T) {
	// 7/10 = 0.70 < 0.80 → not hot
	results := make([]bool, 10)
	for i := 0; i < 7; i++ {
		results[i] = true
	}
	fd := &repository.RecentFormData{Results: results}
	sd := &repository.StreakData{}
	pm := &repository.PointsMovement{MatchCount: 8}
	hs := hotColdHighlights(uid1, "Alice", sd, fd, pm)
	assert.Nil(t, findType(hs, model.TypeHotPlayer))
}

func TestHotColdHighlights_Cold(t *testing.T) {
	fd := &repository.RecentFormData{Results: make([]bool, 10)} // 0% wins
	sd := &repository.StreakData{CurrentLose: 5}
	pm := &repository.PointsMovement{MatchCount: 3}
	hs := hotColdHighlights(uid1, "Alice", sd, fd, pm)
	h := findType(hs, model.TypeColdPlayer)
	require.NotNil(t, h)
	assert.Equal(t, 65, h.Priority)
}

func TestHotColdHighlights_ColdLoseStreakJustBelow(t *testing.T) {
	fd := &repository.RecentFormData{Results: make([]bool, 10)}
	sd := &repository.StreakData{CurrentLose: 4} // below ColdPlayerMinLoseStreak=5
	pm := &repository.PointsMovement{}
	hs := hotColdHighlights(uid1, "Alice", sd, fd, pm)
	assert.Nil(t, findType(hs, model.TypeColdPlayer))
}

// ─── fastHighlights ───────────────────────────────────────────────────────────

func TestFastHighlights_BelowThreshold(t *testing.T) {
	hr := &repository.PointsMovement{PointsGained: 49, PointsLost: 49}
	assert.Empty(t, fastHighlights(uid1, "Alice", hr))
}

func TestFastHighlights_FastClimb(t *testing.T) {
	hr := &repository.PointsMovement{PointsGained: 50}
	hs := fastHighlights(uid1, "Alice", hr)
	h := findType(hs, model.TypeFastClimbHour)
	require.NotNil(t, h)
	assert.Equal(t, model.SectionTrending, h.Section)
	assert.Equal(t, 90, h.Priority)
}

func TestFastHighlights_FastCollapse(t *testing.T) {
	hr := &repository.PointsMovement{PointsLost: 50}
	hs := fastHighlights(uid1, "Alice", hr)
	h := findType(hs, model.TypeFastCollapseHour)
	require.NotNil(t, h)
	assert.Equal(t, model.SectionCompetitive, h.Section)
	assert.Equal(t, 65, h.Priority)
}

// ─── milestoneHighlights ──────────────────────────────────────────────────────

func TestMilestoneHighlights_NoMatch(t *testing.T) {
	td := &repository.TotalsData{TotalWins: 99, CurrentScore: 999, TotalMatches: 499}
	rd := &repository.RankData{}
	assert.Empty(t, milestoneHighlights(uid1, "Alice", td, rd))
}

func TestMilestoneHighlights_WinsExact100(t *testing.T) {
	td := &repository.TotalsData{TotalWins: 100}
	rd := &repository.RankData{}
	hs := milestoneHighlights(uid1, "Alice", td, rd)
	h := findType(hs, model.TypeMilestoneWins)
	require.NotNil(t, h)
	assert.Equal(t, 100.0, h.Value)
	assert.Equal(t, 100, h.Priority)
}

func TestMilestoneHighlights_WinsJustAbove(t *testing.T) {
	td := &repository.TotalsData{TotalWins: 101}
	rd := &repository.RankData{}
	hs := milestoneHighlights(uid1, "Alice", td, rd)
	assert.Nil(t, findType(hs, model.TypeMilestoneWins))
}

func TestMilestoneHighlights_ScoreMilestone(t *testing.T) {
	td := &repository.TotalsData{CurrentScore: 1000}
	rd := &repository.RankData{}
	hs := milestoneHighlights(uid1, "Alice", td, rd)
	h := findType(hs, model.TypeMilestonePoints)
	require.NotNil(t, h)
	assert.Equal(t, 1000.0, h.Value)
}

func TestMilestoneHighlights_MatchesMilestone(t *testing.T) {
	td := &repository.TotalsData{TotalMatches: 500}
	rd := &repository.RankData{}
	hs := milestoneHighlights(uid1, "Alice", td, rd)
	h := findType(hs, model.TypeMilestoneMatches)
	require.NotNil(t, h)
	assert.Equal(t, 500.0, h.Value)
}

// ─── socialHighlights ─────────────────────────────────────────────────────────

func TestSocialHighlights_Cooking(t *testing.T) {
	results := make([]bool, 10)
	for i := 0; i < 9; i++ {
		results[i] = true
	}
	fd := &repository.RecentFormData{Results: results} // 0.90 wr, count=10
	sd := &repository.StreakData{}
	pm := &repository.PointsMovement{MatchCount: 5}
	rd := &repository.RankData{CurrentRank: 5, YesterdayRank: 5}
	hs := socialHighlights(uid1, "Alice", sd, fd, pm, rd)
	h := findType(hs, model.TypeSocialCooking)
	require.NotNil(t, h)
	assert.Equal(t, model.SectionSocial, h.Section)
}

func TestSocialHighlights_Tilt(t *testing.T) {
	fd := &repository.RecentFormData{Results: make([]bool, 5)}
	sd := &repository.StreakData{CurrentLose: 3}
	pm := &repository.PointsMovement{MatchCount: 5}
	rd := &repository.RankData{}
	hs := socialHighlights(uid1, "Alice", sd, fd, pm, rd)
	h := findType(hs, model.TypeSocialTilt)
	require.NotNil(t, h)
	assert.Equal(t, float64(3), h.Value)
}

func TestSocialHighlights_TiltRequiresMinMatches(t *testing.T) {
	fd := &repository.RecentFormData{Results: make([]bool, 5)}
	sd := &repository.StreakData{CurrentLose: 5}
	pm := &repository.PointsMovement{MatchCount: 4} // below threshold
	rd := &repository.RankData{}
	hs := socialHighlights(uid1, "Alice", sd, fd, pm, rd)
	assert.Nil(t, findType(hs, model.TypeSocialTilt))
}

func TestSocialHighlights_Grinder(t *testing.T) {
	fd := &repository.RecentFormData{Results: make([]bool, 5)}
	sd := &repository.StreakData{}
	pm := &repository.PointsMovement{MatchCount: 15} // GrinderMinMatches
	rd := &repository.RankData{}
	hs := socialHighlights(uid1, "Alice", sd, fd, pm, rd)
	assert.NotNil(t, findType(hs, model.TypeSocialGrinder))
	assert.Nil(t, findType(hs, model.TypeSocialTryhard))
}

func TestSocialHighlights_Tryhard(t *testing.T) {
	fd := &repository.RecentFormData{Results: make([]bool, 5)}
	sd := &repository.StreakData{}
	pm := &repository.PointsMovement{MatchCount: 10} // TryhardMinMatches, below Grinder
	rd := &repository.RankData{}
	hs := socialHighlights(uid1, "Alice", sd, fd, pm, rd)
	assert.NotNil(t, findType(hs, model.TypeSocialTryhard))
	assert.Nil(t, findType(hs, model.TypeSocialGrinder))
}

func TestSocialHighlights_Sneaky(t *testing.T) {
	fd := &repository.RecentFormData{Results: make([]bool, 5)}
	sd := &repository.StreakData{}
	pm := &repository.PointsMovement{MatchCount: 3} // <= SneakyMaxMatches
	rd := &repository.RankData{CurrentRank: 5, YesterdayRank: 8} // climbed 3 >= SneakyMinRankClimb
	hs := socialHighlights(uid1, "Alice", sd, fd, pm, rd)
	h := findType(hs, model.TypeSocialSneaky)
	require.NotNil(t, h)
	assert.Equal(t, 3.0, h.Value) // rank climb of 3
}

func TestSocialHighlights_SneakyTooManyMatches(t *testing.T) {
	fd := &repository.RecentFormData{Results: make([]bool, 5)}
	sd := &repository.StreakData{}
	pm := &repository.PointsMovement{MatchCount: 6} // > SneakyMaxMatches
	rd := &repository.RankData{CurrentRank: 3, YesterdayRank: 8}
	hs := socialHighlights(uid1, "Alice", sd, fd, pm, rd)
	assert.Nil(t, findType(hs, model.TypeSocialSneaky))
}

// ─── streakBreakerHighlights ──────────────────────────────────────────────────

func TestStreakBreakerHighlights_Empty(t *testing.T) {
	assert.Empty(t, streakBreakerHighlights(nil))
}

func TestStreakBreakerHighlights_OneBreaker(t *testing.T) {
	breakers := []*repository.StreakBreakerData{
		{BreakerID: uid1, BreakerName: "Alice", VictimID: uid2, VictimName: "Bob", StreakLength: 5},
	}
	hs := streakBreakerHighlights(breakers)
	require.Len(t, hs, 1)
	assert.Equal(t, model.TypeStreakBrokenWin, hs[0].Type)
	assert.Equal(t, uid1, hs[0].PlayerID)
	assert.Equal(t, 5.0, hs[0].Value)
	assert.Equal(t, model.SectionSocial, hs[0].Section)
}

// ─── groupAndCap ──────────────────────────────────────────────────────────────

func TestGroupAndCap_EmptyInput(t *testing.T) {
	resp := groupAndCap(nil)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Trending)
	assert.Empty(t, resp.DailyRecap)
	assert.Empty(t, resp.Competitive)
	assert.Empty(t, resp.Social)
}

func TestGroupAndCap_RoutesToCorrectSection(t *testing.T) {
	hs := []model.Highlight{
		{Type: "a", Section: model.SectionTrending, Priority: 50},
		{Type: "b", Section: model.SectionDailyRecap, Priority: 50},
		{Type: "c", Section: model.SectionCompetitive, Priority: 50},
		{Type: "d", Section: model.SectionSocial, Priority: 50},
	}
	resp := groupAndCap(hs)
	assert.Len(t, resp.Trending, 1)
	assert.Len(t, resp.DailyRecap, 1)
	assert.Len(t, resp.Competitive, 1)
	assert.Len(t, resp.Social, 1)
}

func TestGroupAndCap_DropsUnknownSection(t *testing.T) {
	hs := []model.Highlight{
		{Type: "x", Section: "unknown_section", Priority: 99},
	}
	resp := groupAndCap(hs)
	assert.Empty(t, resp.Trending)
	assert.Empty(t, resp.DailyRecap)
	assert.Empty(t, resp.Competitive)
	assert.Empty(t, resp.Social)
}

func TestGroupAndCap_SortsByPriorityDesc(t *testing.T) {
	hs := []model.Highlight{
		{Type: "low", Section: model.SectionTrending, Priority: 40},
		{Type: "high", Section: model.SectionTrending, Priority: 90},
		{Type: "med", Section: model.SectionTrending, Priority: 60},
	}
	resp := groupAndCap(hs)
	require.Len(t, resp.Trending, 3)
	assert.Equal(t, "high", resp.Trending[0].Type)
	assert.Equal(t, "med", resp.Trending[1].Type)
	assert.Equal(t, "low", resp.Trending[2].Type)
}

func TestGroupAndCap_CapsAtFivePerSection(t *testing.T) {
	hs := make([]model.Highlight, 8)
	for i := range hs {
		hs[i] = model.Highlight{Section: model.SectionTrending, Priority: i}
	}
	resp := groupAndCap(hs)
	assert.Len(t, resp.Trending, maxPerSection)
}

func TestGroupAndCap_CapKeepsHighestPriority(t *testing.T) {
	hs := make([]model.Highlight, 8)
	for i := range hs {
		hs[i] = model.Highlight{Type: string(rune('a' + i)), Section: model.SectionTrending, Priority: i * 10}
	}
	resp := groupAndCap(hs)
	require.Len(t, resp.Trending, maxPerSection)
	// Highest priorities are 70,60,50,40,30 — not 0,10,20
	assert.Equal(t, 70, resp.Trending[0].Priority)
	assert.Equal(t, 30, resp.Trending[4].Priority)
}
