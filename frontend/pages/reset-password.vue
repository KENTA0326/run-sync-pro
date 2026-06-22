<template>
  <div class="flex min-h-screen flex-col items-center justify-center bg-gray-100">
    <div class="w-96 rounded bg-white p-8 shadow-md">
      <h1 class="mb-6 text-center text-2xl font-bold text-gray-800">新しいパスワード</h1>
      <div v-if="!token" class="text-sm text-red-600">
        リセット用のリンクが無効です。もう一度
        <NuxtLink to="/forgot-password" class="text-blue-600 hover:underline">パスワード再設定</NuxtLink>
        からやり直してください。
      </div>
      <div v-else class="space-y-4">
        <input
          v-model="newPassword"
          type="password"
          placeholder="新しいパスワード（6文字以上）"
          class="w-full rounded border p-2 text-gray-800"
        />
        <input
          v-model="confirmPassword"
          type="password"
          placeholder="新しいパスワード（確認）"
          class="w-full rounded border p-2 text-gray-800"
        />
        <button
          type="button"
          class="w-full rounded bg-blue-500 p-2 text-white hover:bg-blue-600 disabled:opacity-50"
          :disabled="loading"
          @click="handleConfirm"
        >
          {{ loading ? '更新中...' : 'パスワードを更新' }}
        </button>
        <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
        <p v-if="message" class="text-sm text-green-600">{{ message }}</p>
      </div>
      <p class="mt-4 text-center text-sm text-gray-600">
        <NuxtLink to="/login" class="text-blue-600 hover:underline">ログインへ</NuxtLink>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { apiPath } from '~/composables/apiPaths'
import type { PasswordResetConfirmBody, PasswordResetConfirmResponse } from '~/types/api'

definePageMeta({ layout: false })

const route = useRoute()
const token = computed(() => (typeof route.query.token === 'string' ? route.query.token : ''))

const newPassword = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const error = ref('')
const message = ref('')
const api = useApi()

async function handleConfirm() {
  error.value = ''
  message.value = ''
  if (newPassword.value.length < 6) {
    error.value = 'パスワードは6文字以上で入力してください。'
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    error.value = '確認用パスワードが一致しません。'
    return
  }
  loading.value = true
  try {
    const body: PasswordResetConfirmBody = {
      token: token.value,
      new_password: newPassword.value,
    }
    const data = await api.post<PasswordResetConfirmResponse>(apiPath.authPasswordResetConfirm, body)
    message.value = data.message
    setTimeout(() => navigateTo('/login'), 2000)
  } catch (err) {
    error.value = api.getErrorMessage(err)
  } finally {
    loading.value = false
  }
}
</script>
