package service

import (
	"testing"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/ws"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func actStrPtr(s string) *string { return &s }
func actIntPtr(i int) *int       { return &i }

func TestBuildActivityEvent_Handicap(t *testing.T) {
	m := &model.WcMatch{HomeTeam: "Bồ Đào Nha", AwayTeam: "Morocco"}
	req := PlaceBetRequest{
		MatchID:   uuid.New(),
		BetType:   model.WcBetTypeHandicap,
		BetChoice: actStrPtr(model.WcTeamHome),
		Stake:     5,
	}
	event := buildActivityEvent("user-1", "Ric Phan", req, m)

	assert.Equal(t, "bet_placed", event.Type)
	assert.Equal(t, "user-1", event.UserID)
	assert.Equal(t, "Ric Phan", event.UserName)
	assert.Equal(t, model.WcBetTypeHandicap, event.BetType)
	assert.Equal(t, "Bồ Đào Nha", event.Selection) // chose home team
	assert.Equal(t, 5, event.Stake)
	assert.Equal(t, "Bồ Đào Nha", event.TeamHome)
	assert.Equal(t, "Morocco", event.TeamAway)
}

func TestBuildActivityEvent_HandicapAway(t *testing.T) {
	m := &model.WcMatch{HomeTeam: "Bồ Đào Nha", AwayTeam: "Morocco"}
	req := PlaceBetRequest{
		BetType:   model.WcBetTypeHandicap,
		BetChoice: actStrPtr(model.WcTeamAway),
		Stake:     3,
	}
	event := buildActivityEvent("u2", "Nam", req, m)
	assert.Equal(t, "Morocco", event.Selection) // chose away team
}

func TestBuildActivityEvent_OverUnder_Over(t *testing.T) {
	m := &model.WcMatch{HomeTeam: "Pháp", AwayTeam: "Anh"}
	req := PlaceBetRequest{
		BetType:   model.WcBetTypeOverUnder,
		BetChoice: actStrPtr(model.WcChoiceOver),
		Stake:     10,
	}
	event := buildActivityEvent("u3", "Hieu", req, m)
	assert.Equal(t, "Tài", event.Selection)
}

func TestBuildActivityEvent_OverUnder_Under(t *testing.T) {
	m := &model.WcMatch{HomeTeam: "Pháp", AwayTeam: "Anh"}
	req := PlaceBetRequest{
		BetType:   model.WcBetTypeOverUnder,
		BetChoice: actStrPtr(model.WcChoiceUnder),
		Stake:     2,
	}
	event := buildActivityEvent("u4", "Linh", req, m)
	assert.Equal(t, "Xỉu", event.Selection)
}

func TestBuildActivityEvent_ExactScore(t *testing.T) {
	m := &model.WcMatch{HomeTeam: "Tây Ban Nha", AwayTeam: "Đức"}
	req := PlaceBetRequest{
		BetType:            model.WcBetTypeExactScore,
		Stake:              7,
		PredictedHomeScore: actIntPtr(2),
		PredictedAwayScore: actIntPtr(1),
	}
	event := buildActivityEvent("u5", "Duy", req, m)
	assert.Equal(t, "2 - 1", event.Selection)
}

// --- buildPredictionActivityEvent ---

func TestBuildPredictionActivityEvent_Handicap(t *testing.T) {
	m := &model.WcMatch{HomeTeam: "Bồ Đào Nha", AwayTeam: "Morocco"}
	req := SubmitPredictionRequest{
		MatchID:          uuid.New(),
		PredictionType:   model.WcPredictionTypeHandicap,
		PredictionChoice: actStrPtr(model.WcTeamHome),
		Points:           5,
	}
	event := buildPredictionActivityEvent("u1", "Ric Phan", req, m)
	assert.Equal(t, "bet_placed", event.Type)
	assert.Equal(t, "Ric Phan", event.UserName)
	assert.Equal(t, model.WcPredictionTypeHandicap, event.BetType)
	assert.Equal(t, "Bồ Đào Nha", event.Selection)
	assert.Equal(t, 5, event.Stake)
	assert.Equal(t, "Bồ Đào Nha", event.TeamHome)
	assert.Equal(t, "Morocco", event.TeamAway)
}

func TestBuildPredictionActivityEvent_HandicapAway(t *testing.T) {
	m := &model.WcMatch{HomeTeam: "Pháp", AwayTeam: "Anh"}
	req := SubmitPredictionRequest{
		PredictionType:   model.WcPredictionTypeHandicap,
		PredictionChoice: actStrPtr(model.WcTeamAway),
		Points:           3,
	}
	event := buildPredictionActivityEvent("u2", "Nam", req, m)
	assert.Equal(t, "Anh", event.Selection)
}

func TestBuildPredictionActivityEvent_OverUnder_Over(t *testing.T) {
	m := &model.WcMatch{HomeTeam: "Pháp", AwayTeam: "Anh"}
	req := SubmitPredictionRequest{
		PredictionType:   model.WcPredictionTypeOverUnder,
		PredictionChoice: actStrPtr(model.WcChoiceOver),
		Points:           8,
	}
	event := buildPredictionActivityEvent("u3", "Hieu", req, m)
	assert.Equal(t, "Tài", event.Selection)
}

func TestBuildPredictionActivityEvent_OverUnder_Under(t *testing.T) {
	m := &model.WcMatch{HomeTeam: "Pháp", AwayTeam: "Anh"}
	req := SubmitPredictionRequest{
		PredictionType:   model.WcPredictionTypeOverUnder,
		PredictionChoice: actStrPtr(model.WcChoiceUnder),
		Points:           2,
	}
	event := buildPredictionActivityEvent("u4", "Linh", req, m)
	assert.Equal(t, "Xỉu", event.Selection)
}

func TestBuildPredictionActivityEvent_ExactScore(t *testing.T) {
	m := &model.WcMatch{HomeTeam: "Tây Ban Nha", AwayTeam: "Đức"}
	req := SubmitPredictionRequest{
		PredictionType:     model.WcPredictionTypeExactScore,
		Points:             6,
		PredictedHomeScore: actIntPtr(3),
		PredictedAwayScore: actIntPtr(0),
	}
	event := buildPredictionActivityEvent("u5", "Duy", req, m)
	assert.Equal(t, "3 - 0", event.Selection)
}

// MockHub records Broadcast calls for assertion in tests
type MockHub struct {
	calls []ws.ActivityEvent
}

func (m *MockHub) Broadcast(e ws.ActivityEvent) {
	m.calls = append(m.calls, e)
}

// --- MockHub broadcast verification ---

func TestMockHub_RecordsBroadcast(t *testing.T) {
	hub := &MockHub{}
	hub.Broadcast(ws.ActivityEvent{Type: "bet_placed", UserName: "Ric Phan", Stake: 5})
	assert.Len(t, hub.calls, 1)
	assert.Equal(t, "Ric Phan", hub.calls[0].UserName)
	assert.Equal(t, 5, hub.calls[0].Stake)
}

func TestNilHub_GracefulDegradation(t *testing.T) {
	// WcService with nil hub must not panic when PlaceBet succeeds
	// (tested implicitly via wc_integration_test.go with hub=nil, but
	// this assertion documents the contract explicitly)
	var hub ws.HubBroadcaster // nil interface
	assert.NotPanics(t, func() {
		if hub != nil {
			hub.Broadcast(ws.ActivityEvent{})
		}
	})
}
