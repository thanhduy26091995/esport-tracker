<template>
  <div class="page-wrapper">
    <div class="not-found-container">
      <div class="not-found-code">404</div>
      <div class="not-found-title">{{ t('common.notFound') }}</div>
      <div class="not-found-desc">{{ t('common.notFoundDesc') }}</div>
      <router-link to="/">
        <el-button type="primary">{{ t('common.backHome') }}</el-button>
      </router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<style scoped>
.not-found-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: 12px;
  text-align: center;
  padding: 40px;
}

.not-found-code {
  font-size: 80px;
  font-weight: 800;
  color: var(--text-muted);
  line-height: 1;
}

.not-found-title {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
}

.not-found-desc {
  font-size: 14px;
  color: var(--text-muted);
  margin-bottom: 8px;
}
</style>
