package service

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"

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
	Name         string           `json:"name" binding:"required"`
	MatchType    string           `json:"match_type"`    // "1v1" or "2v2"; auto-set to "2v2" for round_robin_top4
	Format       string           `json:"format"`        // "classic" (default) | "round_robin_top4"
	PlayerIDs    []uuid.UUID      `json:"player_ids"`    // used for classic format
	Teams        []TeamInputEntry `json:"teams"`         // used for round_robin_top4; if omitted, auto-assigned
	AffectsScore *bool            `json:"affects_score"` // pointer so omitting defaults to true
	EntryFee     int              `json:"entry_fee"`
	KnockoutSize int              `json:"knockout_size"` // 2 = final only; 4 = semis+final+3rd (default); 0 = use default
}

// TeamInputEntry is a pair of player IDs forming a fixed team
type TeamInputEntry struct {
	Player1ID uuid.UUID `json:"player1_id" binding:"required"`
	Player2ID uuid.UUID `json:"player2_id" binding:"required"`
}

// TeamStanding holds computed group-stage standings for one team
type TeamStanding struct {
	TeamID  uuid.UUID  `json:"team_id"`
	Player1 model.User `json:"player1"`
	Player2 model.User `json:"player2"`
	Played  int        `json:"played"`
	Won     int        `json:"won"`
	Drawn   int        `json:"drawn"`
	Lost    int        `json:"lost"`
	GF      int        `json:"gf"`
	GA      int        `json:"ga"`
	GD      int        `json:"gd"`
	Points  int        `json:"points"`
	Seed    int        `json:"seed"` // 1-4 for top 4; 0 = not qualified
}

// TournamentDetailResponse wraps Tournament with computed standings
type TournamentDetailResponse struct {
	*model.Tournament
	Standings []TeamStanding `json:"standings,omitempty"`
}

// resolvedAffectsScore returns a non-nil *bool (defaults to true if nil)
func (r *CreateTournamentRequest) resolvedAffectsScore() *bool {
	if r.AffectsScore == nil {
		t := true
		return &t
	}
	return r.AffectsScore
}

// CreateTournament creates a new tournament with a generated schedule
func (s *TournamentService) CreateTournament(req *CreateTournamentRequest) (*model.Tournament, error) {
	if req.Format == "round_robin_top4" {
		return s.createRoundRobinTop4(req)
	}
	return s.createClassic(req)
}

func (s *TournamentService) createClassic(req *CreateTournamentRequest) (*model.Tournament, error) {
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

	seen := make(map[uuid.UUID]bool)
	for _, playerID := range req.PlayerIDs {
		if seen[playerID] {
			return nil, fmt.Errorf("duplicate player ID %s in tournament", playerID)
		}
		seen[playerID] = true
	}

	users := make([]*model.UserWithStats, 0, len(req.PlayerIDs))
	for _, id := range req.PlayerIDs {
		u, err := s.userRepo.GetByID(id)
		if err != nil {
			return nil, fmt.Errorf("player %s not found", id)
		}
		users = append(users, u)
	}

	participants := make([]model.TournamentParticipant, len(users))
	for i, u := range users {
		participants[i] = model.TournamentParticipant{
			UserID:               u.ID,
			TierSnapshot:         u.Tier,
			HandicapRateSnapshot: u.HandicapRate,
		}
	}

	baseUsers := make([]*model.User, len(users))
	for i, u := range users {
		baseUsers[i] = &u.User
	}

	var tournamentMatches []model.TournamentMatch
	var err error
	if req.MatchType == "1v1" {
		tournamentMatches, err = s.generate1v1Schedule(baseUsers)
	} else {
		tournamentMatches, err = s.generate2v2Schedule(baseUsers)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to generate schedule: %w", err)
	}

	tournament := &model.Tournament{
		Name:         req.Name,
		MatchType:    req.MatchType,
		Format:       "classic",
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

func (s *TournamentService) createRoundRobinTop4(req *CreateTournamentRequest) (*model.Tournament, error) {
	if len(req.Teams) < 4 {
		return nil, errors.New("round_robin_top4 requires at least 4 teams")
	}

	// Validate no duplicate players across teams
	seen := make(map[uuid.UUID]bool)
	for _, t := range req.Teams {
		for _, pid := range []uuid.UUID{t.Player1ID, t.Player2ID} {
			if seen[pid] {
				return nil, fmt.Errorf("player %s appears in more than one team", pid)
			}
			seen[pid] = true
		}
	}

	// Fetch all players and build TournamentTeam records
	teams := make([]model.TournamentTeam, 0, 5)
	for _, entry := range req.Teams {
		p1, err := s.userRepo.GetByID(entry.Player1ID)
		if err != nil {
			return nil, fmt.Errorf("player %s not found", entry.Player1ID)
		}
		p2, err := s.userRepo.GetByID(entry.Player2ID)
		if err != nil {
			return nil, fmt.Errorf("player %s not found", entry.Player2ID)
		}
		teams = append(teams, model.TournamentTeam{
			ID:                      uuid.New(),
			Player1ID:               p1.ID,
			Player2ID:               p2.ID,
			Player1HandicapSnapshot: p1.HandicapRate,
			Player2HandicapSnapshot: p2.HandicapRate,
			Player1:                 &p1.User,
			Player2:                 &p2.User,
		})
	}

	// Generate group-stage schedule
	slots := GenerateTeamSchedule(teams)
	matches := make([]model.TournamentMatch, 0, len(slots))
	for _, slot := range slots {
		t1id := slot.Team1.ID
		t2id := slot.Team2.ID
		h1 := teamHandicap([]uuid.UUID{slot.Team1.Player1ID, slot.Team1.Player2ID}, nil)
		h2 := teamHandicap([]uuid.UUID{slot.Team2.Player1ID, slot.Team2.Player2ID}, nil)
		// Use snapshots for handicap
		h1 = max64(slot.Team1.Player1HandicapSnapshot, slot.Team1.Player2HandicapSnapshot)
		h2 = max64(slot.Team2.Player1HandicapSnapshot, slot.Team2.Player2HandicapSnapshot)

		p2id := slot.Team1.Player2ID
		p4id := slot.Team2.Player2ID
		matches = append(matches, model.TournamentMatch{
			Round:          slot.Round,
			MatchOrder:     slot.Order,
			Stage:          "group",
			Team1TeamID:    &t1id,
			Team2TeamID:    &t2id,
			Team1Player1ID: slot.Team1.Player1ID,
			Team1Player2ID: &p2id,
			Team2Player1ID: slot.Team2.Player1ID,
			Team2Player2ID: &p4id,
			HandicapTeam1:  h1,
			HandicapTeam2:  h2,
			Status:         "pending",
		})
	}

	ks := req.KnockoutSize
	if ks != 2 && ks != 4 {
		ks = 4
	}

	tournament := &model.Tournament{
		Name:         req.Name,
		MatchType:    "2v2",
		Format:       "round_robin_top4",
		Status:       "active",
		AffectsScore: req.resolvedAffectsScore(),
		EntryFee:     req.EntryFee,
		KnockoutSize: ks,
		Teams:        teams,
		Matches:      matches,
	}

	if err := s.repo.Create(tournament); err != nil {
		return nil, fmt.Errorf("failed to create tournament: %w", err)
	}
	return s.repo.GetByID(tournament.ID)
}

func max64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
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

// GenerateKnockouts creates semifinal matches once all group matches are complete
func (s *TournamentService) GenerateKnockouts(tournamentID uuid.UUID) (*TournamentDetailResponse, error) {
	t, err := s.repo.GetByID(tournamentID)
	if err != nil {
		return nil, errors.New("tournament not found")
	}
	if t.Format != "round_robin_top4" {
		return nil, errors.New("generate-knockouts is only available for round_robin_top4 format")
	}

	// Check all group matches are completed
	for _, m := range t.Matches {
		if m.Stage == "group" && m.Status != "completed" {
			return nil, errors.New("all group stage matches must be completed before generating knockouts")
		}
	}

	// Check knockouts don't already exist
	hasKnockout, err := s.repo.HasKnockoutMatches(tournamentID)
	if err != nil {
		return nil, err
	}
	if hasKnockout {
		return nil, errors.New("knockout matches already generated for this tournament")
	}

	ks := t.KnockoutSize
	if ks != 2 && ks != 4 {
		ks = 4
	}

	standings := ComputeStandings(t.Teams, t.Matches, ks)
	if len(standings) < ks {
		return nil, fmt.Errorf("not enough teams to generate knockouts (need %d, have %d)", ks, len(standings))
	}

	// Tie check at the knockout boundary (only if more teams than the cutoff)
	if len(standings) > ks {
		boundary := standings[ks-1]
		next := standings[ks]
		if boundary.Points == next.Points && boundary.GD == next.GD && boundary.GF == next.GF {
			return nil, fmt.Errorf(
				"teams '%s & %s' and '%s & %s' are tied — resolve manually before generating knockouts",
				boundary.Player1.Name, boundary.Player2.Name,
				next.Player1.Name, next.Player2.Name,
			)
		}
	}

	buildKOMatch := func(stage string, s1, s2 TeamStanding, round, order int) (model.TournamentMatch, error) {
		t1id := s1.TeamID
		t2id := s2.TeamID
		tt1 := findTeam(t.Teams, s1.TeamID)
		tt2 := findTeam(t.Teams, s2.TeamID)
		if tt1 == nil || tt2 == nil {
			return model.TournamentMatch{}, errors.New("team not found when building knockout match")
		}
		h1 := max64(tt1.Player1HandicapSnapshot, tt1.Player2HandicapSnapshot)
		h2 := max64(tt2.Player1HandicapSnapshot, tt2.Player2HandicapSnapshot)
		p2id := tt1.Player2ID
		p4id := tt2.Player2ID
		return model.TournamentMatch{
			TournamentID:   tournamentID,
			Round:          round,
			MatchOrder:     order,
			Stage:          stage,
			Team1TeamID:    &t1id,
			Team2TeamID:    &t2id,
			Team1Player1ID: tt1.Player1ID,
			Team1Player2ID: &p2id,
			Team2Player1ID: tt2.Player1ID,
			Team2Player2ID: &p4id,
			HandicapTeam1:  h1,
			HandicapTeam2:  h2,
			Status:         "pending",
		}, nil
	}

	if ks == 2 {
		// Top 2: create final directly
		final, err := buildKOMatch("final", standings[0], standings[1], 6, 1)
		if err != nil {
			return nil, err
		}
		if err := s.db.Create(&final).Error; err != nil {
			return nil, fmt.Errorf("failed to create final match: %w", err)
		}
	} else {
		// Top 4: two semis (1st vs 4th, 2nd vs 3rd)
		pairs := []struct{ top, bottom int }{{0, 3}, {1, 2}}
		for order, pair := range pairs {
			semi, err := buildKOMatch("semi", standings[pair.top], standings[pair.bottom], 6, order+1)
			if err != nil {
				return nil, err
			}
			if err := s.db.Create(&semi).Error; err != nil {
				return nil, fmt.Errorf("failed to create semi match: %w", err)
			}
		}
	}

	return s.GetTournament(tournamentID)
}

func findTeam(teams []model.TournamentTeam, id uuid.UUID) *model.TournamentTeam {
	for i := range teams {
		if teams[i].ID == id {
			return &teams[i]
		}
	}
	return nil
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

	// Lock: block editing group matches once knockouts exist
	if tournament.Format == "round_robin_top4" && tm.Stage == "group" {
		hasKnockout, err := s.repo.HasKnockoutMatches(tournamentID)
		if err != nil {
			return nil, err
		}
		if hasKnockout {
			return nil, errors.New("group stage results are locked after knockout generation")
		}
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
		_ = s.matchService.DeleteMatch(match.ID)
		return nil, fmt.Errorf("failed to save match result: %w", err)
	}

	// After a semi is recorded: maybe create final + 3rd-place
	if tournament.Format == "round_robin_top4" && tm.Stage == "semi" {
		if err := s.maybeCreateFinalMatches(tournamentID); err != nil {
			return nil, err
		}
	}

	// After final is recorded: set champion
	if tournament.Format == "round_robin_top4" && tm.Stage == "final" {
		if err := s.setChampion(tournament, tm); err != nil {
			return nil, err
		}
	}

	return tm, nil
}

func (s *TournamentService) maybeCreateFinalMatches(tournamentID uuid.UUID) error {
	// Idempotent: skip if final already exists
	var finalCount int64
	s.db.Model(&model.TournamentMatch{}).
		Where("tournament_id = ? AND stage IN ('final','third_place')", tournamentID).
		Count(&finalCount)
	if finalCount > 0 {
		return nil
	}

	semis, err := s.repo.GetMatchesByStage(tournamentID, "semi")
	if err != nil {
		return err
	}
	if len(semis) < 2 || semis[0].Status != "completed" || semis[1].Status != "completed" {
		return nil
	}

	// Determine winner/loser team IDs for each semi
	winnerOf := func(m *model.TournamentMatch) *uuid.UUID {
		if m.EffectiveWinner == 1 {
			return m.Team1TeamID
		}
		return m.Team2TeamID
	}
	loserOf := func(m *model.TournamentMatch) *uuid.UUID {
		if m.EffectiveWinner == 1 {
			return m.Team2TeamID
		}
		return m.Team1TeamID
	}

	t, err := s.repo.GetByID(tournamentID)
	if err != nil {
		return err
	}

	buildMatch := func(stage string, teamAID, teamBID *uuid.UUID, order int) model.TournamentMatch {
		ttA := findTeam(t.Teams, *teamAID)
		ttB := findTeam(t.Teams, *teamBID)
		h1 := max64(ttA.Player1HandicapSnapshot, ttA.Player2HandicapSnapshot)
		h2 := max64(ttB.Player1HandicapSnapshot, ttB.Player2HandicapSnapshot)
		p2id := ttA.Player2ID
		p4id := ttB.Player2ID
		return model.TournamentMatch{
			TournamentID:   tournamentID,
			Round:          7,
			MatchOrder:     order,
			Stage:          stage,
			Team1TeamID:    teamAID,
			Team2TeamID:    teamBID,
			Team1Player1ID: ttA.Player1ID,
			Team1Player2ID: &p2id,
			Team2Player1ID: ttB.Player1ID,
			Team2Player2ID: &p4id,
			HandicapTeam1:  h1,
			HandicapTeam2:  h2,
			Status:         "pending",
		}
	}

	finalMatch := buildMatch("final", winnerOf(semis[0]), winnerOf(semis[1]), 1)
	thirdMatch := buildMatch("third_place", loserOf(semis[0]), loserOf(semis[1]), 2)

	if err := s.db.Create(&finalMatch).Error; err != nil {
		return fmt.Errorf("failed to create final match: %w", err)
	}
	if err := s.db.Create(&thirdMatch).Error; err != nil {
		return fmt.Errorf("failed to create third place match: %w", err)
	}
	return nil
}

func (s *TournamentService) setChampion(tournament *model.Tournament, finalMatch *model.TournamentMatch) error {
	var championTeamID *uuid.UUID
	if finalMatch.EffectiveWinner == 1 {
		championTeamID = finalMatch.Team1TeamID
	} else if finalMatch.EffectiveWinner == 2 {
		championTeamID = finalMatch.Team2TeamID
	}
	if championTeamID == nil {
		return nil // draw — no champion set
	}
	tournament.ChampionTeamID = championTeamID
	return s.repo.Update(tournament)
}

// GetTournament returns a tournament by ID; for round_robin_top4 includes standings
func (s *TournamentService) GetTournament(id uuid.UUID) (*TournamentDetailResponse, error) {
	t, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	resp := &TournamentDetailResponse{Tournament: t}
	if t.Format == "round_robin_top4" {
		ks := t.KnockoutSize
		if ks != 2 && ks != 4 {
			ks = 4
		}
		resp.Standings = ComputeStandings(t.Teams, t.Matches, ks)
	}
	return resp, nil
}

// ComputeStandings calculates group-stage standings from completed matches
func ComputeStandings(teams []model.TournamentTeam, matches []model.TournamentMatch, knockoutSize int) []TeamStanding {
	m := make(map[uuid.UUID]*TeamStanding, len(teams))
	for _, t := range teams {
		p1 := model.User{}
		p2 := model.User{}
		if t.Player1 != nil {
			p1 = *t.Player1
		}
		if t.Player2 != nil {
			p2 = *t.Player2
		}
		m[t.ID] = &TeamStanding{TeamID: t.ID, Player1: p1, Player2: p2}
	}

	for _, match := range matches {
		if match.Stage != "group" || match.Status != "completed" {
			continue
		}
		if match.Team1TeamID == nil || match.Team2TeamID == nil {
			continue
		}
		s1 := m[*match.Team1TeamID]
		s2 := m[*match.Team2TeamID]
		if s1 == nil || s2 == nil {
			continue
		}
		if match.ActualScore1 != nil && match.ActualScore2 != nil {
			s1.GF += *match.ActualScore1
			s1.GA += *match.ActualScore2
			s2.GF += *match.ActualScore2
			s2.GA += *match.ActualScore1
		}
		s1.Played++
		s2.Played++
		switch match.EffectiveWinner {
		case 1:
			s1.Won++; s2.Lost++; s1.Points += 3
		case 2:
			s2.Won++; s1.Lost++; s2.Points += 3
		default:
			s1.Drawn++; s2.Drawn++; s1.Points++; s2.Points++
		}
	}

	result := make([]TeamStanding, 0, len(teams))
	for _, s := range m {
		s.GD = s.GF - s.GA
		result = append(result, *s)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Points != result[j].Points {
			return result[i].Points > result[j].Points
		}
		if result[i].GD != result[j].GD {
			return result[i].GD > result[j].GD
		}
		return result[i].GF > result[j].GF
	})
	for i := range result {
		if i < knockoutSize {
			result[i].Seed = i + 1
		}
	}
	return result
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
