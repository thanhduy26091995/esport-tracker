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
  submit: [name: string]
  cancel: []
}>()

const formRef = ref<FormInstance>()
const formData = ref({
  name: ''
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
  } else {
    isEdit.value = false
    formData.value.name = ''
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
      emit('submit', formData.value.name)
    }
  })
}

const handleCancel = () => {
  emit('cancel')
  emit('update:modelValue', false)
}

const resetForm = () => {
  formData.value.name = ''
  formRef.value?.clearValidate()
}
</script>
