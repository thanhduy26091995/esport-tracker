/**
 * Validate user name
 * Rules: 2-100 characters, trimmed
 */
export function validateName(name: string): { valid: boolean; error?: string } {
  const trimmed = name.trim()

  if (trimmed.length === 0) {
    return { valid: false, error: 'Name is required' }
  }

  if (trimmed.length < 2) {
    return { valid: false, error: 'Name must be at least 2 characters' }
  }

  if (trimmed.length > 100) {
    return { valid: false, error: 'Name must be less than 100 characters' }
  }

  return { valid: true }
}

/**
 * Validate amount (for deposits/withdrawals)
 * Rules: Positive number, optional minimum
 */
export function validateAmount(
  amount: number,
  minAmount: number = 1
): { valid: boolean; error?: string } {
  if (isNaN(amount)) {
    return { valid: false, error: 'Amount must be a number' }
  }

  if (amount < minAmount) {
    return { valid: false, error: `Amount must be at least ${minAmount}` }
  }

  return { valid: true }
}

/**
 * Validate config value based on key
 */
export function validateConfigValue(
  key: string,
  value: string
): { valid: boolean; error?: string } {
  const numValue = parseInt(value)

  if (isNaN(numValue)) {
    return { valid: false, error: 'Value must be a number' }
  }

  switch (key) {
    case 'debt_threshold':
      if (numValue > 0) {
        return { valid: false, error: 'Debt threshold must be negative or zero' }
      }
      break

    case 'point_to_vnd':
      if (numValue <= 0) {
        return { valid: false, error: 'Point to VND must be positive' }
      }
      break

    case 'fund_split_percent':
      if (numValue < 0 || numValue > 100) {
        return { valid: false, error: 'Fund split percent must be between 0 and 100' }
      }
      break
  }

  return { valid: true }
}

/**
 * Validate team composition for match
 */
export function validateTeams(
  team1: string[],
  team2: string[],
  matchType: '1v1' | '2v2'
): { valid: boolean; error?: string } {
  const expectedSize = matchType === '1v1' ? 1 : 2

  if (team1.length !== expectedSize) {
    return { valid: false, error: `Team 1 must have ${expectedSize} player(s)` }
  }

  if (team2.length !== expectedSize) {
    return { valid: false, error: `Team 2 must have ${expectedSize} player(s)` }
  }

  // Check for duplicate players
  const allPlayers = [...team1, ...team2]
  const uniquePlayers = new Set(allPlayers)

  if (uniquePlayers.size !== allPlayers.length) {
    return { valid: false, error: 'Players cannot be on both teams' }
  }

  return { valid: true }
}
