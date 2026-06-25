import { onMounted, onUnmounted, ref, watch } from 'vue'
import { ElNotification } from 'element-plus'
import { storeToRefs } from 'pinia'
import { useWcAuthStore } from '@/stores/wcAuthStore'
import { useChatStore } from '@/stores/chatStore'
import type { ChatMessageEvent, ChatMentionEvent, ChatSendFrame, ChatErrorFrame } from '@/types/chat'

export function useChatWs() {
  const auth = useWcAuthStore()
  const chatStore = useChatStore()
  const { isPanelOpen } = storeToRefs(chatStore)
  const isConnected = ref(false)

  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let destroyed = false

  function connect() {
    if (destroyed) return

    const base = import.meta.env.VITE_WS_BASE_URL
    if (!base) return

    let url = `${base}/ws/chat`
    if (auth.token) {
      url += `?token=${encodeURIComponent(auth.token)}`
    }

    ws = new WebSocket(url)

    ws.onopen = () => {
      isConnected.value = true
    }

    ws.onmessage = (e: MessageEvent) => {
      try {
        const frame = JSON.parse(e.data) as ChatMessageEvent | ChatMentionEvent | ChatErrorFrame
        if (frame.type === 'chat_message') {
          const msg = frame as ChatMessageEvent
          chatStore.appendMessage({
            id: msg.id,
            user_id: msg.user_id,
            user_name: msg.user_name,
            avatar_url: msg.avatar_url,
            message: msg.message,
            created_at: msg.created_at,
          })
        } else if (frame.type === 'chat_mention') {
          const mention = frame as ChatMentionEvent
          chatStore.unreadMentionCount++
          ElNotification({
            title: `💬 ${mention.sender_name} nhắc đến bạn`,
            message: mention.message.length > 80 ? mention.message.slice(0, 80) + '…' : mention.message,
            duration: 6000,
            position: isPanelOpen.value ? 'top-right' : 'bottom-right',
            type: 'success',
          })
        } else if (frame.type === 'error') {
          const err = frame as ChatErrorFrame
          ElNotification({ title: 'Chat error', message: err.message, type: 'error', duration: 4000 })
        }
      } catch {
        // Ignore malformed frames
      }
    }

    ws.onclose = () => {
      isConnected.value = false
      if (!destroyed) {
        reconnectTimer = setTimeout(connect, 3000)
      }
    }

    ws.onerror = () => {
      ws?.close()
    }
  }

  // If the token becomes available after mount (page reload, token loaded from storage),
  // close the current guest connection and reconnect as an authenticated user.
  watch(() => auth.token, (newToken, oldToken) => {
    if (newToken === oldToken) return
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (ws) {
      // Null out handlers first so close() doesn't trigger auto-reconnect for this stale conn.
      ws.onclose = null
      ws.onerror = null
      ws.close()
      ws = null
      isConnected.value = false
    }
    if (newToken) connect()
  })

  function sendMessage(text: string, mentions: string[] = []) {
    if (!ws || ws.readyState !== WebSocket.OPEN) return
    const frame: ChatSendFrame = { type: 'chat_send', message: text }
    if (mentions.length > 0) frame.mentions = mentions
    ws.send(JSON.stringify(frame))
  }

  onMounted(connect)

  onUnmounted(() => {
    destroyed = true
    if (reconnectTimer) clearTimeout(reconnectTimer)
    ws?.close()
  })

  return { isConnected, sendMessage }
}
