package service

import (
	"fmt"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/ws"
)

// buildActivityEvent constructs the WebSocket payload broadcast after a successful bet.
func buildActivityEvent(userID, userName string, req PlaceBetRequest, m *model.WcMatch) ws.ActivityEvent {
	selection := buildSelection(req, m)
	return ws.ActivityEvent{
		Type:      "bet_placed",
		UserID:    userID,
		UserName:  userName,
		BetType:   req.BetType,
		Selection: selection,
		Stake:     req.Stake,
		MatchID:   req.MatchID.String(),
		TeamHome:  m.HomeTeam,
		TeamAway:  m.AwayTeam,
	}
}

func buildSelection(req PlaceBetRequest, m *model.WcMatch) string {
	return buildSelectionFromParts(req.BetType, req.BetChoice, req.PredictedHomeScore, req.PredictedAwayScore, m)
}

// buildPredictionActivityEvent constructs the WebSocket payload after a successful prediction.
func buildPredictionActivityEvent(userID, userName string, req SubmitPredictionRequest, m *model.WcMatch) ws.ActivityEvent {
	selection := buildSelectionFromParts(req.PredictionType, req.PredictionChoice, req.PredictedHomeScore, req.PredictedAwayScore, m)
	return ws.ActivityEvent{
		Type:      "bet_placed",
		UserID:    userID,
		UserName:  userName,
		BetType:   req.PredictionType,
		Selection: selection,
		Stake:     req.Points,
		MatchID:   req.MatchID.String(),
		TeamHome:  m.HomeTeam,
		TeamAway:  m.AwayTeam,
	}
}

// buildCancelActivityEvent constructs the WebSocket payload broadcast after a bet is deleted.
func buildCancelActivityEvent(userID, userName, betType, teamHome, teamAway, matchID string) ws.ActivityEvent {
	return ws.ActivityEvent{
		Type:     "bet_cancelled",
		UserID:   userID,
		UserName: userName,
		BetType:  betType,
		MatchID:  matchID,
		TeamHome: teamHome,
		TeamAway: teamAway,
	}
}

func buildSelectionFromParts(betType string, choice *string, homeScore, awayScore *int, m *model.WcMatch) string {
	switch betType {
	case model.WcPredictionTypeHandicap:
		if choice != nil && *choice == model.WcTeamHome {
			return m.HomeTeam
		}
		return m.AwayTeam
	case model.WcPredictionTypeOverUnder:
		if choice != nil && *choice == model.WcChoiceOver {
			return "Tài"
		}
		return "Xỉu"
	case model.WcPredictionTypeExactScore:
		if homeScore != nil && awayScore != nil {
			return fmt.Sprintf("%d - %d", *homeScore, *awayScore)
		}
	}
	return ""
}
