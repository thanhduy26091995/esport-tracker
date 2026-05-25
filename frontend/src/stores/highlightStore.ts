import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Highlight, HighlightsResponse } from '@/types/highlight'
import { highlightService } from '@/services/highlightService'

export const useHighlightStore = defineStore('highlight', () => {
  const trending = ref<Highlight[]>([])
  const dailyRecap = ref<Highlight[]>([])
  const competitive = ref<Highlight[]>([])
  const social = ref<Highlight[]>([])
  const loading = ref(false)
  const generatedAt = ref<string | null>(null)

  async function fetchHighlights() {
    loading.value = true
    try {
      const data: HighlightsResponse = await highlightService.getHighlights()
      trending.value = data.trending ?? []
      dailyRecap.value = data.daily_recap ?? []
      competitive.value = data.competitive ?? []
      social.value = data.social ?? []
      generatedAt.value = data.generated_at
    } catch {
      // non-critical — silently leave arrays empty
    } finally {
      loading.value = false
    }
  }

  return { trending, dailyRecap, competitive, social, loading, generatedAt, fetchHighlights }
})
