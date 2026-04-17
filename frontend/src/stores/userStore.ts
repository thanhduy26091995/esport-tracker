import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { User } from '@/types/user'
import { userService } from '@/services/userService'
import { ElMessage } from 'element-plus'
import { getErrorMessage, translate } from '@/utils/i18n'

export const useUserStore = defineStore('user', () => {
  const users = ref<User[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Fetch all users
  async function fetchUsers() {
    loading.value = true
    error.value = null
    try {
      users.value = await userService.getAll()
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch users'
      if (error.value) ElMessage.error(error.value)
    } finally {
      loading.value = false
    }
  }

  // Create a new user
  async function createUser(name: string, tier?: string, handicapRate?: number) {
    // Validate name before sending
    if (!name || name.trim().length === 0) {
      ElMessage.error(translate('validation.nameRequired'))
      throw new Error('Name is required')
    }
    if (name.trim().length < 2) {
      ElMessage.error(translate('validation.nameMin'))
      throw new Error('Name too short')
    }
    if (name.trim().length > 100) {
      ElMessage.error(translate('validation.nameMax'))
      throw new Error('Name too long')
    }

    loading.value = true
    error.value = null
    try {
      const newUser = await userService.create({
        name: name.trim(),
        ...(tier !== undefined && { tier }),
        ...(handicapRate !== undefined && { handicap_rate: handicapRate }),
      })
      users.value.push(newUser)
      // Sort by score DESC
      users.value.sort((a, b) => b.current_score - a.current_score)
      ElMessage.success(translate('toast.userCreated', { name: newUser.name }))
      return newUser
    } catch (err: any) {
      const errorMsg = getErrorMessage(err)
      error.value = errorMsg
      ElMessage.error(errorMsg)
      throw err
    } finally {
      loading.value = false
    }
  }

  // Update a user
  async function updateUser(id: string, name: string, tier?: string, handicapRate?: number) {
    // Validate name before sending
    if (!name || name.trim().length === 0) {
      ElMessage.error(translate('validation.nameRequired'))
      throw new Error('Name is required')
    }
    if (name.trim().length < 2) {
      ElMessage.error(translate('validation.nameMin'))
      throw new Error('Name too short')
    }
    if (name.trim().length > 100) {
      ElMessage.error(translate('validation.nameMax'))
      throw new Error('Name too long')
    }

    loading.value = true
    error.value = null
    try {
      const updatedUser = await userService.update(id, {
        name: name.trim(),
        ...(tier !== undefined && { tier }),
        ...(handicapRate !== undefined && { handicap_rate: handicapRate }),
      })
      const index = users.value.findIndex(u => u.id === id)
      if (index !== -1) {
        users.value[index] = updatedUser
      }
      ElMessage.success(translate('toast.userUpdated', { name: updatedUser.name }))
      return updatedUser
    } catch (err: any) {
      const errorMsg = getErrorMessage(err)
      error.value = errorMsg
      ElMessage.error(errorMsg)
      throw err
    } finally {
      loading.value = false
    }
  }

  // Delete a user
  async function deleteUser(id: string) {
    if (!id) {
      ElMessage.error(translate('validation.invalidUserId'))
      throw new Error('User ID is required')
    }

    loading.value = true
    error.value = null
    try {
      await userService.delete(id)
      const deletedUser = users.value.find(u => u.id === id)
      users.value = users.value.filter(u => u.id !== id)
      ElMessage.success(translate('toast.userDeleted', { name: deletedUser?.name || translate('common.unknown') }))
    } catch (err: any) {
      const errorMsg = getErrorMessage(err)
      error.value = errorMsg
      ElMessage.error(errorMsg)
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    users,
    loading,
    error,
    fetchUsers,
    createUser,
    updateUser,
    deleteUser,
  }
})
