<template>
  <div class="flex flex-col items-center justify-center min-h-screen bg-gray-100">
    <div class="p-8 bg-white rounded shadow-md w-96">
      <h1 class="mb-6 text-2xl font-bold text-center text-gray-800">RunSync ログイン</h1>
      <div class="space-y-4">
        <input v-model="email" type="email" placeholder="メールアドレス" class="w-full p-2 border rounded text-gray-800" />
        <input v-model="password" type="password" placeholder="パスワード" class="w-full p-2 border rounded text-gray-800" />
        
        <button @click="handleLogin" class="w-full p-2 text-white bg-blue-500 rounded hover:bg-blue-600">
          ログイン
        </button>

        <p v-if="error" class="text-red-500 text-sm mt-2">{{ error }}</p>
      </div>

      <p class="mt-4 text-center text-gray-600 text-sm">
        <NuxtLink to="/signup" class="text-blue-600 hover:underline">新規登録はこちら</NuxtLink>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { LoginRequest, LoginResponse } from '~/types/api'

const email = ref('')
const password = ref('')
const error = ref('')
const auth = useAuthStore()
const api = useApi()

const handleLogin = async () => {
  error.value = ''
  try {
    const body: LoginRequest = {
      email: email.value,
      password: password.value,
    }
    const data = await api.post<LoginResponse>('/login', body)

    auth.setToken(data.token)
    alert('ログイン成功！')
    navigateTo('/')
  } catch (err) {
    console.error(err)
    error.value = api.getErrorMessage(err) || 'ログインに失敗しました。メールアドレスとパスワードを確認してください。'
  }
}
</script>