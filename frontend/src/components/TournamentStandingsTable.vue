<template>
  <div class="standings-wrap">
    <el-table :data="standings" size="small" class="standings-table" :row-class-name="({ rowIndex }: { rowIndex: number }) => rowIndex < (knockoutSize ?? 4) ? 'row-qualified' : ''">
      <el-table-column label="#" width="36" align="center">
        <template #default="{ $index }">
          <span :class="$index < (knockoutSize ?? 4) ? 'seed-qualified' : 'seed-out'">{{ $index + 1 }}</span>
        </template>
      </el-table-column>

      <el-table-column :label="t('tournaments.standings.team')" min-width="130">
        <template #default="{ row }: { row: TeamStanding }">
          <div class="team-cell">
            <span class="team-names">{{ row.player1?.name }} &amp; {{ row.player2?.name }}</span>
            <span v-if="row.seed > 0" class="seed-badge">S{{ row.seed }}</span>
          </div>
        </template>
      </el-table-column>

      <el-table-column prop="played" :label="t('tournaments.standings.played', 'P')" width="38" align="center" />
      <el-table-column prop="won"    :label="t('tournaments.standings.won', 'W')"    width="38" align="center" />
      <el-table-column prop="drawn"  :label="t('tournaments.standings.drawn', 'D')"  width="38" align="center" />
      <el-table-column prop="lost"   :label="t('tournaments.standings.lost', 'L')"   width="38" align="center" />
      <el-table-column prop="gf"     :label="t('tournaments.standings.gf', 'GF')"    width="40" align="center" />
      <el-table-column prop="ga"     :label="t('tournaments.standings.ga', 'GA')"    width="40" align="center" />

      <el-table-column :label="t('tournaments.standings.gd', 'GD')" width="48" align="center">
        <template #default="{ row }: { row: TeamStanding }">
          <span :class="row.gd >= 0 ? 'text-success' : 'text-danger'">
            {{ row.gd >= 0 ? '+' : '' }}{{ row.gd }}
          </span>
        </template>
      </el-table-column>

      <el-table-column :label="t('tournaments.standings.points', 'Pts')" width="48" align="center">
        <template #default="{ row }: { row: TeamStanding }">
          <strong>{{ row.points }}</strong>
        </template>
      </el-table-column>
    </el-table>

    <div v-if="standings.length === 0" class="empty-state">
      {{ t('tournaments.standings.noResults', 'No results yet') }}
    </div>

    <p class="qualification-note">
      {{ t('tournaments.standings.qualificationNote', { n: knockoutSize }, `Top ${knockoutSize} teams advance to knockout stage`) }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { TeamStanding } from '@/types/tournament'

defineProps<{ standings: TeamStanding[]; knockoutSize?: number }>()
const { t } = useI18n()
</script>

<style scoped>
.standings-wrap {
  width: 100%;
  overflow-x: auto;
}

.standings-table {
  min-width: 420px;
}

:deep(.el-table__row.row-qualified) {
  background-color: color-mix(in srgb, var(--el-color-success) 6%, transparent);
}

.team-cell {
  display: flex;
  align-items: center;
  gap: 6px;
}

.team-names {
  font-size: 13px;
  font-weight: 500;
}

.seed-badge {
  font-size: 10px;
  font-weight: 700;
  background: var(--el-color-success);
  color: #fff;
  border-radius: 3px;
  padding: 1px 4px;
}

.seed-qualified {
  font-weight: 700;
  color: var(--el-color-success);
}

.seed-out {
  color: var(--text-muted, #9ca3af);
}

.text-success { color: var(--el-color-success); }
.text-danger  { color: var(--el-color-danger); }

.empty-state {
  text-align: center;
  color: var(--text-muted, #9ca3af);
  padding: 16px 0;
  font-size: 13px;
}

.qualification-note {
  font-size: 11px;
  color: var(--text-muted, #9ca3af);
  margin-top: 8px;
  text-align: center;
}
</style>
