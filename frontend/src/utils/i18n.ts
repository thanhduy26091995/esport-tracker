import type { ApiError } from '@/types/api'
import { i18n } from '@/plugins/i18n'

export type TranslationValues = Record<string, string | number | boolean | null | undefined>

const ERROR_CODE_KEYS: Record<string, string> = {
  VALIDATION_ERROR: 'errors.validation',
  INTERNAL_ERROR: 'errors.internal',
  NOT_FOUND: 'errors.notFound',
  CONFLICT: 'errors.conflict',
  INVALID_ID: 'errors.invalidId',
  INVALID_UUID: 'errors.invalidId',
  INVALID_MATCH_TYPE: 'errors.invalidMatchType',
  INVALID_WINNER_TEAM: 'errors.invalidWinnerTeam',
  DUPLICATE_PLAYER: 'errors.duplicatePlayer',
  INVALID_TEAM_SIZE: 'errors.invalidTeamSize',
  USER_NOT_FOUND: 'errors.userNotFound',
  MATCH_LOCKED: 'errors.matchLocked',
  INSUFFICIENT_BALANCE: 'errors.insufficientBalance',
  CREATE_FAILED: 'errors.actionFailed',
  DELETE_FAILED: 'errors.actionFailed',
  COMPLETE_FAILED: 'errors.actionFailed',
  RECORD_FAILED: 'errors.actionFailed',
}

export function translate(key: string, values?: TranslationValues): string {
  return i18n.global.t(key, values ?? {})
}

export function translateError(code?: string | null, fallbackMessage?: string | null): string {
  if (code) {
    const key = ERROR_CODE_KEYS[code]
    if (key) {
      return translate(key)
    }
  }

  if (fallbackMessage && fallbackMessage.trim().length > 0) {
    return fallbackMessage
  }

  return translate('errors.unknown')
}

export function getApiError(error: unknown): ApiError | undefined {
  const candidate = (error as { response?: { data?: { error?: ApiError; code?: string; message?: string } } })?.response?.data

  if (candidate?.error) {
    return candidate.error
  }

  if (candidate?.code || candidate?.message) {
    return {
      code: candidate.code || 'UNKNOWN',
      message: candidate.message || '',
    }
  }

  return undefined
}

export function getErrorMessage(error: unknown): string {
  const apiError = getApiError(error)
  if (apiError) {
    return translateError(apiError.code, apiError.message)
  }

  const message = (error as { message?: string })?.message
  return translateError(undefined, message)
}