<template>
  <el-dialog
    :model-value="modelValue"
    :title="t('matches.bonus.dialogTitle')"
    @update:model-value="$emit('update:modelValue', $event)"
    width="500px"
    :close-on-click-modal="false"
    destroy-on-close
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-width="120px"
      @submit.prevent="handleSubmit"
    >
      <!-- Player -->
      <el-form-item :label="t('matches.bonus.player')" prop="user_id">
        <el-select
          v-model="formData.user_id"
          :placeholder="t('matches.bonus.playerPlaceholder')"
          filterable
          class="w-full"
        >
          <el-option
            v-for="user in users"
            :key="user.id"
            :label="`${user.name} (${user.current_score > 0 ? '+' : ''}${user.current_score})`"
            :value="user.id"
          />
        </el-select>
      </el-form-item>

      <!-- Points -->
      <el-form-item :label="t('matches.bonus.points')" prop="points">
        <el-input-number
          v-model="formData.points"
          :min="1"
          :step="1"
          :controls="true"
          class="w-full"
          :precision="0"
        />
      </el-form-item>

      <!-- Description -->
      <el-form-item :label="t('matches.bonus.description')" prop="description">
        <el-input
          v-model="formData.description"
          type="textarea"
          :rows="2"
          :placeholder="t('matches.bonus.descriptionPlaceholder')"
          maxlength="200"
          show-word-limit
        />
      </el-form-item>

      <!-- Date (Optional) -->
      <el-form-item :label="t('matches.bonus.date')">
        <el-date-picker
          v-model="formData.bonus_date"
          type="datetime"
          :placeholder="t('matches.bonus.datePlaceholder')"
          class="w-full"
          format="DD/MM/YYYY HH:mm"
          :disabled-date="(d: Date) => d > new Date()"
        />
      </el-form-item>

      <el-alert type="info" :closable="false" show-icon class="mb-2">
        <template #title>{{ t('matches.bonus.hint') }}</template>
      </el-alert>
    </el-form>

    <template #footer>
      <el-button @click="$emit('update:modelValue', false)">{{ t('common.cancel') }}</el-button>
      <el-button type="warning" @click="handleSubmit" :loading="loading" plain>
        {{ t('matches.bonus.submit') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import type { FormInstance, FormRules } from 'element-plus'
import type { User } from '@/types/user'
import type { CreateScoreBonusRequest } from '@/types/scoreBonus'

const { t } = useI18n()

const props = defineProps<{
  modelValue: boolean
  users: User[]
  loading: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', v: boolean): void
  (e: 'submit', req: CreateScoreBonusRequest): void
}>()

const formRef = ref<FormInstance>()

const formData = reactive({
  user_id: '',
  points: 1,
  description: '',
  bonus_date: null as Date | null,
})

const rules: FormRules = {
  user_id: [{ required: true, message: t('matches.bonus.playerRequired'), trigger: 'change' }],
  points: [{ required: true, type: 'number', min: 1, message: t('matches.bonus.pointsRequired'), trigger: 'change' }],
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate((valid) => {
    if (!valid) return
    const req: CreateScoreBonusRequest = {
      user_id: formData.user_id,
      points: formData.points,
      description: formData.description,
    }
    if (formData.bonus_date) {
      req.bonus_date = (formData.bonus_date as Date).toISOString()
    }
    emit('submit', req)
  })
}
</script>
