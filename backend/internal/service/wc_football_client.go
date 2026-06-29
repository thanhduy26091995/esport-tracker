package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
)

const footballDataBaseURL = "https://api.football-data.org/v4"

type footballClient struct {
	apiKey     string
	httpClient *http.Client
}

func newFootballClient() *footballClient {
	return &footballClient{
		apiKey:     os.Getenv("FOOTBALL_DATA_API_KEY"),
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// footballMatchesResponse is the raw API response shape.
type footballMatchesResponse struct {
	Matches []footballMatch `json:"matches"`
}

type footballMatch struct {
	ID       int    `json:"id"`
	Status   string `json:"status"`
	Stage    string `json:"stage"`
	Group    string `json:"group"`
	UtcDate  string `json:"utcDate"`
	HomeTeam struct {
		Name string `json:"name"`
		TLA  string `json:"tla"`
	} `json:"homeTeam"`
	AwayTeam struct {
		Name string `json:"name"`
		TLA  string `json:"tla"`
	} `json:"awayTeam"`
	Score struct {
		Duration    string `json:"duration"` // REGULAR | EXTRA_TIME | PENALTY_SHOOTOUT
		FullTime    struct {
			Home *int `json:"home"`
			Away *int `json:"away"`
		} `json:"fullTime"`
		RegularTime struct {
			Home *int `json:"home"`
			Away *int `json:"away"`
		} `json:"regularTime"`
	} `json:"score"`
	Venue string `json:"venue"`
}

func (c *footballClient) FetchWCMatches() ([]model.WcMatch, error) {
	req, err := http.NewRequest("GET", footballDataBaseURL+"/competitions/WC/matches", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("X-Auth-Token", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch matches: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("football-data.org returned %d: %s", resp.StatusCode, string(body))
	}

	var raw footballMatchesResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	matches := make([]model.WcMatch, 0, len(raw.Matches))
	for _, m := range raw.Matches {
		matchDate, err := time.Parse(time.RFC3339, m.UtcDate)
		if err != nil {
			continue
		}

		status := mapStatus(m.Status)
		stage := mapStage(m.Stage)

		// For knockout matches that go to extra time or penalties, use regularTime
		// (90-min score) so bets are settled on the result within regular time only.
		homeScore, awayScore := selectBettingScore(
			m.Score.Duration,
			m.Score.FullTime.Home, m.Score.FullTime.Away,
			m.Score.RegularTime.Home, m.Score.RegularTime.Away,
		)

		wm := model.WcMatch{
			ExternalID:   fmt.Sprintf("%d", m.ID),
			HomeTeam:     m.HomeTeam.Name,
			AwayTeam:     m.AwayTeam.Name,
			HomeTeamCode: m.HomeTeam.TLA,
			AwayTeamCode: m.AwayTeam.TLA,
			MatchDate:    matchDate,
			GroupName:    normalizeGroupName(m.Group),
			Stage:        stage,
			Venue:        m.Venue,
			Status:       status,
			HomeScore:    homeScore,
			AwayScore:    awayScore,
			PredictionsLockedAt: &matchDate, // auto-lock at kickoff time
		}
		matches = append(matches, wm)
	}
	return matches, nil
}

// selectBettingScore returns the score used for bet settlement.
// For matches that went to extra time or penalties, regularTime (90-min score) is used
// so bets are settled on the result within regulation only.
func selectBettingScore(duration string, ftHome, ftAway, rtHome, rtAway *int) (*int, *int) {
	if (duration == "EXTRA_TIME" || duration == "PENALTY_SHOOTOUT") && rtHome != nil && rtAway != nil {
		return rtHome, rtAway
	}
	return ftHome, ftAway
}

// normalizeGroupName converts API format "GROUP_A" → "Group A".
func normalizeGroupName(group string) string {
	if group == "" {
		return ""
	}
	parts := strings.SplitN(strings.ToUpper(group), "_", 2)
	if len(parts) == 2 && parts[0] == "GROUP" {
		return "Group " + parts[1]
	}
	return group
}

func mapStatus(s string) string {
	switch strings.ToUpper(s) {
	case "IN_PLAY", "PAUSED", "HALFTIME":
		return model.WcStatusLive
	case "FINISHED":
		return model.WcStatusCompleted
	case "CANCELLED", "POSTPONED", "SUSPENDED":
		return model.WcStatusCancelled
	default:
		return model.WcStatusScheduled
	}
}

func mapStage(s string) string {
	switch strings.ToUpper(s) {
	case "GROUP_STAGE":
		return model.WcStageGroup
	case "ROUND_OF_32":
		return model.WcStageR32
	case "ROUND_OF_16":
		return model.WcStageR16
	case "QUARTER_FINALS":
		return model.WcStageQF
	case "SEMI_FINALS":
		return model.WcStageSF
	case "FINAL":
		return model.WcStageFinal
	case "THIRD_PLACE":
		return model.WcStageThirdPlace
	default:
		return model.WcStageGroup
	}
}
