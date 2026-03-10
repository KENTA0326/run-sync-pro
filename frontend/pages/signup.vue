<template>
  <div class="flex flex-col items-center justify-center min-h-screen bg-gray-100">
    <div class="p-8 bg-white rounded shadow-md w-96">
      <h1 class="mb-6 text-2xl font-bold text-center text-gray-800">RunSync Pro 新規登録</h1>
      <div class="space-y-4">
        <input
          v-model="name"
          type="text"
          placeholder="名前"
          class="w-full p-2 border rounded text-gray-800"
        />
        <input
          v-model="email"
          type="email"
          placeholder="メールアドレス"
          class="w-full p-2 border rounded text-gray-800"
        />
        <input
          v-model="password"
          type="password"
          placeholder="パスワード（6文字以上）"
          class="w-full p-2 border rounded text-gray-800"
        />

        <button
          type="button"
          class="w-full p-2 text-white bg-blue-500 rounded hover:bg-blue-600"
          @click="handleSignUp"
        >
          登録する
        </button>

        <p v-if="error" class="text-red-500 text-sm mt-2">{{ error }}</p>
        <p v-if="success" class="text-green-600 text-sm mt-2">{{ success }}</p>
      </div>

      <p class="mt-4 text-center text-gray-600 text-sm">
        <NuxtLink to="/login" class="text-blue-600 hover:underline">ログインはこちら</NuxtLink>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { SignUpRequest, SignUpResponse } from '~/types/api'

definePageMeta({
  layout: false,
})

const name = ref('')
const email = ref('')
const password = ref('')
const error = ref('')
const success = ref('')
const api = useApi()

const handleSignUp = async () => {
  error.value = ''
  success.value = ''
  try {
    const body: SignUpRequest = {
      name: name.value,
      email: email.value,
      password: password.value,
    }
    await api.post<SignUpResponse>('/signup', body)
    success.value = '登録が完了しました。ログインしてください。'
    name.value = ''
    email.value = ''
    password.value = ''
  } catch (err) {
    console.error(err)
    error.value = api.getErrorMessage(err) || '登録に失敗しました。'
  }
}
</script>
