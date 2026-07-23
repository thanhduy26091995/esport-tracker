<template>
  <div class="wc-admin-panel">
    <!-- Feature Toggle -->
    <div class="card card-body wc-admin-section">
      <div class="wc-admin-section-title">{{ t("wc.featureToggle") }}</div>
      <div class="wc-feature-toggle-row">
        <span
          class="wc-feature-status"
          :class="configEnabled ? 'wc-feat--on' : 'wc-feat--off'"
        >
          {{ configEnabled ? t("wc.featureEnabled") : t("wc.featureDisabled") }}
        </span>
        <el-switch
          v-model="configEnabled"
          :loading="togglingFeature"
          @change="handleFeatureToggle"
        />
      </div>
    </div>

    <!-- Bet Limits -->
    <div class="card card-body wc-admin-section">
      <div class="wc-admin-section-title">Giới hạn dự đoán</div>
      <div class="wc-feature-toggle-row" style="gap: 16px; flex-wrap: wrap;">
        <el-form-item label="Min điểm" style="margin: 0">
          <el-input-number
            v-model="betLimitForm.minPoints"
            :min="1"
            :max="betLimitForm.maxPoints"
            controls-position="right"
            style="width: 100px"
          />
        </el-form-item>
        <el-form-item label="Max điểm" style="margin: 0">
          <el-input-number
            v-model="betLimitForm.maxPoints"
            :min="betLimitForm.minPoints"
            controls-position="right"
            style="width: 100px"
          />
        </el-form-item>
        <el-button type="primary" :loading="savingBetLimits" @click="handleSaveBetLimits">Lưu</el-button>
      </div>
    </div>

    <!-- Penalty Config -->
    <div class="card card-body wc-admin-section">
      <div class="wc-admin-section-title">{{ t('wc.penaltyConfig') }}</div>
      <div style="display: flex; flex-direction: column; gap: 12px;">
        <div class="wc-feature-toggle-row" style="gap: 16px; flex-wrap: wrap; align-items: center;">
          <el-form-item :label="t('wc.cancelPenaltyEnabledLabel')" style="margin: 0">
            <el-switch v-model="penaltyForm.cancelPenaltyEnabled" />
          </el-form-item>
          <el-form-item :label="t('wc.cancelPenaltyPercentLabel')" style="margin: 0">
            <el-input-number
              v-model="penaltyForm.cancelPenaltyPercent"
              :min="0"
              :max="100"
              controls-position="right"
              style="width: 100px"
            />
            <span style="margin-left: 6px; color: #888;">%</span>
          </el-form-item>
        </div>
        <div class="wc-feature-toggle-row" style="gap: 16px; flex-wrap: wrap; align-items: center;">
          <el-form-item :label="t('wc.reduceMaxPercentLabel')" style="margin: 0">
            <el-input-number
              v-model="penaltyForm.betReduceMaxPercent"
              :min="0"
              :max="100"
              controls-position="right"
              style="width: 100px"
            />
            <span style="margin-left: 6px; color: #888;">%</span>
          </el-form-item>
          <el-form-item :label="t('wc.reducePenaltyPercentLabel')" style="margin: 0">
            <el-input-number
              v-model="penaltyForm.betReducePenaltyPercent"
              :min="0"
              :max="100"
              controls-position="right"
              style="width: 100px"
            />
            <span style="margin-left: 6px; color: #888;">%</span>
          </el-form-item>
          <el-button type="primary" :loading="savingPenaltyConfig" @click="handleSavePenaltyConfig">Lưu</el-button>
        </div>
        <div style="display: flex; align-items: center; gap: 12px; padding-top: 4px; border-top: 1px solid var(--border-default);">
          <div style="font-size: 13px; color: var(--text-secondary); flex: 1">
            {{ t('wc.backfillOriginalPointsDesc') }}
          </div>
          <el-button type="warning" :loading="backfillingOriginalPoints" @click="handleBackfillOriginalPoints">
            {{ t('wc.backfillOriginalPointsBtn') }}
          </el-button>
        </div>
      </div>
    </div>

    <!-- Match Management -->
    <div class="card card-body wc-admin-section">
      <div class="wc-admin-section-title">Quản lý trận đấu</div>
      <div class="wc-admin-row">
        <template v-if="!isAc">
          <el-button
            type="primary"
            :loading="syncing"
            @click="handleSync"
            :icon="Refresh"
          >
            {{ t("wc.syncMatches") }}
          </el-button>
          <el-button type="info" plain @click="mappingDialogRef?.open()">
            Setup StatsAPI Mapping
          </el-button>
        </template>
        <el-button v-if="isAc" type="primary" plain @click="createMatchDialogVisible = true">
          ➕ Thêm trận đấu
        </el-button>
        <el-button
          type="success"
          plain
          :loading="previewLoading && pendingAction === 'finalize-all'"
          @click="handleFinalizeAll"
        >
          Tính điểm toàn bộ
        </el-button>
        <el-button
          type="warning"
          plain
          :loading="previewLoading && pendingAction === 'refinalize-all'"
          @click="handleRefinalizeAll"
        >
          Tính lại toàn bộ (fix điểm float)
        </el-button>
      </div>

      <!-- Admin match filter bar -->
      <div class="wc-filter-bar">
        <el-input
          v-model="adminSearch"
          placeholder="Tìm đội bóng..."
          clearable
          size="small"
          class="wc-filter-search"
        >
          <template #prefix>🔍</template>
        </el-input>
        <div class="wc-filter-pills">
          <button
            v-for="f in adminFilterOptions"
            :key="f.key"
            class="wc-filter-pill"
            :class="{ 'wc-filter-pill--active': adminFilter === f.key }"
            @click="adminFilter = f.key"
          >
            {{ f.label }}
            <span v-if="f.count > 0" class="wc-filter-count">{{
              f.count
            }}</span>
          </button>
        </div>
      </div>

      <div class="wc-admin-match-list">
        <div v-if="adminFiltered.length === 0" class="wc-admin-empty">
          Không tìm thấy trận đấu nào.
        </div>
        <div
          v-for="match in adminFiltered"
          :key="match.id"
          class="wc-admin-match-row"
        >
          <div class="wc-admin-match-name">
            {{ match.home_team }} vs {{ match.away_team }}
            <span class="wc-admin-match-date">{{
              formatDate(match.match_date)
            }}</span>
            <span v-if="match.statsapi_fixture_id" class="wc-sync-chip wc-sync-chip--ok">
              Mapped
            </span>
            <span v-if="match.odds_synced_at" class="wc-sync-chip wc-sync-chip--synced">
              HDP {{ formatSyncTime(match.odds_synced_at) }}
            </span>
            <span v-if="match.ou_synced_at" class="wc-sync-chip wc-sync-chip--synced">
              O/U {{ formatSyncTime(match.ou_synced_at) }}
            </span>
            <span v-if="match.poisson_synced_at" class="wc-sync-chip wc-sync-chip--synced">
              Poisson {{ formatSyncTime(match.poisson_synced_at) }}
            </span>
            <span v-if="match.handicap_value != null" class="wc-val-chip">
              HDP
              {{ match.handicap_team === "away" ? match.away_team : match.home_team }}
              {{ match.handicap_value }}
              <template v-if="match.odds_handicap_home != null">
                @{{ match.odds_handicap_home?.toFixed(2) }}/{{ match.odds_handicap_away?.toFixed(2) }}
              </template>
            </span>
            <span v-if="match.ou_line != null" class="wc-val-chip">
              T/X {{ match.ou_line }}
              <template v-if="match.odds_over != null">
                @{{ match.odds_over?.toFixed(2) }}/{{ match.odds_under?.toFixed(2) }}
              </template>
            </span>
            <span v-if="(match.custom_bet_count ?? 0) > 0" class="wc-val-chip wc-val-chip--custom">
              {{ match.custom_bet_count }} kèo phụ
            </span>
          </div>
          <div class="wc-admin-match-actions">
            <el-button
              size="small"
              plain
              @click="openEditMatchDialog(match)"
            >
              ✏️ Sửa
            </el-button>
            <el-button
              v-if="!match.predictions_open"
              size="small"
              type="success"
              plain
              @click="handleOpen(match.id)"
            >
              🔓 Mở dự đoán
            </el-button>
            <el-button
              v-else
              size="small"
              type="warning"
              plain
              @click="handleClose(match.id)"
              :icon="Lock"
            >
              Đóng dự đoán
            </el-button>
            <el-button
              plain
              size="small"
              type="success"
              @click="handleSettle(match.id)"
              :disabled="match.status !== 'completed' || match.home_score == null || match.away_score == null"
              :title="match.home_score == null ? 'Cần nhập tỉ số trước' : ''"
            >
              {{ t("wc.finalizeMatch") }}
            </el-button>
            <el-button
              plain
              size="small"
              type="warning"
              @click="openScoreMultipliersDialog(match)"
            >
              {{ t("wc.scoreMultipliers") }}
            </el-button>
            <el-button
              plain
              size="small"
              type="info"
              @click="openHandicapDialog(match)"
            >
              Chấp điểm
            </el-button>
            <el-button
              plain
              size="small"
              type="info"
              @click="openOUDialog(match)"
            >
              Tài Xỉu
            </el-button>
            <template v-if="!isAc">
              <el-button
                plain
                size="small"
                type="primary"
                @click="importHandicapDialogRef?.open(match)"
              >
                HDP API
              </el-button>
              <el-button
                plain
                size="small"
                type="primary"
                @click="importOUDialogRef?.open(match)"
              >
                O/U API
              </el-button>
              <el-button
                plain
                size="small"
                @click="poissonDialogRef?.open(match)"
              >
                Poisson
              </el-button>
            </template>
            <el-button
              plain
              size="small"
              type="primary"
              @click="customBetPanelRef?.open(match.id, `${match.home_team} vs ${match.away_team}`)"
            >
              Kèo phụ
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- StatsAPI Sync Logs (WC only) -->
    <div v-if="!isAc" class="card card-body wc-admin-section">
      <div class="wc-admin-section-title">StatsAPI Sync</div>
      <WcSyncLogsPanel ref="syncLogsRef" />
    </div>

    <!-- User & Wallet Management -->
    <div class="card card-body wc-admin-section">
      <div class="wc-admin-section-title">{{ t("wc.userManagement") }}</div>
      <div class="wc-user-table">
        <div v-for="user in store.allUsers" :key="user.id" class="wc-user-row">
          <div class="wc-user-info-col">
            <span class="wc-user-name-col">{{ user.name }}</span>
            <span v-if="user.is_admin" class="wc-admin-tag">Admin</span>
            <span v-if="user.is_blocked" class="wc-blocked-tag">Bị khóa</span>
            <span v-if="user.is_bot" class="wc-bot-tag">Bot</span>
          </div>
          <div class="wc-user-wallet-col">
            <span class="wc-user-balance">
              {{ fmtPts(walletBalance(user.id)) }} pts
            </span>
          </div>
          <div class="wc-user-actions-col">
            <el-button size="small" @click="openTopUpDialog(user)">
              {{ t("wc.topUp") }}
            </el-button>
            <el-button
              size="small"
              :type="user.is_admin ? 'danger' : 'primary'"
              text
              @click="handleRoleToggle(user)"
            >
              {{ user.is_admin ? t("wc.removeAdmin") : t("wc.makeAdmin") }}
            </el-button>
            <el-button
              size="small"
              :type="user.is_blocked ? 'success' : 'danger'"
              text
              :disabled="user.id === authStore.user?.id"
              @click="handleToggleBlock(user)"
            >
              {{ user.is_blocked ? 'Mở khóa' : 'Khóa' }}
            </el-button>
            <el-button
              size="small"
              text
              @click="handleBotToggle(user)"
            >
              {{ user.is_bot ? 'Bỏ Bot' : 'Đánh dấu Bot' }}
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- House P&L -->
    <WcHousePnL ref="pnlRef" />

    <!-- Champion Prediction Admin -->
    <WcChampionAdminPanel />

    <!-- Settlement Panel -->
    <div class="card card-body wc-admin-section">
      <div class="wc-admin-section-title">{{ t("wc.settlementPanel") }}</div>
      <el-alert v-if="isAc" type="warning" :closable="false" style="margin-bottom:12px">
        Quyết toán sẽ reset <strong>toàn bộ ví</strong> của tất cả người dùng (bao gồm cả điểm World Cup). Ví WC và AC dùng chung một số dư.
      </el-alert>
      <el-tabs v-model="settlementTab">
        <el-tab-pane :label="t('wc.previewSettlement')" name="preview">
          <WcSettlementPreview />
        </el-tab-pane>
        <el-tab-pane :label="t('wc.settlementHistory')" name="history">
          <WcSettlementHistory :settlements="store.settlements" />
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- Finalize Preview Dialog -->
    <WcFinalizePreviewDialog
      v-model="previewDialogVisible"
      :title="previewTitle"
      :preview="previewData"
      :loading="previewLoading"
      :confirming="previewConfirming"
      @confirm="handleConfirmPreview"
      @cancel="previewDialogVisible = false"
    />

    <!-- Top-Up Dialog -->
    <el-dialog v-model="topUpVisible" :title="t('wc.topUp')" width="360px">
      <div v-if="topUpTarget" class="wc-topup-header">
        {{ topUpTarget.name }}
      </div>
      <el-form :model="topUpForm" label-position="top">
        <el-form-item :label="t('wc.topUpDelta')">
          <el-input-number
            v-model="topUpForm.delta"
            style="width: 100%"
            controls-position="right"
          />
        </el-form-item>
        <el-form-item :label="t('wc.topUpNote')">
          <el-input v-model="topUpForm.note" :placeholder="t('wc.topUpNote')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="topUpVisible = false">{{
          t("common.cancel")
        }}</el-button>
        <el-button type="primary" :loading="topping" @click="handleTopUp">
          {{ t("wc.topUp") }}
        </el-button>
      </template>
    </el-dialog>

    <!-- Handicap Dialog -->
    <el-dialog
      v-model="handicapVisible"
      title="Cấu hình chấp điểm"
      width="440px"
    >
      <div class="wc-so-match-name" v-if="handicapMatch">
        {{ handicapMatch.home_team }} vs {{ handicapMatch.away_team }}
      </div>
      <el-form
        :model="handicapForm"
        label-position="top"
        class="wc-handicap-config-form"
      >
        <el-form-item label="Đội chấp (Handicap Team)">
          <el-radio-group
            v-model="handicapForm.handicap_team"
            style="width: 100%"
          >
            <el-radio-button value="home">{{
              handicapMatch?.home_team ?? "Home"
            }}</el-radio-button>
            <el-radio-button value="away">{{
              handicapMatch?.away_team ?? "Away"
            }}</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="Số bàn chấp (Handicap Value)">
          <el-input-number
            v-model="handicapForm.handicap_value"
            :min="0"
            :max="5"
            :step="0.25"
            :precision="2"
            style="width: 100%"
          />
        </el-form-item>
        <div class="wc-handicap-odds-row">
          <el-form-item label="Kèo Home" style="flex: 1">
            <el-input-number
              v-model="handicapForm.odds_handicap_home"
              :min="1.01"
              :step="0.05"
              :precision="2"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item label="Kèo Away" style="flex: 1">
            <el-input-number
              v-model="handicapForm.odds_handicap_away"
              :min="1.01"
              :step="0.05"
              :precision="2"
              style="width: 100%"
            />
          </el-form-item>
        </div>
      </el-form>
      <template #footer>
        <el-button @click="handicapVisible = false">{{
          t("common.cancel")
        }}</el-button>
        <el-button
          type="primary"
          :loading="savingHandicap"
          @click="handleSaveHandicap"
        >
          Lưu chấp điểm
        </el-button>
      </template>
    </el-dialog>

    <!-- O/U Config Dialog -->
    <el-dialog v-model="ouVisible" title="Cấu hình Tài Xỉu" width="400px">
      <div class="wc-so-match-name" v-if="ouMatch">
        {{ ouMatch.home_team }} vs {{ ouMatch.away_team }}
      </div>
      <el-form :model="ouForm" label-position="top" class="wc-handicap-config-form">
        <el-form-item label="Đường kèo (O/U Line)">
          <el-input-number
            v-model="ouForm.ou_line"
            :min="0.5"
            :max="10"
            :step="0.25"
            :precision="2"
            style="width: 100%"
          />
        </el-form-item>
        <div class="wc-handicap-odds-row">
          <el-form-item label="Kèo Tài (Over)" style="flex: 1">
            <el-input-number
              v-model="ouForm.odds_over"
              :min="1.01"
              :step="0.05"
              :precision="2"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item label="Kèo Xỉu (Under)" style="flex: 1">
            <el-input-number
              v-model="ouForm.odds_under"
              :min="1.01"
              :step="0.05"
              :precision="2"
              style="width: 100%"
            />
          </el-form-item>
        </div>
      </el-form>
      <template #footer>
        <el-button @click="ouVisible = false">{{ t("common.cancel") }}</el-button>
        <el-button type="primary" :loading="savingOU" @click="handleSaveOU">
          Lưu Tài Xỉu
        </el-button>
      </template>
    </el-dialog>

    <!-- StatsAPI Dialogs (WC only) -->
    <template v-if="!isAc">
      <WcSetupMappingDialog ref="mappingDialogRef" @mapped="handleOddsImported" />
      <WcImportHandicapDialog ref="importHandicapDialogRef" @imported="handleOddsImported" />
      <WcImportOUDialog ref="importOUDialogRef" @imported="handleOddsImported" />
      <WcGeneratePoissonDialog ref="poissonDialogRef" @saved="handleOddsImported" />
    </template>

    <!-- Create Match Dialog (AC only) -->
    <el-dialog v-model="createMatchDialogVisible" title="Thêm trận đấu" width="480px">
      <el-form label-position="top" @submit.prevent>
        <el-form-item label="Đội nhà">
          <el-input v-model="createMatchForm.home_team" placeholder="Ví dụ: Vietnam" />
        </el-form-item>
        <el-form-item label="Đội khách">
          <el-input v-model="createMatchForm.away_team" placeholder="Ví dụ: Thailand" />
        </el-form-item>
        <el-form-item label="Mã đội nhà (tùy chọn)">
          <el-input v-model="createMatchForm.home_team_code" placeholder="VIE" />
        </el-form-item>
        <el-form-item label="Mã đội khách (tùy chọn)">
          <el-input v-model="createMatchForm.away_team_code" placeholder="THA" />
        </el-form-item>
        <el-form-item label="Ngày giờ thi đấu">
          <el-date-picker
            v-model="createMatchForm.match_date"
            type="datetime"
            placeholder="Chọn ngày giờ"
            format="DD/MM/YYYY HH:mm"
            value-format="YYYY-MM-DDTHH:mm:ss[Z]"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="Bảng / Giai đoạn (tùy chọn)">
          <el-input v-model="createMatchForm.group_name" placeholder="Bảng A" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createMatchDialogVisible = false">Hủy</el-button>
        <el-button type="primary" :loading="creatingMatch" @click="handleCreateMatch">Thêm trận</el-button>
      </template>
    </el-dialog>

    <!-- Edit Match Dialog -->
    <el-dialog v-model="editMatchDialogVisible" title="Sửa trận đấu" width="520px">
      <el-form v-if="editMatchTarget" label-position="top" @submit.prevent>
        <div style="display:flex;gap:12px">
          <el-form-item label="Đội nhà" style="flex:1">
            <el-input v-model="editMatchForm.home_team" />
          </el-form-item>
          <el-form-item label="Đội khách" style="flex:1">
            <el-input v-model="editMatchForm.away_team" />
          </el-form-item>
        </div>
        <div style="display:flex;gap:12px">
          <el-form-item label="Mã đội nhà" style="flex:1">
            <el-input v-model="editMatchForm.home_team_code" placeholder="VIE" />
          </el-form-item>
          <el-form-item label="Mã đội khách" style="flex:1">
            <el-input v-model="editMatchForm.away_team_code" placeholder="THA" />
          </el-form-item>
        </div>
        <el-form-item label="Ngày giờ thi đấu">
          <el-date-picker
            v-model="editMatchForm.match_date"
            type="datetime"
            placeholder="Chọn ngày giờ"
            format="DD/MM/YYYY HH:mm"
            value-format="YYYY-MM-DDTHH:mm:ss[Z]"
            style="width:100%"
          />
        </el-form-item>
        <el-divider>Kết quả</el-divider>
        <div style="display:flex;gap:12px;align-items:center">
          <el-form-item :label="`Bàn thắng – ${editMatchForm.home_team || 'Đội nhà'}`" style="flex:1">
            <el-input-number v-model="editMatchForm.home_score" :min="0" :precision="0" style="width:100%" />
          </el-form-item>
          <span style="margin-top:22px;font-size:18px;font-weight:700;color:var(--text-muted)">–</span>
          <el-form-item :label="`Bàn thắng – ${editMatchForm.away_team || 'Đội khách'}`" style="flex:1">
            <el-input-number v-model="editMatchForm.away_score" :min="0" :precision="0" style="width:100%" />
          </el-form-item>
        </div>
        <el-form-item label="Trạng thái">
          <el-select v-model="editMatchForm.status" style="width:100%">
            <el-option value="scheduled" label="Chưa diễn ra" />
            <el-option value="live" label="Đang diễn ra" />
            <el-option value="completed" label="Đã kết thúc" />
            <el-option value="cancelled" label="Hủy" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editMatchDialogVisible = false">Hủy</el-button>
        <el-button type="primary" :loading="savingEditMatch" @click="handleSaveEditMatch">Lưu</el-button>
      </template>
    </el-dialog>

    <WcAdminCustomBetPanel ref="customBetPanelRef" />

    <!-- Score Multipliers Dialog -->
    <el-dialog
      v-model="scoreMultipliersVisible"
      :title="t('wc.scoreMultipliers')"
      width="480px"
    >
      <div class="wc-so-match-name" v-if="scoreMultipliersMatch">
        {{ scoreMultipliersMatch.home_team }} vs {{ scoreMultipliersMatch.away_team }}
      </div>
      <div class="wc-so-list">
        <div v-for="so in currentScoreMultipliers" :key="so.id" class="wc-so-row mb-2 mt-2" >
          <span class="wc-so-score"
            >{{ so.home_score }}–{{ so.away_score }}</span
          >
          <el-input-number
            v-model="so.multiplier"
            :min="1.01"
            :step="0.05"
            :precision="2"
            size="small"
            style="width: 120px"
            @change="handleUpdateMultiplier(so.id, so.multiplier)"
          />
          <el-button
            size="small"
            type="danger"
            text
            @click="handleDeleteScoreMultiplier(so.id)"
          >
            {{ t("common.delete") }}
          </el-button>
        </div>
      </div>
      <el-divider />
      <div class="wc-so-add-form">
        <span class="wc-so-add-label">Thêm tỉ số:</span>
        <el-input-number
          v-model="newSo.homeScore"
          :min="0"
          :max="20"
          size="small"
          style="width: 80px"
        />
        <span>–</span>
        <el-input-number
          v-model="newSo.awayScore"
          :min="0"
          :max="20"
          size="small"
          style="width: 80px"
        />
        <el-input-number
          v-model="newSo.multiplier"
          :min="1.01"
          :step="0.05"
          :precision="2"
          size="small"
          style="width: 100px"
        />
        <el-button type="primary" size="small" @click="handleAddScoreMultiplier">
          {{ t("common.create") }}
        </el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";
import { Refresh, Lock } from "@element-plus/icons-vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { useWcStore } from "@/stores/wcStore";
import { useWcAuthStore } from "@/stores/wcAuthStore";
import { wcService } from "@/services/wcService";
import { useMatchFilter } from "@/composables/useMatchFilter";
import type { WcUser, WcMatch, WcScoreMultiplier, FinalizePreviewResult } from "@/types/wc";
import WcSettlementPreview from "./WcSettlementPreview.vue";
import WcSettlementHistory from "./WcSettlementHistory.vue";
import WcHousePnL from "./WcHousePnL.vue";
import WcSetupMappingDialog from "./WcSetupMappingDialog.vue";
import WcImportHandicapDialog from "./WcImportHandicapDialog.vue";
import WcImportOUDialog from "./WcImportOUDialog.vue";
import WcGeneratePoissonDialog from "./WcGeneratePoissonDialog.vue";
import WcSyncLogsPanel from "./WcSyncLogsPanel.vue";
import WcChampionAdminPanel from "./WcChampionAdminPanel.vue";
import WcFinalizePreviewDialog from "./WcFinalizePreviewDialog.vue";
import WcAdminCustomBetPanel from "./WcAdminCustomBetPanel.vue";

const { t } = useI18n();
const route = useRoute();
const isAc = computed(() => route.meta?.tournamentType === 'asean_cup');
const store = useWcStore();
const authStore = useWcAuthStore();

const storeMatches = computed(() => store.matches);
const {
  search: adminSearch,
  activeFilter: adminFilter,
  filtered: adminFiltered,
  counts: adminCounts,
} = useMatchFilter(storeMatches, "incoming");

const adminFilterOptions = computed(() => [
  {
    key: "incoming" as const,
    label: "Sắp tới",
    count: adminCounts.value.incoming,
  },
  { key: "open" as const, label: "Mở dự đoán", count: adminCounts.value.open },
  { key: "live" as const, label: "Đang diễn", count: adminCounts.value.live },
  { key: "locked" as const, label: "Đã khóa", count: adminCounts.value.locked },
  {
    key: "completed" as const,
    label: "Đã kết thúc",
    count: adminCounts.value.completed,
  },
  { key: "all" as const, label: "Tất cả", count: adminCounts.value.all },
]);

const settlementTab = ref("preview");
const pnlRef = ref<InstanceType<typeof WcHousePnL> | null>(null);
const mappingDialogRef = ref<InstanceType<typeof WcSetupMappingDialog> | null>(null);
const importHandicapDialogRef = ref<InstanceType<typeof WcImportHandicapDialog> | null>(null);
const importOUDialogRef = ref<InstanceType<typeof WcImportOUDialog> | null>(null);
const poissonDialogRef = ref<InstanceType<typeof WcGeneratePoissonDialog> | null>(null);
const syncLogsRef = ref<InstanceType<typeof WcSyncLogsPanel> | null>(null);
const customBetPanelRef = ref<InstanceType<typeof WcAdminCustomBetPanel> | null>(null);
const syncing = ref(false);
const togglingFeature = ref(false);
const createMatchDialogVisible = ref(false);
const creatingMatch = ref(false);
const createMatchForm = ref({
  home_team: '',
  away_team: '',
  home_team_code: '',
  away_team_code: '',
  match_date: '',
  group_name: '',
});
async function handleCreateMatch() {
  if (!createMatchForm.value.home_team || !createMatchForm.value.away_team || !createMatchForm.value.match_date) {
    ElMessage.warning('Vui lòng nhập đủ thông tin trận đấu');
    return;
  }
  creatingMatch.value = true;
  try {
    await (wcService as any).createMatch({
      home_team: createMatchForm.value.home_team,
      away_team: createMatchForm.value.away_team,
      home_team_code: createMatchForm.value.home_team_code || undefined,
      away_team_code: createMatchForm.value.away_team_code || undefined,
      match_date: createMatchForm.value.match_date,
      group_name: createMatchForm.value.group_name || undefined,
    });
    ElMessage.success('Đã thêm trận đấu');
    createMatchDialogVisible.value = false;
    createMatchForm.value = { home_team: '', away_team: '', home_team_code: '', away_team_code: '', match_date: '', group_name: '' };
    await store.fetchMatches();
  } finally {
    creatingMatch.value = false;
  }
}
// ── Edit Match ──────────────────────────────────────────────
const editMatchDialogVisible = ref(false);
const savingEditMatch = ref(false);
const editMatchTarget = ref<WcMatch | null>(null);
const editMatchForm = ref({
  home_team: '',
  away_team: '',
  home_team_code: '',
  away_team_code: '',
  match_date: '',
  home_score: null as number | null,
  away_score: null as number | null,
  status: 'scheduled' as string,
});

function openEditMatchDialog(match: WcMatch) {
  editMatchTarget.value = match;
  editMatchForm.value = {
    home_team: match.home_team,
    away_team: match.away_team,
    home_team_code: match.home_team_code ?? '',
    away_team_code: match.away_team_code ?? '',
    match_date: match.match_date,
    home_score: match.home_score ?? null,
    away_score: match.away_score ?? null,
    status: match.status,
  };
  editMatchDialogVisible.value = true;
}

async function handleSaveEditMatch() {
  if (!editMatchTarget.value) return;
  if (editMatchForm.value.status === 'completed' &&
      (editMatchForm.value.home_score === null || editMatchForm.value.away_score === null)) {
    ElMessage.warning('Vui lòng nhập tỉ số trước khi đặt trạng thái Đã kết thúc');
    return;
  }
  savingEditMatch.value = true;
  try {
    const fields: Record<string, unknown> = {
      home_team: editMatchForm.value.home_team,
      away_team: editMatchForm.value.away_team,
      match_date: editMatchForm.value.match_date,
      status: editMatchForm.value.status,
    };
    if (editMatchForm.value.home_team_code) fields.home_team_code = editMatchForm.value.home_team_code;
    if (editMatchForm.value.away_team_code) fields.away_team_code = editMatchForm.value.away_team_code;
    if (editMatchForm.value.status === 'completed') {
      fields.home_score = editMatchForm.value.home_score;
      fields.away_score = editMatchForm.value.away_score;
    }
    await wcService.updateMatch(editMatchTarget.value.id, fields);
    ElMessage.success('Đã cập nhật trận đấu');
    editMatchDialogVisible.value = false;
    await store.fetchMatches();
  } catch {
    ElMessage.error('Lỗi khi cập nhật trận đấu');
  } finally {
    savingEditMatch.value = false;
  }
}
// ────────────────────────────────────────────────────────────

const configEnabled = ref(store.config?.is_enabled ?? false);

const betLimitForm = ref({ minPoints: store.minPoints, maxPoints: store.maxPoints });
const savingBetLimits = ref(false);
async function handleSaveBetLimits() {
  savingBetLimits.value = true;
  try {
    await wcService.updateBetLimits(betLimitForm.value.minPoints, betLimitForm.value.maxPoints);
    await store.fetchPublicConfig();
    ElMessage.success('Đã cập nhật giới hạn dự đoán');
  } catch {
    ElMessage.error('Lỗi khi cập nhật giới hạn');
  } finally {
    savingBetLimits.value = false;
  }
}

const penaltyForm = ref({
  cancelPenaltyEnabled: store.config?.cancel_penalty_enabled ?? false,
  cancelPenaltyPercent: store.config?.cancel_penalty_percent ?? 20,
  betReduceMaxPercent: store.config?.bet_reduce_max_percent ?? 50,
  betReducePenaltyPercent: store.config?.bet_reduce_penalty_percent ?? 20,
});
const savingPenaltyConfig = ref(false);
async function handleSavePenaltyConfig() {
  savingPenaltyConfig.value = true;
  try {
    await wcService.updatePenaltyConfig({
      cancel_penalty_enabled: penaltyForm.value.cancelPenaltyEnabled,
      cancel_penalty_percent: penaltyForm.value.cancelPenaltyPercent,
      bet_reduce_max_percent: penaltyForm.value.betReduceMaxPercent,
      bet_reduce_penalty_percent: penaltyForm.value.betReducePenaltyPercent,
    });
    await store.fetchPublicConfig();
    ElMessage.success('Đã cập nhật cài đặt phạt cược');
  } catch {
    ElMessage.error('Lỗi khi cập nhật');
  } finally {
    savingPenaltyConfig.value = false;
  }
}

const backfillingOriginalPoints = ref(false);
async function handleBackfillOriginalPoints() {
  try {
    await ElMessageBox.confirm(
      t('wc.backfillOriginalPointsConfirm'),
      t('wc.backfillOriginalPointsBtn'),
      { confirmButtonText: t('wc.confirm'), cancelButtonText: t('wc.cancel'), type: 'warning' },
    );
  } catch {
    return;
  }
  backfillingOriginalPoints.value = true;
  try {
    const res = await wcService.backfillOriginalPoints();
    ElMessage.success(t('wc.backfillOriginalPointsDone', { count: res.rows_updated }));
  } catch {
    ElMessage.error('Lỗi khi backfill');
  } finally {
    backfillingOriginalPoints.value = false;
  }
}

type PendingAction = 'finalize-match' | 'finalize-all' | 'refinalize-all'
const previewDialogVisible = ref(false);
const previewData = ref<FinalizePreviewResult | null>(null);
const previewTitle = ref('');
const previewLoading = ref(false);
const previewConfirming = ref(false);
const pendingAction = ref<PendingAction | null>(null);
const pendingMatchId = ref<string | null>(null);

const topUpVisible = ref(false);
const topUpTarget = ref<WcUser | null>(null);
const topUpForm = ref({ delta: 0, note: "" });
const topping = ref(false);

const scoreMultipliersVisible = ref(false);
const scoreMultipliersMatch = ref<WcMatch | null>(null);
const currentScoreMultipliers = ref<WcScoreMultiplier[]>([]);
const newSo = ref({ homeScore: 0, awayScore: 0, multiplier: 3.0 });


const handicapVisible = ref(false);
const handicapMatch = ref<WcMatch | null>(null);
const savingHandicap = ref(false);
const handicapForm = ref({
  handicap_team: "home",
  handicap_value: 0.5,
  odds_handicap_home: 1.9,
  odds_handicap_away: 1.9,
});

const ouVisible = ref(false);
const ouMatch = ref<WcMatch | null>(null);
const savingOU = ref(false);
const ouForm = ref({ ou_line: 2.5, odds_over: 1.9, odds_under: 1.9 });

const walletMap = computed(() => {
  const m: Record<string, number> = {};
  for (const w of store.allWallets) m[w.wc_user_id] = w.balance;
  return m;
});

function walletBalance(userId: string) {
  return walletMap.value[userId] ?? 0;
}

function fmtPts(v: number): string {
  return parseFloat(v.toFixed(2)).toString()
}

async function handleSync() {
  syncing.value = true;
  try {
    await store.syncMatches();
  } finally {
    syncing.value = false;
  }
}

async function handleFeatureToggle(val: boolean) {
  togglingFeature.value = true;
  try {
    await store.updateConfig(val);
    if (store.config) store.config.is_enabled = val;
  } finally {
    togglingFeature.value = false;
  }
}

async function handleOpen(matchId: string) {
  await store.openMatch(matchId)
}

async function handleClose(matchId: string) {
  await store.closeMatch(matchId)
}

async function handleSettle(matchId: string) {
  previewTitle.value = t('wc.finalizeMatch');
  pendingAction.value = 'finalize-match';
  pendingMatchId.value = matchId;
  previewData.value = null;
  previewLoading.value = true;
  previewDialogVisible.value = true;
  try {
    previewData.value = await wcService.previewFinalizeMatch(matchId);
  } catch (e: unknown) {
    ElMessage.error((e as Error)?.message ?? 'Không thể tải preview');
    previewDialogVisible.value = false;
  } finally {
    previewLoading.value = false;
  }
}

async function handleFinalizeAll() {
  previewTitle.value = 'Tính điểm toàn bộ';
  pendingAction.value = 'finalize-all';
  pendingMatchId.value = null;
  previewData.value = null;
  previewLoading.value = true;
  previewDialogVisible.value = true;
  try {
    previewData.value = await wcService.previewFinalizeAll();
  } catch (e: unknown) {
    ElMessage.error((e as Error)?.message ?? 'Không thể tải preview');
    previewDialogVisible.value = false;
  } finally {
    previewLoading.value = false;
  }
}

async function handleRefinalizeAll() {
  previewTitle.value = 'Tính lại toàn bộ';
  pendingAction.value = 'refinalize-all';
  pendingMatchId.value = null;
  previewData.value = null;
  previewLoading.value = true;
  previewDialogVisible.value = true;
  try {
    previewData.value = await wcService.previewRefinalizeAll();
  } catch (e: unknown) {
    ElMessage.error((e as Error)?.message ?? 'Không thể tải preview');
    previewDialogVisible.value = false;
  } finally {
    previewLoading.value = false;
  }
}

async function handleConfirmPreview() {
  previewConfirming.value = true;
  try {
    if (pendingAction.value === 'finalize-match' && pendingMatchId.value) {
      await store.finalizeMatch(pendingMatchId.value);
    } else if (pendingAction.value === 'finalize-all') {
      await store.finalizeAll();
    } else if (pendingAction.value === 'refinalize-all') {
      await store.refinalizeAll();
    }
    previewDialogVisible.value = false;
    pnlRef.value?.load();
  } catch (e: unknown) {
    ElMessage.error((e as Error)?.message ?? 'Lỗi khi tính điểm');
  } finally {
    previewConfirming.value = false;
  }
}

async function handleToggleBlock(user: WcUser) {
  try {
    if (user.is_blocked) {
      await wcService.unblockUser(user.id);
      ElMessage.success(`Đã mở khóa ${user.name}`);
    } else {
      const res = await wcService.blockUser(user.id);
      ElMessage.success(`Đã khóa ${user.name}` + (res.voided_bets > 0 ? ` (void ${res.voided_bets} cược)` : ''));
    }
    await store.fetchAllUsers();
  } catch (e: unknown) {
    ElMessage.error((e as Error)?.message ?? 'Lỗi khi thay đổi trạng thái');
  }
}

function openTopUpDialog(user: WcUser) {
  topUpTarget.value = user;
  topUpForm.value = { delta: 0, note: "" };
  topUpVisible.value = true;
}

async function handleTopUp() {
  if (!topUpTarget.value) return;
  topping.value = true;
  try {
    await store.topUp(
      topUpTarget.value.id,
      topUpForm.value.delta,
      topUpForm.value.note,
    );
    topUpVisible.value = false;
  } finally {
    topping.value = false;
  }
}

async function handleRoleToggle(user: WcUser) {
  await store.setUserRole(user.id, !user.is_admin);
}

async function handleBotToggle(user: WcUser) {
  await store.setUserBot(user.id, !user.is_bot);
}

function openHandicapDialog(match: WcMatch) {
  handicapMatch.value = match;
  handicapForm.value = {
    handicap_team: match.handicap_team ?? "home",
    handicap_value: match.handicap_value ?? 0.5,
    odds_handicap_home: match.odds_handicap_home ?? 1.9,
    odds_handicap_away: match.odds_handicap_away ?? 1.9,
  };
  handicapVisible.value = true;
}

async function handleSaveHandicap() {
  if (!handicapMatch.value) return;
  savingHandicap.value = true;
  try {
    await wcService.updateMatch(handicapMatch.value.id, {
      handicap_team: handicapForm.value.handicap_team,
      handicap_value: handicapForm.value.handicap_value,
      odds_handicap_home: handicapForm.value.odds_handicap_home,
      odds_handicap_away: handicapForm.value.odds_handicap_away,
    });
    ElMessage.success("Đã lưu chấp điểm");
    handicapVisible.value = false;
    await store.fetchMatches();
  } finally {
    savingHandicap.value = false;
  }
}

function openOUDialog(match: WcMatch) {
  ouMatch.value = match;
  ouForm.value = {
    ou_line: match.ou_line ?? 2.5,
    odds_over: match.odds_over ?? 1.9,
    odds_under: match.odds_under ?? 1.9,
  };
  ouVisible.value = true;
}

async function handleSaveOU() {
  if (!ouMatch.value) return;
  savingOU.value = true;
  try {
    await wcService.updateMatch(ouMatch.value.id, {
      ou_line: ouForm.value.ou_line,
      odds_over: ouForm.value.odds_over,
      odds_under: ouForm.value.odds_under,
    });
    ElMessage.success("Đã lưu Tài Xỉu");
    ouVisible.value = false;
    await store.fetchMatches();
  } finally {
    savingOU.value = false;
  }
}

async function openScoreMultipliersDialog(match: WcMatch) {
  scoreMultipliersMatch.value = match;
  const multipliers = await wcService.getScoreMultipliers(match.id);
  currentScoreMultipliers.value = multipliers;
  scoreMultipliersVisible.value = true;
}

async function handleAddScoreMultiplier() {
  if (!scoreMultipliersMatch.value) return;
  const so = await wcService.addScoreMultiplier(
    scoreMultipliersMatch.value.id,
    newSo.value.homeScore,
    newSo.value.awayScore,
    newSo.value.multiplier,
  );
  currentScoreMultipliers.value.push(so);
  newSo.value = { homeScore: 0, awayScore: 0, multiplier: 3.0 };
}

async function handleUpdateMultiplier(id: string, multiplier: number) {
  await wcService.updateScoreMultiplier(id, multiplier);
}

async function handleDeleteScoreMultiplier(id: string) {
  await wcService.deleteScoreMultiplier(id);
  currentScoreMultipliers.value = currentScoreMultipliers.value.filter((so) => so.id !== id);
}

async function handleOddsImported() {
  await store.fetchMatches();
  syncLogsRef.value?.load();
}

function formatDate(s: string) {
  return new Date(s).toLocaleDateString("vi-VN", {
    day: "2-digit",
    month: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function formatSyncTime(s: string) {
  return new Date(s).toLocaleTimeString("vi-VN", {
    hour: "2-digit",
    minute: "2-digit",
  });
}

onMounted(async () => {
  await Promise.all([
    store.fetchConfig(),
    store.fetchMatches(),
    store.fetchAllUsers(),
    store.fetchAllWallets(),
    store.fetchSettlements(),
  ]);
  configEnabled.value = store.config?.is_enabled ?? false;
  if (store.config) {
    betLimitForm.value = { minPoints: store.config.min_points, maxPoints: store.config.max_points };
    penaltyForm.value = {
      cancelPenaltyEnabled: store.config.cancel_penalty_enabled,
      cancelPenaltyPercent: store.config.cancel_penalty_percent,
      betReduceMaxPercent: store.config.bet_reduce_max_percent,
      betReducePenaltyPercent: store.config.bet_reduce_penalty_percent,
    };
  }
});
</script>

<style scoped>
.wc-admin-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding-top: 8px;
}

.wc-admin-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.wc-admin-section-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-subtle);
}

.wc-feature-toggle-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.wc-feature-status {
  font-size: 13px;
  font-weight: 700;
  padding: 3px 10px;
  border-radius: 8px;
}

.wc-feat--on {
  background: rgba(22, 163, 74, 0.12);
  color: #16a34a;
}

.wc-feat--off {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.wc-admin-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.wc-filter-bar {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.wc-filter-search {
  max-width: 260px;
}

.wc-filter-pills {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.wc-filter-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 10px;
  border-radius: 20px;
  border: 1px solid var(--border-default);
  background: var(--surface-card);
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
}

.wc-filter-pill:hover {
  border-color: #16a34a60;
  color: var(--text-primary);
}

.wc-filter-pill--active {
  background: #16a34a;
  border-color: #16a34a;
  color: #fff;
}

.wc-filter-count {
  font-size: 11px;
  font-weight: 700;
  background: rgba(255, 255, 255, 0.25);
  border-radius: 8px;
  padding: 0 5px;
  line-height: 1.4;
}

.wc-filter-pill:not(.wc-filter-pill--active) .wc-filter-count {
  background: var(--surface-page);
  color: var(--text-muted);
}

.wc-admin-empty {
  font-size: 13px;
  color: var(--text-muted);
  padding: 16px 0;
  text-align: center;
}

.wc-admin-match-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 480px;
  overflow-y: auto;
}

.wc-admin-match-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  background: var(--surface-page);
  border-radius: 8px;
  flex-wrap: wrap;
}

.wc-admin-match-name {
  flex: 1;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  min-width: 150px;
}

.wc-admin-match-date {
  display: block;
  font-size: 11px;
  color: var(--text-muted);
  font-weight: 400;
}

.wc-sync-chip {
  display: inline-block;
  font-size: 10px;
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 8px;
  margin-left: 4px;
  vertical-align: middle;
}

.wc-sync-chip--ok {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
}

.wc-sync-chip--synced {
  background: rgba(22, 163, 74, 0.1);
  color: #16a34a;
}

.wc-val-chip {
  display: inline-block;
  font-size: 10px;
  font-weight: 700;
  padding: 1px 6px;
  border-radius: 8px;
  margin-left: 4px;
  vertical-align: middle;
  background: rgba(100, 116, 139, 0.12);
  color: var(--text-secondary);
}

.wc-val-chip--custom {
  background: rgba(217, 119, 6, 0.12);
  color: #d97706;
}

.wc-admin-match-actions {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.wc-user-table {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 480px;
  overflow-y: auto;
}

.wc-user-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  background: var(--surface-page);
  border-radius: 8px;
  flex-wrap: wrap;
}

.wc-user-info-col {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 120px;
}

.wc-user-name-col {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.wc-admin-tag {
  font-size: 10px;
  font-weight: 700;
  background: rgba(217, 119, 6, 0.12);
  color: #d97706;
  padding: 1px 6px;
  border-radius: 4px;
}

.wc-blocked-tag {
  font-size: 10px;
  font-weight: 700;
  background: rgba(239, 68, 68, 0.12);
  color: #ef4444;
  padding: 1px 6px;
  border-radius: 4px;
}

.wc-bot-tag {
  font-size: 10px;
  font-weight: 700;
  background: rgba(107, 114, 128, 0.12);
  color: #6b7280;
  padding: 1px 6px;
  border-radius: 4px;
}

.wc-user-wallet-col {
  flex-shrink: 0;
}

.wc-user-balance {
  font-size: 14px;
  font-weight: 700;
  tabular-nums: true;
  color: var(--text-primary);
}

.wc-user-actions-col {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.wc-topup-header {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 16px;
}

.wc-so-match-name {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 12px;
  text-align: center;
}

.wc-so-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 240px;
  overflow-y: auto;
}

.wc-so-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.wc-so-score {
  font-size: 16px;
  font-weight: 800;
  color: var(--text-primary);
  width: 50px;
  text-align: center;
}

.wc-so-add-form {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.wc-so-add-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
}

.wc-handicap-config-form {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.wc-handicap-odds-row {
  display: flex;
  gap: 16px;
}
</style>
