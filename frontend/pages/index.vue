<template>
  <div class="flex flex-col items-center justify-center min-h-screen bg-gray-100 p-8">
    <div class="p-8 bg-white rounded shadow-md w-full max-w-md space-y-4">
      <h1 class="text-xl font-bold text-gray-800">RunSync Pro</h1>
      <p class="text-gray-600 text-sm">認証付きAPIの動作確認用です。</p>

      <button
        type="button"
        class="w-full p-2 text-white bg-blue-500 rounded hover:bg-blue-600"
        @click="fetchMe"
      >
        認証情報を取得（/auth/me）
      </button>

      <div v-if="me" class="p-3 bg-gray-50 rounded text-sm text-gray-800">
        <p><strong>user_id:</strong> {{ me.user_id }}</p>
        <p><strong>message:</strong> {{ me.message }}</p>
      </div>
      <p v-if="meError" class="text-red-500 text-sm">{{ meError }}</p>

      <NuxtLink to="/login" class="block text-center text-blue-600 text-sm">ログインへ</NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { AuthMeResponse } from '~/types/api'

const api = useApi()
const me = ref<AuthMeResponse | null>(null)
const meError = ref('')

async function fetchMe() {
  me.value = null
  meError.value = ''
  try {
    const data = await api.get<AuthMeResponse>('/auth/me')
    me.value = data
  } catch (err) {
    meError.value = api.getErrorMessage(err)
  }
}
</script>
