<template>
  <div class="wc-standings-card">
    <div class="wc-standings-header">{{ standing.group_name }}</div>
    <div class="wc-standings-table-wrap">
      <table class="wc-standings-table">
        <thead>
          <tr>
            <th class="col-rank">{{ t('wc.standings.rank') }}</th>
            <th class="col-team">{{ t('wc.standings.team') }}</th>
            <th class="col-stat">{{ t('wc.standings.played') }}</th>
            <th class="col-stat">{{ t('wc.standings.won') }}</th>
            <th class="col-stat">{{ t('wc.standings.drawn') }}</th>
            <th class="col-stat">{{ t('wc.standings.lost') }}</th>
            <th class="col-stat">{{ t('wc.standings.goalDiff') }}</th>
            <th class="col-stat col-pts">{{ t('wc.standings.points') }}</th>
            <th class="col-form">{{ t('wc.standings.form') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(team, idx) in standing.teams"
            :key="team.team_name"
            :class="rowClass(idx)"
          >
            <td class="col-rank">{{ idx + 1 }}</td>
            <td class="col-team">
              <span class="team-flag">{{ teamCodeToFlag(team.team_code) }}</span>
              <span class="team-code">{{ team.team_code }}</span>
              <span class="team-name">{{ team.team_name }}</span>
            </td>
            <td class="col-stat">{{ team.played }}</td>
            <td class="col-stat">{{ team.won }}</td>
            <td class="col-stat">{{ team.drawn }}</td>
            <td class="col-stat">{{ team.lost }}</td>
            <td class="col-stat">{{ formatGD(team.goal_difference) }}</td>
            <td class="col-stat col-pts">{{ team.points }}</td>
            <td class="col-form">
              <span
                v-for="(result, fi) in team.form"
                :key="fi"
                class="form-badge"
                :class="`form-badge--${result.toLowerCase()}`"
              >{{ result }}</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { WcGroupStanding } from '@/types/wc'

const { t } = useI18n()

defineProps<{
  standing: WcGroupStanding
}>()

// Row highlight: 1-2 green (direct), 3 yellow (best-3rd potential), 4 none
function rowClass(idx: number): string {
  if (idx < 2) return 'row-qualify'
  if (idx === 2) return 'row-potential'
  return ''
}

function formatGD(gd: number): string {
  if (gd > 0) return `+${gd}`
  return `${gd}`
}

// Map 3-letter FIFA TLA codes to emoji flags via ISO 3166-1 alpha-2
function isoAlpha2ToFlag(alpha2: string): string {
  return [...alpha2.toUpperCase()]
    .map(c => String.fromCodePoint(0x1f1e6 + c.charCodeAt(0) - 65))
    .join('')
}

const TLA_TO_ALPHA2: Record<string, string> = {
  // Group A
  ARG: 'AR', ECU: 'EC', CAN: 'CA', CHI: 'CL',
  // Group B
  MEX: 'MX', POL: 'PL', SAU: 'SA', BUL: 'BG',
  // Group C
  USA: 'US', URU: 'UY', PAN: 'PA', BOL: 'BO',
  // Group D
  ENG: 'GB', SEN: 'SN', NGA: 'NG', RSA: 'ZA',
  // Group E
  GER: 'DE', JPN: 'JP', AUS: 'AU', TUN: 'TN',
  // Group F
  POR: 'PT', CRO: 'HR', MAR: 'MA', CMR: 'CM',
  // Group G
  BRA: 'BR', COL: 'CO', VEN: 'VE', PAR: 'PY',
  // Group H
  FRA: 'FR', BEL: 'BE', ALG: 'DZ', CIV: 'CI',
  // Group I
  ESP: 'ES', SUI: 'CH', NED: 'NL', SRB: 'RS',
  // Group J
  KOR: 'KR', IRN: 'IR', QAT: 'QA', NZL: 'NZ',
  // Group K
  ITA: 'IT', AUT: 'AT', MEX2: 'MX', CRC: 'CR',
  // Group L
  PER: 'PE', TUR: 'TR', NOR: 'NO', HON: 'HN',
  // Additional common codes
  CHN: 'CN', EGY: 'EG', GHA: 'GH', MLI: 'ML', SEN2: 'SN',
  GRE: 'GR', CZE: 'CZ', POL2: 'PL', UKR: 'UA', SLO: 'SI',
  DEN: 'DK', SWE: 'SE', IRQ: 'IQ', SYR: 'SY', LBN: 'LB',
  WAL: 'GB', SCO: 'GB', NIR: 'GB',
  // Fallback common codes from football-data.org TLA
  ALB: 'AL', AND: 'AD', ARM: 'AM', AZE: 'AZ', BLR: 'BY',
  BIH: 'BA', CYP: 'CY', EST: 'EE', FIN: 'FI', GEO: 'GE',
  HUN: 'HU', ISL: 'IS', ISR: 'IL', KAZ: 'KZ', KOS: 'XK',
  LAT: 'LV', LIE: 'LI', LTU: 'LT', LUX: 'LU', MKD: 'MK',
  MLT: 'MT', MDA: 'MD', MNE: 'ME', ROU: 'RO', RUS: 'RU',
  SMR: 'SM', SVK: 'SK', SVN: 'SI',
  CHL: 'CL', COL2: 'CO', CUB: 'CU', DOM: 'DO', GTM: 'GT',
  GUY: 'GY', HAI: 'HT', JAM: 'JM', TRI: 'TT', USA2: 'US',
  AFG: 'AF', BHR: 'BH', BAN: 'BD', IDN: 'ID', IND: 'IN',
  JOR: 'JO', KWT: 'KW', LKA: 'LK', MYS: 'MY', NPL: 'NP',
  OMN: 'OM', PAK: 'PK', PHI: 'PH', THA: 'TH', UAE: 'AE',
  UZB: 'UZ', VIE: 'VN', YEM: 'YE',
  AGO: 'AO', BEN: 'BJ', BFA: 'BF', BWA: 'BW', CPV: 'CV',
  COM: 'KM', COD: 'CD', COG: 'CG', DJI: 'DJ', ERI: 'ER',
  ETH: 'ET', GAB: 'GA', GAM: 'GM', GIN: 'GN', GNB: 'GW',
  GNQ: 'GQ', KEN: 'KE', LBR: 'LR', LBA: 'LY', MDG: 'MG',
  MWI: 'MW', MRT: 'MR', MUS: 'MU', MOZ: 'MZ', NAM: 'NA',
  NER: 'NE', RWA: 'RW', SLE: 'SL', SOM: 'SO', SSD: 'SS',
  SDN: 'SD', SWZ: 'SZ', TAN: 'TZ', TOG: 'TG', UGA: 'UG',
  ZAM: 'ZM', ZIM: 'ZW', ZAN: 'TZ',
}

function teamCodeToFlag(code: string): string {
  const alpha2 = TLA_TO_ALPHA2[code?.toUpperCase()]
  if (!alpha2) return '🏳️'
  return isoAlpha2ToFlag(alpha2)
}
</script>

<style scoped>
.wc-standings-card {
  background: var(--surface-card);
  border: 1px solid var(--border-subtle);
  border-radius: 12px;
  overflow: hidden;
}

.wc-standings-header {
  font-size: 12px;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  padding: 10px 14px 8px;
  border-bottom: 1px solid var(--border-subtle);
  background: var(--surface-raised, var(--surface-card));
}

.wc-standings-table-wrap {
  overflow-x: auto;
}

.wc-standings-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.wc-standings-table thead tr {
  background: transparent;
}

.wc-standings-table th {
  padding: 6px 8px;
  font-weight: 600;
  color: var(--text-muted);
  text-align: center;
  border-bottom: 1px solid var(--border-subtle);
  white-space: nowrap;
}

.wc-standings-table td {
  padding: 7px 8px;
  text-align: center;
  color: var(--text-primary);
  border-bottom: 1px solid var(--border-subtle);
}

.wc-standings-table tbody tr:last-child td {
  border-bottom: none;
}

/* Row highlights */
.row-qualify {
  background: rgba(22, 163, 74, 0.07);
}

.row-potential {
  background: rgba(234, 179, 8, 0.07);
}

/* Rank column */
.col-rank {
  width: 28px;
  font-weight: 700;
  font-size: 11px;
}

/* Team column */
.col-team {
  text-align: left !important;
  min-width: 120px;
  max-width: 180px;
}

.team-flag {
  font-size: 14px;
  margin-right: 4px;
}

.team-code {
  font-weight: 700;
  font-size: 11px;
  letter-spacing: 0.03em;
  margin-right: 4px;
  color: var(--text-secondary);
}

.team-name {
  font-size: 11px;
  color: var(--text-secondary);
  display: none; /* hidden in grid mode by default — shown in single-group mode via parent class */
}

/* Stats columns */
.col-stat {
  width: 28px;
  font-variant-numeric: tabular-nums;
}

.col-pts {
  font-weight: 700;
  color: var(--text-primary);
}

/* Form column */
.col-form {
  min-width: 64px;
  white-space: nowrap;
}

.form-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  border-radius: 3px;
  font-size: 9px;
  font-weight: 800;
  margin-right: 2px;
  line-height: 1;
}

.form-badge--w {
  background: rgba(22, 163, 74, 0.15);
  color: #16a34a;
}

.form-badge--d {
  background: rgba(100, 116, 139, 0.15);
  color: #64748b;
}

.form-badge--l {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}

/* Single-group mode: parent adds .wc-standings-single class */
:global(.wc-standings-single) .team-name {
  display: inline;
}

/* Mobile */
@media (max-width: 480px) {
  .col-form {
    display: none;
  }
}
</style>
