<template>
  <div class="page-wrapper">
    <div class="page-container" style="max-width: 700px">
      <div class="page-header">
        <div class="page-header-left">
          <el-button text @click="router.back()" :icon="ArrowLeft">Back</el-button>
          <h1 class="page-title">Create Tournament</h1>
        </div>
      </div>

      <div class="card">
        <div class="card-body">
          <el-form
            ref="formRef"
            :model="form"
            label-width="140px"
            @submit.prevent="handleSubmit"
          >
            <el-form-item
              label="Name"
              prop="name"
              :rules="[{ required: true, message: 'Name is required', trigger: 'blur' }]"
            >
              <el-input v-model="form.name" placeholder="e.g. Weekly Tourney #5" />
            </el-form-item>

            <el-form-item label="Match Type">
              <el-radio-group v-model="form.match_type">
                <el-radio-button value="1v1">1v1</el-radio-button>
                <el-radio-button value="2v2">2v2</el-radio-button>
              </el-radio-group>
            </el-form-item>

            <el-form-item label="Players">
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
                <div class="mt-2" style="font-size: 12px; color: #909399">
                  Selected: {{ form.player_ids.length }} players
                  <el-tag
                    v-if="form.match_type === '2v2' && form.player_ids.length % 2 !== 0"
                    type="warning"
                    size="small"
                    class="ml-2"
                  >
                    2v2 needs even count
                  </el-tag>
                </div>
              </div>
            </el-form-item>

            <el-form-item label="Affects Score">
              <el-switch
                v-model="form.affects_score"
                active-text="Yes (updates player scores)"
                inactive-text="No (tournament only)"
              />
            </el-form-item>

            <el-form-item label="Entry Fee (VND)">
              <el-input-number
                v-model="form.entry_fee"
                :min="0"
                :step="10000"
                placeholder="0 = free"
              />
            </el-form-item>

            <el-form-item>
              <el-button
                type="primary"
                native-type="submit"
                :loading="store.loading"
                size="large"
              >
                Generate Tournament
              </el-button>
              <el-button @click="router.push('/tournaments')" size="large">Cancel</el-button>
            </el-form-item>
          </el-form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'
import type { FormInstance } from 'element-plus'
import { useTournamentStore } from '@/stores/tournamentStore'
import { useUserStore } from '@/stores/userStore'
import PlayerTierBadge from '@/components/PlayerTierBadge.vue'

const router = useRouter()
const store = useTournamentStore()
const userStore = useUserStore()
const formRef = ref<FormInstance>()

const form = ref({
  name: '',
  match_type: '1v1' as '1v1' | '2v2',
  player_ids: [] as string[],
  affects_score: true,
  entry_fee: 0,
})

onMounted(() => userStore.fetchUsers())

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
</style>
