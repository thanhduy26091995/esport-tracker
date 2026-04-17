<template>
  <el-dialog
    :model-value="modelValue"
    :title="t('matches.form.recordTitle')"
    @update:model-value="$emit('update:modelValue', $event)"
    width="90%"
    style="max-width: 600px"
    :close-on-click-modal="false"
    destroy-on-close
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-position="top"
      @submit.prevent="handleSubmit"
    >
      <!-- Match Type -->
      <el-form-item :label="t('matches.form.type')" prop="match_type">
        <el-radio-group v-model="formData.match_type" @change="handleMatchTypeChange">
          <el-radio-button value="1v1">{{ t('matches.types.oneVsOne') }}</el-radio-button>
          <el-radio-button value="2v2">{{ t('matches.types.twoVsTwo') }}</el-radio-button>
        </el-radio-group>
      </el-form-item>

      <!-- Team 1 -->
      <el-form-item :label="t('matches.form.team1')" prop="team1">
        <el-select
          v-model="formData.team1"
          multiple
          :multiple-limit="formData.match_type === '1v1' ? 1 : 2"
          :placeholder="t('matches.form.selectPlayersPlaceholder')"
          class="w-full"
          filterable
          :disabled="loading || !availableUsers.length"
        >
          <el-option
            v-for="user in availableUsersForTeam1"
            :key="user.id"
            :label="`${user.name} (${user.current_score > 0 ? '+' : ''}${user.current_score})`"
            :value="user.id"
            :disabled="formData.team2.includes(user.id)"
          >
            <div class="flex items-center justify-between">
              <span>{{ user.name }}</span>
              <el-tag
                :type="user.current_score >= 0 ? 'success' : 'danger'"
                size="small"
              >
                {{ user.current_score > 0 ? '+' : '' }}{{ user.current_score }}
              </el-tag>
            </div>
          </el-option>
        </el-select>
      </el-form-item>

      <!-- Team 2 -->
      <el-form-item :label="t('matches.form.team2')" prop="team2">
        <el-select
          v-model="formData.team2"
          multiple
          :multiple-limit="formData.match_type === '1v1' ? 1 : 2"
          :placeholder="t('matches.form.selectPlayersPlaceholder')"
          class="w-full"
          filterable
          :disabled="loading || !availableUsers.length"
        >
          <el-option
            v-for="user in availableUsersForTeam2"
            :key="user.id"
            :label="`${user.name} (${user.current_score > 0 ? '+' : ''}${user.current_score})`"
            :value="user.id"
            :disabled="formData.team1.includes(user.id)"
          >
            <div class="flex items-center justify-between">
              <span>{{ user.name }}</span>
              <el-tag
                :type="user.current_score >= 0 ? 'success' : 'danger'"
                size="small"
              >
                {{ user.current_score > 0 ? '+' : '' }}{{ user.current_score }}
              </el-tag>
            </div>
          </el-option>
        </el-select>
      </el-form-item>

      <!-- Winner Selection -->
      <el-form-item :label="t('matches.form.winner')" prop="winner_team">
        <el-radio-group v-model="formData.winner_team">
          <el-radio :value="1" :disabled="!isTeam1Valid">
            {{ t('matches.team1') }}
            <span v-if="team1Names" class="text-sm text-gray-500 ml-2">
              ({{ team1Names }})
            </span>
          </el-radio>
          <el-radio :value="2" :disabled="!isTeam2Valid">
            {{ t('matches.team2') }}
            <span v-if="team2Names" class="text-sm text-gray-500 ml-2">
              ({{ team2Names }})
            </span>
          </el-radio>
        </el-radio-group>
      </el-form-item>

      <!-- Match Date (Optional) -->
      <el-form-item :label="t('matches.form.matchDate')" prop="match_date">
        <el-date-picker
          v-model="formData.match_date"
          type="datetime"
          :placeholder="t('matches.form.matchDatePlaceholder')"
          class="w-full"
          format="DD/MM/YYYY HH:mm"
          :disabled-date="disabledDate"
        />
      </el-form-item>

      <!-- Points Per Win -->
      <el-form-item :label="t('matches.form.pointsPerWin')" prop="points_per_win">
        <el-input-number
          v-model="formData.points_per_win"
          :min="1"
          :max="99"
          controls-position="right"
          class="w-full"
        />
        <div class="pts-preview">
          <span class="pts-win">{{ t('matches.form.pointsPreviewWinner', { value: formData.points_per_win }) }}</span>
          <span class="pts-sep">/</span>
          <span class="pts-lose">{{ t('matches.form.pointsPreviewLoser', { value: formData.points_per_win }) }}</span>
        </div>
      </el-form-item>

      <!-- Warning Messages -->
      <el-alert
        v-if="hasDebtWarning"
        type="warning"
        :closable="false"
        show-icon
        class="mb-4"
      >
        <template #title>
          {{ debtWarningMessage }}
        </template>
      </el-alert>

      <el-alert
        v-if="hasDuplicatePlayers"
        type="error"
        :closable="false"
        show-icon
        class="mb-4"
      >
        <template #title>
          {{ t('matches.form.duplicatePlayers') }}
        </template>
      </el-alert>
    </el-form>

    <template #footer>
      <div class="flex justify-between items-center">
        <div class="text-sm text-gray-500">
          <span v-if="isValid">✓ {{ t('matches.form.ready') }}</span>
        </div>
        <div class="space-x-2">
          <el-button @click="handleCancel">{{ t('common.cancel') }}</el-button>
          <el-button
            type="primary"
            @click="handleSubmit"
            :loading="loading"
            :disabled="!isValid"
          >
            {{ t('matches.form.submit') }}
          </el-button>
        </div>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { FormInstance, FormRules } from 'element-plus'
import type { User } from '@/types/user'
import type { CreateMatchRequest, MatchType } from '@/types/match'

interface Props {
  modelValue: boolean
  users: User[]
  loading?: boolean
  debtThreshold?: number
  pointsPerWin?: number
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  debtThreshold: -6,
  pointsPerWin: 1
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  submit: [data: CreateMatchRequest]
  cancel: []
}>()

const formRef = ref<FormInstance>()
const { t } = useI18n()
const formData = ref<{
  match_type: MatchType
  team1: string[]
  team2: string[]
  winner_team: 1 | 2
  match_date: Date | null
  points_per_win: number
}>({
  match_type: '1v1',
  team1: [],
  team2: [],
  winner_team: 1,
  match_date: null,
  points_per_win: props.pointsPerWin
})

// Validation rules
const rules = computed<FormRules>(() => ({
  match_type: [{ required: true, message: t('validation.selectMatchType') }],
  team1: [{ required: true, message: t('validation.selectTeam1') }],
  team2: [{ required: true, message: t('validation.selectTeam2') }],
  winner_team: [{ required: true, message: t('validation.selectWinner') }]
}))

// Computed
const availableUsers = computed(() => props.users.filter((u) => u.is_active))

const availableUsersForTeam1 = computed(() => availableUsers.value)
const availableUsersForTeam2 = computed(() => availableUsers.value)

const teamSize = computed(() => (formData.value.match_type === '1v1' ? 1 : 2))

const isTeam1Valid = computed(() => formData.value.team1.length === teamSize.value)
const isTeam2Valid = computed(() => formData.value.team2.length === teamSize.value)

const hasDuplicatePlayers = computed(() => {
  const allPlayers = [...formData.value.team1, ...formData.value.team2]
  return new Set(allPlayers).size !== allPlayers.length
})

const isValid = computed(() => {
  return (
    isTeam1Valid.value &&
    isTeam2Valid.value &&
    !hasDuplicatePlayers.value &&
    formData.value.winner_team > 0
  )
})

const team1Names = computed(() => {
  if (!formData.value.team1.length) return ''
  return formData.value.team1
    .map((id) => props.users.find((u) => u.id === id)?.name)
    .filter(Boolean)
    .join(', ')
})

const team2Names = computed(() => {
  if (!formData.value.team2.length) return ''
  return formData.value.team2
    .map((id) => props.users.find((u) => u.id === id)?.name)
    .filter(Boolean)
    .join(', ')
})

// Debt warning
const hasDebtWarning = computed(() => {
  const allPlayerIds = [...formData.value.team1, ...formData.value.team2]
  return allPlayerIds.some((id) => {
    const user = props.users.find((u) => u.id === id)
    return user && user.current_score <= props.debtThreshold
  })
})

const debtWarningMessage = computed(() => {
  const allPlayerIds = [...formData.value.team1, ...formData.value.team2]
  const debtPlayers = allPlayerIds
    .map((id) => props.users.find((u) => u.id === id))
    .filter((u) => u && u.current_score <= props.debtThreshold)
    .map((u) => u!.name)

  if (debtPlayers.length === 1) {
    return t('matches.form.debtWarningOne', { name: debtPlayers[0] })
  }
  return t('matches.form.debtWarningMany', { names: debtPlayers.join(', ') })
})

// Handle match type change
const handleMatchTypeChange = () => {
  // Reset teams when switching match type
  formData.value.team1 = []
  formData.value.team2 = []
  formData.value.winner_team = 1
}

// Watch for model value changes
watch(
  () => props.modelValue,
  (newValue) => {
    if (!newValue) {
      resetForm()
    }
  }
)

// Disable future dates
const disabledDate = (date: Date) => {
  return date > new Date()
}

const handleSubmit = async () => {
  if (!formRef.value || !isValid.value) return

  await formRef.value.validate((valid) => {
    if (valid) {
      const data: CreateMatchRequest = {
        match_type: formData.value.match_type,
        team1: formData.value.team1.slice(0, teamSize.value),
        team2: formData.value.team2.slice(0, teamSize.value),
        winner_team: formData.value.winner_team,
        points_per_win: formData.value.points_per_win
      }

      // Add match_date if set
      if (formData.value.match_date) {
        data.match_date = formData.value.match_date.toISOString()
      }

      emit('submit', data)
    }
  })
}

const handleCancel = () => {
  emit('cancel')
  emit('update:modelValue', false)
}

const resetForm = () => {
  formData.value = {
    match_type: '1v1',
    team1: [],
    team2: [],
    winner_team: 1,
    match_date: null,
    points_per_win: props.pointsPerWin
  }
  formRef.value?.clearValidate()
}
</script>

<style scoped>
:deep(.el-select__tags) {
  max-width: 100%;
}

:deep(.el-input-number) {
  width: 100%;
}

.pts-preview {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 6px;
  font-size: 12px;
  font-weight: 600;
}
.pts-win  { color: var(--color-success); }
.pts-sep  { color: var(--text-muted); }
.pts-lose { color: var(--color-danger); }
</style>
