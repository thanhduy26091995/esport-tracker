package service

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
)

const openfootballURL = "https://raw.githubusercontent.com/openfootball/worldcup.json/master/2026/worldcup.json"

type WcOpenFootballClient struct {
	httpClient *http.Client
}

func NewWcOpenFootballClient() *WcOpenFootballClient {
	return &WcOpenFootballClient{httpClient: &http.Client{Timeout: 15 * time.Second}}
}

// WcOpenFootballData holds all analytics derived from a single GET of worldcup.json.
type WcOpenFootballData struct {
	GoalTiming        []model.WcGoalTimingBucket
	HalfTimeStats     model.WcHalfTimeStats
	TeamStats         []model.WcTeamStat
	GoalsByGroup      []model.WcGroupGoals
	TopScoringMatches []model.WcMatchDetail
	VenueStats        []model.WcVenueStat
}

func (c *WcOpenFootballClient) GetWCData() (*WcOpenFootballData, error) {
	resp, err := c.httpClient.Get(openfootballURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw ofWorldcup
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	buckets := map[string]int{
		"1-15": 0, "16-30": 0, "31-45": 0, "45+": 0,
		"46-60": 0, "61-75": 0, "76-90": 0, "90+": 0,
	}
	var ownGoals, penaltyGoals, firstHalf, secondHalf int
	var comebacks, heldLead int

	teamGoalsFor := map[string]int{}
	teamGoalsAgainst := map[string]int{}
	teamMatches := map[string]int{}
	groupGoals := map[string]int{}
	groupMatches := map[string]int{}
	venueGoals := map[string]int{}
	venueMatches := map[string]int{}

	type scoredMatch struct {
		detail model.WcMatchDetail
		total  int
	}
	var allMatches []scoredMatch

	for _, m := range raw.Matches {
		if m.Score == nil || len(m.Score.FT) < 2 {
			continue // upcoming — no score yet
		}

		ft := m.Score.FT
		homeGoals := ft[0]
		awayGoals := ft[1]
		totalGoals := homeGoals + awayGoals

		teamGoalsFor[m.Team1] += homeGoals
		teamGoalsAgainst[m.Team1] += awayGoals
		teamMatches[m.Team1]++
		teamGoalsFor[m.Team2] += awayGoals
		teamGoalsAgainst[m.Team2] += homeGoals
		teamMatches[m.Team2]++

		groupGoals[m.Group] += totalGoals
		groupMatches[m.Group]++
		venueGoals[m.Ground] += totalGoals
		venueMatches[m.Ground]++

		if len(m.Score.HT) == 2 {
			ht := m.Score.HT
			if (ht[0] < ht[1] && ft[0] > ft[1]) || (ht[1] < ht[0] && ft[1] > ft[0]) {
				comebacks++
			}
			if (ht[0] > ht[1] && ft[0] > ft[1]) || (ht[1] > ht[0] && ft[1] > ft[0]) {
				heldLead++
			}
		}

		for _, g := range m.Goals1 {
			if g.OwnGoal {
				ownGoals++
				continue
			}
			if g.Penalty {
				penaltyGoals++
			}
			buckets[parseBucket(g.Minute)]++
			if isFirstHalfMinute(g.Minute) {
				firstHalf++
			} else {
				secondHalf++
			}
		}
		for _, g := range m.Goals2 {
			if g.OwnGoal {
				ownGoals++
				continue
			}
			if g.Penalty {
				penaltyGoals++
			}
			buckets[parseBucket(g.Minute)]++
			if isFirstHalfMinute(g.Minute) {
				firstHalf++
			} else {
				secondHalf++
			}
		}

		allMatches = append(allMatches, scoredMatch{
			detail: model.WcMatchDetail{
				HomeTeam: m.Team1, AwayTeam: m.Team2,
				HomeScore: homeGoals, AwayScore: awayGoals,
				TotalGoals: totalGoals,
				Group: m.Group, Round: m.Round, Date: m.Date, Venue: m.Ground,
			},
			total: totalGoals,
		})
	}

	bucketOrder := []string{"1-15", "16-30", "31-45", "45+", "46-60", "61-75", "76-90", "90+"}
	timingSlice := make([]model.WcGoalTimingBucket, len(bucketOrder))
	for i, label := range bucketOrder {
		timingSlice[i] = model.WcGoalTimingBucket{Label: label, Goals: buckets[label]}
	}

	teamSlice := make([]model.WcTeamStat, 0, len(teamMatches))
	for name, matches := range teamMatches {
		teamSlice = append(teamSlice, model.WcTeamStat{
			TeamName:     name,
			GoalsFor:     teamGoalsFor[name],
			GoalsAgainst: teamGoalsAgainst[name],
			Matches:      matches,
		})
	}
	sort.Slice(teamSlice, func(i, j int) bool {
		return teamSlice[i].GoalsFor > teamSlice[j].GoalsFor
	})

	groupSlice := make([]model.WcGroupGoals, 0, len(groupGoals))
	for g, goals := range groupGoals {
		groupSlice = append(groupSlice, model.WcGroupGoals{Group: g, Matches: groupMatches[g], Goals: goals})
	}
	sort.Slice(groupSlice, func(i, j int) bool { return groupSlice[i].Group < groupSlice[j].Group })

	venueSlice := make([]model.WcVenueStat, 0, len(venueGoals))
	for v, goals := range venueGoals {
		venueSlice = append(venueSlice, model.WcVenueStat{Venue: v, Matches: venueMatches[v], Goals: goals})
	}
	sort.Slice(venueSlice, func(i, j int) bool { return venueSlice[i].Goals > venueSlice[j].Goals })

	sort.Slice(allMatches, func(i, j int) bool { return allMatches[i].total > allMatches[j].total })
	top5 := make([]model.WcMatchDetail, 0, 5)
	for i, m := range allMatches {
		if i >= 5 {
			break
		}
		top5 = append(top5, m.detail)
	}

	return &WcOpenFootballData{
		GoalTiming: timingSlice,
		HalfTimeStats: model.WcHalfTimeStats{
			FirstHalfGoals:  firstHalf,
			SecondHalfGoals: secondHalf,
			OwnGoals:        ownGoals,
			PenaltyGoals:    penaltyGoals,
			Comebacks:       comebacks,
			HeldLead:        heldLead,
		},
		TeamStats:         teamSlice,
		GoalsByGroup:      groupSlice,
		TopScoringMatches: top5,
		VenueStats:        venueSlice,
	}, nil
}

// parseBucket maps a goal minute string to one of 8 display buckets.
func parseBucket(minuteStr string) string {
	parts := strings.SplitN(minuteStr, "+", 2)
	base, _ := strconv.Atoi(parts[0])
	injury := len(parts) == 2
	switch {
	case injury && base == 45:
		return "45+"
	case injury && base >= 90:
		return "90+"
	case base <= 15:
		return "1-15"
	case base <= 30:
		return "16-30"
	case base <= 45:
		return "31-45"
	case base <= 60:
		return "46-60"
	case base <= 75:
		return "61-75"
	default:
		return "76-90"
	}
}

// isFirstHalfMinute returns true for minutes 1-45 and 45+X injury time.
func isFirstHalfMinute(minuteStr string) bool {
	parts := strings.SplitN(minuteStr, "+", 2)
	base, _ := strconv.Atoi(parts[0])
	return base <= 45
}

// fixMojibake corrects accented player names stored as Latin-1 codepoints in the JSON.
func fixMojibake(s string) string {
	runes := []rune(s)
	b := make([]byte, len(runes))
	for i, r := range runes {
		if r > 0xFF {
			return s
		}
		b[i] = byte(r)
	}
	return string(b)
}

// --- raw JSON structs ---

type ofWorldcup struct {
	Matches []ofMatch `json:"matches"`
}

type ofMatch struct {
	Round  string   `json:"round"`
	Date   string   `json:"date"`
	Team1  string   `json:"team1"`
	Team2  string   `json:"team2"`
	Score  *ofScore `json:"score"`
	Goals1 []ofGoal `json:"goals1"`
	Goals2 []ofGoal `json:"goals2"`
	Group  string   `json:"group"`
	Ground string   `json:"ground"`
}

type ofScore struct {
	FT []int `json:"ft"`
	HT []int `json:"ht"`
}

type ofGoal struct {
	Name    string `json:"name"`
	Minute  string `json:"minute"`
	OwnGoal bool   `json:"owngoal"`
	Penalty bool   `json:"penalty"`
}
