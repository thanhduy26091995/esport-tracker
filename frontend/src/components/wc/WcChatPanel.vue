<template>
  <div v-if="chatStore.isPanelOpen" class="chat-panel">
    <div class="chat-header">
      <span class="chat-title">{{ t('wc.chat.title') }}</span>
      <el-button text @click="chatStore.closePanel()" class="close-btn">✕</el-button>
    </div>

    <div ref="listRef" class="chat-messages" @scroll="onScroll">
      <div v-if="chatStore.isLoadingMore" class="load-more-spinner">
        <el-icon class="is-loading"><Loading /></el-icon>
      </div>
      <div v-if="chatStore.messages.length === 0" class="chat-empty">
        {{ t('wc.chat.empty') }}
      </div>
      <div
        v-for="msg in chatStore.messages"
        :key="msg.id"
        class="chat-message"
        :class="{ 'is-own': msg.user_id === auth.user?.id }"
      >
        <el-avatar :src="msg.avatar_url || undefined" :size="28" class="msg-avatar">
          {{ msg.user_name.charAt(0).toUpperCase() }}
        </el-avatar>
        <div class="msg-body">
          <div class="msg-meta">
            <span class="msg-name">{{ msg.user_name }}</span>
            <span class="msg-time">{{ formatTime(msg.created_at) }}</span>
          </div>
          <div class="msg-text">{{ msg.message }}</div>
        </div>
      </div>
    </div>

    <div class="chat-input-area">
      <template v-if="auth.isLoggedIn">
        <el-input
          v-model="draft"
          :placeholder="t('wc.chat.placeholder')"
          :maxlength="500"
          @keydown.enter.exact.prevent="submit"
          :disabled="!chatWs.isConnected.value"
          class="chat-input"
          size="default"
        />
        <el-button
          type="primary"
          size="default"
          :disabled="!draft.trim() || !chatWs.isConnected.value"
          @click="submit"
          class="send-btn"
        >
          {{ t('wc.chat.send') }}
        </el-button>
      </template>
      <div v-else class="login-prompt">{{ t('wc.chat.loginPrompt') }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { Loading } from '@element-plus/icons-vue'
import axios from 'axios'
import { useChatStore } from '@/stores/chatStore'
import { useWcAuthStore } from '@/stores/wcAuthStore'
import { useChatWs } from '@/composables/useChatWs'
import type { ChatMessage } from '@/types/chat'

const { t } = useI18n()
const chatStore = useChatStore()
const auth = useWcAuthStore()
const chatWs = useChatWs()

const draft = ref('')
const listRef = ref<HTMLElement | null>(null)
// Guard: don't trigger loadMore until initial history is rendered and scrolled
const readyToLoadMore = ref(false)

const BASE = (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1') + '/wc'
const PAGE_SIZE = 20

async function loadHistory() {
  readyToLoadMore.value = false
  try {
    const res = await axios.get<{ messages: ChatMessage[]; has_more: boolean }>(
      `${BASE}/chat/messages`,
      { params: { limit: PAGE_SIZE } },
    )
    chatStore.setHistory(res.data.messages ?? [], res.data.has_more ?? false)
    // Double nextTick + rAF ensures DOM height is fully calculated before scrolling
    await nextTick()
    requestAnimationFrame(() => {
      if (listRef.value) {
        listRef.value.scrollTop = listRef.value.scrollHeight
      }
      // Allow loadMore only after initial scroll is done
      readyToLoadMore.value = true
    })
  } catch {
    readyToLoadMore.value = true
  }
}

async function loadMore() {
  if (!readyToLoadMore.value || !chatStore.hasMore || chatStore.isLoadingMore) return
  const oldest = chatStore.messages[0]
  if (!oldest) return

  chatStore.isLoadingMore = true
  const el = listRef.value
  const prevScrollHeight = el?.scrollHeight ?? 0

  try {
    const res = await axios.get<{ messages: ChatMessage[]; has_more: boolean }>(
      `${BASE}/chat/messages`,
      { params: { limit: PAGE_SIZE, before: oldest.created_at } },
    )
    chatStore.prependHistory(res.data.messages ?? [], res.data.has_more ?? false)
    await nextTick()
    if (el) {
      el.scrollTop = el.scrollHeight - prevScrollHeight
    }
  } catch {
    // ignore
  } finally {
    chatStore.isLoadingMore = false
  }
}

function onScroll() {
  if (!listRef.value || !readyToLoadMore.value) return
  if (listRef.value.scrollTop < 40) {
    loadMore()
  }
}

function scrollToBottom() {
  if (listRef.value) {
    listRef.value.scrollTop = listRef.value.scrollHeight
  }
}

function submit() {
  const text = draft.value.trim()
  if (!text) return
  chatWs.sendMessage(text)
  draft.value = ''
}

function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

// Load history and scroll to bottom every time panel opens
watch(
  () => chatStore.isPanelOpen,
  (open) => {
    if (open) loadHistory()
    else readyToLoadMore.value = false
  },
)

// Scroll to bottom only when a NEW message is appended via WS
watch(
  () => chatStore.messages[chatStore.messages.length - 1]?.id,
  (newId, oldId) => {
    // Only fire for genuinely new messages, not when history is set/prepended
    if (newId && oldId && readyToLoadMore.value) {
      nextTick(scrollToBottom)
    }
  },
)
</script>

<style scoped>
.chat-panel {
  position: fixed;
  bottom: 72px;
  right: 20px;
  width: 320px;
  height: 420px;
  background: var(--surface-card, #fff);
  border: 1px solid var(--border-color, #e4e7ed);
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  display: flex;
  flex-direction: column;
  z-index: 2000;
}

.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  border-bottom: 1px solid var(--border-color, #e4e7ed);
  border-radius: 12px 12px 0 0;
}

.chat-title {
  font-weight: 600;
  font-size: 14px;
}

.close-btn {
  padding: 0 4px;
  color: var(--text-secondary, #909399);
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.load-more-spinner {
  display: flex;
  justify-content: center;
  padding: 6px 0;
  color: var(--text-secondary, #909399);
  font-size: 16px;
}

.chat-empty {
  text-align: center;
  color: var(--text-secondary, #909399);
  font-size: 13px;
  margin-top: 40px;
}

.chat-message {
  display: flex;
  gap: 8px;
  align-items: flex-start;
}

.chat-message.is-own {
  flex-direction: row-reverse;
}

.chat-message.is-own .msg-body {
  align-items: flex-end;
}

.chat-message.is-own .msg-text {
  background: var(--el-color-primary, #409eff);
  color: #fff;
}

.msg-avatar {
  flex-shrink: 0;
  font-size: 12px;
}

.msg-body {
  display: flex;
  flex-direction: column;
  gap: 2px;
  max-width: 210px;
}

.msg-meta {
  display: flex;
  gap: 6px;
  align-items: baseline;
}

.msg-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary, #303133);
}

.msg-time {
  font-size: 11px;
  color: var(--text-secondary, #909399);
}

.msg-text {
  font-size: 13px;
  background: var(--surface-ground, #f5f7fa);
  padding: 6px 10px;
  border-radius: 8px;
  word-break: break-word;
  line-height: 1.4;
}

.chat-input-area {
  padding: 8px 10px;
  border-top: 1px solid var(--border-color, #e4e7ed);
  display: flex;
  gap: 6px;
  align-items: center;
}

.chat-input {
  flex: 1;
}

.send-btn {
  flex-shrink: 0;
}

.login-prompt {
  width: 100%;
  text-align: center;
  font-size: 13px;
  color: var(--text-secondary, #909399);
  padding: 4px 0;
}
</style>
