import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { ChatMessage, WcUserForMention } from '@/types/chat'
import { wcService } from '@/services/wcService'

export const useChatStore = defineStore('chat', () => {
  const messages = ref<ChatMessage[]>([])
  const unreadCount = ref(0)
  const unreadMentionCount = ref(0)
  const isPanelOpen = ref(false)
  const hasMore = ref(true)
  const isLoadingMore = ref(false)
  const wcUsers = ref<WcUserForMention[]>([])
  const wcUsersLoaded = ref(false)

  function openPanel() {
    isPanelOpen.value = true
    unreadCount.value = 0
  }

  function closePanel() {
    isPanelOpen.value = false
  }

  function appendMessage(msg: ChatMessage) {
    // Deduplicate by id (handles race between REST history and WS stream)
    if (messages.value.some((m) => m.id === msg.id)) return
    messages.value.push(msg)
    if (!isPanelOpen.value) {
      unreadCount.value++
    }
  }

  function setHistory(history: ChatMessage[], more: boolean) {
    messages.value = history
    hasMore.value = more
  }

  function prependHistory(older: ChatMessage[], more: boolean) {
    const ids = new Set(messages.value.map((m) => m.id))
    const fresh = older.filter((m) => !ids.has(m.id))
    messages.value = [...fresh, ...messages.value]
    hasMore.value = more
  }

  async function fetchUnreadMentionCount() {
    try {
      const data = await wcService.getUnreadMentionCount()
      unreadMentionCount.value = data.count
    } catch {
      // ignore
    }
  }

  async function markMentionsRead() {
    if (unreadMentionCount.value === 0) return
    try {
      await wcService.markMentionsRead()
      unreadMentionCount.value = 0
    } catch {
      // ignore
    }
  }

  async function loadWcUsers() {
    if (wcUsersLoaded.value) return
    try {
      const data = await wcService.getWcUsersForMention()
      wcUsers.value = data.users
      wcUsersLoaded.value = true
    } catch {
      // ignore
    }
  }

  return {
    messages,
    unreadCount,
    unreadMentionCount,
    isPanelOpen,
    hasMore,
    isLoadingMore,
    wcUsers,
    wcUsersLoaded,
    openPanel,
    closePanel,
    appendMessage,
    setHistory,
    prependHistory,
    fetchUnreadMentionCount,
    markMentionsRead,
    loadWcUsers,
  }
})
