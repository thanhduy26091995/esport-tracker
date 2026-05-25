export type HighlightSection = 'trending' | 'daily_recap' | 'competitive' | 'social'

export interface Highlight {
  player_id: string
  player_name: string
  second_name?: string
  type: string
  section: HighlightSection
  emoji: string
  message: string
  value: number
  priority: number
}

export interface HighlightsResponse {
  trending: Highlight[]
  daily_recap: Highlight[]
  competitive: Highlight[]
  social: Highlight[]
  generated_at: string
}
