<template>
  <el-dialog
    :model-value="modelValue"
    :title="isEdit ? t('users.editUser') : t('users.addUser')"
    @update:model-value="$emit('update:modelValue', $event)"
    width="92%"
    style="max-width: 500px"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-position="top"
      @submit.prevent="handleSubmit"
    >
      <!-- Avatar upload (edit mode only) -->
      <el-form-item v-if="isEdit" label="Avatar">
        <div class="avatar-upload-row">
          <UserAvatar :avatar-url="avatarPreview" :name="formData.name || '?'" size="lg" />
          <div class="avatar-upload-actions">
            <input ref="fileInput" type="file" accept="image/jpeg,image/png,image/gif,image/webp" class="hidden-file-input" @change="onFileChange" />
            <el-button size="small" :loading="avatarUploading" @click="fileInput?.click()">
              {{ avatarUploading ? 'Đang upload...' : 'Chọn ảnh' }}
            </el-button>
            <el-button v-if="avatarPreview" size="small" text type="danger" @click="handleDeleteAvatar">Xoá</el-button>
            <span class="avatar-hint">JPG, PNG, GIF, WebP — tối đa 5 MB</span>
          </div>
        </div>
      </el-form-item>

      <el-form-item :label="t('users.form.name')" prop="name">
        <el-input
          v-model="formData.name"
          :placeholder="t('users.form.namePlaceholder')"
          maxlength="100"
          show-word-limit
          autofocus
        />
      </el-form-item>

      <!-- Club picker (edit mode only) -->
      <el-form-item v-if="isEdit" label="Club yêu thích">
        <el-select v-model="formData.favorite_club" placeholder="Chọn club" clearable class="w-full">
          <el-option v-for="club in CLUBS" :key="club.slug" :label="club.name" :value="club.slug" />
        </el-select>
      </el-form-item>

      <el-form-item :label="t('users.form.tier')" prop="tier">
        <el-select v-model="formData.tier" :placeholder="t('users.form.tierPlaceholder')" class="w-full">
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
          class="w-full"
          placeholder="0.0"
        />
        <span class="el-form-item__helper">
          {{ t('users.form.handicapHint') }}
        </span>
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="user-form-footer">
        <el-button @click="handleCancel">{{ t('common.cancel') }}</el-button>
        <el-button
          type="primary"
          @click="handleSubmit"
          :loading="loading"
        >
          {{ isEdit ? t('common.update') : t('common.create') }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import type { User } from '@/types/user'
import { CLUBS } from '@/config/clubs'
import { userService } from '@/services/userService'
import UserAvatar from '@/components/shared/UserAvatar.vue'

interface Props {
  modelValue: boolean
  user?: User | null
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  user: null,
  loading: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  submit: [data: { name: string; tier: string; handicap_rate: number; favorite_club: string }]
  cancel: []
}>()

const formRef = ref<FormInstance>()
const fileInput = ref<HTMLInputElement>()
const avatarUploading = ref(false)
const avatarPreview = ref<string | null>(null)

const formData = ref({
  name: '',
  tier: 'normal',
  handicap_rate: 0,
  favorite_club: '',
})

const isEdit = ref(false)
const { t } = useI18n()

const rules = computed<FormRules>(() => ({
  name: [
    { required: true, message: t('validation.nameRequired'), trigger: 'blur' },
    { min: 2, max: 100, message: t('validation.nameMinMax'), trigger: 'blur' },
  ],
}))

watch(() => props.user, (newUser) => {
  if (newUser) {
    isEdit.value = true
    formData.value.name = newUser.name
    formData.value.tier = newUser.tier || 'normal'
    formData.value.handicap_rate = newUser.handicap_rate ?? 0
    formData.value.favorite_club = newUser.favorite_club ?? ''
    avatarPreview.value = newUser.avatar_url ?? null
  } else {
    isEdit.value = false
    formData.value.name = ''
    formData.value.tier = 'normal'
    formData.value.handicap_rate = 0
    formData.value.favorite_club = ''
    avatarPreview.value = null
  }
}, { immediate: true })

watch(() => props.modelValue, (val) => {
  if (!val) resetForm()
})

async function onFileChange(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file || !props.user) return
  if (file.size > 5 * 1024 * 1024) {
    ElMessage.error('File quá lớn (tối đa 5 MB)')
    return
  }
  avatarUploading.value = true
  try {
    const url = await userService.uploadAvatar(props.user.id, file)
    avatarPreview.value = url
    ElMessage.success('Upload avatar thành công')
  } catch {
    ElMessage.error('Upload thất bại')
  } finally {
    avatarUploading.value = false
    if (fileInput.value) fileInput.value.value = ''
  }
}

async function handleDeleteAvatar() {
  if (!props.user) return
  try {
    await userService.deleteAvatar(props.user.id)
    avatarPreview.value = null
    ElMessage.success('Đã xoá avatar')
  } catch {
    ElMessage.error('Xoá thất bại')
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate((valid) => {
    if (valid) {
      emit('submit', {
        name: formData.value.name,
        tier: formData.value.tier,
        handicap_rate: formData.value.handicap_rate,
        favorite_club: formData.value.favorite_club,
      })
    }
  })
}

const handleCancel = () => {
  emit('cancel')
  emit('update:modelValue', false)
}

const resetForm = () => {
  formData.value = { name: '', tier: 'normal', handicap_rate: 0, favorite_club: '' }
  avatarPreview.value = null
  formRef.value?.clearValidate()
}
</script>

<style scoped>
:deep(.el-input-number) {
  width: 100%;
}

.user-form-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.avatar-upload-row {
  display: flex;
  align-items: center;
  gap: 16px;
}

.avatar-upload-actions {
  display: flex;
  flex-direction: column;
  gap: 6px;
  align-items: flex-start;
}

.hidden-file-input {
  display: none;
}

.avatar-hint {
  font-size: 11px;
  color: var(--text-muted);
}

@media (max-width: 640px) {
  .user-form-footer {
    width: 100%;
  }

  .user-form-footer .el-button {
    flex: 1;
  }
}
</style>
