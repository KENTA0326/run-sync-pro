<template>
  <div class="flex min-h-screen flex-col items-center justify-center bg-gray-100">
    <div class="w-96 rounded bg-white p-8 shadow-md">
      <h1 class="mb-6 text-center text-2xl font-bold text-gray-800">パスワード再設定</h1>
      <p class="mb-4 text-sm text-gray-600">
        登録済みのメールアドレスを入力してください。
      </p>
      <div class="space-y-4">
        <input
          v-model="email"
          type="email"
          placeholder="メールアドレス"
          class="w-full rounded border p-2 text-gray-800"
        />
        <button
          type="button"
          class="w-full rounded bg-blue-500 p-2 text-white hover:bg-blue-600 disabled:opacity-50"
          :disabled="loading"
          @click="handleRequest"
        >
          {{ loading ? '送信中...' : '再設定メールを送信' }}
        </button>
        <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
        <p v-if="message" class="text-sm text-green-600">{{ message }}</p>
        <div v-if="resetUrl" class="rounded border border-amber-200 bg-amber-50 p-3 text-sm">
          <p class="mb-2 font-medium text-amber-900">開発環境: リセットリンク</p>
          <NuxtLink :to="resetPath" class="break-all text-blue-600 hover:underline">{{ resetUrl }}</NuxtLink>
        </div>
      </div>
      <p class="mt-4 text-center text-sm text-gray-600">
        <NuxtLink to="/login" class="text-blue-600 hover:underline">ログインに戻る</NuxtLink>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { apiPath } from '~/composables/apiPaths'
import type { PasswordResetRequestBody, PasswordResetRequestResponse } from '~/types/api'

definePageMeta({ layout: false })

const email = ref('')
const loading = ref(false)
const error = ref('')
const message = ref('')
const resetUrl = ref('')
const resetPath = computed(() => {
  if (!resetUrl.value) return '/reset-password'
  try {
    const u = new URL(resetUrl.value)
    return u.pathname + u.search
  } catch {
    return '/reset-password'
  }
})

const api = useApi()

async function handleRequest() {
  error.value = ''
  message.value = ''
  resetUrl.value = ''
  loading.value = true
  try {
    const body: PasswordResetRequestBody = { email: email.value }
    const data = await api.post<PasswordResetRequestResponse>(apiPath.authPasswordResetRequest, body)
    message.value = data.message
    if (data.reset_url) {
      resetUrl.value = data.reset_url
    }
  } catch (err) {
    error.value = api.getErrorMessage(err)
  } finally {
    loading.value = false
  }
}
</script>
