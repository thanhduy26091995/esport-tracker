package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
)

const fdAPIBaseURL = "https://api.football-data.org/v4"

type WcFootballDataClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewWcFootballDataClient(apiKey string) *WcFootballDataClient {
	return &WcFootballDataClient{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *WcFootballDataClient) GetWCScorers(limit int) ([]model.WcTournamentScorer, time.Time, error) {
	url := fmt.Sprintf("%s/competitions/WC/scorers?limit=%d", fdAPIBaseURL, limit)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, time.Time{}, err
	}
	req.Header.Set("X-Auth-Token", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, time.Time{}, fmt.Errorf("football-data.org returned %d", resp.StatusCode)
	}

	var result fdScorersResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, time.Time{}, err
	}

	scorers := make([]model.WcTournamentScorer, 0, len(result.Scorers))
	for i, s := range result.Scorers {
		scorers = append(scorers, model.WcTournamentScorer{
			Rank:          i + 1,
			PlayerName:    s.Player.Name,
			TeamName:      s.Team.Name,
			TeamCode:      s.Team.TLA,
			TeamCrest:     s.Team.Crest,
			Goals:         s.Goals,
			Assists:       s.Assists,
			PlayedMatches: s.PlayedMatches,
		})
	}
	return scorers, time.Now(), nil
}

type fdScorersResponse struct {
	Scorers []fdScorer `json:"scorers"`
}

type fdScorer struct {
	Player        fdPlayer `json:"player"`
	Team          fdTeam   `json:"team"`
	Goals         int      `json:"goals"`
	Assists       *int     `json:"assists"`
	PlayedMatches int      `json:"playedMatches"`
}

type fdPlayer struct {
	Name string `json:"name"`
}

type fdTeam struct {
	Name  string `json:"name"`
	TLA   string `json:"tla"`
	Crest string `json:"crest"`
}
