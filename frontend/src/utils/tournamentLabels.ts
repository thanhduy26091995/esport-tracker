import type { MatchType } from '@/types/match'
import type { TournamentStatus } from '@/types/tournament'
import { translate } from '@/utils/i18n'

export function getTournamentStatusLabel(status: TournamentStatus): string {
  return status === 'completed'
    ? translate('tournaments.statusCompleted')
    : translate('tournaments.statusActive')
}

export function getMatchTypeLabel(matchType: MatchType): string {
  return matchType === '2v2'
    ? translate('matches.types.twoVsTwo')
    : translate('matches.types.oneVsOne')
}

export function getTournamentAffectsScoreLabel(affectsScore: boolean): string {
  return affectsScore
    ? translate('tournaments.affectsScoreYes')
    : translate('tournaments.affectsScoreNo')
}