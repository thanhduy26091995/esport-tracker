export type ConfigKey = 'debt_threshold' | 'point_to_vnd' | 'fund_split_percent' | 'auto_settlement' | 'points_per_win' | 'min_matches_for_tier'

export interface Config {
  key: ConfigKey
  value: string
  description: string
}

export interface UpdateConfigRequest {
  value: string
}
