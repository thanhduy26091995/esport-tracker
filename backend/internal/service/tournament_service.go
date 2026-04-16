package service

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TournamentService struct {
	repo         *repository.TournamentRepository
	userRepo     *repository.UserRepository
	matchService *MatchService
	db           *gorm.DB
}

func NewTournamentService(
	repo *repository.TournamentRepository,
	userRepo *repository.UserRepository,
	matchService *MatchService,
	db *gorm.DB,
) *TournamentService {
	return &TournamentService{repo: repo, userRepo: userRepo, matchService: matchService, db: db}
}

// CreateTournamentRequest is the input for creating a tournament
type CreateTournamentRequest struct {
	Name         string      `json:"name" binding:"required"`
	MatchType    string      `json:"match_type" binding:"required"` // "1v1" or "2v2"
	PlayerIDs    []uuid.UUID `json:"player_ids" binding:"required"`
	AffectsScore *bool       `json:"affects_score"` // pointer so omitting defaults to true
	EntryFee     int         `json:"entry_fee"`
}

// resolvedAffectsScore returns a non-nil *bool (defaults to true if nil)
func (r *CreateTournamentRequest) resolvedAffectsScore() *bool {
	if r.AffectsScore == nil {
		t := true
		return &t
	}
	return r.AffectsScore
}

// CreateTournament creates a new tournament with a generated round-robin schedule
func (s *TournamentService) CreateTournament(req *CreateTournamentRequest) (*model.Tournament, error) {
	if req.MatchType != "1v1" && req.MatchType != "2v2" {
		return nil, errors.New("match_type must be '1v1' or '2v2'")
	}
	if len(req.PlayerIDs) < 3 {
		return nil, errors.New("tournament requires at least 3 players")
	}
	if len(req.PlayerIDs) > 16 {
		return nil, errors.New("tournament supports at most 16 players")
	}
	if req.MatchType == "2v2" && len(req.PlayerIDs)%2 != 0 {
		return nil, errors.New("2v2 requires an even number of players")
	}
	if req.MatchType == "2v2" && len(req.PlayerIDs) < 4 {
		return nil, errors.New("2v2 requires at least 4 players")
	}

	// Check for duplicate player IDs
	seen := make(map[uuid.UUID]bool)
	for _, playerID := range req.PlayerIDs {
		if seen[playerID] {
			return nil, fmt.Errorf("duplicate player ID %s in tournament", playerID)
		}
		seen[playerID] = true
	}

	// Fetch all players
	users := make([]*model.User, 0, len(req.PlayerIDs))
	for _, id := range req.PlayerIDs {
		u, err := s.userRepo.GetByID(id)
		if err != nil {
			return nil, fmt.Errorf("player %s not found", id)
		}
		users = append(users, u)
	}

	// Build participants with tier snapshot
	participants := make([]model.TournamentParticipant, len(users))
	for i, u := range users {
		participants[i] = model.TournamentParticipant{
			UserID:               u.ID,
			TierSnapshot:         u.Tier,
			HandicapRateSnapshot: u.HandicapRate,
		}
	}

	// Generate schedule based on match type
	var tournamentMatches []model.TournamentMatch
	var err error

	if req.MatchType == "1v1" {
		tournamentMatches, err = s.generate1v1Schedule(users)
	} else {
		tournamentMatches, err = s.generate2v2Schedule(users)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to generate schedule: %w", err)
	}

	tournament := &model.Tournament{
		Name:         req.Name,
		MatchType:    req.MatchType,
		Status:       "active",
		AffectsScore: req.resolvedAffectsScore(),
		EntryFee:     req.EntryFee,
		Participants: participants,
		Matches:      tournamentMatches,
	}

	if err := s.repo.Create(tournament); err != nil {
		return nil, fmt.Errorf("failed to create tournament: %w", err)
	}

	return s.repo.GetByID(tournament.ID)
}

func (s *TournamentService) generate1v1Schedule(users []*model.User) ([]model.TournamentMatch, error) {
	shuffled := make([]*model.User, len(users))
	copy(shuffled, users)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

	n := len(shuffled)
	rounds := GenerateRoundRobin(n)

	var matches []model.TournamentMatch
	for ri, round := range rounds {
		for mi, pair := range round {
			p1 := shuffled[pair.A]
			p2 := shuffled[pair.B]
			matches = append(matches, model.TournamentMatch{
				Round:          ri + 1,
				MatchOrder:     mi + 1,
				Team1Player1ID: p1.ID,
				Team2Player1ID: p2.ID,
				HandicapTeam1:  p1.HandicapRate,
				HandicapTeam2:  p2.HandicapRate,
				Status:         "pending",
			})
		}
	}
	return matches, nil
}

func (s *TournamentService) generate2v2Schedule(users []*model.User) ([]model.TournamentMatch, error) {
	if len(users) < 4 {
		return nil, errors.New("2v2 requires at least 4 players")
	}

	userMap := make(map[uuid.UUID]*model.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	slots := GenerateSchedule2v2(users)
	if len(slots) == 0 {
		return nil, errors.New("failed to generate 2v2 schedule")
	}

	matches := make([]model.TournamentMatch, 0, len(slots))
	for ri, slot := range slots {
		t1IDs := slot.Team1[:]
		t2IDs := slot.Team2[:]

		h1 := teamHandicap(t1IDs, userMap)
		h2 := teamHandicap(t2IDs, userMap)

		p2ID := slot.Team1[1]
		p4ID := slot.Team2[1]
		m := model.TournamentMatch{
			Round:          ri + 1,
			MatchOrder:     1,
			Team1Player1ID: slot.Team1[0],
			Team1Player2ID: &p2ID,
			Team2Player1ID: slot.Team2[0],
			Team2Player2ID: &p4ID,
			HandicapTeam1:  h1,
			HandicapTeam2:  h2,
			Status:         "pending",
		}
		matches = append(matches, m)
	}
	return matches, nil
}

// teamHandicap returns the maximum (most penalizing) handicap_rate in a team.
// handicap_rate is stored as a positive number; higher value = more penalty.
// Using max ensures the Pro player's handicap applies to the whole team.
func teamHandicap(playerIDs []uuid.UUID, userMap map[uuid.UUID]*model.User) float64 {
	max := 0.0
	for _, id := range playerIDs {
		if u, ok := userMap[id]; ok && u.HandicapRate > max {
			max = u.HandicapRate
		}
	}
	return max
}

// RecordMatchResultRequest is the input for recording a result
type RecordMatchResultRequest struct {
	ActualScore1 int    `json:"actual_score1"`
	ActualScore2 int    `json:"actual_score2"`
	RecordedBy   string `json:"recorded_by"`
}

// RecordMatchResult records the result of a tournament match
func (s *TournamentService) RecordMatchResult(tournamentID, matchID uuid.UUID, req *RecordMatchResultRequest) (*model.TournamentMatch, error) {
	tournament, err := s.repo.GetByID(tournamentID)
	if err != nil {
		return nil, errors.New("tournament not found")
	}

	tm, err := s.repo.GetMatch(matchID)
	if err != nil {
		return nil, errors.New("tournament match not found")
	}
	if tm.TournamentID != tournamentID {
		return nil, errors.New("match does not belong to this tournament")
	}

	// Revert previous linked regular match if re-recording
	if tm.Status == "completed" && tm.MatchID != nil {
		if err := s.matchService.DeleteMatch(*tm.MatchID); err != nil {
			return nil, fmt.Errorf("failed to revert previous match: %w", err)
		}
		tm.MatchID = nil
	}

	effectiveWinner := EffectiveWinner(req.ActualScore1, req.ActualScore2, tm.HandicapTeam1, tm.HandicapTeam2)

	matchWinnerTeam := 0
	if (tournament.AffectsScore == nil || *tournament.AffectsScore) && effectiveWinner != 0 {
		matchWinnerTeam = effectiveWinner
	}

	team1 := []uuid.UUID{tm.Team1Player1ID}
	if tm.Team1Player2ID != nil {
		team1 = append(team1, *tm.Team1Player2ID)
	}
	team2 := []uuid.UUID{tm.Team2Player1ID}
	if tm.Team2Player2ID != nil {
		team2 = append(team2, *tm.Team2Player2ID)
	}

	matchReq := &CreateMatchRequest{
		MatchType:         tournament.MatchType,
		Team1:             team1,
		Team2:             team2,
		WinnerTeam:        matchWinnerTeam,
		TournamentMatchID: &tm.ID,
	}
	match, matchErr := s.matchService.CreateMatch(matchReq)
	if matchErr != nil {
		return nil, fmt.Errorf("failed to create regular match record: %w", matchErr)
	}
	tm.MatchID = &match.ID

	tm.ActualScore1 = &req.ActualScore1
	tm.ActualScore2 = &req.ActualScore2
	tm.EffectiveWinner = effectiveWinner
	tm.Status = "completed"

	if err := s.repo.SaveMatch(tm); err != nil {
		// Compensate: revert the just-created regular match to avoid orphaned score changes
		_ = s.matchService.DeleteMatch(match.ID)
		return nil, fmt.Errorf("failed to save match result: %w", err)
	}

	return tm, nil
}

// GetTournament returns a tournament by ID with all relations
func (s *TournamentService) GetTournament(id uuid.UUID) (*model.Tournament, error) {
	return s.repo.GetByID(id)
}

// GetAllTournaments returns all tournaments
func (s *TournamentService) GetAllTournaments() ([]*model.Tournament, error) {
	return s.repo.GetAll()
}

// CompleteTournament marks a tournament as completed
func (s *TournamentService) CompleteTournament(id uuid.UUID) (*model.Tournament, error) {
	tournament, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("tournament not found")
	}
	if tournament.Status == "completed" {
		return tournament, nil
	}
	tournament.Status = "completed"
	if err := s.repo.Update(tournament); err != nil {
		return nil, fmt.Errorf("failed to complete tournament: %w", err)
	}
	return tournament, nil
}

// DeleteTournament deletes a tournament, reverting all linked regular matches
func (s *TournamentService) DeleteTournament(id uuid.UUID) error {
	tournament, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("tournament not found")
	}

	for _, tm := range tournament.Matches {
		if tm.MatchID != nil {
			if err := s.matchService.DeleteMatch(*tm.MatchID); err != nil {
				return fmt.Errorf("failed to revert match %s: %w", *tm.MatchID, err)
			}
		}
	}

	return s.repo.Delete(id)
}
