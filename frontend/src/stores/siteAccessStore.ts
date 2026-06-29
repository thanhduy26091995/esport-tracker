import { defineStore } from 'pinia'
import { getQuestion, validateAnswer } from '@/services/siteAccessService'

const TOKEN_KEY = 'site_access_token'

export const useSiteAccessStore = defineStore('siteAccess', {
  state: () => ({
    token: localStorage.getItem(TOKEN_KEY) ?? null as string | null,
    question: '',
    enabled: false,
    checked: false,
  }),

  getters: {
    isGranted: (state): boolean => !state.enabled || !!state.token,
  },

  actions: {
    async init() {
      try {
        const { question, enabled } = await getQuestion()
        this.question = question
        this.enabled = enabled
      } catch {
        // if we can't reach the server at all, don't block the app
        this.enabled = false
      } finally {
        this.checked = true
      }
    },

    async submit(answer: string): Promise<void> {
      const token = await validateAnswer(answer)
      this.token = token
      localStorage.setItem(TOKEN_KEY, token)
    },

    invalidate() {
      this.token = null
      localStorage.removeItem(TOKEN_KEY)
    },
  },
})
