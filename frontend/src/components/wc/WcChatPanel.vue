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
          <!-- eslint-disable-next-line vue/no-v-html -->
          <div class="msg-text" v-html="renderMessage(msg.message)"></div>
        </div>
      </div>
    </div>

    <div class="chat-input-area">
      <template v-if="auth.isLoggedIn">
        <div class="input-wrap">
          <el-input
            v-model="draft"
            :placeholder="t('wc.chat.placeholder')"
            :maxlength="500"
            @keydown.enter.exact.prevent="submit"
            @keydown.escape="closeMentionDropdown"
            @input="onInput"
            :disabled="!chatWs.isConnected.value"
            class="chat-input"
            size="default"
          />
          <!-- @mention autocomplete dropdown -->
          <div v-if="mentionDropdown.open" class="mention-dropdown">
            <div
              v-for="user in mentionDropdown.filtered"
              :key="user.id"
              class="mention-item"
              @mousedown.prevent="selectMention(user)"
            >
              <el-avatar :src="user.avatar_url || undefined" :size="20" class="mention-avatar">
                {{ user.name.charAt(0).toUpperCase() }}
              </el-avatar>
              <span class="mention-name">{{ user.name }}</span>
            </div>
            <div v-if="mentionDropdown.filtered.length === 0" class="mention-empty">
              Không tìm thấy
            </div>
          </div>
        </div>
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
import { ref, reactive, watch, nextTick, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Loading } from '@element-plus/icons-vue'
import axios from 'axios'
import { useChatStore } from '@/stores/chatStore'
import { useWcAuthStore } from '@/stores/wcAuthStore'
import { useChatWs } from '@/composables/useChatWs'
import type { ChatMessage, WcUserForMention } from '@/types/chat'

const { t } = useI18n()
const chatStore = useChatStore()
const auth = useWcAuthStore()
const chatWs = useChatWs()

const draft = ref('')
const listRef = ref<HTMLElement | null>(null)
const readyToLoadMore = ref(false)

const BASE = (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1') + '/wc'
const PAGE_SIZE = 20

// --- mention state ---
const pendingMentions = ref<WcUserForMention[]>([]) // users selected in this draft

const mentionDropdown = reactive({
  open: false,
  query: '',
  filtered: [] as WcUserForMention[],
  triggerIndex: -1, // position of '@' in draft
})

function onInput() {
  const val = draft.value
  const cursorPos = val.length // input cursor is always at end for simplicity
  // Find last '@' that has not been resolved yet
  const lastAt = val.lastIndexOf('@')
  if (lastAt === -1) {
    mentionDropdown.open = false
    return
  }
  const afterAt = val.slice(lastAt + 1)
  // Close if there's a space after the '@' (user typed past a name)
  if (afterAt.includes(' ')) {
    mentionDropdown.open = false
    return
  }
  mentionDropdown.triggerIndex = lastAt
  mentionDropdown.query = afterAt.toLowerCase()
  mentionDropdown.filtered = chatStore.wcUsers
    .filter((u) => u.name.toLowerCase().includes(mentionDropdown.query))
    .slice(0, 6)
  mentionDropdown.open = mentionDropdown.filtered.length > 0 || mentionDropdown.query.length === 0

  // Load users lazily on first '@'
  if (!chatStore.wcUsersLoaded) {
    chatStore.loadWcUsers().then(() => {
      mentionDropdown.filtered = chatStore.wcUsers
        .filter((u) => u.name.toLowerCase().includes(mentionDropdown.query))
        .slice(0, 6)
    })
  }
}

function selectMention(user: WcUserForMention) {
  // Replace from the '@' trigger to cursor with '@Name '
  const before = draft.value.slice(0, mentionDropdown.triggerIndex)
  draft.value = before + '@' + user.name + ' '
  pendingMentions.value.push(user)
  mentionDropdown.open = false
}

function closeMentionDropdown() {
  mentionDropdown.open = false
}

// --- highlight mentions in rendered messages ---
const mentionPattern = computed(() => {
  if (chatStore.wcUsers.length === 0) return null
  const names = chatStore.wcUsers.map((u) => u.name.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'))
  return new RegExp(`@(${names.join('|')})(?=\\s|$|[^\\w])`, 'g')
})

function renderMessage(text: string): string {
  // Escape HTML first
  const escaped = text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
  if (!mentionPattern.value) return escaped
  return escaped.replace(mentionPattern.value, '<span class="mention-highlight">@$1</span>')
}

// --- history / scroll ---
async function loadHistory() {
  readyToLoadMore.value = false
  try {
    const res = await axios.get<{ messages: ChatMessage[]; has_more: boolean }>(
      `${BASE}/chat/messages`,
      { params: { limit: PAGE_SIZE } },
    )
    chatStore.setHistory(res.data.messages ?? [], res.data.has_more ?? false)
    await nextTick()
    requestAnimationFrame(() => {
      if (listRef.value) {
        listRef.value.scrollTop = listRef.value.scrollHeight
      }
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
  const mentionIDs = pendingMentions.value.map((u) => u.id)
  chatWs.sendMessage(text, mentionIDs)
  draft.value = ''
  pendingMentions.value = []
  mentionDropdown.open = false
}

function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

watch(
  () => chatStore.isPanelOpen,
  (open) => {
    if (open) {
      loadHistory()
      chatStore.markMentionsRead()
      chatStore.loadWcUsers()
    } else {
      readyToLoadMore.value = false
    }
  },
)

watch(
  () => chatStore.messages[chatStore.messages.length - 1]?.id,
  (newId, oldId) => {
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

:deep(.mention-highlight) {
  color: var(--el-color-primary, #409eff);
  font-weight: 600;
}

.chat-message.is-own :deep(.mention-highlight) {
  color: #fff;
  opacity: 0.9;
}

.chat-input-area {
  padding: 8px 10px;
  border-top: 1px solid var(--border-color, #e4e7ed);
  display: flex;
  gap: 6px;
  align-items: center;
  position: relative;
}

.input-wrap {
  flex: 1;
  position: relative;
}

.chat-input {
  width: 100%;
}

.mention-dropdown {
  position: absolute;
  bottom: calc(100% + 4px);
  left: 0;
  right: 0;
  background: var(--surface-card, #fff);
  border: 1px solid var(--border-color, #e4e7ed);
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  max-height: 180px;
  overflow-y: auto;
  z-index: 2100;
}

.mention-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 12px;
  cursor: pointer;
  font-size: 13px;
  transition: background 0.12s;
}

.mention-item:hover {
  background: var(--surface-page, #f5f7fa);
}

.mention-avatar {
  flex-shrink: 0;
  font-size: 10px;
}

.mention-name {
  font-weight: 500;
  color: var(--text-primary, #303133);
}

.mention-empty {
  padding: 10px 12px;
  font-size: 12px;
  color: var(--text-muted, #909399);
  text-align: center;
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
