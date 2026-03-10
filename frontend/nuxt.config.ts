// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  modules: ['@nuxtjs/tailwindcss', '@pinia/nuxt'],
  runtimeConfig: {
    public: {
      // 既定値。環境変数 NUXT_PUBLIC_API_BASE があれば Nuxt が自動で上書きします
      apiBase: 'http://localhost:8080',
    },
  },
})