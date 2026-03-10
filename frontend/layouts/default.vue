<template>
  <div class="min-h-screen bg-gray-100">
    <header v-if="isAuthed" class="sticky top-0 z-20 w-full border-b border-gray-200 bg-white">
      <div class="relative flex w-full items-center justify-between px-4 py-3">
        <div class="h-10 w-10 shrink-0" aria-hidden="true" />
        <NuxtLink
          to="/"
          class="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 text-base font-bold text-gray-800"
        >
          RunSync Pro
        </NuxtLink>
        <button
          type="button"
          class="flex h-10 w-10 shrink-0 items-center justify-center rounded text-gray-700 hover:bg-gray-100"
          aria-label="メニューを開く"
          @click="menuOpen = true"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            aria-hidden="true"
          >
            <line x1="3" y1="6" x2="21" y2="6" />
            <line x1="3" y1="12" x2="21" y2="12" />
            <line x1="3" y1="18" x2="21" y2="18" />
          </svg>
        </button>
      </div>
    </header>

    <!-- Backdrop -->
    <div
      v-if="isAuthed && menuOpen"
      class="fixed inset-0 z-30 bg-black/30"
      @click="menuOpen = false"
    />

    <!-- Drawer -->
    <aside
      v-if="isAuthed"
      class="fixed right-0 top-0 z-40 h-full w-72 transform bg-white shadow-lg transition-transform"
      :class="menuOpen ? 'translate-x-0' : 'translate-x-full'"
      aria-label="navigation drawer"
    >
      <div class="flex items-center justify-between border-b border-gray-200 px-4 py-3">
        <p class="text-sm font-semibold text-gray-800">メニュー</p>
        <button
          type="button"
          class="rounded px-2 py-1 text-sm text-gray-700 hover:bg-gray-100"
          @click="menuOpen = false"
        >
          閉じる
        </button>
      </div>

      <nav class="p-2">
        <NuxtLink
          to="/"
          class="block rounded px-3 py-2 text-sm text-gray-800 hover:bg-gray-100"
          @click="menuOpen = false"
        >
          トップ
        </NuxtLink>
        <NuxtLink
          to="/shoes"
          class="block rounded px-3 py-2 text-sm text-gray-800 hover:bg-gray-100"
          @click="menuOpen = false"
        >
          シューズ
        </NuxtLink>
        <NuxtLink
          to="/training-logs"
          class="block rounded px-3 py-2 text-sm text-gray-800 hover:bg-gray-100"
          @click="menuOpen = false"
        >
          走行ログ
        </NuxtLink>
        <NuxtLink
          to="/vdot"
          class="block rounded px-3 py-2 text-sm text-gray-800 hover:bg-gray-100"
          @click="menuOpen = false"
        >
          VDOT
        </NuxtLink>
        <NuxtLink
          to="/analysis"
          class="block rounded px-3 py-2 text-sm text-gray-800 hover:bg-gray-100"
          @click="menuOpen = false"
        >
          解析
        </NuxtLink>

        <button
          type="button"
          class="mt-2 w-full rounded px-3 py-2 text-left text-sm text-red-600 hover:bg-red-50"
          @click="logout"
        >
          ログアウト
        </button>
      </nav>
    </aside>

    <main class="mx-auto max-w-4xl px-4 py-6">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
const auth = useAuthStore()
const menuOpen = ref(false)
const isAuthed = computed(() => !!auth.tokenOrStorage)

async function logout() {
  auth.clearToken()
  menuOpen.value = false
  await navigateTo('/login')
}
</script>

