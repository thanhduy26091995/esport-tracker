import { onMounted, onUnmounted } from 'vue'
import { storeToRefs } from 'pinia'
import { ElNotification } from 'element-plus'
import { useWcAuthStore } from '@/stores/wcAuthStore'
import { useChatStore } from '@/stores/chatStore'
import type { ActivityEvent } from '@/types/activity'

const BET_TYPE_LABEL: Record<string, string> = {
  handicap: 'Kèo Chấp',
  exact_score: 'Tỉ Số',
  over_under: 'Tài Xỉu',
}

function formatMessage(event: ActivityEvent): string {
  const betLabel = BET_TYPE_LABEL[event.bet_type] ?? event.bet_type
  return `${event.user_name} vừa đặt ${betLabel} — ${event.selection} (${event.stake} ly)`
}

export function useActivityFeed() {
  const auth = useWcAuthStore()
  const { isPanelOpen } = storeToRefs(useChatStore())
  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let destroyed = false

  function connect() {
    if (destroyed) return

    const base = import.meta.env.VITE_WS_BASE_URL
    if (!base) return

    ws = new WebSocket(`${base}/ws`)

    ws.onmessage = (e: MessageEvent) => {
      try {
        const event = JSON.parse(e.data) as ActivityEvent
        if (event.type !== 'bet_placed') return

        // Self-suppression: don't show own bets
        if (auth.user?.id && event.user_id === auth.user.id) return

        ElNotification({
          title: '🎯 Hoạt động',
          message: formatMessage(event),
          duration: 5000,
          position: isPanelOpen.value ? 'top-right' : 'bottom-right',
          type: 'info',
        })
      } catch {
        // Ignore malformed frames
      }
    }

    ws.onclose = () => {
      if (!destroyed) {
        reconnectTimer = setTimeout(connect, 3000)
      }
    }

    ws.onerror = () => {
      ws?.close()
    }
  }

  onMounted(connect)

  onUnmounted(() => {
    destroyed = true
    if (reconnectTimer) clearTimeout(reconnectTimer)
    ws?.close()
  })
}
