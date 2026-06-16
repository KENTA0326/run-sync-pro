<template>
  <div class="space-y-6">
    <div class="rounded-lg bg-white p-6 shadow-md">
      <h1 class="mb-2 text-xl font-bold text-gray-800">走行ログの登録</h1>
      <p class="mb-4 text-sm text-gray-600">日々の走行を記録し、シューズごとの累計距離も自動で更新します。</p>

      <div class="grid gap-4 sm:grid-cols-3">
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">日付</label>
          <input
            v-model="trainingDate"
            type="date"
            class="w-full rounded border border-gray-300 p-2 text-gray-800"
          />
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">距離 (km)</label>
          <input
            v-model.number="distance"
            type="number"
            step="0.1"
            min="0.1"
            class="w-full rounded border border-gray-300 p-2 text-gray-800"
          />
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">時間 (分:秒)</label>
          <div class="flex gap-2">
            <input
              v-model.number="durationMin"
              type="number"
              min="0"
              class="w-20 rounded border border-gray-300 p-2 text-gray-800"
              placeholder="分"
            />
            <span class="self-center text-gray-500">:</span>
            <input
              v-model.number="durationSec"
              type="number"
              min="0"
              max="59"
              class="w-20 rounded border border-gray-300 p-2 text-gray-800"
              placeholder="秒"
            />
          </div>
        </div>
      </div>

      <div class="mt-4 grid gap-4 sm:grid-cols-3">
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">ペース表記</label>
          <input
            v-model="pace"
            type="text"
            class="w-full rounded border border-gray-300 p-2 text-gray-800"
            placeholder="例: 5:15"
          />
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">種類</label>
          <select
            v-model.number="kind"
            class="w-full rounded border border-gray-300 p-2 text-gray-800"
          >
            <option :value="0">ジョグ</option>
            <option :value="1">LSD</option>
            <option :value="2">ペース走</option>
            <option :value="3">インターバル</option>
          </select>
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">シューズ</label>
          <select
            v-model.number="shoeId"
            class="w-full rounded border border-gray-300 p-2 text-gray-800"
          >
            <option :value="0">選択してください</option>
            <option v-for="shoe in shoes" :key="shoe.id" :value="shoe.id">
              {{ shoe.brand }} {{ shoe.model }} ({{ shoe.total_distance.toFixed(0) }}km)
            </option>
          </select>
        </div>
      </div>

      <div class="mt-4">
        <label class="mb-1 block text-sm font-medium text-gray-700">メモ</label>
        <textarea
          v-model="memo"
          rows="2"
          class="w-full rounded border border-gray-300 p-2 text-gray-800"
          placeholder="今日の感想やポイントなど"
        />
      </div>

      <div class="mt-4 flex items-center gap-3">
        <button
          type="button"
          class="rounded bg-blue-500 px-4 py-2 text-sm font-medium text-white hover:bg-blue-600"
          @click="handleCreate"
        >
          走行ログを追加
        </button>
        <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
        <p v-if="success" class="text-sm text-green-600">{{ success }}</p>
      </div>
    </div>

    <div class="rounded-lg bg-white p-6 shadow-md">
      <h2 class="mb-4 text-lg font-bold text-gray-800">CSV 一括取込</h2>

      <div class="flex flex-wrap items-center gap-3">
        <input
          ref="csvFileInput"
          type="file"
          accept=".csv,text/csv"
          class="block max-w-full text-sm text-gray-700 file:mr-3 file:rounded file:border-0 file:bg-gray-100 file:px-3 file:py-2 file:text-sm file:font-medium file:text-gray-800 hover:file:bg-gray-200"
          @change="onCsvFileChange"
        />
        <button
          type="button"
          class="rounded bg-emerald-600 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-50"
          :disabled="csvImporting || !csvFile"
          @click="handleCsvImport"
        >
          {{ csvImporting ? '取込中...' : 'CSV を取り込む' }}
        </button>
        <button
          type="button"
          class="rounded border border-gray-300 px-3 py-2 text-sm text-gray-700 hover:bg-gray-50"
          @click="downloadCsvTemplate"
        >
          テンプレートをダウンロード
        </button>
      </div>

      <p v-if="csvError" class="mt-3 text-sm text-red-600">{{ csvError }}</p>
      <p v-if="csvSuccess" class="mt-3 text-sm text-green-600">{{ csvSuccess }}</p>
    </div>

    <div class="rounded-lg bg-white p-6 shadow-md">
      <div class="mb-3 flex flex-wrap items-center justify-between gap-3">
        <h2 class="text-lg font-bold text-gray-800">走行ログ一覧</h2>
        <button
          type="button"
          class="rounded border border-gray-300 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50"
          :disabled="csvExporting || logsLoading"
          @click="handleCsvExport"
        >
          {{ csvExporting ? 'ダウンロード中...' : 'CSV をダウンロード' }}
        </button>
      </div>
      <p v-if="exportError" class="mb-2 text-sm text-red-600">{{ exportError }}</p>
      <p v-if="logsLoading" class="text-sm text-gray-600">読み込み中...</p>
      <p v-else-if="logs.length === 0" class="text-sm text-gray-600">まだ走行ログがありません。</p>

      <table v-else class="mt-2 w-full border-collapse text-sm">
        <thead>
          <tr class="border-b border-gray-200 bg-gray-50">
            <th class="p-2 text-left font-medium text-gray-700">日付</th>
            <th class="p-2 text-right font-medium text-gray-700">距離 (km)</th>
            <th class="p-2 text-right font-medium text-gray-700">時間</th>
            <th class="p-2 text-right font-medium text-gray-700">ペース</th>
            <th class="p-2 text-left font-medium text-gray-700">種類</th>
            <th class="p-2 text-left font-medium text-gray-700">シューズ</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="log in logs" :key="log.id" class="border-b border-gray-100">
            <td class="p-2 text-gray-800">{{ formatDate(log.training_date) }}</td>
            <td class="p-2 text-right font-mono">{{ log.distance.toFixed(1) }}</td>
            <td class="p-2 text-right font-mono">{{ formatDuration(log.duration) }}</td>
            <td class="p-2 text-right font-mono">{{ log.pace }}</td>
            <td class="p-2 text-gray-700">{{ formatKind(log.kind) }}</td>
            <td class="p-2 text-gray-700">
              <span v-if="log.shoe">{{ log.shoe.brand }} {{ log.shoe.model }}</span>
              <span v-else>-</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { apiPath } from '~/composables/apiPaths'
import type {
  CreateTrainingLogRequest,
  ImportTrainingLogsResponse,
  ListShoesResponse,
  PaginatedTrainingLogsResponse,
  Shoe,
  TrainingLog,
} from '~/types/api'

definePageMeta({
  middleware: 'auth',
})

const api = useApi()

const trainingDate = ref('')
const distance = ref<number | null>(null)
const durationMin = ref<number | null>(null)
const durationSec = ref<number | null>(null)
const pace = ref('')
const memo = ref('')
const kind = ref(0)
const shoeId = ref(0)

const shoes = ref<Shoe[]>([])
const logs = ref<TrainingLog[]>([])
const logsLoading = ref(false)
const error = ref('')
const success = ref('')

const csvFileInput = ref<HTMLInputElement | null>(null)
const csvFile = ref<File | null>(null)
const csvImporting = ref(false)
const csvError = ref('')
const csvSuccess = ref('')
const csvExporting = ref(false)
const exportError = ref('')

const CSV_HEADER =
  'training_date,distance,duration,pace,kind,shoe_id,memo'

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  return dateStr.slice(0, 10)
}

function formatDuration(totalSeconds: number): string {
  const sec = Math.max(0, Math.round(totalSeconds))
  const h = Math.floor(sec / 3600)
  const m = Math.floor((sec % 3600) / 60)
  const s = sec % 60
  if (h > 0) return `${h}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
  return `${m}:${s.toString().padStart(2, '0')}`
}

function formatKind(k: number): string {
  switch (k) {
    case 0:
      return 'ジョグ'
    case 1:
      return 'LSD'
    case 2:
      return 'ペース走'
    case 3:
      return 'インターバル'
    default:
      return 'その他'
  }
}

async function fetchShoes() {
  try {
    const data = await api.get<ListShoesResponse>(apiPath.shoes)
    shoes.value = data.shoes
  } catch (err) {
    // エラーはフォーム上でまとめて表示するのでここでは握りつぶす
    console.error(err)
  }
}

async function fetchLogs() {
  logsLoading.value = true
  error.value = ''
  try {
    const data = await api.get<PaginatedTrainingLogsResponse>(apiPath.trainingLogs, { limit: 500 })
    logs.value = data.items
  } catch (err) {
    error.value = api.getErrorMessage(err)
  } finally {
    logsLoading.value = false
  }
}

function onCsvFileChange(event: Event) {
  csvError.value = ''
  csvSuccess.value = ''
  const target = event.target as HTMLInputElement
  csvFile.value = target.files?.[0] ?? null
}

function downloadCsvTemplate() {
  const shoeID = shoes.value[0]?.id ?? 1
  const today = new Date().toISOString().slice(0, 10)
  const sample = `${CSV_HEADER}\n${today},10.0,3600,6:00,0,${shoeID},サンプル行\n`
  const blob = new Blob([sample], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'training_logs_template.csv'
  a.click()
  URL.revokeObjectURL(url)
}

async function handleCsvExport() {
  exportError.value = ''
  csvExporting.value = true
  try {
    const today = new Date().toISOString().slice(0, 10).replace(/-/g, '')
    await api.downloadGet(apiPath.trainingLogs, `training_logs_${today}.csv`, {
      Accept: 'text/csv',
    })
  } catch (err) {
    exportError.value = err instanceof Error ? err.message : api.getErrorMessage(err)
  } finally {
    csvExporting.value = false
  }
}

async function handleCsvImport() {
  csvError.value = ''
  csvSuccess.value = ''
  if (!csvFile.value) {
    csvError.value = 'CSV ファイルを選択してください。'
    return
  }

  csvImporting.value = true
  try {
    const form = new FormData()
    form.append('file', csvFile.value, csvFile.value.name)
    const res = await api.postFormData<ImportTrainingLogsResponse>(
      apiPath.trainingLogs,
      form
    )
    csvSuccess.value = `${res.message}（${res.created_count} 件）`
    csvFile.value = null
    if (csvFileInput.value) {
      csvFileInput.value.value = ''
    }
    await fetchLogs()
    await fetchShoes()
  } catch (err) {
    csvError.value = api.getErrorMessage(err)
  } finally {
    csvImporting.value = false
  }
}

async function handleCreate() {
  error.value = ''
  success.value = ''

  if (!trainingDate.value || !distance.value || distance.value <= 0) {
    error.value = '日付と距離を正しく入力してください。'
    return
  }
  const totalSec = (durationMin.value ?? 0) * 60 + (durationSec.value ?? 0)
  if (totalSec <= 0) {
    error.value = '時間を正しく入力してください。'
    return
  }
  if (!shoeId.value) {
    error.value = '使用したシューズを選択してください。'
    return
  }

  try {
    const body: CreateTrainingLogRequest = {
      training_date: trainingDate.value,
      distance: distance.value,
      duration: totalSec,
      pace: pace.value || '0:00',
      memo: memo.value,
      kind: kind.value,
      shoe_id: shoeId.value,
    }
    await api.post<unknown>(apiPath.trainingLogs, body)
    success.value = '走行ログを保存しました。'

    // 入力をリセット
    distance.value = null
    durationMin.value = null
    durationSec.value = null
    pace.value = ''
    memo.value = ''
    kind.value = 0
    // シューズ選択は残す

    await fetchLogs()
    await fetchShoes() // 累計距離が変わるので更新
  } catch (err) {
    error.value = api.getErrorMessage(err)
  }
}

onMounted(() => {
  const today = new Date().toISOString().slice(0, 10)
  trainingDate.value = today
  fetchShoes()
  fetchLogs()
})
</script>

