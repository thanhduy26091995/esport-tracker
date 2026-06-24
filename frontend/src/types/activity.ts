export interface ActivityEvent {
  type: 'bet_placed'
  user_id: string
  user_name: string
  bet_type: 'handicap' | 'exact_score' | 'over_under'
  selection: string
  stake: number
  match_id: string
  team_home: string
  team_away: string
}
