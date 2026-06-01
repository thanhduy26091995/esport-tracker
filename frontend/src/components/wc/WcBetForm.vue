<template>
  <el-dialog
    v-model="visible"
    :title="t('wc.betForm')"
    width="520px"
    destroy-on-close
    @close="reset"
  >
    <div v-if="match" class="wc-bet-match-header">
      <span class="wc-bet-team">{{ match.home_team }}</span>
      <span class="wc-bet-vs">vs</span>
      <span class="wc-bet-team">{{ match.away_team }}</span>
    </div>

    <el-tabs v-model="activeTab" class="wc-bet-tabs">
      <!-- HANDICAP TAB -->
      <el-tab-pane :label="t('wc.betTypeHandicap')" name="handicap">
        <div v-if="!hasHandicap" class="wc-bet-empty">
          <el-icon><InfoFilled /></el-icon>
          Chưa có kèo chấp cho trận này.
        </div>
        <div v-else class="wc-handicap-form">
          <div class="wc-hc-sides">
            <div
              class="wc-hc-side"
              :class="{ 'wc-hc-side--selected': handicapChoice === 'home' }"
              @click="selectHandicap('home')"
            >
              <div class="wc-hc-team">{{ match!.home_team }}</div>
              <div class="wc-hc-odds">
                <span v-if="homeGives" class="wc-hc-handi"
                  >-{{ match!.handicap_value }}</span
                >
                <span v-else class="wc-hc-handi wc-hc-handi--receive"
                  >+{{ match!.handicap_value }}</span
                >
                <span class="wc-hc-rate"
                  >@ {{ match!.odds_handicap_home?.toFixed(2) }}</span
                >
              </div>
            </div>

            <div
              class="wc-hc-side"
              :class="{ 'wc-hc-side--selected': handicapChoice === 'away' }"
              @click="selectHandicap('away')"
            >
              <div class="wc-hc-team">{{ match!.away_team }}</div>
              <div class="wc-hc-odds">
                <span v-if="!homeGives" class="wc-hc-handi"
                  >-{{ match!.handicap_value }}</span
                >
                <span v-else class="wc-hc-handi wc-hc-handi--receive"
                  >+{{ match!.handicap_value }}</span
                >
                <span class="wc-hc-rate"
                  >@ {{ match!.odds_handicap_away?.toFixed(2) }}</span
                >
              </div>
            </div>
          </div>

          <div v-if="handicapChoice" class="wc-stake-row">
            <el-input-number
              v-model="handicapStake"
              :min="1"
              :placeholder="t('wc.stake')"
              controls-position="right"
              style="width: 160px"
            />
            <div class="wc-payout-preview">
              <span class="wc-payout-label">{{ t("wc.payout") }}</span>
              <span class="wc-payout-value">{{ handicapPayout }}</span>
            </div>
          </div>
        </div>
      </el-tab-pane>

      <!-- EXACT SCORE TAB -->
      <el-tab-pane :label="t('wc.betTypeExactScore')" name="exact_score">
        <div v-if="!scoreOdds || scoreOdds.length === 0" class="wc-bet-empty">
          <el-icon><InfoFilled /></el-icon>
          Chưa có tỉ số nào được cấu hình cho trận này.
        </div>
        <div v-else class="wc-score-grid">
          <div
            v-for="so in scoreOdds"
            :key="so.id"
            class="wc-score-card"
            :class="{ 'wc-score-card--selected': isScoreSelected(so.id) }"
            @click="toggleScore(so)"
          >
            <div class="wc-sc-score">
              {{ so.home_score }}–{{ so.away_score }}
            </div>
            <div class="wc-sc-odds">x{{ so.odds.toFixed(2) }}</div>
            <div
              v-if="isScoreSelected(so.id)"
              class="wc-sc-stake-row"
              @click.stop
            >
              <el-input-number
                v-model="selectedScores[so.id].stake"
                :min="1"
                controls-position="right"
                size="small"
                style="width: 110px"
              />
              <span class="wc-sc-payout-preview">
                = {{ Math.floor(selectedScores[so.id].stake * so.odds) }}
              </span>
            </div>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>

    <template #footer>
      <div class="wc-bet-footer">
        <span class="wc-bet-count" v-if="totalBetCount > 0">
          {{ totalBetCount }} cược
        </span>
        <el-button @click="visible = false">{{ t("common.cancel") }}</el-button>
        <el-button
          plain
          type="success"
          :loading="submitting"
          :disabled="!canSubmit"
          @click="handleSubmit"
        >
          {{ t("wc.submitBet") }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { useI18n } from "vue-i18n";
import { InfoFilled } from "@element-plus/icons-vue";
import { ElMessage } from "element-plus";
import { wcService } from "@/services/wcService";
import type { WcMatchWithOdds, WcScoreOdds } from "@/types/wc";

const { t } = useI18n();

const props = defineProps<{
  modelValue: boolean;
  match: WcMatchWithOdds | null;
  scoreOdds: WcScoreOdds[];
}>();

const emit = defineEmits<{
  (e: "update:modelValue", v: boolean): void;
  (e: "bet-placed"): void;
}>();

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit("update:modelValue", v),
});

const activeTab = ref<"handicap" | "exact_score">("handicap");
const handicapChoice = ref<"home" | "away" | null>(null);
const handicapStake = ref(100);
const selectedScores = ref<Record<string, { stake: number; odds: number }>>({});
const submitting = ref(false);

const hasHandicap = computed(
  () =>
    !!(
      props.match?.handicap_value &&
      props.match.odds_handicap_home &&
      props.match.odds_handicap_away
    ),
);

const homeGives = computed(() => props.match?.handicap_team === "home");

const handicapOdds = computed(() => {
  if (!props.match) return 1;
  return handicapChoice.value === "home"
    ? (props.match.odds_handicap_home ?? 1)
    : (props.match.odds_handicap_away ?? 1);
});

const handicapPayout = computed(() =>
  Math.floor(handicapStake.value * handicapOdds.value),
);

function selectHandicap(side: "home" | "away") {
  handicapChoice.value = handicapChoice.value === side ? null : side;
}

function isScoreSelected(id: string) {
  return !!selectedScores.value[id];
}

function toggleScore(so: WcScoreOdds) {
  if (selectedScores.value[so.id]) {
    delete selectedScores.value[so.id];
  } else {
    selectedScores.value[so.id] = { stake: 100, odds: so.odds };
  }
}

const totalBetCount = computed(() => {
  let count = 0;
  if (activeTab.value === "handicap" && handicapChoice.value) count++;
  if (activeTab.value === "exact_score")
    count += Object.keys(selectedScores.value).length;
  return count;
});

const canSubmit = computed(() => {
  if (activeTab.value === "handicap")
    return !!handicapChoice.value && handicapStake.value > 0;
  return Object.keys(selectedScores.value).length > 0;
});

function reset() {
  handicapChoice.value = null;
  handicapStake.value = 100;
  selectedScores.value = {};
  activeTab.value = "handicap";
}

async function handleSubmit() {
  if (!props.match) return;
  submitting.value = true;
  try {
    if (activeTab.value === "handicap" && handicapChoice.value) {
      await wcService.placeBet({
        match_id: props.match.id,
        bet_type: "handicap",
        bet_choice: handicapChoice.value,
        stake: handicapStake.value,
      });
    } else if (activeTab.value === "exact_score") {
      const scoreBets = props.scoreOdds.filter(
        (so) => selectedScores.value[so.id],
      );
      for (const so of scoreBets) {
        await wcService.placeBet({
          match_id: props.match.id,
          bet_type: "exact_score",
          predicted_home_score: so.home_score,
          predicted_away_score: so.away_score,
          stake: selectedScores.value[so.id].stake,
        });
      }
    }
    ElMessage.success(t("wc.betSuccess"));
    emit("bet-placed");
    visible.value = false;
  } finally {
    submitting.value = false;
  }
}

watch(
  () => props.modelValue,
  (v) => {
    if (!v) reset();
  },
);
</script>

<style scoped>
.wc-bet-match-header {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 8px 0 16px;
  border-bottom: 1px solid var(--border-subtle);
  margin-bottom: 4px;
}

.wc-bet-team {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
}

.wc-bet-vs {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
}

.wc-bet-tabs :deep(.el-tabs__nav-wrap::after) {
  display: none;
}

.wc-bet-empty {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-muted);
  font-size: 13px;
  padding: 20px 0;
  justify-content: center;
}

.wc-handicap-form {
  padding: 4px 0;
}

.wc-hc-sides {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin-bottom: 16px;
}

.wc-hc-side {
  border: 2px solid var(--border-default);
  border-radius: 12px;
  padding: 14px;
  text-align: center;
  cursor: pointer;
  transition: all 0.15s;
  user-select: none;
}

.wc-hc-side:hover {
  border-color: #16a34a;
  background: rgba(22, 163, 74, 0.04);
}

.wc-hc-side--selected {
  border-color: #16a34a;
  background: rgba(22, 163, 74, 0.08);
  box-shadow: 0 0 0 1px #16a34a;
}

.wc-hc-team {
  font-size: 13px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 6px;
}

.wc-hc-handi {
  font-size: 16px;
  font-weight: 800;
  color: #ef4444;
}

.wc-hc-handi--receive {
  color: #16a34a;
}

.wc-hc-rate {
  display: block;
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 2px;
}

.wc-stake-row {
  display: flex;
  align-items: center;
  gap: 16px;
}

.wc-payout-preview {
  display: flex;
  flex-direction: column;
}

.wc-payout-label {
  font-size: 11px;
  color: var(--text-muted);
}

.wc-payout-value {
  font-size: 20px;
  font-weight: 800;
  color: #16a34a;
  tabular-nums: true;
}

.wc-score-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 8px;
  padding: 4px 0;
  max-height: 300px;
  overflow-y: auto;
}

.wc-score-card {
  border: 2px solid var(--border-default);
  border-radius: 10px;
  padding: 10px;
  text-align: center;
  cursor: pointer;
  transition: all 0.15s;
  user-select: none;
}

.wc-score-card:hover {
  border-color: #16a34a;
  background: rgba(22, 163, 74, 0.04);
}

.wc-score-card--selected {
  border-color: #16a34a;
  background: rgba(22, 163, 74, 0.08);
  grid-column: span 2;
}

.wc-sc-score {
  font-size: 18px;
  font-weight: 800;
  color: var(--text-primary);
  tabular-nums: true;
}

.wc-sc-odds {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 2px;
}

.wc-sc-stake-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  justify-content: center;
  flex-wrap: wrap;
}

.wc-sc-payout-preview {
  font-size: 14px;
  font-weight: 700;
  color: #16a34a;
}

.wc-bet-footer {
  display: flex;
  align-items: center;
  gap: 8px;
}

.wc-bet-count {
  font-size: 12px;
  color: var(--text-muted);
  margin-right: auto;
}
</style>
