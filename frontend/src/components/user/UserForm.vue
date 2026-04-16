<template>
  <el-dialog
    :model-value="modelValue"
    :title="isEdit ? 'Edit User' : 'Add New User'"
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
      <el-form-item label="Name" prop="name">
        <el-input
          v-model="formData.name"
          placeholder="Enter player name"
          maxlength="100"
          show-word-limit
          autofocus
        />
      </el-form-item>
      <el-form-item label="Tier" prop="tier">
        <el-select v-model="formData.tier" placeholder="Select tier">
          <el-option label="Normal" value="normal" />
          <el-option label="Pro" value="pro" />
          <el-option label="Noop" value="noop" />
        </el-select>
      </el-form-item>
      <el-form-item label="Handicap" prop="handicap_rate">
        <el-input-number
          v-model="formData.handicap_rate"
          :min="0"
          :max="5"
          :step="0.5"
          :precision="1"
          placeholder="0.0"
        />
        <span class="el-form-item__helper" style="margin-left: 8px; color: var(--text-muted); font-size: 12px;">
          penalty goals subtracted from score (e.g. 0.5 = must win by 1+ to count as win)
        </span>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="handleCancel">Cancel</el-button>
      <el-button
        type="primary"
        @click="handleSubmit"
        :loading="loading"
      >
        {{ isEdit ? 'Update' : 'Create' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
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

// Rules for form validation
const rules: FormRules = {
  name: [
    { required: true, message: 'Please enter a name', trigger: 'blur' },
    { min: 2, max: 100, message: 'Name should be 2-100 characters', trigger: 'blur' }
  ]
}

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
