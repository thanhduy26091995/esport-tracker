package service

import "github.com/duyb/esport-score-tracker/internal/model"

// TeamMatchSlot is a scheduled match between two fixed teams
type TeamMatchSlot struct {
	Team1 model.TournamentTeam
	Team2 model.TournamentTeam
	Round int
	Order int
}

// GenerateTeamSchedule generates a round-robin schedule for a fixed list of teams.
// Uses the existing GenerateRoundRobin polygon-rotation algorithm and maps
// team indices back to TournamentTeam values.
// For 5 teams: 5 rounds × 2 real matches = 10 matches total; each team plays 4 times.
func GenerateTeamSchedule(teams []model.TournamentTeam) []TeamMatchSlot {
	rounds := GenerateRoundRobin(len(teams))
	var schedule []TeamMatchSlot
	for roundIdx, pairs := range rounds {
		order := 1
		for _, pair := range pairs {
			// GenerateRoundRobin already excludes bye pairs (A==-1 || B==-1)
			schedule = append(schedule, TeamMatchSlot{
				Team1: teams[pair.A],
				Team2: teams[pair.B],
				Round: roundIdx + 1,
				Order: order,
			})
			order++
		}
	}
	return schedule
}
