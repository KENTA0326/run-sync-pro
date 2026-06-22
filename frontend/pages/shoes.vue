<template>
  <div class="space-y-6">
    <div class="rounded-lg bg-white p-6 shadow-md">
      <h1 class="mb-2 text-xl font-bold text-gray-800">シューズ管理</h1>
      <p class="mb-4 text-sm text-gray-600">
        日々の走行で使うシューズを登録・管理します。削除すると一覧からは消えますが、過去の走行ログには残ります。
      </p>

      <div class="grid gap-4 sm:grid-cols-3">
        <div class="sm:col-span-1">
          <label class="mb-1 block text-sm font-medium text-gray-700">ブランド</label>
          <input
            v-model="brand"
            type="text"
            class="w-full rounded border border-gray-300 p-2 text-gray-800"
            placeholder="例: Nike"
          />
        </div>
        <div class="sm:col-span-1">
          <label class="mb-1 block text-sm font-medium text-gray-700">モデル</label>
          <input
            v-model="model"
            type="text"
            class="w-full rounded border border-gray-300 p-2 text-gray-800"
            placeholder="例: Pegasus 41"
          />
        </div>
        <div class="sm:col-span-1">
          <label class="mb-1 block text-sm font-medium text-gray-700">購入日</label>
          <input
            v-model="purchaseDate"
            type="date"
            class="w-full rounded border border-gray-300 p-2 text-gray-800"
          />
        </div>
      </div>

      <div class="mt-4 flex items-center gap-3">
        <button
          type="button"
          class="rounded bg-blue-500 px-4 py-2 text-sm font-medium text-white hover:bg-blue-600"
          @click="handleCreate"
        >
          シューズを追加
        </button>
        <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
        <p v-if="success" class="text-sm text-green-600">{{ success }}</p>
      </div>
    </div>

    <div class="rounded-lg bg-white p-6 shadow-md">
      <h2 class="mb-3 text-lg font-bold text-gray-800">登録済みシューズ</h2>
      <p v-if="loading" class="text-sm text-gray-600">読み込み中...</p>
      <p v-else-if="shoes.length === 0" class="text-sm text-gray-600">まだシューズが登録されていません。</p>

      <table v-else class="mt-2 w-full border-collapse text-sm">
        <thead>
          <tr class="border-b border-gray-200 bg-gray-50">
            <th class="p-2 text-left font-medium text-gray-700">ブランド / モデル</th>
            <th class="p-2 text-left font-medium text-gray-700">購入日</th>
            <th class="p-2 text-right font-medium text-gray-700">累計距離 (km)</th>
            <th class="p-2 text-center font-medium text-gray-700">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="shoe in shoes" :key="shoe.id" class="border-b border-gray-100">
            <td class="p-2 text-gray-800">
              <p class="font-medium">{{ shoe.brand }} {{ shoe.model }}</p>
            </td>
            <td class="p-2 text-gray-700">{{ formatDate(shoe.purchase_date) }}</td>
            <td class="p-2 text-right font-mono">{{ shoe.total_distance.toFixed(1) }}</td>
            <td class="p-2 text-center">
              <button
                type="button"
                class="rounded bg-red-50 px-2 py-1 text-xs font-medium text-red-600 hover:bg-red-100"
                @click="handleDelete(shoe.id)"
              >
                削除
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { apiPath } from '~/composables/apiPaths'
import type { CreateShoeRequest, ListShoesResponse, Shoe } from '~/types/api'

definePageMeta({
  middleware: 'auth',
})

const api = useApi()

const brand = ref('')
const model = ref('')
const purchaseDate = ref('')
const shoes = ref<Shoe[]>([])
const loading = ref(false)
const error = ref('')
const success = ref('')

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  return dateStr.slice(0, 10)
}

async function fetchShoes() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get<ListShoesResponse>(apiPath.shoes, { limit: 500 })
    shoes.value = data.shoes
  } catch (err) {
    error.value = api.getErrorMessage(err)
  } finally {
    loading.value = false
  }
}

async function handleCreate() {
  error.value = ''
  success.value = ''
  if (!brand.value || !model.value || !purchaseDate.value) {
    error.value = 'ブランド・モデル・購入日を入力してください。'
    return
  }
  try {
    const body: CreateShoeRequest = {
      brand: brand.value,
      model: model.value,
      purchase_date: purchaseDate.value,
    }
    await api.post<Shoe>(apiPath.shoes, body)
    success.value = 'シューズを登録しました。'
    brand.value = ''
    model.value = ''
    purchaseDate.value = ''
    await fetchShoes()
  } catch (err) {
    error.value = api.getErrorMessage(err)
  }
}

async function handleDelete(id: number) {
  if (!confirm('このシューズを削除しますか？')) return
  error.value = ''
  success.value = ''
  try {
    await api.delete<unknown>(apiPath.shoe(id))
    success.value = 'シューズを削除しました。'
    await fetchShoes()
  } catch (err) {
    error.value = api.getErrorMessage(err)
  }
}

onMounted(() => {
  fetchShoes()
})
</script>

