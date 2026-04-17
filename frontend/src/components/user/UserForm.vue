<template>
  <el-dialog
    :model-value="modelValue"
    :title="isEdit ? t('users.editUser') : t('users.addUser')"
    @update:model-value="$emit('update:modelValue', $event)"
    width="500px"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-width="100px"
      @submit.prevent="handleSubmit"
    >
      <el-form-item :label="t('users.form.name')" prop="name">
        <el-input
          v-model="formData.name"
          :placeholder="t('users.form.namePlaceholder')"
          maxlength="100"
          show-word-limit
          autofocus
        />
      </el-form-item>
      <el-form-item :label="t('users.form.tier')" prop="tier">
        <el-select v-model="formData.tier" :placeholder="t('users.form.tierPlaceholder')">
          <el-option :label="t('users.tierNormal')" value="normal" />
          <el-option :label="t('users.tierPro')" value="pro" />
          <el-option :label="t('users.tierNoop')" value="noop" />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('users.form.handicap')" prop="handicap_rate">
        <el-input-number
          v-model="formData.handicap_rate"
          :min="0"
          :max="5"
          :step="0.5"
          :precision="1"
          placeholder="0.0"
        />
        <span class="el-form-item__helper" style="margin-left: 8px; color: var(--text-muted); font-size: 12px;">
          {{ t('users.form.handicapHint') }}
        </span>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="handleCancel">{{ t('common.cancel') }}</el-button>
      <el-button
        type="primary"
        @click="handleSubmit"
        :loading="loading"
      >
        {{ isEdit ? t('common.update') : t('common.create') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { FormInstance, FormRules } from 'element-plus'
import type { User } from '@/types/user'

interface Props {
  modelValue: boolean
  user?: User | null
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  user: null,
  loading: false
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  submit: [data: { name: string; tier: string; handicap_rate: number }]
  cancel: []
}>()

const formRef = ref<FormInstance>()
const formData = ref({
  name: '',
  tier: 'normal',
  handicap_rate: 0
})

const isEdit = ref(false)
const { t } = useI18n()

// Rules for form validation
const rules = computed<FormRules>(() => ({
  name: [
    { required: true, message: t('validation.nameRequired'), trigger: 'blur' },
    { min: 2, max: 100, message: t('validation.nameMinMax'), trigger: 'blur' }
  ]
}))

// Watch for user prop changes to populate form
watch(() => props.user, (newUser) => {
  if (newUser) {
    isEdit.value = true
    formData.value.name = newUser.name
    formData.value.tier = newUser.tier || 'normal'
    formData.value.handicap_rate = newUser.handicap_rate ?? 0
  } else {
    isEdit.value = false
    formData.value.name = ''
    formData.value.tier = 'normal'
    formData.value.handicap_rate = 0
  }
}, { immediate: true })

// Watch for dialog close to reset form
watch(() => props.modelValue, (newValue) => {
  if (!newValue) {
    resetForm()
  }
})

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate((valid) => {
    if (valid) {
      emit('submit', {
        name: formData.value.name,
        tier: formData.value.tier,
        handicap_rate: formData.value.handicap_rate,
      })
    }
  })
}

const handleCancel = () => {
  emit('cancel')
  emit('update:modelValue', false)
}

const resetForm = () => {
  formData.value.name = ''
  formData.value.tier = 'normal'
  formData.value.handicap_rate = 0
  formRef.value?.clearValidate()
}
</script>
