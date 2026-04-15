<template>
  <el-dialog
    :model-value="modelValue"
    :title="type === 'deposit' ? 'Deposit to Fund' : 'Withdraw from Fund'"
    @update:model-value="$emit('update:modelValue', $event)"
    width="500px"
    :close-on-click-modal="false"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-width="120px"
      @submit.prevent="handleSubmit"
    >
      <!-- Current Balance (for withdrawal) -->
      <div v-if="type === 'withdrawal'" class="mb-4 p-3 bg-blue-50 rounded-lg">
        <div class="text-sm text-blue-900">
          <span class="font-medium">Current Balance:</span>
          <span class="ml-2 text-lg font-bold">{{ formatVND(currentBalance) }}</span>
        </div>
      </div>

      <!-- Amount -->
      <el-form-item label="Amount (VND)" prop="amount">
        <el-input-number
          v-model="formData.amount"
          :min="1000"
          :max="type === 'withdrawal' ? currentBalance : undefined"
          :step="1000"
          :controls="true"
          class="w-full"
          :precision="0"
        />
      </el-form-item>

      <!-- Description -->
      <el-form-item label="Description" prop="description">
        <el-input
          v-model="formData.description"
          type="textarea"
          :rows="3"
          placeholder="Enter transaction details..."
          maxlength="200"
          show-word-limit
        />
      </el-form-item>

      <!-- Date (Optional) -->
      <el-form-item label="Date">
        <el-date-picker
          v-model="formData.date"
          type="datetime"
          placeholder="Default: Now"
          class="w-full"
          format="DD/MM/YYYY HH:mm"
          :disabled-date="disabledDate"
        />
      </el-form-item>

      <!-- Warning for large withdrawal -->
      <el-alert
        v-if="type === 'withdrawal' && formData.amount > currentBalance * 0.5"
        type="warning"
        :closable="false"
        show-icon
        class="mb-4"
      >
        <template #title>
          You are withdrawing more than 50% of the fund balance
        </template>
      </el-alert>
    </el-form>

    <template #footer>
      <el-button @click="handleCancel">Cancel</el-button>
      <el-button
        :type="type === 'deposit' ? 'success' : 'danger'"
        @click="handleSubmit"
        :loading="loading"
        :disabled="!isValid"
        plain
      >
        {{ type === 'deposit' ? 'Deposit' : 'Withdraw' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { formatVND } from '@/utils/formatters'

interface Props {
  modelValue: boolean
  type: 'deposit' | 'withdrawal'
  currentBalance: number
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  loading: false
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  submit: [data: { amount: number; description: string; date?: string }]
  cancel: []
}>()

const formRef = ref<FormInstance>()
const formData = ref({
  amount: 10000,
  description: '',
  date: null as Date | null
})

// Validation rules
const rules: FormRules = {
  amount: [
    { required: true, message: 'Please enter amount', trigger: 'blur' },
    {
      type: 'number',
      min: 1000,
      message: 'Amount must be at least 1,000 VND',
      trigger: 'blur'
    }
  ],
  description: [
    { required: true, message: 'Please enter description', trigger: 'blur' },
    {
      min: 3,
      max: 200,
      message: 'Description should be 3-200 characters',
      trigger: 'blur'
    }
  ]
}

const isValid = computed(() => {
  if (!formData.value.amount || formData.value.amount < 1000) return false
  if (!formData.value.description || formData.value.description.trim().length < 3) return false
  if (props.type === 'withdrawal' && formData.value.amount > props.currentBalance) return false
  return true
})

// Watch for dialog close
watch(
  () => props.modelValue,
  (newValue) => {
    if (!newValue) {
      resetForm()
    } else {
      // Set default description based on type
      if (!formData.value.description) {
        formData.value.description =
          props.type === 'deposit' ? 'Manual deposit to fund' : 'Manual withdrawal from fund'
      }
    }
  }
)

const disabledDate = (date: Date) => {
  return date > new Date()
}

const handleSubmit = async () => {
  if (!formRef.value || !isValid.value) return

  await formRef.value.validate((valid) => {
    if (valid) {
      const data: { amount: number; description: string; date?: string } = {
        amount: formData.value.amount,
        description: formData.value.description.trim()
      }

      if (formData.value.date) {
        data.date = formData.value.date.toISOString()
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
    amount: 10000,
    description: '',
    date: null
  }
  formRef.value?.clearValidate()
}
</script>
