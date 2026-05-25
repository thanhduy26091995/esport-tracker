package service

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/google/uuid"
)

const maxPerSection = 5

type HighlightService struct {
	repo     *repository.HighlightRepository
	userRepo *repository.UserRepository
}

func NewHighlightService(repo *repository.HighlightRepository, userRepo *repository.UserRepository) *HighlightService {
	return &HighlightService{repo: repo, userRepo: userRepo}
}

// GenerateHighlights computes all highlights from live match data.
func (s *HighlightService) GenerateHighlights() (*model.HighlightsResponse, error) {
	// --- Fetch all data (sequential; can be parallelised with errgroup if needed) ---
	streaks, err := s.repo.GetCurrentStreaks()
	if err != nil {
		return emptyResponse(), nil
	}
	_, err = s.repo.GetLongestStreaks()
	if err != nil {
		return emptyResponse(), nil
	}
	today, err := s.repo.GetPointsMovementToday()
	if err != nil {
		return emptyResponse(), nil
	}
	lastHour, err := s.repo.GetPointsMovementLastHour()
	if err != nil {
		return emptyResponse(), nil
	}
	ranks, err := s.repo.GetRankSnapshot()
	if err != nil {
		return emptyResponse(), nil
	}
	form, err := s.repo.GetRecentForm()
	if err != nil {
		return emptyResponse(), nil
	}
	activity, err := s.repo.GetWeeklyActivity()
	if err != nil {
		return emptyResponse(), nil
	}
	totals, err := s.repo.GetTotals()
	if err != nil {
		return emptyResponse(), nil
	}
	breakers, err := s.repo.GetStreakBreakers()
	if err != nil {
		breakers = nil
	}

	// Build user name map from totals (all active users present)
	names := make(map[uuid.UUID]string)
	users, err := s.userRepo.GetAll()
	if err == nil {
		for _, u := range users {
			names[u.ID] = u.Name
		}
	}

	var all []model.Highlight

	// --- Per-user highlights ---
	for userID, name := range names {
		sd := streaks[userID]
		pm := today[userID]
		hr := lastHour[userID]
		rd := ranks[userID]
		fd := form[userID]
		wa := activity[userID]
		td := totals[userID]

		if sd != nil {
			all = append(all, streakHighlights(userID, name, sd)...)
		}
		if pm != nil {
			all = append(all, pointsHighlights(userID, name, pm)...)
		}
		if rd != nil && pm != nil {
			all = append(all, rankHighlights(userID, name, rd)...)
		}
		if fd != nil {
			all = append(all, formHighlights(userID, name, fd)...)
		}
		if wa != nil {
			all = append(all, activityHighlights(userID, name, wa)...)
		}
		if sd != nil && fd != nil && pm != nil {
			all = append(all, hotColdHighlights(userID, name, sd, fd, pm)...)
		}
		if hr != nil {
			all = append(all, fastHighlights(userID, name, hr)...)
		}
		if td != nil && rd != nil {
			all = append(all, milestoneHighlights(userID, name, td, rd)...)
		}
		if sd != nil && fd != nil && pm != nil && rd != nil {
			all = append(all, socialHighlights(userID, name, sd, fd, pm, rd)...)
		}
	}

	// --- Cross-user highlights (need global max) ---
	all = append(all, fastestClimberToday(today, names)...)
	all = append(all, mostActiveToday(today, names)...)
	all = append(all, biggestCollapseToday(today, names)...)

	// --- Streak-breaker highlights ---
	all = append(all, streakBreakerHighlights(breakers)...)

	return groupAndCap(all), nil
}

// --- Category 1: Streaks ---

func streakHighlights(userID uuid.UUID, name string, sd *repository.StreakData) []model.Highlight {
	var h []model.Highlight
	if sd.CurrentWin >= model.StreakLowPriority {
		h = append(h, hl(userID, name, model.TypeStreakWin, model.SectionTrending, "🔥",
			fmt.Sprintf("%s đang thắng liên tục %d trận", name, sd.CurrentWin),
			float64(sd.CurrentWin), priorityForStreak(sd.CurrentWin)))
	}
	if sd.CurrentLose >= model.StreakLowPriority {
		h = append(h, hl(userID, name, model.TypeStreakLose, model.SectionDailyRecap, "😵",
			fmt.Sprintf("%s đã thua %d trận liên tiếp", name, sd.CurrentLose),
			float64(sd.CurrentLose), priorityForStreak(sd.CurrentLose)))
	}
	// Unbeaten only when unbeaten > win streak (meaning draws are involved, or just a clean unbeaten run)
	if sd.CurrentUnbeaten >= model.StreakLowPriority && sd.CurrentUnbeaten > sd.CurrentWin {
		h = append(h, hl(userID, name, model.TypeStreakUnbeaten, model.SectionTrending, "🛡️",
			fmt.Sprintf("%s bất bại %d trận gần nhất", name, sd.CurrentUnbeaten),
			float64(sd.CurrentUnbeaten), priorityForStreak(sd.CurrentUnbeaten)))
	}
	return h
}

// --- Category 2: Rank / Point Movement ---

func pointsHighlights(userID uuid.UUID, name string, pm *repository.PointsMovement) []model.Highlight {
	var h []model.Highlight
	if pm.PointsGained > 0 {
		p := 50
		if pm.PointsGained >= model.PointsHighGain {
			p = 80
		} else if pm.PointsGained >= model.PointsMedGain {
			p = 65
		}
		h = append(h, hl(userID, name, model.TypePointsGainedToday, model.SectionDailyRecap, "🚀",
			fmt.Sprintf("%s vừa tăng +%d điểm hôm nay", name, pm.PointsGained),
			float64(pm.PointsGained), p))
	}
	if pm.PointsLost > 0 {
		h = append(h, hl(userID, name, model.TypePointsLostToday, model.SectionDailyRecap, "📉",
			fmt.Sprintf("%s mất -%d điểm chỉ trong hôm nay", name, pm.PointsLost),
			float64(pm.PointsLost), 45))
	}
	return h
}

func rankHighlights(userID uuid.UUID, name string, rd *repository.RankData) []model.Highlight {
	var h []model.Highlight
	diff := rd.YesterdayRank - rd.CurrentRank // positive = climbed
	if diff > 0 {
		p := 65
		if rd.CurrentRank <= 3 {
			p = 90
		} else if diff >= 3 {
			p = 75
		}
		h = append(h, hl(userID, name, model.TypeRankClimbed, model.SectionCompetitive, "🏆",
			fmt.Sprintf("%s vừa leo lên hạng %d (tăng %d bậc)", name, rd.CurrentRank, diff),
			float64(diff), p))
		// milestone: entered top 10 today
		if rd.CurrentRank <= 10 && rd.YesterdayRank > 10 {
			h = append(h, hl(userID, name, model.TypeMilestoneTop10, model.SectionCompetitive, "👑",
				fmt.Sprintf("%s vừa lọt vào Top 10 hôm nay!", name),
				float64(rd.CurrentRank), 90))
		}
	} else if diff < 0 {
		dropped := -diff
		h = append(h, hl(userID, name, model.TypeRankDropped, model.SectionCompetitive, "😬",
			fmt.Sprintf("%s tụt %d bậc trên BXH hôm nay", name, dropped),
			float64(dropped), 50))
	}
	return h
}

func fastestClimberToday(today map[uuid.UUID]*repository.PointsMovement, names map[uuid.UUID]string) []model.Highlight {
	var best uuid.UUID
	bestGain := 0
	for uid, pm := range today {
		if pm.PointsGained > bestGain {
			bestGain = pm.PointsGained
			best = uid
		}
	}
	if bestGain <= 0 {
		return nil
	}
	name := names[best]
	return []model.Highlight{hl(best, name, model.TypeFastestClimberToday, model.SectionTrending, "⚡",
		fmt.Sprintf("%s tăng rank nhanh nhất hôm nay (+%d điểm)", name, bestGain),
		float64(bestGain), 85)}
}

func biggestCollapseToday(today map[uuid.UUID]*repository.PointsMovement, names map[uuid.UUID]string) []model.Highlight {
	var best uuid.UUID
	bestLoss := 0
	for uid, pm := range today {
		if pm.PointsLost > bestLoss {
			bestLoss = pm.PointsLost
			best = uid
		}
	}
	if bestLoss <= 0 {
		return nil
	}
	name := names[best]
	return []model.Highlight{hl(best, name, model.TypeBiggestCollapse, model.SectionCompetitive, "📉",
		fmt.Sprintf("%s tụt rank mạnh nhất hôm nay (-%d điểm)", name, bestLoss),
		float64(bestLoss), 60)}
}

// --- Category 3: Recent Form ---

func formHighlights(userID uuid.UUID, name string, fd *repository.RecentFormData) []model.Highlight {
	wr, count := winRateFromForm(fd)
	if count < 5 {
		return nil
	}
	var h []model.Highlight
	if wr >= model.FormHotMinWinRate {
		wins := int(wr * float64(count))
		h = append(h, hl(userID, name, model.TypeFormHot, model.SectionTrending, "🔥",
			fmt.Sprintf("%s thắng %d/%d trận gần nhất", name, wins, count),
			wr, 70))
	} else if wr <= model.FormColdMaxWinRate {
		h = append(h, hl(userID, name, model.TypeFormCold, model.SectionDailyRecap, "🥶",
			fmt.Sprintf("%s đang tụt phong độ (%.0f%% thắng trong %d trận gần nhất)", name, wr*100, count),
			wr, 50))
	} else if wr >= 0.45 && wr <= 0.65 && count >= 8 {
		h = append(h, hl(userID, name, model.TypeFormStable, model.SectionSocial, "🎮",
			fmt.Sprintf("%s là người chơi ổn định nhất tuần", name),
			wr, 50))
	}
	return h
}

// --- Category 4: Activity ---

func activityHighlights(userID uuid.UUID, name string, wa *repository.WeeklyActivityData) []model.Highlight {
	var h []model.Highlight
	if wa.ConsecutiveActiveDays >= 5 {
		h = append(h, hl(userID, name, model.TypeActiveStreakDays, model.SectionSocial, "📅",
			fmt.Sprintf("%s active liên tục %d ngày", name, wa.ConsecutiveActiveDays),
			float64(wa.ConsecutiveActiveDays), 50))
	}
	return h
}

func mostActiveToday(today map[uuid.UUID]*repository.PointsMovement, names map[uuid.UUID]string) []model.Highlight {
	var best uuid.UUID
	bestCount := 0
	for uid, pm := range today {
		if pm.MatchCount > bestCount {
			bestCount = pm.MatchCount
			best = uid
		}
	}
	if bestCount < 5 {
		return nil
	}
	name := names[best]
	h := []model.Highlight{hl(best, name, model.TypeMostActiveToday, model.SectionTrending, "🎮",
		fmt.Sprintf("%s chơi nhiều trận nhất hôm nay (%d trận)", name, bestCount),
		float64(bestCount), 55)}
	if bestCount >= model.MarathonMinMatches {
		h = append(h, hl(best, name, model.TypeMarathon, model.SectionSocial, "🕹️",
			fmt.Sprintf("%s vừa marathon %d trận!", name, bestCount),
			float64(bestCount), 55))
	}
	return h
}

// --- Category 5: Hot / Cold ---

func hotColdHighlights(userID uuid.UUID, name string, sd *repository.StreakData, fd *repository.RecentFormData, pm *repository.PointsMovement) []model.Highlight {
	wr, count := winRateFromForm(fd)
	if count < 10 {
		return nil
	}
	var h []model.Highlight
	if wr >= model.HotPlayerMinWinRate && pm.MatchCount >= model.HotPlayerMinMatchesToday {
		h = append(h, hl(userID, name, model.TypeHotPlayer, model.SectionTrending, "🔥",
			fmt.Sprintf("Không ai cản nổi %s hôm nay (%.0f%% trong 10 trận gần nhất)", name, wr*100),
			wr, 95))
	}
	if sd.CurrentLose >= model.ColdPlayerMinLoseStreak {
		h = append(h, hl(userID, name, model.TypeColdPlayer, model.SectionDailyRecap, "🥶",
			fmt.Sprintf("%s đang gặp khó khăn với %d trận thua liên tiếp", name, sd.CurrentLose),
			float64(sd.CurrentLose), 65))
	}
	return h
}

// --- Category 6: Fast Climb / Collapse (last hour) ---

func fastHighlights(userID uuid.UUID, name string, hr *repository.PointsMovement) []model.Highlight {
	var h []model.Highlight
	if hr.PointsGained >= model.PointsMedGain {
		h = append(h, hl(userID, name, model.TypeFastClimbHour, model.SectionTrending, "⚡",
			fmt.Sprintf("%s vừa gain %d điểm trong 1 giờ qua", name, hr.PointsGained),
			float64(hr.PointsGained), 90))
	}
	if hr.PointsLost >= model.PointsMedGain {
		h = append(h, hl(userID, name, model.TypeFastCollapseHour, model.SectionCompetitive, "📉",
			fmt.Sprintf("%s mất %d điểm trong 1 giờ qua", name, hr.PointsLost),
			float64(hr.PointsLost), 65))
	}
	return h
}

// --- Category 7: Milestones ---

func milestoneHighlights(userID uuid.UUID, name string, td *repository.TotalsData, rd *repository.RankData) []model.Highlight {
	var h []model.Highlight
	for _, threshold := range model.MilestoneWins {
		if td.TotalWins == threshold {
			h = append(h, hl(userID, name, model.TypeMilestoneWins, model.SectionSocial, "🎉",
				fmt.Sprintf("%s cán mốc %d trận thắng!", name, threshold),
				float64(threshold), 100))
		}
	}
	for _, threshold := range model.MilestonePoints {
		if td.CurrentScore == threshold {
			h = append(h, hl(userID, name, model.TypeMilestonePoints, model.SectionSocial, "🏅",
				fmt.Sprintf("%s đạt %d điểm!", name, threshold),
				float64(threshold), 100))
		}
	}
	for _, threshold := range model.MilestoneMatches {
		if td.TotalMatches == threshold {
			h = append(h, hl(userID, name, model.TypeMilestoneMatches, model.SectionSocial, "🎮",
				fmt.Sprintf("%s vừa chơi trận thứ %d!", name, threshold),
				float64(threshold), 100))
		}
	}
	return h
}

// --- Category 8: Social / Fun ---

var socialCookingVariants = []string{
	"%s is cooking 🔥",
	"Không ai cản nổi %s hôm nay",
	"%s đang trong trạng thái thăng hoa",
}
var socialTiltVariants = []string{
	"%s cần reset mental 😅",
	"%s đang tilt nặng rồi",
	"Ai đó hãy gọi %s nghỉ ngơi đi",
}
var socialTryhardVariants = []string{
	"%s vừa bật mode tryhard 💀",
	"%s đang cày cuốc không nghỉ",
	"%s hôm nay không có chỗ cho ai",
}
var socialGrinderVariants = []string{
	"%s đang spam rank 🕹️",
	"%s chơi nhiều hơn cả máy",
	"%s hôm nay không làm gì ngoài chơi game",
}
var socialSneakyVariants = []string{
	"%s đang âm thầm leo rank 👀",
	"Đừng nhìn vào %s, họ đang leo rank",
	"%s ít trận nhưng vẫn leo rank cực nhanh",
}

func socialHighlights(userID uuid.UUID, name string, sd *repository.StreakData, fd *repository.RecentFormData, pm *repository.PointsMovement, rd *repository.RankData) []model.Highlight {
	wr, count := winRateFromForm(fd)
	var h []model.Highlight

	// social_cooking: same condition as hot_player
	if count >= 10 && wr >= model.HotPlayerMinWinRate && pm.MatchCount >= model.HotPlayerMinMatchesToday {
		msg := fmt.Sprintf(randomVariant(socialCookingVariants), name)
		h = append(h, hl(userID, name, model.TypeSocialCooking, model.SectionSocial, "🔥", msg, wr, 60))
	}
	// social_tilt: lose streak >= 3 AND matches today >= 5
	if sd.CurrentLose >= 3 && pm.MatchCount >= 5 {
		msg := fmt.Sprintf(randomVariant(socialTiltVariants), name)
		h = append(h, hl(userID, name, model.TypeSocialTilt, model.SectionSocial, "😅", msg, float64(sd.CurrentLose), 55))
	}
	// social_grinder >= social_tryhard (check grinder first, it's more extreme)
	if pm.MatchCount >= model.GrinderMinMatches {
		msg := fmt.Sprintf(randomVariant(socialGrinderVariants), name)
		h = append(h, hl(userID, name, model.TypeSocialGrinder, model.SectionSocial, "🕹️", msg, float64(pm.MatchCount), 50))
	} else if pm.MatchCount >= model.TryhardMinMatches {
		msg := fmt.Sprintf(randomVariant(socialTryhardVariants), name)
		h = append(h, hl(userID, name, model.TypeSocialTryhard, model.SectionSocial, "💀", msg, float64(pm.MatchCount), 50))
	}
	// social_sneaky: climbed >= 2 ranks AND <= 5 matches today
	rankClimb := rd.YesterdayRank - rd.CurrentRank
	if rankClimb >= model.SneakyMinRankClimb && pm.MatchCount <= model.SneakyMaxMatches {
		msg := fmt.Sprintf(randomVariant(socialSneakyVariants), name)
		h = append(h, hl(userID, name, model.TypeSocialSneaky, model.SectionSocial, "👀", msg, float64(rankClimb), 55))
	}
	return h
}

// --- Streak breaker highlights ---

func streakBreakerHighlights(breakers []*repository.StreakBreakerData) []model.Highlight {
	var h []model.Highlight
	for _, b := range breakers {
		h = append(h, hl(b.BreakerID, b.BreakerName, model.TypeStreakBrokenWin, model.SectionSocial, "💥",
			fmt.Sprintf("%s vừa kết thúc chuỗi %d trận thắng của %s", b.BreakerName, b.StreakLength, b.VictimName),
			float64(b.StreakLength), 65))
	}
	return h
}

// --- Helpers ---

func hl(playerID uuid.UUID, playerName, typ, section, emoji, message string, value float64, priority int) model.Highlight {
	return model.Highlight{
		PlayerID:   playerID,
		PlayerName: playerName,
		Type:       typ,
		Section:    section,
		Emoji:      emoji,
		Message:    message,
		Value:      value,
		Priority:   priority,
	}
}

func winRateFromForm(fd *repository.RecentFormData) (float64, int) {
	if fd == nil || len(fd.Results) == 0 {
		return 0, 0
	}
	wins := 0
	for _, w := range fd.Results {
		if w {
			wins++
		}
	}
	return float64(wins) / float64(len(fd.Results)), len(fd.Results)
}

func priorityForStreak(length int) int {
	if length >= model.StreakHighPriority {
		return 95
	}
	if length >= model.StreakMedPriority {
		return 80
	}
	return 60
}

func randomVariant(variants []string) string {
	return variants[rand.Intn(len(variants))]
}

func groupAndCap(highlights []model.Highlight) *model.HighlightsResponse {
	sections := map[string][]model.Highlight{
		model.SectionTrending:    {},
		model.SectionDailyRecap:  {},
		model.SectionCompetitive: {},
		model.SectionSocial:      {},
	}
	for _, h := range highlights {
		if _, ok := sections[h.Section]; ok {
			sections[h.Section] = append(sections[h.Section], h)
		}
	}
	for sec := range sections {
		s := sections[sec]
		sort.Slice(s, func(i, j int) bool { return s[i].Priority > s[j].Priority })
		if len(s) > maxPerSection {
			s = s[:maxPerSection]
		}
		sections[sec] = s
	}
	return &model.HighlightsResponse{
		Trending:   sections[model.SectionTrending],
		DailyRecap: sections[model.SectionDailyRecap],
		Competitive: sections[model.SectionCompetitive],
		Social:     sections[model.SectionSocial],
	}
}

func emptyResponse() *model.HighlightsResponse {
	return &model.HighlightsResponse{
		Trending:    []model.Highlight{},
		DailyRecap:  []model.Highlight{},
		Competitive: []model.Highlight{},
		Social:      []model.Highlight{},
	}
}
