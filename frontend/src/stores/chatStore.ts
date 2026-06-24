import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { ChatMessage } from '@/types/chat'

export const useChatStore = defineStore('chat', () => {
  const messages = ref<ChatMessage[]>([])
  const unreadCount = ref(0)
  const isPanelOpen = ref(false)
  const hasMore = ref(true)
  const isLoadingMore = ref(false)

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

  return {
    messages,
    unreadCount,
    isPanelOpen,
    hasMore,
    isLoadingMore,
    openPanel,
    closePanel,
    appendMessage,
    setHistory,
    prependHistory,
  }
})
