<template>
  <el-dialog
    v-model="visible"
    :title="t('wc.predictionForm')"
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
      <el-tab-pane :label="t('wc.predictionTypeHandicap')" name="handicap">
        <div v-if="!hasHandicap" class="wc-bet-empty">
          <el-icon><InfoFilled /></el-icon>
          Chưa có chấp điểm cho trận này.
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
                  >-{{ fmtHandicap(match!.handicap_value ?? 0) }}</span
                >
                <span v-else class="wc-hc-handi wc-hc-handi--receive"
                  >+{{ fmtHandicap(match!.handicap_value ?? 0) }}</span
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
                  >-{{ fmtHandicap(match!.handicap_value ?? 0) }}</span
                >
                <span v-else class="wc-hc-handi wc-hc-handi--receive"
                  >+{{ fmtHandicap(match!.handicap_value ?? 0) }}</span
                >
                <span class="wc-hc-rate"
                  >@ {{ match!.odds_handicap_away?.toFixed(2) }}</span
                >
              </div>
            </div>
          </div>

          <div v-if="handicapChoice" class="wc-stake-row">
            <el-input-number
              v-model="handicapPoints"
              :min="1"
              :max="5"
              :placeholder="t('wc.points')"
              controls-position="right"
              style="width: 160px"
            />
            <div v-if="isQuarterHandicap" class="wc-payout-split">
              <div class="wc-payout-split-row wc-payout-split--win">
                <span>Thắng cả</span>
                <span>+{{ handicapSplitPayout.winProfit }}</span>
              </div>
              <div class="wc-payout-split-row wc-payout-split--win-half">
                <span>Thắng nửa</span>
                <span>+{{ handicapSplitPayout.winHalfProfit }}</span>
              </div>
              <div class="wc-payout-split-row wc-payout-split--lose-half">
                <span>Thua nửa</span>
                <span>-{{ handicapSplitPayout.loseHalfLoss }}</span>
              </div>
              <div class="wc-payout-split-row wc-payout-split--lose">
                <span>Thua cả</span>
                <span>-{{ handicapPoints }}</span>
              </div>
            </div>
            <div v-else class="wc-payout-preview">
              <span class="wc-payout-label">{{ t("wc.pointsEarned") }}</span>
              <span class="wc-payout-value">+{{ handicapPayoutPreview }}</span>
            </div>
          </div>
        </div>
      </el-tab-pane>

      <!-- EXACT SCORE TAB -->
      <el-tab-pane :label="t('wc.predictionTypeExactScore')" name="exact_score">
        <div v-if="!scoreMultipliers || scoreMultipliers.length === 0" class="wc-bet-empty">
          <el-icon><InfoFilled /></el-icon>
          Chưa có tỉ số nào được cấu hình cho trận này.
        </div>
        <div v-else class="wc-score-grid">
          <div
            v-for="so in scoreMultipliers"
            :key="so.id"
            class="wc-score-card"
            :class="{ 'wc-score-card--selected': isScoreSelected(so.id) }"
            @click="toggleScore(so)"
          >
            <div class="wc-sc-score">
              {{ so.home_score }}–{{ so.away_score }}
            </div>
            <div class="wc-sc-odds">x{{ so.multiplier.toFixed(2) }}</div>
            <div
              v-if="isScoreSelected(so.id)"
              class="wc-sc-stake-row"
              @click.stop
            >
              <el-input-number
                v-model="selectedScores[so.id].points"
                :min="1"
                :max="5"
                controls-position="right"
                size="small"
                style="width: 110px"
              />
              <span class="wc-sc-payout-preview">
                +{{ +(selectedScores[so.id].points * so.multiplier - selectedScores[so.id].points).toFixed(2) }}
              </span>
            </div>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>

    <template #footer>
      <div class="wc-bet-footer">
        <span class="wc-bet-count" v-if="totalPredictionCount > 0">
          {{ totalPredictionCount }} dự đoán
        </span>
        <el-button @click="visible = false">{{ t("common.cancel") }}</el-button>
        <el-button
          plain
          type="success"
          :loading="submitting"
          :disabled="!canSubmit"
          @click="handleSubmit"
        >
          {{ t("wc.submitPrediction") }}
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
import type { WcMatchWithOdds, WcScoreMultiplier, WcPredictionWithMatch } from "@/types/wc";

const { t } = useI18n();

const props = defineProps<{
  modelValue: boolean;
  match: WcMatchWithOdds | null;
  scoreMultipliers: WcScoreMultiplier[];
  existingPredictions?: WcPredictionWithMatch[];
}>();

const emit = defineEmits<{
  (e: "update:modelValue", v: boolean): void;
  (e: "prediction-placed"): void;
}>();

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit("update:modelValue", v),
});

const activeTab = ref<"handicap" | "exact_score">("handicap");
const handicapChoice = ref<"home" | "away" | null>(null);
const handicapPoints = ref(2);
const selectedScores = ref<Record<string, { points: number; multiplier: number }>>({});
const submitting = ref(false);

const existingHandicapPrediction = ref<WcPredictionWithMatch | null>(null);
const existingScorePredictionMap = ref<Record<string, WcPredictionWithMatch>>({});

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

const handicapPayoutPreview = computed(() =>
  +(handicapPoints.value * handicapOdds.value - handicapPoints.value).toFixed(2),
);

function fmtHandicap(v: number): string {
  return parseFloat(v.toFixed(2)).toString()
}

const isQuarterHandicap = computed(() => {
  const h = props.match?.handicap_value ?? 0
  return Math.abs((Math.abs(h) % 0.5) - 0.25) < 0.001
})

const handicapSplitPayout = computed(() => {
  const s = handicapPoints.value
  const odds = handicapOdds.value
  const half = s / 2
  return {
    winProfit: +(s * odds - s).toFixed(2),
    winHalfProfit: +(half * odds + half - s).toFixed(2),
    loseHalfLoss: +(s - half).toFixed(2),
  }
})

function selectHandicap(side: "home" | "away") {
  handicapChoice.value = handicapChoice.value === side ? null : side;
}

function isScoreSelected(id: string) {
  return !!selectedScores.value[id];
}

function toggleScore(so: WcScoreMultiplier) {
  if (selectedScores.value[so.id]) {
    delete selectedScores.value[so.id];
  } else {
    selectedScores.value[so.id] = { points: 2, multiplier: so.multiplier };
  }
}

const totalPredictionCount = computed(() => {
  let count = 0;
  if (activeTab.value === "handicap" && handicapChoice.value) count++;
  if (activeTab.value === "exact_score")
    count += Object.keys(selectedScores.value).length;
  return count;
});

const canSubmit = computed(() => {
  if (activeTab.value === "handicap")
    return !!handicapChoice.value && handicapPoints.value > 0;
  return Object.keys(selectedScores.value).length > 0;
});

function reset() {
  handicapChoice.value = null;
  handicapPoints.value = 2;
  selectedScores.value = {};
  activeTab.value = "handicap";
  existingHandicapPrediction.value = null;
  existingScorePredictionMap.value = {};
}

function populateFromExisting() {
  const predictions = props.existingPredictions ?? [];
  let hasHc = false;
  let hasScore = false;
  for (const pred of predictions) {
    if (pred.prediction_type === "handicap" && pred.prediction_choice) {
      existingHandicapPrediction.value = pred;
      handicapChoice.value = pred.prediction_choice as "home" | "away";
      handicapPoints.value = pred.points;
      hasHc = true;
    } else if (pred.prediction_type === "exact_score") {
      const so = props.scoreMultipliers.find(
        (s) =>
          s.home_score === pred.predicted_home_score &&
          s.away_score === pred.predicted_away_score,
      );
      if (so) {
        existingScorePredictionMap.value[so.id] = pred;
        selectedScores.value[so.id] = { points: pred.points, multiplier: so.multiplier };
        hasScore = true;
      }
    }
  }
  if (!hasHc && hasScore) activeTab.value = "exact_score";
}

async function handleSubmit() {
  if (!props.match) return;
  submitting.value = true;
  try {
    if (activeTab.value === "handicap" && handicapChoice.value) {
      const existing = existingHandicapPrediction.value;
      if (existing && existing.prediction_choice === handicapChoice.value) {
        if (existing.points !== handicapPoints.value) {
          await wcService.updatePredictionPoints(existing.id, handicapPoints.value);
        }
      } else {
        if (existing) await wcService.deletePrediction(existing.id);
        await wcService.submitPrediction({
          match_id: props.match.id,
          prediction_type: "handicap",
          prediction_choice: handicapChoice.value,
          points: handicapPoints.value,
        });
      }
    } else if (activeTab.value === "exact_score") {
      for (const so of props.scoreMultipliers.filter((s) => selectedScores.value[s.id])) {
        const existing = existingScorePredictionMap.value[so.id];
        const points = selectedScores.value[so.id].points;
        if (existing) {
          if (existing.points !== points) {
            await wcService.updatePredictionPoints(existing.id, points);
          }
        } else {
          await wcService.submitPrediction({
            match_id: props.match.id,
            prediction_type: "exact_score",
            predicted_home_score: so.home_score,
            predicted_away_score: so.away_score,
            points,
          });
        }
      }
    }
    ElMessage.success(t("wc.predictionSuccess"));
    emit("prediction-placed");
    visible.value = false;
  } finally {
    submitting.value = false;
  }
}

watch(
  () => props.modelValue,
  (v) => {
    reset();
    if (v && props.existingPredictions?.length) populateFromExisting();
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
  font-variant-numeric: tabular-nums;
}

.wc-score-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
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
  font-variant-numeric: tabular-nums;
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

.wc-payout-split {
  display: flex;
  flex-direction: column;
  gap: 3px;
  flex: 1;
}

.wc-payout-split-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 4px;
  font-variant-numeric: tabular-nums;
}

.wc-payout-split--win {
  background: rgba(22, 163, 74, 0.12);
  color: #16a34a;
}

.wc-payout-split--win-half {
  background: rgba(22, 163, 74, 0.07);
  color: #16a34a;
}

.wc-payout-split--lose-half {
  background: rgba(239, 68, 68, 0.07);
  color: #ef4444;
}

.wc-payout-split--lose {
  background: rgba(239, 68, 68, 0.12);
  color: #ef4444;
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
