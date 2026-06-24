import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { ChatMessage } from '@/types/chat'

export const useChatStore = defineStore('chat', () => {
  const messages = ref<ChatMessage[]>([])
  const unreadCount = ref(0)
  const isPanelOpen = ref(false)

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

  function setHistory(history: ChatMessage[]) {
    messages.value = history
  }

  return { messages, unreadCount, isPanelOpen, openPanel, closePanel, appendMessage, setHistory }
})
