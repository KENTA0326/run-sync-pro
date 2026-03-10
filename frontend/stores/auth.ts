import { defineStore } from 'pinia'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: null as string | null,
  }),
  getters: {
    /** ストア + localStorage の両方を見る（SSR時は null） */
    tokenOrStorage(): string | null {
      if (import.meta.client && this.token) return this.token
      if (import.meta.client) return localStorage.getItem('auth_token')
      return this.token
    },
  },
  actions: {
    setToken(newToken: string) {
      this.token = newToken
      if (import.meta.client) localStorage.setItem('auth_token', newToken)
    },
    clearToken() {
      this.token = null
      if (import.meta.client) localStorage.removeItem('auth_token')
    },
  },
})