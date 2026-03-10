<template>
  <div class="rounded-lg bg-white p-6 shadow-md">
    <h1 class="mb-1 text-xl font-bold text-gray-800">ユーザートップ</h1>
    <p class="text-sm text-gray-600">ログイン済みのユーザー向けトップページです。</p>

    <div class="mt-4 rounded bg-gray-50 p-4 text-sm text-gray-800">
      <p class="font-semibold">次にやること</p>
      <ul class="mt-2 list-disc space-y-1 pl-5 text-gray-700">
        <li>メニューから「VDOT」を開いて計算してみる</li>
        <li>今後はここに「直近の走行ログ」や「目標タイム」などを表示していく想定</li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: 'auth',
})
</script>
