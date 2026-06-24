import { onMounted, onUnmounted, ref } from 'vue'
import { useWcAuthStore } from '@/stores/wcAuthStore'
import { useChatStore } from '@/stores/chatStore'
import type { ChatMessageEvent, ChatSendFrame } from '@/types/chat'

export function useChatWs() {
  const auth = useWcAuthStore()
  const chatStore = useChatStore()
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
        const frame = JSON.parse(e.data) as ChatMessageEvent
        if (frame.type !== 'chat_message') return
        chatStore.appendMessage({
          id: frame.id,
          user_id: frame.user_id,
          user_name: frame.user_name,
          avatar_url: frame.avatar_url,
          message: frame.message,
          created_at: frame.created_at,
        })
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

  function sendMessage(text: string) {
    if (!ws || ws.readyState !== WebSocket.OPEN) return
    const frame: ChatSendFrame = { type: 'chat_send', message: text }
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
