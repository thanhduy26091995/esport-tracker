<template>
  <div class="page-wrapper">
    <div class="page-container">
      <div class="page-header">
        <div class="page-header-left">
          <h1 class="page-title">Settings</h1>
          <p class="page-subtitle">Application configuration and preferences</p>
        </div>
      </div>

      <div class="card">
        <div class="card-body">
          <div class="empty-state">
            <el-icon :size="40" class="empty-state-icon"><Setting /></el-icon>
            <p class="empty-state-title">Configuration coming soon</p>
            <p class="empty-state-desc">Settings will be available in a future update.</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Setting } from '@element-plus/icons-vue'
</script>
