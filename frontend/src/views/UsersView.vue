<template>
  <div class="page-wrapper">
    <div class="page-container">
      <!-- Header -->
      <div class="page-header">
        <div class="page-header-left">
          <h1 class="page-title">Players</h1>
          <p class="page-subtitle">Manage FC25 players and view their scores</p>
        </div>
        <el-button type="primary" @click="handleAdd" :icon="Plus" size="large">
          Add Player
        </el-button>
      </div>

      <!-- Stats -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">
        <StatCard
          title="Total Players"
          :value="userStore.users.length"
          :icon="UserIcon"
          :loading="userStore.loading"
          type="info"
        />
        <StatCard
          title="Top Score"
          :value="topScore >= 0 ? `+${topScore}` : topScore"
          :icon="Trophy"
          :loading="userStore.loading"
          type="success"
        />
        <StatCard
          title="Players in Debt"
          :value="playersInDebt"
          :icon="Warning"
          :loading="userStore.loading"
          :type="playersInDebt > 0 ? 'danger' : 'default'"
        />
      </div>

      <!-- Player Table -->
      <div class="card">
        <div class="card-body">
          <UserTable
            :users="userStore.users"
            :loading="userStore.loading"
            :conversion-rate="configStore.pointToVnd"
            :debt-threshold="configStore.debtThreshold"
            @edit="handleEdit"
            @delete="handleDeleteConfirm"
            @trigger-settlement="handleTriggerSettlement"
          />
        </div>
      </div>

      <!-- User Form Dialog -->
      <UserForm
        v-model="showDialog"
        :user="selectedUser"
        :loading="userStore.loading"
        @submit="handleSubmit"
        @cancel="handleCancel"
      />

      <!-- Settlement Trigger Dialog -->
      <SettlementTriggerDialog
        v-model="showSettlementDialog"
        :debtor="settlementDebtor"
        :users="userStore.users"
        :point-to-vnd="configStore.pointToVnd"
        :fund-split-percent="configStore.fundSplitPercent"
        :loading="settlementStore.loading"
        @confirm="handleSettlementConfirm"
        @cancel="handleSettlementCancel"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Plus, User as UserIcon, Trophy, Warning } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/userStore'
import { useConfigStore } from '@/stores/configStore'
import { useSettlementStore } from '@/stores/settlementStore'
import UserTable from '@/components/user/UserTable.vue'
import UserForm from '@/components/user/UserForm.vue'
import StatCard from '@/components/shared/StatCard.vue'
import SettlementTriggerDialog from '@/components/settlement/SettlementTriggerDialog.vue'
import type { User } from '@/types/user'

const userStore = useUserStore()
const configStore = useConfigStore()
const settlementStore = useSettlementStore()
const showDialog = ref(false)
const selectedUser = ref<User | null>(null)
const showSettlementDialog = ref(false)
const settlementDebtor = ref<User | null>(null)

const topScore = computed(() => {
  if (userStore.users.length === 0) return 0
  return Math.max(...userStore.users.map((u: User) => u.current_score))
})

const playersInDebt = computed(() =>
  userStore.users.filter((u: User) => u.current_score < 0).length
)

onMounted(async () => {
  await Promise.all([userStore.fetchUsers(), configStore.fetchConfigs()])
})

const handleAdd = () => {
  selectedUser.value = null
  showDialog.value = true
}

const handleEdit = (user: User) => {
  selectedUser.value = user
  showDialog.value = true
}

const handleSubmit = async (data: { name: string; tier: string; handicap_rate: number }) => {
  try {
    if (selectedUser.value) {
      await userStore.updateUser(selectedUser.value.id, data.name, data.tier, data.handicap_rate)
    } else {
      await userStore.createUser(data.name)
    }
    showDialog.value = false
  } catch {}
}

const handleCancel = () => {
  selectedUser.value = null
  showDialog.value = false
}

const handleDeleteConfirm = (user: User) => {
  ElMessageBox.confirm(
    `Are you sure you want to delete "${user.name}"? This action cannot be undone.`,
    'Delete Player',
    {
      confirmButtonText: 'Delete',
      cancelButtonText: 'Cancel',
      type: 'warning',
      confirmButtonClass: 'el-button--danger'
    }
  )
    .then(() => { userStore.deleteUser(user.id) })
    .catch(() => {})
}

const handleTriggerSettlement = (user: User) => {
  settlementDebtor.value = user
  showSettlementDialog.value = true
}

const handleSettlementConfirm = async (winnerIds: string[]) => {
  if (!settlementDebtor.value) return
  
  try {
    await settlementStore.triggerSettlement(settlementDebtor.value.id, winnerIds)
    await userStore.fetchUsers() // Refresh user data
    ElMessage.success(`Settlement triggered for ${settlementDebtor.value.name}`)
    showSettlementDialog.value = false
    settlementDebtor.value = null
  } catch (error) {
    // Error already handled by store
  }
}

const handleSettlementCancel = () => {
  settlementDebtor.value = null
  showSettlementDialog.value = false
}
</script>
