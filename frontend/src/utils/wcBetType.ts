import { useI18n } from 'vue-i18n'

export const WC_BET_TYPES = ['handicap', 'exact_score', 'over_under', 'corner', 'custom'] as const
export type WcBetType = (typeof WC_BET_TYPES)[number]

// Single source of truth for bet type → i18n key mapping.
// Adding a new type: add one entry here + one key in vi.json under wc.betType*.
export const BET_TYPE_I18N_KEYS: Record<WcBetType, string> = {
  handicap: 'wc.betTypeHandicap',
  exact_score: 'wc.betTypeExactScore',
  over_under: 'wc.betTypeOverUnder',
  corner: 'wc.betTypeCorner',
  custom: 'wc.betTypeCustom',
}

export function useWcBetTypeLabel() {
  const { t } = useI18n()
  return (type: string) => {
    const key = BET_TYPE_I18N_KEYS[type as WcBetType]
    return key ? t(key) : type
  }
}
