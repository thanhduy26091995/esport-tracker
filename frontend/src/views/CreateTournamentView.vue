<template>
  <div class="page-wrapper">
    <div class="page-container" style="max-width: 700px">
      <div class="page-header">
        <div class="page-header-left">
          <el-button text @click="router.back()" :icon="ArrowLeft">{{ t('common.back') }}</el-button>
          <h1 class="page-title">{{ t('tournaments.form.title') }}</h1>
        </div>
      </div>

      <div class="card">
        <div class="card-body">
          <el-form
            ref="formRef"
            :model="form"
            label-position="top"
            @submit.prevent="handleSubmit"
          >
            <el-form-item
              :label="t('tournaments.form.name')"
              prop="name"
              :rules="[{ required: true, message: t('validation.tournamentNameRequired'), trigger: 'blur' }]"
            >
              <el-input v-model="form.name" :placeholder="t('tournaments.form.namePlaceholder')" />
            </el-form-item>

            <el-form-item :label="t('tournaments.form.matchType')">
              <el-radio-group v-model="form.match_type">
                <el-radio-button value="1v1">{{ t('matches.types.oneVsOne') }}</el-radio-button>
                <el-radio-button value="2v2">{{ t('matches.types.twoVsTwo') }}</el-radio-button>
              </el-radio-group>
            </el-form-item>

            <el-form-item :label="t('tournaments.form.players')">
              <div class="player-selector">
                <el-checkbox-group v-model="form.player_ids">
                  <div
                    v-for="user in userStore.users.filter(u => u.is_active)"
                    :key="user.id"
                    class="player-checkbox"
                  >
                    <el-checkbox :value="user.id">
                      {{ user.name }}
                      <PlayerTierBadge :tier="user.tier || 'normal'" style="margin-left: 6px" />
                      <span
                        v-if="user.handicap_rate > 0"
                        style="font-size: 11px; color: #909399; margin-left: 4px"
                      >
                        (-{{ user.handicap_rate }})
                      </span>
                    </el-checkbox>
                  </div>
                </el-checkbox-group>
                <el-button :icon="Plus" class="quick-create-player-button" @click="showQuickCreatePlayer = true" :title="t('players.quickCreate')">
                  {{ t('players.quickCreate') }}
                </el-button>
                <div class="el-form-item__helper mt-2">
                  {{ t('tournaments.form.selectedCount', { count: form.player_ids.length }) }}
                  <el-tag
                    v-if="form.match_type === '2v2' && form.player_ids.length % 2 !== 0"
                    type="warning"
                    size="small"
                    class="ml-2"
                  >
                    {{ t('tournaments.form.evenPlayerCountRequired') }}
                  </el-tag>
                </div>
              </div>
            </el-form-item>

            <el-form-item :label="t('tournaments.form.affectsScore')">
              <el-switch
                v-model="form.affects_score"
                :active-text="t('tournaments.form.affectsScoreActive')"
                :inactive-text="t('tournaments.form.affectsScoreInactive')"
              />
            </el-form-item>

            <el-form-item :label="t('tournaments.form.entryFee')">
              <el-input-number
                v-model="form.entry_fee"
                :min="0"
                :step="10000"
                :placeholder="t('tournaments.form.entryFeePlaceholder')"
              />
            </el-form-item>

            <el-form-item>
              <el-button
                type="primary"
                native-type="submit"
                :loading="store.loading"
                size="large"
              >
                {{ t('tournaments.form.submit') }}
              </el-button>
              <el-button @click="router.push('/tournaments')" size="large">{{ t('common.cancel') }}</el-button>
            </el-form-item>
          </el-form>

          <UserForm
            v-model="showQuickCreatePlayer"
            :loading="quickCreateLoading"
            @submit="handlePlayerCreated"
            @cancel="showQuickCreatePlayer = false"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ArrowLeft, Plus } from '@element-plus/icons-vue'
import type { FormInstance } from 'element-plus'
import { useTournamentStore } from '@/stores/tournamentStore'
import { useUserStore } from '@/stores/userStore'
import PlayerTierBadge from '@/components/PlayerTierBadge.vue'
import UserForm from '@/components/user/UserForm.vue'

const router = useRouter()
const { t } = useI18n()
const store = useTournamentStore()
const userStore = useUserStore()
const formRef = ref<FormInstance>()
const showQuickCreatePlayer = ref(false)
const quickCreateLoading = ref(false)

const form = ref({
  name: '',
  match_type: '1v1' as '1v1' | '2v2',
  player_ids: [] as string[],
  affects_score: true,
  entry_fee: 0,
})

onMounted(() => userStore.fetchUsers())

const handlePlayerCreated = async (data: { name: string; tier: string; handicap_rate: number }) => {
  quickCreateLoading.value = true
  try {
    await userStore.createUser(data.name, data.tier, data.handicap_rate)
    await userStore.fetchUsers()
    showQuickCreatePlayer.value = false
  } finally {
    quickCreateLoading.value = false
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      const tournament = await store.createTournament(form.value)
      router.push(`/tournaments/${tournament.id}`)
    } catch {}
  })
}
</script>

<style scoped>
.player-selector {
  width: 100%;
}

.player-checkbox {
  margin-bottom: 8px;
}

.quick-create-player-button {
  margin-top: 8px;
}

@media (max-width: 640px) {
  :deep(.el-radio-group) {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  :deep(.el-input-number) {
    width: 100%;
  }

  .quick-create-player-button {
    width: 100%;
    justify-content: center;
  }
}
</style>
